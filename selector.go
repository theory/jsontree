package jsontree

import (
	"math"
	"strconv"
	"strings"
)

// selector represents a single selector in a RFC 9535 JSONPath query.
type selector interface {
	// writeTo writes a string representation of a selector to buf.
	writeTo(buf *strings.Builder)
}

// Name is a key name selector, e.g., .name or ["name"].
type Name string

func (n Name) writeTo(buf *strings.Builder) {
	buf.WriteString(strconv.Quote(string(n)))
}

// wc is the underlying nil value used by [Wildcard].
type wc struct{}

// Wildcard is a wildcard selector, e.g., * or [*].
//
//nolint:gochecknoglobals
var Wildcard = wc{}

// writeTo "*" to buf.
func (wc) writeTo(buf *strings.Builder) { buf.WriteByte('*') }

// Index is an array index selector, e.g., [3].
type Index int

// writeTo writes a string representation of i to buf.
func (i Index) writeTo(buf *strings.Builder) {
	buf.WriteString(strconv.FormatInt(int64(i), 10))
}

// SliceSelector is a slice selector, e.g., [0:100:5].
type SliceSelector struct {
	// Start of the slice; defaults to 0.
	start int
	// End of the slice; defaults to math.MaxInt.
	end int
	// Steps between start and end; defaults to 0.
	step int
}

// Slice creates a new SliceSelector. Pass up to three integers or nils for
// the start, end, and step arguments. Subsequent arguments are ignored.
func Slice(args ...any) SliceSelector {
	const (
		startArg = 0
		endArg   = 1
		stepArg  = 2
	)
	// Set defaults.
	s := SliceSelector{0, math.MaxInt, 1}
	switch len(args) - 1 {
	case stepArg:
		//nolint:gosec // disable G602
		switch step := args[stepArg].(type) {
		case int:
			s.step = step
		case nil:
			// Nothing to do
		default:
			panic("Third value passed to NewSlice is not an integer")
		}
		fallthrough
	case endArg:
		//nolint:gosec // disable G602
		switch end := args[endArg].(type) {
		case int:
			s.end = end
		case nil:
			// Negative step: end with minimum int.
			if s.step < 0 {
				s.end = math.MinInt
			}
		default:
			panic("Second value passed to NewSlice is not an integer")
		}
		fallthrough
	case startArg:
		switch start := args[startArg].(type) {
		case int:
			s.start = start
		case nil:
			// Negative step: start with maximum int.
			if s.step < 0 {
				s.start = math.MaxInt
			}
		default:
			panic("First value passed to NewSlice is not an integer")
		}
	}
	return s
}

// writeTo writes a string representation of s to buf.
func (s SliceSelector) writeTo(buf *strings.Builder) {
	if s.start != 0 {
		buf.WriteString(strconv.FormatInt(int64(s.start), 10))
	}
	buf.WriteByte(':')
	if s.end != math.MaxInt {
		buf.WriteString(strconv.FormatInt(int64(s.end), 10))
	}
	if s.step != 1 {
		buf.WriteByte(':')
		buf.WriteString(strconv.FormatInt(int64(s.step), 10))
	}
}

// bounds returns the lower and upper bounds for selecting from a slice of
// length.
func (s SliceSelector) bounds(length int) (int, int) {
	start := normalize(s.start, length)
	end := normalize(s.end, length)
	switch {
	case s.step > 0:
		return max(min(start, length), 0), max(min(end, length), 0)
	case s.step < 0:
		return max(min(end, length-1), -1), max(min(start, length-1), -1)
	default:
		return 0, 0
	}
}

// normalize normalizes index i relative to a slice of length.
func normalize(i, length int) int {
	if i >= 0 {
		return i
	}

	return length + i
}
