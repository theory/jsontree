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

// Contains returns true if seg contains sel and false if it does not.
// Accounts for [spec.Index]es in [spec.SliceSelector]s, [spec.SliceSelector]
// overlap, and compares [*spec.FilterSelector] strings.
func (seg *Segment) Contains(sel spec.Selector) bool {
	if len(seg.selectors) == 1 {
		// A wildcard selector should always be the only selector.
		if _, ok := seg.selectors[0].(spec.WildcardSelector); ok {
			return true
		}
	}

	// Search for the segment by type.
	switch sel := sel.(type) {
	case spec.WildcardSelector:
		return false
	case spec.Name:
		if seg.containsName(sel) {
			return true
		}
	case spec.Index:
		if seg.containsIndex(sel) {
			return true
		}
	case spec.SliceSelector:
		if seg.containsSlice(sel) {
			return true
		}
	case *spec.FilterSelector:
		if seg.containsFilter(sel) {
			return true
		}
	}

	return false
}

// containsName returns true if seg contains name.
func (seg *Segment) containsName(name spec.Name) bool {
	for _, s := range seg.selectors {
		if s, ok := s.(spec.Name); ok {
			if s == name {
				return true
			}
		}
	}
	return false
}

// containsIndex returns true if seg contains idx. It evaluates both
// [spec.Index] values and [spec.SliceSelector]s with positive start and end
// values and positive steps or a -1 step where end < start. Supports both
// positive and negative idx values within those constraints.
func (seg *Segment) containsIndex(idx spec.Index) bool {
	for _, s := range seg.selectors {
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

// containsSlice returns true if seg contains slice. To qualify, slice's start
// and end must come between the start and end of a slice in seg, and the step
// of that slice must be a multiple of slice's step. Or, slice must select a
// single element that is the same as a [spec.Index] in seg.
//
// In theory, all the spec.Index values in seg could account for all the
// indexes in the slice. But this is good enough for now.
func (seg *Segment) containsSlice(slice spec.SliceSelector) bool {
	if slice.Step() == 0 ||
		slice.Start() == slice.End() ||
		(slice.Step() > 0 && slice.Start() > slice.End()) ||
		(slice.Step() < 0 && slice.Start() < slice.End()) {
		// Never selects anything, so true.
		return true
	}

	for _, s := range seg.selectors {
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

// containsFilter returns true if seg contains filter. Currently relies on
// string comparison, but could be improved by implementing [sort.Interface]
// for [spec.LogicalOr] and [spec.LogicalAnd], as well as operand and operator
// normalization in [spec.ComparisonExpr].
//
// [sort.Interface]: https://pkg.go.dev/sort#Interface
func (seg *Segment) containsFilter(filter *spec.FilterSelector) bool {
	for _, s := range seg.selectors {
		if s, ok := s.(*spec.FilterSelector); ok {
			if s.String() == filter.String() {
				return true
			}
		}
	}

	return false
}

// String returns a string representation of seg, including all of its child
// segments in as a tree diagram.
func (seg *Segment) String() string {
	buf := new(strings.Builder)
	lastIndex := len(seg.children) - 1
	for i, seg := range seg.children {
		seg.writeTo(buf, "", i == lastIndex)
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
