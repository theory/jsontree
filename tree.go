/*
Package jsontree selects RFC 9535 JSONPaths from one JSON value into a new value.
*/
package jsontree

import (
	"fmt"
	"slices"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

// Tree represents a tree of JSONPath query expressions.
type Tree struct {
	root *Segment
}

// selectorsFor returns the selectors from seg, eliminating duplicates. Slices
// are listed first, so that subsequent indexes can be checked for inclusion
// in them. It also returns true if the returned selectors are a wildcard.
func selectorsFor(seg *spec.Segment) ([]spec.Selector, bool) {
	// Sort wildcards and slices first.
	selectors := seg.Selectors()
	slices.SortFunc(selectors, func(a, b spec.Selector) int {
		switch a.(type) {
		case spec.WildcardSelector:
			return -1
		case spec.SliceSelector:
			if _, ok := b.(spec.SliceSelector); ok {
				return 0
			}
			return -1
		}

		switch b.(type) {
		case spec.WildcardSelector:
			return 1
		case spec.SliceSelector:
			return 1
		}

		return 0
	})

	ret := make([]spec.Selector, 0, len(selectors))
	for _, sel := range selectors {
		if _, ok := sel.(spec.WildcardSelector); ok {
			// Wildcard trumps all other selectors.
			return []spec.Selector{spec.Wildcard}, true
		}
		if !selectorsContain(ret, sel) {
			ret = append(ret, sel)
		}
	}
	return ret, false
}

// New compiles paths into a Tree that can be used to select all of their
// paths.
//
//nolint:gocognit
func New(paths ...*jsonpath.Path) *Tree {
	root := Child()
	cur := root

PATH:
	for _, path := range paths {
		// Iterate over the sequence of spec.Segments in the path.
		segs := path.Query().Segments()
	SEG:
		for i, seg := range segs {
			selectors, isWild := selectorsFor(seg)
			if isWild && i == len(segs)-1 {
				// Trailing wildcard is the same as selecting the parent, so
				// discard it and continue with the next path.
				continue
			}

			// Compare the path to each of the children.
			for _, child := range cur.children {
				switch {
				case child.descendant == seg.IsDescendant():
					switch {
					case child.isBranch(segs[i+1:]):
						// Sub-branches equal; merge selectors and continue.
						cur = child.mergeSelectors(selectors)
						continue SEG

					case child.hasSameSelectors(selectors):
						// Same selectors, diff branches; discard sub-segments?
						switch {
						case len(child.children) == 0:
							// Discard remaining segments and go to next path.
							continue PATH
						case i == len(segs)-1:
							// Discard existing children and go to next path.
							child.children = []*Segment{}
							continue PATH
						default:
							// Branches continue in sub-segments.
							cur = child
							continue SEG
						}
					}
				case isWild && !child.descendant && child.isWildcard() && child.isBranch(segs[i+1:]):
					// Descendant wildcard with same descendants wins.
					child.descendant = true
					cur = child
					continue SEG
				}
				// Nothing to merge, continue with the next child.
			}

			// No matching child, append a new one.
			cur = newChild(cur, seg, selectors)
		}

		// Continue to the next path.
		cur = root
	}

	root.deduplicate()

	return &Tree{root: root}
}

func newChild(cur *Segment, seg *spec.Segment, selectors []spec.Selector) *Segment {
	child := Child(selectors...)
	child.descendant = seg.IsDescendant()
	cur.Append(child)
	return child
}

// String returns a string representation of seg, starting from "$" for the
// root, and including all of its child segments in as a tree diagram.
func (tree *Tree) String() string {
	return "$\n" + tree.root.String()
}

// Select selects tree's paths from the from JSON value into a new value. For
// a root-only JSONTree that contains no children, from will simply be
// returned. All other JSONTree queries will select from the from value if
// it's an array ([]any) or object (map[string]any), and return nil for any
// other values.
func (tree *Tree) Select(from any) any {
	if len(tree.root.children) == 0 {
		return from
	}

	switch entity := from.(type) {
	case map[string]any:
		ret := map[string]any{}
		tree.selectObjectSegment(tree.root, entity, entity, ret)
		return ret
	case []any:
		ret := make([]any, 0, cap(entity))
		if sel := tree.selectArraySegment(tree.root, entity, entity, ret); sel != nil {
			return sel
		}
		return ret
	default:
		// Cannot select from any other type. Following RFC 9535, return nil.
		return nil
	}
}

// selectObjectSegment uses the selectors in seg to select paths from src into
// dst and recurses into its children.
func (tree *Tree) selectObjectSegment(seg *Segment, root any, cur, dst map[string]any) {
	tree.selectObject(seg, root, cur, dst)
	for _, seg := range seg.children {
		tree.selectObject(seg, root, cur, dst)
	}
}

// selectObject uses the selectors in seg to select paths from src to dst. If
// seg is a descendant Segment, it recursively selects from seg into all of
// src's values.
func (tree *Tree) selectObject(seg *Segment, root any, cur, dst map[string]any) {
	for _, sel := range seg.selectors {
		switch sel := sel.(type) {
		case spec.Name:
			tree.processKey(string(sel), seg, root, cur, dst)
		case spec.WildcardSelector:
			for k := range cur {
				tree.processKey(k, seg, root, cur, dst)
			}
		case *spec.FilterSelector:
			for k, v := range cur {
				if sel.Eval(v, root) {
					tree.processKey(k, seg, root, cur, dst)
				}
			}
		}
	}
	if seg.descendant {
		tree.descendObject(seg, root, cur, dst)
	}
}

// descendObject selects the paths from seg from each value from src into
// dst.
func (tree *Tree) descendObject(seg *Segment, root any, cur, dst map[string]any) {
	for k, v := range cur {
		switch v := v.(type) {
		case map[string]any:
			if sub := tree.dispatchObject(seg, root, v, dst[k]); sub != nil {
				dst[k] = sub
			}
		case []any:
			if sub := tree.dispatchArray(seg, root, v, dst[k]); sub != nil {
				dst[k] = sub
			}
		}
	}
}

// processKey fetches the value for key from src and, if the value exists,
// stores it in dst. If the value is a JSON object (map[string]any) or array
// ([]any), it dispatches selection for that value so that seg's children can
// select from the value.
func (tree *Tree) processKey(key string, seg *Segment, root any, cur, dst map[string]any) {
	// Do we have a value?
	val, ok := cur[key]
	if !ok {
		return
	}

	// Keep the value if it's the end of the path.
	if len(seg.children) == 0 {
		dst[key] = val
		return
	}

	// Allow the child segments to select from an object or array.
	switch val := val.(type) {
	case map[string]any:
		sub := tree.dispatchObject(seg, root, val, dst[key])
		if sub != nil {
			dst[key] = sub
		}
	case []any:
		sub := tree.dispatchArray(seg, root, val, dst[key])
		if sub != nil {
			dst[key] = sub
		}
	}
}

// dispatchObject determines whether dst is a map or is nil to decide what to
// submit to selectObject. If dst is nil it creates a new map, calls
// selectObject, and returns the result if it contains any values and nil when
// it does not. Otherwise it converts dst to a map and calls selectObject.
func (tree *Tree) dispatchObject(seg *Segment, root any, cur map[string]any, dst any) map[string]any {
	var sub map[string]any
	if dst != nil {
		var ok bool
		if sub, ok = dst.(map[string]any); !ok {
			// This should not happen.
			panic(fmt.Sprintf("jsontree: expected destination object but got %T", dst))
		}
		tree.selectObjectSegment(seg, root, cur, sub)
		return sub
	}

	// Set up the destination object.
	sub = map[string]any{}
	tree.selectObjectSegment(seg, root, cur, sub)

	// Don't bother to keep the object if there's nothing in it.
	if len(sub) > 0 {
		return sub
	}

	return nil
}

// processIndex fetches the value for idx from src and, if the value exists,
// stores it in dst. Indexes are always preserved: when idx is 2, the value
// will be stored in index 2 in dst, even if its length was 0.
//
// If the value is a JSON object (map[string]any) or array ([]any), it
// dispatches selection for that value so that seg's children can select from
// the value. Returns the updated dst.
//
// Note: cap(dst) MUST be equal to len(src). Callers should create dst like
// so:
//
//	dst := make([]any, 0, cap(src))
func (tree *Tree) processIndex(idx int, seg *Segment, root any, cur, dst []any) []any {
	prevLen := len(dst)
	// Grow the destination to the index, if necessary.
	if idx >= prevLen {
		dst = dst[:idx+1]
	} else {
		prevLen = -1
	}

	// Keep the value if it's the end of the path.
	if len(seg.children) == 0 {
		dst[idx] = cur[idx]
		return dst
	}

	// Allow the child segments to select from an object or array. Return the
	// updated dst.
	switch val := cur[idx].(type) {
	case map[string]any:
		if sub := tree.dispatchObject(seg, root, val, dst[idx]); sub != nil {
			dst[idx] = sub
			return dst
		}
	case []any:
		if sub := tree.dispatchArray(seg, root, val, dst[idx]); sub != nil {
			dst[idx] = sub
			return dst
		}
	}

	// Nothing found, restore dst to its original length and return.
	if prevLen > -1 {
		dst = dst[:prevLen]
	}
	return dst
}

// dispatchArray determines whether dst is a slice or is nil to decide what to
// submit to selectArray. If dst is nil it creates a new slice and passes it
// to selectArray. Otherwise it converts dst to a slice and passes it to
// selectArray.
func (tree *Tree) dispatchArray(seg *Segment, root any, cur []any, dstVal any) []any {
	var sub []any
	if dstVal == nil {
		// Set up the destination slice.
		sub = make([]any, 0, cap(cur))
	} else {
		// Make sure dst is a slice.
		var ok bool
		if sub, ok = dstVal.([]any); !ok {
			// This should not happen.
			panic(fmt.Sprintf("jsontree: expected destination array but got %T", dstVal))
		}
	}

	// Select seg from src into sub and return the updated slice.
	return tree.selectArraySegment(seg, root, cur, sub)
}

// selectArraySegment uses the selectors in seg to select paths from src into
// dst and recurses into its children. Returns the updated dst or nil if it's
// empty.
func (tree *Tree) selectArraySegment(seg *Segment, root any, cur, dst []any) []any {
	dst = tree.selectArray(seg, root, cur, dst)
	for _, seg := range seg.children {
		dst = tree.selectArray(seg, root, cur, dst)
	}

	if len(dst) == 0 {
		return nil
	}
	return dst
}

// selectArray uses the selectors in seg to select paths from src to dst. If
// seg is a descendant Segment, it recursively selects from seg into all of
// src's values.
func (tree *Tree) selectArray(n *Segment, root any, cur, dst []any) []any {
	for _, sel := range n.selectors {
		switch sel := sel.(type) {
		case spec.Index:
			idx := int(sel)
			if idx < len(cur) {
				dst = tree.processIndex(idx, n, root, cur, dst)
			}
		case spec.WildcardSelector:
			for i := range cur {
				dst = tree.processIndex(i, n, root, cur, dst)
			}
		case spec.SliceSelector:
			dst = tree.processSlice(n, sel, root, cur, dst)
		case *spec.FilterSelector:
			for i, v := range cur {
				if sel.Eval(v, root) {
					dst = tree.processIndex(i, n, root, cur, dst)
				}
			}
		}
	}

	if n.descendant {
		dst = tree.descendArray(n, root, cur, dst)
	}

	return dst
}

// processSlice iterates over the list of array indexes from sel and
// dispatches them to [processIndex].
//
// Note: cap(dst) MUST be equal to len(src). Callers should create dst like
// so:
//
//	dst := make([]any, 0, cap(src))
func (tree *Tree) processSlice(seg *Segment, sel spec.SliceSelector, root any, cur, dst []any) []any {
	// When step == 0, no elements are selected.
	switch {
	case sel.Step() > 0:
		lower, upper := sel.Bounds(len(cur))
		for i := lower; i < upper; i += sel.Step() {
			dst = tree.processIndex(i, seg, root, cur, dst)
		}
	case sel.Step() < 0:
		lower, upper := sel.Bounds(len(cur))
		for i := upper; lower < i; i += sel.Step() {
			dst = tree.processIndex(i, seg, root, cur, dst)
		}
	}
	return dst
}

// descendArray selects the paths from seg from each value from src into
// dst.
func (tree *Tree) descendArray(seg *Segment, root any, cur, dst []any) []any {
	dstLen := len(dst)
	for i, v := range cur {
		// Grab the destination array if it exists.
		var subDest any
		if i < dstLen {
			subDest = dst[i]
		}
		switch v := v.(type) {
		case map[string]any:
			if sub := tree.dispatchObject(seg, root, v, subDest); sub != nil {
				// We have data. Resize the array and save it.
				if i >= dstLen {
					dst = dst[:i+1]
				}
				dst[i] = sub
			}
		case []any:
			if sub := tree.dispatchArray(seg, root, v, subDest); sub != nil {
				// We have data. Resize the array and save it.
				if i >= dstLen {
					dst = dst[:i+1]
				}
				dst[i] = sub
			}
		}
	}

	// Always return the updated slice.
	return dst
}
