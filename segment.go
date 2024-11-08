package jsontree

import (
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
