/*
Package jsontree selects paths from one entity into a new entity.
*/
package jsontree

import (
	"fmt"
)

// JSONTree selects a subset of values in an entity and returns it. It preserves
// the full paths to the selected entities.
type JSONTree struct {
	root *Segment
}

// New creates a JSONTree query that selects seg and their children from a
// JSON value.
func New(seg ...*Segment) *JSONTree {
	return &JSONTree{root: Child().Append(seg...)}
}

// String returns a string representation of seg, starting from "$" for the
// root, and including all of its child segments in as a tree diagram.
func (jt *JSONTree) String() string {
	return "$\n" + jt.root.String()
}

// Select selects jt's tree of paths from the `from` JSON value into a new
// value. For a root-only JSONTree that contains no children, from will simply
// be returned. All other JSONTree queries will select from the from value if
// it's an array ([]any) or object (map[string]any). Returns nil for any other
// values.
func (jt *JSONTree) Select(from any) any {
	if len(jt.root.children) == 0 {
		return from
	}

	switch entity := from.(type) {
	case map[string]any:
		ret := map[string]any{}
		jt.selectObjectSegment(jt.root, entity, ret)
		return ret
	case []any:
		ret := make([]any, 0, cap(entity))
		if sel := jt.selectArraySegment(jt.root, entity, ret); sel != nil {
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
func (jt *JSONTree) selectObjectSegment(seg *Segment, src, dst map[string]any) {
	jt.selectObject(seg, src, dst)
	for _, seg := range seg.children {
		jt.selectObject(seg, src, dst)
	}
}

// selectObject uses the selectors in seg to select paths from src to dst. If
// seg is a descendant Segment, it recursively selects from seg into all of
// src's values.
func (jt *JSONTree) selectObject(seg *Segment, src, dst map[string]any) {
	for _, sel := range seg.selectors {
		switch sel := sel.(type) {
		case Name:
			jt.processKey(string(sel), seg, src, dst)
		case wc:
			for k := range src {
				jt.processKey(k, seg, src, dst)
			}
		}
	}
	if seg.descendant {
		jt.descendObject(seg, src, dst)
	}
}

// descendObject selects the paths from seg from each value from src into
// dst.
func (jt *JSONTree) descendObject(seg *Segment, src, dst map[string]any) {
	for k, v := range src {
		switch v := v.(type) {
		case map[string]any:
			if sub := jt.dispatchObject(seg, v, dst[k]); sub != nil {
				dst[k] = sub
			}
		case []any:
			if sub := jt.dispatchArray(seg, v, dst[k]); sub != nil {
				dst[k] = sub
			}
		}
	}
}

// processKey fetches the value for key from src and, if the value exists,
// stores it in dst. If the value is a JSON object (map[string]any) or array
// ([]any), it dispatches selection for that value so that seg's children can
// select from the value.
func (jt *JSONTree) processKey(key string, seg *Segment, src, dst map[string]any) {
	// Do we have a value?
	val, ok := src[key]
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
		sub := jt.dispatchObject(seg, val, dst[key])
		if sub != nil {
			dst[key] = sub
		}
	case []any:
		sub := jt.dispatchArray(seg, val, dst[key])
		if sub != nil {
			dst[key] = sub
		}
	}
}

// dispatchObject determines whether dst is a map or is nil to decide what to
// submit to selectObject. If dst is nil it creates a new map, calls
// selectObject, and returns the result if it contains any values and nil when
// it does not. Otherwise it converts dst to a map and calls selectObject.
func (jt *JSONTree) dispatchObject(seg *Segment, src map[string]any, dst any) map[string]any {
	var sub map[string]any
	if dst != nil {
		var ok bool
		if sub, ok = dst.(map[string]any); !ok {
			// This should not happen.
			panic(fmt.Sprintf("jsontree: expected destination object but got %T", dst))
		}
		jt.selectObjectSegment(seg, src, sub)
		return sub
	}

	// Set up the destination object.
	sub = map[string]any{}
	jt.selectObjectSegment(seg, src, sub)

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
func (jt *JSONTree) processIndex(idx int, seg *Segment, src, dst []any) []any {
	prevLen := len(dst)
	// Grow the destination to the index, if necessary.
	if idx >= prevLen {
		dst = dst[:idx+1]
	} else {
		prevLen = -1
	}

	// Keep the value if it's the end of the path.
	if len(seg.children) == 0 {
		dst[idx] = src[idx]
		return dst
	}

	// Allow the child segments to select from an object or array. Return the
	// updated dst.
	switch val := src[idx].(type) {
	case map[string]any:
		if sub := jt.dispatchObject(seg, val, dst[idx]); sub != nil {
			dst[idx] = sub
			return dst
		}
	case []any:
		if sub := jt.dispatchArray(seg, val, dst[idx]); sub != nil {
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
func (jt *JSONTree) dispatchArray(seg *Segment, src []any, dstVal any) []any {
	var sub []any
	if dstVal == nil {
		// Set up the destination slice.
		sub = make([]any, 0, cap(src))
	} else {
		// Make sure dst is a slice.
		var ok bool
		if sub, ok = dstVal.([]any); !ok {
			// This should not happen.
			panic(fmt.Sprintf("jsontree: expected destination array but got %T", dstVal))
		}
	}

	// Select seg from src into sub and return the updated slice.
	return jt.selectArraySegment(seg, src, sub)
}

// selectArraySegment uses the selectors in seg to select paths from src into
// dst and recurses into its children. Returns the updated dst or nil if it's
// empty.
func (jt *JSONTree) selectArraySegment(seg *Segment, src, dst []any) []any {
	dst = jt.selectArray(seg, src, dst)
	for _, seg := range seg.children {
		dst = jt.selectArray(seg, src, dst)
	}

	if len(dst) == 0 {
		return nil
	}
	return dst
}

// selectArray uses the selectors in seg to select paths from src to dst. If
// seg is a descendant Segment, it recursively selects from seg into all of
// src's values.
func (jt *JSONTree) selectArray(n *Segment, src, dst []any) []any {
	for _, sel := range n.selectors {
		switch sel := sel.(type) {
		case Index:
			idx := int(sel)
			if idx < len(src) {
				dst = jt.processIndex(idx, n, src, dst)
			}
		case wc:
			for i := range src {
				dst = jt.processIndex(i, n, src, dst)
			}
		case SliceSelector:
			dst = jt.processSlice(n, sel, src, dst)
		}
	}

	if n.descendant {
		dst = jt.descendArray(n, src, dst)
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
func (jt *JSONTree) processSlice(seg *Segment, sel SliceSelector, src, dst []any) []any {
	// When step == 0, no elements are selected.
	switch {
	case sel.step > 0:
		lower, upper := sel.bounds(len(src))
		for i := lower; i < upper; i += sel.step {
			dst = jt.processIndex(i, seg, src, dst)
		}
	case sel.step < 0:
		lower, upper := sel.bounds(len(src))
		for i := upper; lower < i; i += sel.step {
			dst = jt.processIndex(i, seg, src, dst)
		}
	}
	return dst
}

// descendArray selects the paths from seg from each value from src into
// dst.
func (jt *JSONTree) descendArray(seg *Segment, src, dst []any) []any {
	dstLen := len(dst)
	for i, v := range src {
		// Grab the destination array if it exists.
		var subDest any
		if i < dstLen {
			subDest = dst[i]
		}
		switch v := v.(type) {
		case map[string]any:
			if sub := jt.dispatchObject(seg, v, subDest); sub != nil {
				// We have data. Resize the array and save it.
				if i >= dstLen {
					dst = dst[:i+1]
				}
				dst[i] = sub
			}
		case []any:
			if sub := jt.dispatchArray(seg, v, subDest); sub != nil {
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
