package jsontree

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectorInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		tok  any
	}{
		{"name", Name("hi")},
		{"index", Index(42)},
		{"slice", Slice()},
		{"wildcard", Wildcard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*selector)(nil), tc.tok)
		})
	}
}

func TestSelectorWriteTo(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		tok  selector
		str  string
	}{
		{
			name: "name",
			tok:  Name("hi"),
			str:  `"hi"`,
		},
		{
			name: "name_space",
			tok:  Name("hi there"),
			str:  `"hi there"`,
		},
		{
			name: "name_quote",
			tok:  Name(`hi "there"`),
			str:  `"hi \"there\""`,
		},
		{
			name: "name_unicode",
			tok:  Name(`hi 😀`),
			str:  `"hi 😀"`,
		},
		{
			name: "name_digits",
			tok:  Name(`42`),
			str:  `"42"`,
		},
		{
			name: "index",
			tok:  Index(42),
			str:  "42",
		},
		{
			name: "index_big",
			tok:  Index(math.MaxUint32),
			str:  "4294967295",
		},
		{
			name: "index_zero",
			tok:  Index(0),
			str:  "0",
		},
		{
			name: "slice_0_4",
			tok:  Slice(0, 4),
			str:  ":4",
		},
		{
			name: "slice_4_5",
			tok:  Slice(4, 5),
			str:  "4:5",
		},
		{
			name: "slice_end_42",
			tok:  Slice(nil, 42),
			str:  ":42",
		},
		{
			name: "slice_start_4",
			tok:  Slice(4),
			str:  "4:",
		},
		{
			name: "slice_start_end_step",
			tok:  Slice(4, 7, 2),
			str:  "4:7:2",
		},
		{
			name: "slice_start_step",
			tok:  Slice(4, nil, 2),
			str:  "4::2",
		},
		{
			name: "slice_end_step",
			tok:  Slice(nil, 4, 2),
			str:  ":4:2",
		},
		{
			name: "slice_step",
			tok:  Slice(nil, nil, 3),
			str:  "::3",
		},
		{
			name: "wildcard",
			tok:  Wildcard,
			str:  "*",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(strings.Builder)
			tc.tok.writeTo(buf)
			a.Equal(tc.str, buf.String())
		})
	}
}

func TestSliceBounds(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	json := []any{"a", "b", "c", "d", "e", "f", "g"}

	extract := func(s SliceSelector) []any {
		lower, upper := s.bounds(len(json))
		res := make([]any, 0, len(json))
		switch {
		case s.step > 0:
			for i := lower; i < upper; i += s.step {
				res = append(res, json[i])
			}
		case s.step < 0:
			for i := upper; lower < i; i += s.step {
				res = append(res, json[i])
			}
		}
		return res
	}

	type lenCase struct {
		length int
		lower  int
		upper  int
	}

	for _, tc := range []struct {
		name  string
		slice SliceSelector
		cases []lenCase
		exp   []any
	}{
		{
			name:  "defaults",
			slice: Slice(),
			exp:   json,
			cases: []lenCase{
				{10, 0, 10},
				{3, 0, 3},
				{99, 0, 99},
			},
		},
		{
			name:  "step_0",
			slice: Slice(nil, nil, 0),
			exp:   []any{},
			cases: []lenCase{
				{10, 0, 0},
				{3, 0, 0},
				{99, 0, 0},
			},
		},
		{
			name:  "3_8_2",
			slice: Slice(3, 8, 2),
			exp:   []any{"d", "f"},
			cases: []lenCase{
				{10, 3, 8},
				{3, 3, 3},
				{99, 3, 8},
			},
		},
		{
			name:  "1_3_1",
			slice: Slice(1, 3, 1),
			exp:   []any{"b", "c"},
			cases: []lenCase{
				{10, 1, 3},
				{2, 1, 2},
				{99, 1, 3},
			},
		},
		{
			name:  "5_defaults",
			slice: Slice(5),
			exp:   []any{"f", "g"},
			cases: []lenCase{
				{10, 5, 10},
				{8, 5, 8},
				{99, 5, 99},
			},
		},
		{
			name:  "1_5_2",
			slice: Slice(1, 5, 2),
			exp:   []any{"b", "d"},
			cases: []lenCase{
				{10, 1, 5},
				{4, 1, 4},
				{99, 1, 5},
			},
		},
		{
			name:  "5_1_neg2",
			slice: Slice(5, 1, -2),
			exp:   []any{"f", "d"},
			cases: []lenCase{
				{10, 1, 5},
				{4, 1, 3},
				{99, 1, 5},
			},
		},
		{
			name:  "def_def_neg1",
			slice: Slice(nil, nil, -1),
			exp:   []any{"g", "f", "e", "d", "c", "b", "a"},
			cases: []lenCase{
				{10, -1, 9},
				{4, -1, 3},
				{99, -1, 98},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, lc := range tc.cases {
				lower, upper := tc.slice.bounds(lc.length)
				a.Equal(lc.lower, lower)
				a.Equal(lc.upper, upper)
			}
			a.Equal(tc.exp, extract(tc.slice))
		})
	}
}

func TestSlicePanic(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	a.PanicsWithValue(
		"First value passed to NewSlice is not an integer",
		func() { Slice("hi") },
	)
	a.PanicsWithValue(
		"Second value passed to NewSlice is not an integer",
		func() { Slice(nil, "hi") },
	)
	a.PanicsWithValue(
		"Third value passed to NewSlice is not an integer",
		func() { Slice(nil, 42, "hi") },
	)
}
