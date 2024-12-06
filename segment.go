package jsontree

import (
	"math"
	"strings"

	"github.com/theory/jsonpath/spec"
)

// Segment represents a single segment in a JSONTree query.
type Segment struct {
	selectors  []spec.Selector
	children   []*Segment
	descendant bool
}

// Child creates and returns a child ([<selectors>]) Segment.
func Child(sel ...spec.Selector) *Segment {
	return &Segment{selectors: sel, children: []*Segment{}}
}

// Descendant creates and returns a descendant (..[<selectors>]) Segment.
func Descendant(sel ...spec.Selector) *Segment {
	return &Segment{selectors: sel, descendant: true, children: []*Segment{}}
}

// Append appends child as child segments of seg.
func (seg *Segment) Append(child ...*Segment) *Segment {
	seg.children = append(seg.children, child...)
	return seg
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// hasSelector returns true if seg contains sel and false if it does not.
// Accounts for [spec.Index]es in [spec.SliceSelector]s, [spec.SliceSelector]
// overlap, and compares [*spec.FilterSelector] strings.
func (seg *Segment) hasSelector(sel spec.Selector) bool {
	return selectorsContain(seg.selectors, sel)
}

// hasSelectors returns true selectors is a subset of seg.selectors.
func (seg *Segment) hasSelectors(selectors []spec.Selector) bool {
	for _, sel := range selectors {
		if !seg.hasSelector(sel) {
			return false
		}
	}
	return true
}

// hasSameSelectors returns true if seg's selectors are the same as selectors.
func (seg *Segment) hasSameSelectors(selectors []spec.Selector) bool {
	return len(seg.selectors) == len(selectors) && seg.hasSelectors(selectors)
}

// selectorsContain returns true if selectors contains sel and false if it
// does not. Accounts for [spec.Index]es in [spec.SliceSelector]s,
// [spec.SliceSelector] overlap, and compares [*spec.FilterSelector] strings.
func selectorsContain(selectors []spec.Selector, sel spec.Selector) bool {
	if len(selectors) == 1 {
		// A wildcard selector should always be the only selector.
		if _, ok := selectors[0].(spec.WildcardSelector); ok {
			return true
		}
	}

	// Search for the segment by type.
	switch sel := sel.(type) {
	case spec.WildcardSelector:
		return false
	case spec.Name:
		if containsName(selectors, sel) {
			return true
		}
	case spec.Index:
		if containsIndex(selectors, sel) {
			return true
		}
	case spec.SliceSelector:
		if containsSlice(selectors, sel) {
			return true
		}
	case *spec.FilterSelector:
		if containsFilter(selectors, sel) {
			return true
		}
	}

	return false
}

// hasExactSelector returns true if seg's selectors contains the same selector
// as sel and false if it does not. [spec.Index]es do not match
// [spec.SliceSelector]s, [spec.SliceSelector]s must be identical, and
// compares [*spec.FilterSelector] strings.
func (seg *Segment) hasExactSelector(sel spec.Selector) bool {
	// Search for the segment by type.
	switch sel := sel.(type) {
	case spec.WildcardSelector:
		return seg.isWildcard()

	case spec.Name:
		if containsName(seg.selectors, sel) {
			return true
		}
	case spec.Index:
		for _, s := range seg.selectors {
			if s, ok := s.(spec.Index); ok && s == sel {
				return true
			}
		}
	case spec.SliceSelector:
		for _, s := range seg.selectors {
			if s, ok := s.(spec.SliceSelector); ok && s == sel {
				return true
			}
		}

	case *spec.FilterSelector:
		if containsFilter(seg.selectors, sel) {
			return true
		}
	}

	return false
}

// hasExactSelectors returns true seg contains exactly selectors.
// [spec.Index]es do not match [spec.SliceSelector]s, [spec.SliceSelector]s
// must be identical, and compares [*spec.FilterSelector] strings.
func (seg *Segment) hasExactSelectors(selectors []spec.Selector) bool {
	if len(seg.selectors) != len(selectors) {
		return false
	}
	for _, sel := range selectors {
		if !seg.hasExactSelector(sel) {
			return false
		}
	}
	return true
}

// containsName returns true if selectors contains name.
func containsName(selectors []spec.Selector, name spec.Name) bool {
	for _, s := range selectors {
		if s, ok := s.(spec.Name); ok {
			if s == name {
				return true
			}
		}
	}
	return false
}

// containsIndex returns true if selectors contains idx. It evaluates both
// [spec.Index] values and [spec.SliceSelector]s with positive start and end
// values and positive steps or a -1 step where end < start. Supports both
// positive and negative idx values within those constraints.
func containsIndex(selectors []spec.Selector, idx spec.Index) bool {
	for _, s := range selectors {
		switch s := s.(type) {
		case spec.Index:
			if s == idx {
				return true
			}
		case spec.SliceSelector:
			// Negative bounds and backward slice without -1 step depend on
			// input length, so cannot be determined independently.
			if s.Start() < 0 || s.End() < 0 || (s.End() < s.Start() && s.Step() != -1) {
				return false
			}

			// Set sized based on the slice params and determine the bounds.
			sel := int(idx)
			size := max(abs(sel), s.Start(), s.End())
			if size != math.MaxInt {
				size++
			}
			lower, upper := s.Bounds(size)

			if sel < 0 {
				// Subtract the index from the upper bound.
				sel = upper + sel
			}

			step := s.Step()
			switch {
			// step == 0 never selects values.
			case step > 0 && sel >= lower && sel < upper && ((sel-lower)%step == 0):
				return true
			case step == -1 && sel <= upper && sel > lower:
				// All other negative steps depend on input length.
				return true
			}
		}
	}

	return false
}

// containsSlice returns true if selectors contains slice. To qualify, slice's
// start and end must come between the start and end of a slice in seg, and
// the step of that slice must be a multiple of slice's step. Or, slice must
// select a single element that is the same as a [spec.Index] in seg.
//
// In theory, all the spec.Index values in seg could account for all the
// indexes in the slice. But this is good enough for now.
func containsSlice(selectors []spec.Selector, slice spec.SliceSelector) bool {
	if slice.Step() == 0 ||
		slice.Start() == slice.End() ||
		(slice.Step() > 0 && slice.Start() > slice.End()) ||
		(slice.Step() < 0 && slice.Start() < slice.End()) {
		// Never selects anything, so true.
		return true
	}

	for _, s := range selectors {
		switch s := s.(type) {
		case spec.SliceSelector:
			if ok := sliceInSlice(slice, s); ok {
				return true
			}
		case spec.Index:
			idx := int(s)
			if (slice.Start() == idx && slice.End() == idx+1 && slice.Step() > 0) ||
				(slice.Start() == idx+1 && slice.End() == idx && slice.Step() < 0) {
				return true
			}
		}
	}

	return false
}

// sliceInSlice returns true if sub is a subset of or equal to sup. Always
// returns false if sub.step is not a multiple of sup.step. Accounts for
// logical subsets where the steps for one slice are positive and the other
// negative.
func sliceInSlice(sub, sup spec.SliceSelector) bool {
	if sub.Step()%sup.Step() != 0 {
		// Non overlapping if s.Step() is not a multiple of slice.Step()1.
		return false
	}

	switch {
	case sub.Step() > 0 && sup.Step() > 0:
		// Most common case: is sub between sup start and end?
		if sub.Start() >= sup.Start() && sub.End() <= sup.End() {
			return true
		}
	case sub.Step() < 0 && sup.Step() < 0:
		// Both step backward: is sub between sup end and start?
		if sub.Start() <= sup.Start() && sub.End() >= sup.End() {
			return true
		}
	case sub.Step() <= 1 && sup.Step() > 0:
		// sub backward vs sup forward: is sub between sup end and start?
		if sub.Start() < sup.End() && sub.End() >= sup.Start()-1 {
			return true
		}
	case sub.Step() > 0 && sup.Step() < 0:
		// sub forward vs sup backward: is sub between sup start and end?
		if sub.End() > sup.Start() && sub.Start()-1 >= sup.End() {
			return true
		}
	}
	return false
}

// containsFilter returns true if selectors contains filter. Currently relies on
// string comparison, but could be improved by implementing [sort.Interface]
// for [spec.LogicalOr] and [spec.LogicalAnd], as well as operand and operator
// normalization in [spec.ComparisonExpr].
//
// [sort.Interface]: https://pkg.go.dev/sort#Interface
func containsFilter(selectors []spec.Selector, filter *spec.FilterSelector) bool {
	for _, s := range selectors {
		if s, ok := s.(*spec.FilterSelector); ok {
			if s.String() == filter.String() {
				return true
			}
		}
	}

	return false
}

// isBranch returns true if seg's descendants constitute a single branch with
// the same selectors as specSeg.
func (seg *Segment) isBranch(specSeg []*spec.Segment) bool {
	cur := seg
	size := len(specSeg)
	for i, c := range specSeg {
		if i >= size || len(cur.children) != 1 {
			return false
		}

		cur = cur.children[0]
		if cur.descendant != c.IsDescendant() {
			return false
		}

		if len(cur.selectors) != len(c.Selectors()) || !cur.hasSelectors(c.Selectors()) {
			return false
		}
	}
	return len(cur.children) == 0
}

// mergeSelectors merges selectors into seg.selectors and return seg.
func (seg *Segment) mergeSelectors(selectors []spec.Selector) *Segment {
	for _, sel := range selectors {
		if !seg.hasSelector(sel) {
			seg.selectors = append(seg.selectors, sel)
		}
	}
	return seg
}

// deduplicate recursively deduplicates seg. In other words, for any child
// segment with all of its selectors and descendant branches also held by
// another child segment, the former will be merged into the latter. It also
// merges slice selectors where one slice is clearly a subset of another.
func (seg *Segment) deduplicate() {
	merged := []*Segment{}

	for _, child := range seg.children {
		child.deduplicate()
		skip := false
		for i, prev := range merged {
			if !prev.sameBranches(child) {
				continue
			}
			// Can probably merge.
			switch {
			case prev.descendant == child.descendant:
				// Merge.
				prev.mergeSelectors(child.selectors)
				skip = true
			case child.descendant:
				// Remove common selectors from prev.
				if skip = child.removeCommonSelectorsFrom(prev); skip {
					// Replace prev with child
					merged[i] = child
				}
			case prev.descendant:
				// Remove common selectors from child
				skip = prev.removeCommonSelectorsFrom(child)
			}
		}
		if !skip {
			merged = append(merged, child)
		}
	}

	if len(merged) != len(seg.children) {
		// XXX Shrink merged cap to its len?
		seg.children = merged
	}
	seg.mergeSlices()
}

// mergeSlices compares [spec.SliceSelector]s in seg.selectors, and eliminates
// any that are clear subsets of another.
func (seg *Segment) mergeSlices() {
	merged := []spec.Selector{}
SEL:
	for _, sel := range seg.selectors {
		if sel, ok := sel.(spec.SliceSelector); ok {
			for i, ss := range merged {
				if ss, ok := ss.(spec.SliceSelector); ok {
					if sliceInSlice(sel, ss) {
						continue SEL
					}
					if sliceInSlice(ss, sel) {
						merged[i] = sel
						continue SEL
					}
				}
			}
		}
		merged = append(merged, sel)
	}

	if len(merged) != len(seg.selectors) {
		// XXX Shrink merged cap to its len?
		seg.selectors = merged
	}
}

// removeCommonSelectorsFrom removes selectors from seg2 that are present in
// seg. Returns true if all selectors are removed from seg2.
func (seg *Segment) removeCommonSelectorsFrom(seg2 *Segment) bool {
	// Prune common selectors.
	uniqueSel := []spec.Selector{}
	for _, sel := range seg2.selectors {
		if !seg.hasSelector(sel) {
			uniqueSel = append(uniqueSel, sel)
		}
	}

	switch len(uniqueSel) {
	case len(seg2.selectors):
		// None in common.
		return false
	case 0:
		// All merged
		// XXX Shrink merged cap to its len?
		seg2.selectors = uniqueSel
		return true
	default:
		// Save only retained selectors.
		// XXX Shrink merged cap to its len?
		seg2.selectors = uniqueSel
		return false
	}
}

// sameBranches returns true if seg has the same branches as seg2. It
// recursively compares seg's children to seg2's children to ensure they have
// the exactly the same selectors and descendants. [spec.Index]es do not match
// [spec.SliceSelector]s, [spec.SliceSelector]s must be identical, and
// compares [*spec.FilterSelector] strings.
func (seg *Segment) sameBranches(seg2 *Segment) bool {
	if len(seg.children) != len(seg2.children) {
		// Let leaf nodes merge?
		return false
	}

C1:
	for _, c1 := range seg.children {
		for _, c2 := range seg2.children {
			if c1.hasExactSelectors(c2.selectors) && c1.sameBranches(c2) {
				continue C1
			}
		}
		return false
	}
	return true
}

// isWildcard returns true if seg is a wildcard selector.
func (seg *Segment) isWildcard() bool {
	if len(seg.selectors) != 1 {
		return false
	}
	_, ok := seg.selectors[0].(spec.WildcardSelector)
	return ok
}

// String returns a string representation of seg's child segments in as a tree
// diagram.
func (seg *Segment) String() string {
	buf := new(strings.Builder)
	lastIndex := len(seg.children) - 1
	for i, c := range seg.children {
		c.writeTo(buf, "", i == lastIndex)
	}
	return buf.String()
}

const (
	elbow = "└── "
	pipe  = "│   "
	tee   = "├── "
	blank = "    "
)

// writeTo writes the string representation of seg to buf.
func (seg *Segment) writeTo(buf *strings.Builder, prefix string, last bool) {
	buf.WriteString(prefix)
	if last {
		buf.WriteString(elbow)
	} else {
		buf.WriteString(tee)
	}

	if seg.descendant {
		buf.WriteString("..")
	}
	buf.WriteByte('[')
	for i, sel := range seg.selectors {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(sel.String())
	}
	buf.WriteString("]\n")

	lastIndex := len(seg.children) - 1
	for i, sub := range seg.children {
		if last {
			sub.writeTo(buf, prefix+blank, i == lastIndex)
		} else {
			sub.writeTo(buf, prefix+pipe, i == lastIndex)
		}
	}
}
