package jsontree

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

func TestWriteNode(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		segs []*Segment
		str  string
	}{
		{
			name: "root_only",
			str:  "$\n",
		},
		{
			name: "wildcard",
			segs: []*Segment{Child(spec.Wildcard)},
			str:  "$\n└── [*]\n",
		},
		{
			name: "one_key",
			segs: []*Segment{Child(spec.Name("foo"))},
			str:  "$\n└── [\"foo\"]\n",
		},
		{
			name: "two_keys",
			segs: []*Segment{Child(spec.Name("foo"), spec.Name("bar"))},
			str:  "$\n└── [\"foo\",\"bar\"]\n",
		},
		{
			name: "two_segments",
			segs: []*Segment{Child(spec.Name("foo")), Child(spec.Name("bar"))},
			str:  "$\n├── [\"foo\"]\n└── [\"bar\"]\n",
		},
		{
			name: "two_keys_and_sub_keys",
			segs: []*Segment{
				Child(spec.Name("foo")).Append(
					Child(spec.Name("x")),
					Child(spec.Name("y")),
					Descendant(spec.Name("z")),
				),
				Child(spec.Name("bar")).Append(
					Child(spec.Name("a"), spec.Index(42), spec.Slice(0, 8, 2)),
					Child(spec.Name("b")),
					Child(spec.Name("c")),
				),
			},
			str: `$
├── ["foo"]
│   ├── ["x"]
│   ├── ["y"]
│   └── ..["z"]
└── ["bar"]
    ├── ["a",42,:8:2]
    ├── ["b"]
    └── ["c"]
`,
		},
		{
			name: "mixed_and_deep",
			segs: []*Segment{
				Child(spec.Name("foo")).Append(
					Child(spec.Name("x")),
					Child(spec.Name("y")).Append(
						Child(spec.Wildcard).Append(
							Child(spec.Name("a")),
							Child(spec.Name("b")),
						),
					),
				),
				Child(spec.Name("bar")).Append(
					Child(spec.Name("go")),
					Child(spec.Name("z")).Append(
						Child(spec.Wildcard).Append(
							Child(spec.Name("c")),
							Child(spec.Name("d")).Append(
								Child(spec.Slice(2, 3)),
							),
						),
					),
					Child(spec.Name("hi")),
				),
			},
			str: `$
├── ["foo"]
│   ├── ["x"]
│   └── ["y"]
│       └── [*]
│           ├── ["a"]
│           └── ["b"]
└── ["bar"]
    ├── ["go"]
    ├── ["z"]
    │   └── [*]
    │       ├── ["c"]
    │       └── ["d"]
    │           └── [2:3]
    └── ["hi"]
`,
		},
		{
			name: "wildcard",
			segs: []*Segment{Child(spec.Wildcard)},
			str:  "$\n└── [*]\n",
		},
		{
			name: "one_index",
			segs: []*Segment{Child(spec.Index(0))},
			str:  "$\n└── [0]\n",
		},
		{
			name: "two_indexes",
			segs: []*Segment{Child(spec.Index(0), spec.Index(2))},
			str:  "$\n└── [0,2]\n",
		},
		{
			name: "other_two_indexes",
			segs: []*Segment{Child(spec.Index(0)), Child(spec.Index(2))},
			str:  "$\n├── [0]\n└── [2]\n",
		},
		{
			name: "index_index",
			segs: []*Segment{Child(spec.Index(0)).Append(Child(spec.Index(2)))},
			str:  "$\n└── [0]\n    └── [2]\n",
		},
		{
			name: "two_keys_and_sub_indexes",
			segs: []*Segment{
				Child(spec.Name("foo")).Append(
					Child(spec.Index(0)),
					Child(spec.Index(1)),
					Child(spec.Index(2)),
				),
				Child(spec.Name("bar")).Append(
					Child(spec.Index(3)),
					Child(spec.Index(4)),
					Child(spec.Index(5)),
				),
			},
			str: `$
├── ["foo"]
│   ├── [0]
│   ├── [1]
│   └── [2]
└── ["bar"]
    ├── [3]
    ├── [4]
    └── [5]
`,
		},
		{
			name: "filter",
			segs: []*Segment{Child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Paren(spec.LogicalOr{spec.LogicalAnd{
					spec.Existence(spec.Query(true, []*spec.Segment{})),
				}}),
			}}))},
			str: "$\n└── [?($)]\n",
		},
		{
			name: "filter_and_key",
			segs: []*Segment{
				Child(spec.Name("x")),
				Child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
					spec.Paren(spec.LogicalOr{spec.LogicalAnd{
						spec.Existence(spec.Query(true, []*spec.Segment{})),
					}}),
				}})),
			},
			str: "$\n├── [\"x\"]\n└── [?($)]\n",
		},
		{
			name: "filter_and_key_segment",
			segs: []*Segment{
				Child(
					spec.Name("x"),
					spec.Filter(spec.LogicalOr{spec.LogicalAnd{
						spec.Paren(spec.LogicalOr{spec.LogicalAnd{
							spec.Existence(spec.Query(true, []*spec.Segment{})),
						}}),
					}}),
				),
			},
			str: "$\n└── [\"x\",?($)]\n",
		},
		{
			name: "nested_filter",
			segs: []*Segment{Child(spec.Name("x")).Append(
				Child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
					spec.Paren(spec.LogicalOr{spec.LogicalAnd{
						spec.Existence(spec.Query(true, []*spec.Segment{})),
					}}),
				}})),
			)},
			str: "$\n└── [\"x\"]\n    └── [?($)]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n := Tree{root: &Segment{children: tc.segs}}
			a.Equal(tc.str, n.String())
		})
	}
}

func TestIsWildcard(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *Segment
		exp  bool
	}{
		{"empty", &Segment{}, false},
		{"name", &Segment{selectors: []spec.Selector{spec.Name("x")}}, false},
		{"index", &Segment{selectors: []spec.Selector{spec.Index(0)}}, false},
		{"slice", &Segment{selectors: []spec.Selector{spec.Slice()}}, false},
		{"filter", &Segment{selectors: []spec.Selector{mkFilter("$[?@]")}}, false},
		{"wildcard", &Segment{selectors: []spec.Selector{spec.Wildcard}}, true},
		{"multiples", &Segment{selectors: []spec.Selector{spec.Wildcard, spec.Index(0)}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.seg.isWildcard())
		})
	}
}

func TestHasSelector(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		name  string
		list  []spec.Selector
		sel   spec.Selector
		exp   bool
		exact bool
	}{
		{
			name: "wildcard_empty",
			list: []spec.Selector{},
			sel:  spec.Wildcard,
		},
		{
			name: "wildcard_name",
			list: []spec.Selector{spec.Name("foo")},
			sel:  spec.Wildcard,
		},
		{
			name: "wildcard_index",
			list: []spec.Selector{spec.Index(1)},
			sel:  spec.Wildcard,
		},
		{
			name:  "wildcard_wildcard",
			list:  []spec.Selector{spec.Wildcard},
			sel:   spec.Wildcard,
			exp:   true,
			exact: true,
		},
		{
			name: "name_empty",
			list: []spec.Selector{},
			sel:  spec.Name("foo"),
		},
		{
			name:  "name_exists",
			list:  []spec.Selector{spec.Name("foo")},
			sel:   spec.Name("foo"),
			exp:   true,
			exact: true,
		},
		{
			name:  "name_exists_list",
			list:  []spec.Selector{spec.Name("foo"), spec.Name("bar"), spec.Index(0)},
			sel:   spec.Name("foo"),
			exp:   true,
			exact: true,
		},
		{
			name: "name_not_exists_list",
			list: []spec.Selector{spec.Name("foo"), spec.Name("bar"), spec.Index(0)},
			sel:  spec.Name("hello"),
		},
		{
			name: "name_wildcard",
			list: []spec.Selector{spec.Wildcard},
			sel:  spec.Name("foo"),
			exp:  true,
		},
		{
			name: "name_index",
			list: []spec.Selector{spec.Index(0)},
			sel:  spec.Name("foo"),
		},
		{
			name: "index_empty",
			list: []spec.Selector{},
			sel:  spec.Index(1),
		},
		{
			name:  "index_exists",
			list:  []spec.Selector{spec.Index(1)},
			sel:   spec.Index(1),
			exp:   true,
			exact: true,
		},
		{
			name: "index_wildcard",
			list: []spec.Selector{spec.Wildcard},
			sel:  spec.Index(1),
			exp:  true,
		},
		{
			name: "index_not_exists",
			list: []spec.Selector{spec.Index(2)},
			sel:  spec.Index(1),
		},
		{
			name:  "index_in_list",
			list:  []spec.Selector{spec.Name("foo"), spec.Index(0), spec.Index(1)},
			sel:   spec.Index(1),
			exp:   true,
			exact: true,
		},
		{
			name: "index_not_in_list",
			list: []spec.Selector{spec.Name("foo"), spec.Index(0), spec.Index(1)},
			sel:  spec.Index(2),
		},
		{
			name: "index_in_default_slice",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(2),
			exp:  true,
		},
		{
			name: "index_in_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(2),
			exp:  true,
		},
		{
			name: "index_in_explicit_slice_step",
			list: []spec.Selector{spec.Slice(1, 4, 2)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			name: "index_not_in_explicit_slice_step",
			list: []spec.Selector{spec.Slice(1, 4, 2)},
			sel:  spec.Index(2),
		},
		{
			name: "index_not_in_backwards_slice",
			list: []spec.Selector{spec.Slice(4, 1)},
			sel:  spec.Index(2),
		},
		{
			name: "index_start_of_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(1),
			exp:  true,
		},
		{
			name: "index_end_of_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			name: "index_gt_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(5),
		},
		{
			name: "index_lt_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(0),
		},
		{
			name: "index_not_in_neg_slice",
			list: []spec.Selector{spec.Slice(-4, -1)},
			sel:  spec.Index(2),
		},
		{
			name: "neg_index_in_default",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(-5),
			exp:  true,
		},
		{
			name: "neg_one_in_default",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(-1),
			exp:  true,
		},
		{
			name: "neg_one_in_explicit",
			list: []spec.Selector{spec.Slice(0, 5)},
			sel:  spec.Index(-1),
			exp:  true,
		},
		{
			name: "neg_just_in_explicit",
			list: []spec.Selector{spec.Slice(0, 2)},
			sel:  spec.Index(-2),
			exp:  true,
		},
		{
			name: "neg_not_in_explicit",
			list: []spec.Selector{spec.Slice(0, 2)},
			sel:  spec.Index(-3),
		},
		{
			name: "in_neg_step",
			list: []spec.Selector{spec.Slice(5, 2, -1)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			name: "not_in_neg_two_step",
			list: []spec.Selector{spec.Slice(6, 2, -2)},
			sel:  spec.Index(4),
		},
		{
			name: "not_in_neg_three_step",
			list: []spec.Selector{spec.Slice(6, 1, -3)},
			sel:  spec.Index(2),
		},
		{
			name: "slice_empty",
			list: []spec.Selector{},
			sel:  spec.Slice(),
		},
		{
			name: "slice_step_0",
			list: []spec.Selector{},
			sel:  spec.Slice(1, 3, 0),
			exp:  true,
		},
		{
			name: "slice_start_stop_equal",
			list: []spec.Selector{},
			sel:  spec.Slice(3, 3),
			exp:  true,
		},
		{
			name:  "same_slice",
			list:  []spec.Selector{spec.Slice(1, 3)},
			sel:   spec.Slice(1, 3),
			exp:   true,
			exact: true,
		},
		{
			name: "within_slice_start",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(2, 3),
			exp:  true,
		},
		{
			name: "before_start",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(0, 2),
		},
		{
			name: "after_end",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(1, 4),
		},
		{
			name: "no_selectors",
			list: []spec.Selector{},
			sel:  simpleExists,
		},
		{
			name:  "has_filter",
			list:  []spec.Selector{simpleExists},
			sel:   simpleExists,
			exp:   true,
			exact: true,
		},
		{
			name: "not_has_filter",
			list: []spec.Selector{simpleExists},
			sel:  diffExists,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, selectorsContain(tc.list, tc.sel))
			seg := &Segment{selectors: tc.list}
			a.Equal(tc.exp, seg.hasSelector(tc.sel))
			a.Equal(tc.exact, seg.hasExactSelector(tc.sel))
		})
	}
}

func TestHasSelectors(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// simpleExists := mkFilter("$[?@]")
	// diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		name      string
		seg       *Segment
		selectors []spec.Selector
		exp       bool
		same      bool
		exact     bool
	}{
		{
			name:      "empty",
			seg:       Child(),
			selectors: []spec.Selector{},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			name:      "a_name",
			seg:       Child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			name:      "diff_name",
			seg:       Child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("y")},
		},
		{
			name:      "diff_length",
			seg:       Child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x"), spec.Name("y")},
		},
		{
			name:      "diff_length_has_ok",
			seg:       Child(spec.Name("x"), spec.Name("y")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
		},
		{
			name:      "same_not_exact",
			seg:       Child(spec.Name("x"), spec.Slice(0)),
			selectors: []spec.Selector{spec.Name("x"), spec.Index(0)},
			exp:       true,
			same:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.seg.hasSelectors(tc.selectors))
			a.Equal(tc.same, tc.seg.hasSameSelectors(tc.selectors))
			a.Equal(tc.exact, tc.seg.hasExactSelectors(tc.selectors))
		})
	}
}

func TestContainsIndex(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		slice spec.SliceSelector
		idx   int
		exp   bool
	}{
		{
			name:  "in_slice_end",
			slice: spec.Slice(nil, 6),
			idx:   5,
			exp:   true,
		},
		{
			name:  "not_in_slice_end",
			slice: spec.Slice(nil, 6),
			idx:   6,
			exp:   false,
		},
		{
			name:  "in_bounded",
			slice: spec.Slice(2, 6),
			idx:   5,
			exp:   true,
		},
		{
			name:  "not_in_bounded",
			slice: spec.Slice(2, 6),
			idx:   6,
			exp:   false,
		},
		{
			name:  "in_bounded_step_two",
			slice: spec.Slice(2, 6, 2),
			idx:   4,
			exp:   true,
		},
		{
			name:  "not_in_bounded_step_two",
			slice: spec.Slice(2, 6, 2),
			idx:   3,
			exp:   false,
		},
		{
			name:  "index_not_in_backwards_slice",
			slice: spec.Slice(4, 1),
			idx:   2,
			exp:   false,
		},
		{
			name:  "at_slice_start",
			slice: spec.Slice(1, 4),
			idx:   1,
			exp:   true,
		},
		{
			name:  "at_slice_default_start",
			slice: spec.Slice(nil, 4),
			idx:   0,
			exp:   true,
		},
		{
			name:  "before_slice_start",
			slice: spec.Slice(2, 4),
			idx:   1,
			exp:   false,
		},
		{
			name:  "neg_start",
			slice: spec.Slice(-4, 20),
			idx:   2,
			exp:   false,
		},
		{
			name:  "neg_end",
			slice: spec.Slice(0, -1),
			idx:   2,
			exp:   false,
		},
		{
			name:  "both_neg",
			slice: spec.Slice(-4, -1),
			idx:   0,
			exp:   false,
		},
		{
			name:  "end_lt_start",
			slice: spec.Slice(12, 0),
			idx:   5,
			exp:   false,
		},
		{
			name:  "in_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   3,
			exp:   true,
		},
		{
			name:  "not_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   1,
			exp:   false,
		},
		{
			name:  "exclude_end_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   2,
			exp:   false,
		},
		{
			name:  "in_neg_one_step_start",
			slice: spec.Slice(5, 2, -1),
			idx:   5,
			exp:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, containsIndex([]spec.Selector{tc.slice}, spec.Index(tc.idx)))

			// Test with actual data.
			size := 1 + max(abs(tc.idx), abs(tc.slice.Start()), abs(tc.slice.End()))
			input := make([]any, size)
			switch {
			case tc.idx >= 0:
				input[tc.idx] = true
			case tc.idx < 0:
				input[len(input)+tc.idx] = true
			}
			res := tc.slice.Select(input, nil)
			a.Equal(tc.exp, slices.Contains(res, true))
		})
	}
}

func TestContainsFilter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		name   string
		list   []spec.Selector
		filter *spec.FilterSelector
		exp    bool
	}{
		{
			name:   "no_selectors",
			list:   []spec.Selector{},
			filter: simpleExists,
			exp:    false,
		},
		{
			name:   "has_filter",
			list:   []spec.Selector{simpleExists},
			filter: simpleExists,
			exp:    true,
		},
		{
			name:   "not_has_filter",
			list:   []spec.Selector{simpleExists},
			filter: diffExists,
			exp:    false,
		},
		{
			name:   "same_operands",
			list:   []spec.Selector{mkFilter("$[?@.x > @.y]")},
			filter: mkFilter("$[?@.x > @.y]"),
			exp:    true,
		},
		{
			name:   "diff_cmp_operands",
			list:   []spec.Selector{mkFilter("$[?@.a > @.y]")},
			filter: mkFilter("$[?@.x > @.y]"),
			exp:    false,
		},
		{
			name:   "diff_and_operands",
			list:   []spec.Selector{mkFilter("$[?@.a && @.y]")},
			filter: mkFilter("$[?@.x && @.y]"),
			exp:    false,
		},
		{
			name:   "diff_or_operands",
			list:   []spec.Selector{mkFilter("$[?@.a || @.y]")},
			filter: mkFilter("$[?@.x || @.y]"),
			exp:    false,
		},
		// These examples could be equivalent if spec.ComparisonExpr
		// normalized its expressions for deterministic order.
		{
			name:   "reversed_operands",
			list:   []spec.Selector{mkFilter("$[?@.x > @.y]")},
			filter: mkFilter("$[?@.y > @.x]"),
			exp:    false,
		},
		{
			name:   "reversed_eq_operands",
			list:   []spec.Selector{mkFilter("$[?@.x == @.y]")},
			filter: mkFilter("$[?@.y == @.x]"),
			exp:    false,
		},
		{
			name:   "reversed_ne_operands",
			list:   []spec.Selector{mkFilter("$[?@.x != @.y]")},
			filter: mkFilter("$[?@.y != @.x]"),
			exp:    false,
		},
		// These examples could be equal if spec.LogicalOr and spec.LogicalAnd
		// normalized adopted https://pkg.go.dev/sort#Interface.
		{
			name:   "reversed_or_operands",
			list:   []spec.Selector{mkFilter("$[?@.x || @.y]")},
			filter: mkFilter("$[?@.y || @.x]"),
			exp:    false,
		},
		{
			name:   "reversed_and_operands",
			list:   []spec.Selector{mkFilter("$[?@.x && @.y]")},
			filter: mkFilter("$[?@.y && @.x]"),
			exp:    false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, containsFilter(tc.list, tc.filter))
		})
	}
}

func mkFilter(f string) *spec.FilterSelector {
	p := jsonpath.MustParse(f)
	filter := p.Query().Segments()[0].Selectors()[0]
	if f, ok := filter.(*spec.FilterSelector); ok {
		return f
	}
	panic(fmt.Sprintf("Expected *spec.FilterSelector, got %T", filter))
}

func TestContainsSlice(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		list  []spec.Selector
		slice spec.SliceSelector
		exp   bool
	}{
		{
			name:  "no_selectors",
			list:  []spec.Selector{},
			slice: spec.Slice(),
			exp:   false,
		},
		{
			name:  "step_0",
			list:  []spec.Selector{},
			slice: spec.Slice(1, 3, 0),
			exp:   true,
		},
		{
			name:  "start_stop_equal",
			list:  []spec.Selector{},
			slice: spec.Slice(3, 3),
			exp:   true,
		},
		{
			name:  "same_slice",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 3),
			exp:   true,
		},
		{
			name:  "within_start",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(2, 3),
			exp:   true,
		},
		{
			name:  "within_end",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 2),
			exp:   true,
		},
		{
			name:  "before_start",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(0, 2),
			exp:   false,
		},
		{
			name:  "after_end",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 4),
			exp:   false,
		},
		{
			name:  "multiple_of_step",
			list:  []spec.Selector{spec.Slice(1, 3, 2)},
			slice: spec.Slice(1, 3, 4),
			exp:   true,
		},
		{
			name:  "out_of_step",
			list:  []spec.Selector{spec.Slice(1, 3, 2)},
			slice: spec.Slice(1, 3, 3),
			exp:   false,
		},
		{
			name:  "over_step",
			list:  []spec.Selector{spec.Slice(1, 3, 4)},
			slice: spec.Slice(1, 3, 2),
			exp:   false,
		},

		{
			name:  "same_backward_slice",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(3, 1, -1),
			exp:   true,
		},
		{
			name:  "within_start_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(3, 2, -1),
			exp:   true,
		},
		{
			name:  "within_end_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(2, 1, -1),
			exp:   true,
		},
		{
			name:  "after_end_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(2, 0, -1),
			exp:   false,
		},
		{
			name:  "before_start_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(4, 1, -1),
			exp:   false,
		},
		{
			name:  "multiple_of_backward_step",
			list:  []spec.Selector{spec.Slice(3, 1, -2)},
			slice: spec.Slice(3, 1, -4),
			exp:   true,
		},
		{
			name:  "out_of_step_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -2)},
			slice: spec.Slice(3, 1, -3),
			exp:   false,
		},
		{
			name:  "over_step_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -4)},
			slice: spec.Slice(3, 1, -2),
			exp:   false,
		},
		{
			name: "opposite_step_not_in_range",
			// $[1:3:1] != $[3:1:-1]
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(1, 3, 1),
			exp:   false,
		},
		{
			// $[2:0:-1] == $[1:3:1]
			name:  "opposites",
			list:  []spec.Selector{spec.Slice(1, 3, 1)},
			slice: spec.Slice(2, 0, -1),
			exp:   true,
		},
		{
			// $[1:3:1] == $[2:0:-1]
			name:  "inverted_opposites",
			list:  []spec.Selector{spec.Slice(2, 0, -1)},
			slice: spec.Slice(1, 3, 1),
			exp:   true,
		},
		{
			// $[2:0:-2] mod within $[1:3:1]
			name:  "opposite_mod_step",
			list:  []spec.Selector{spec.Slice(1, 3, 1)},
			slice: spec.Slice(2, 0, -2),
			exp:   true,
		},
		{
			// $[1:3:1] not mod within $[2:0:-2]
			name:  "opposite_not_mod_step",
			list:  []spec.Selector{spec.Slice(2, 0, -2)},
			slice: spec.Slice(1, 3, 1),
			exp:   false,
		},
		{
			// $[2:0:-1] within $[1:5:1]
			name:  "within_opposite",
			list:  []spec.Selector{spec.Slice(1, 5, 1)},
			slice: spec.Slice(2, 0, -1),
			exp:   true,
		},
		{
			name:  "equals_index",
			list:  []spec.Selector{spec.Index(3)},
			slice: spec.Slice(3, 4),
			exp:   true,
		},
		{
			name:  "equals_index_inverted",
			list:  []spec.Selector{spec.Index(3)},
			slice: spec.Slice(4, 3, -1),
			exp:   true,
		},
		{
			name:  "not_equals_index",
			list:  []spec.Selector{spec.Index(4)},
			slice: spec.Slice(3, 4),
			exp:   false,
		},
		{
			name:  "not_equals_index_inverted",
			list:  []spec.Selector{spec.Index(4)},
			slice: spec.Slice(4, 3, -1),
			exp:   false,
		},
		{
			// XXX Compare all indexes to slice range?
			name:  "equals_all_indexes",
			list:  []spec.Selector{spec.Index(3), spec.Index(4)},
			slice: spec.Slice(3, 5),
			exp:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, containsSlice(tc.list, tc.slice))
		})
	}
}

func TestIsBranch(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		seg    *Segment
		branch []*spec.Segment
		exp    bool
	}{
		{
			name:   "empty",
			seg:    &Segment{},
			branch: []*spec.Segment{},
			exp:    true,
		},
		{
			name:   "empty_is_not_name",
			seg:    &Segment{},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			name:   "eq_name",
			seg:    &Segment{children: []*Segment{Child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    true,
		},
		{
			name:   "ne_name",
			seg:    &Segment{children: []*Segment{Child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("y"))},
			exp:    false,
		},
		{
			name: "size",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")),
				Child(spec.Name("y")),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Name("y")),
			},
			exp: false,
		},
		{
			name: "eq_branch_mixed",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Index(0), spec.Slice(1)).Append(
						Child(spec.Wildcard).Append(
							Child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1)),
				spec.Child(spec.Wildcard),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: true,
		},
		{
			name: "ne_child_selectors",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Index(0), spec.Slice(1), spec.Name("x")).Append(
						Child(spec.Wildcard).Append(
							Child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1)),
				spec.Child(spec.Wildcard),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: false,
		},
		{
			name: "ne_spec_selectors",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Index(0), spec.Slice(1)).Append(
						Child(spec.Wildcard).Append(
							Child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1), spec.Name("x")),
				spec.Child(spec.Wildcard),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: false,
		},
		{
			name: "diff_child_length",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("y")).Append(
					Child(spec.Name("z")),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("y")),
			},
			exp: false,
		},
		{
			name: "diff_spec_length",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("y")),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("y")),
				spec.Child(spec.Name("z")),
			},
			exp: false,
		},
		{
			name:   "ne_spec_desc",
			seg:    &Segment{children: []*Segment{Child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Descendant(spec.Name("x"))},
			exp:    false,
		},
		{
			name:   "ne_child_desc",
			seg:    &Segment{children: []*Segment{Descendant(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			name: "ne_sub_spec_desc",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Name("y")),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Descendant(spec.Name("y")),
			},
			exp: false,
		},
		{
			name: "ne_sub_child_desc",
			seg: &Segment{children: []*Segment{
				Child(spec.Name("x")).Append(
					Descendant(spec.Name("y")),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Name("y")),
			},
			exp: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.seg.isBranch(tc.branch))
		})
	}
}

func TestMergeSelectors(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name      string
		selectors []spec.Selector
		merge     []spec.Selector
		exp       []spec.Selector
	}{
		{
			name: "none",
		},
		{
			name:  "name_into_empty",
			merge: []spec.Selector{spec.Name("x")},
			exp:   []spec.Selector{spec.Name("x")},
		},
		{
			name:      "name_into_existing",
			selectors: []spec.Selector{spec.Name("x")},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Name("x")},
		},
		{
			name:      "name_into_index",
			selectors: []spec.Selector{spec.Index(0)},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Index(0), spec.Name("x")},
		},
		{
			name:      "name_into_index_no_dupe",
			selectors: []spec.Selector{spec.Index(0), spec.Name("x")},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Index(0), spec.Name("x")},
		},
		{
			name:      "index_slice",
			selectors: []spec.Selector{spec.Slice(2), spec.Name("x")},
			merge:     []spec.Selector{spec.Index(2), spec.Name("y")},
			exp:       []spec.Selector{spec.Slice(2), spec.Name("x"), spec.Name("y")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seg := &Segment{selectors: tc.selectors}
			seg.mergeSelectors(tc.merge)
			a.Equal(tc.exp, seg.selectors)
		})
	}
}

func TestMergeChildren(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name     string
		children []*Segment
		expect   []*Segment
	}{
		{
			name:     "empty",
			children: []*Segment{},
			expect:   []*Segment{},
		},
		{
			name:     "one_child",
			children: []*Segment{Child(spec.Name("x"))},
			expect:   []*Segment{Child(spec.Name("x"))},
		},
		{
			name:     "merge_name",
			children: []*Segment{Child(spec.Name("x")), Child(spec.Name("x"))},
			expect:   []*Segment{Child(spec.Name("x"))},
		},
		{
			name:     "merge_selectors",
			children: []*Segment{Child(spec.Name("x")), Child(spec.Name("y"))},
			expect:   []*Segment{Child(spec.Name("x"), spec.Name("y"))},
		},
		{
			name: "no_merge_diff_child",
			children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Name("y")),
				),
				Child(spec.Name("a")).Append(
					Child(spec.Name("z")),
				),
			},
			expect: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Name("y")),
				),
				Child(spec.Name("a")).Append(
					Child(spec.Name("z")),
				),
			},
		},
		{
			name: "merge_same_child",
			children: []*Segment{
				Child(spec.Name("x")).Append(
					Child(spec.Name("y")),
				),
				Child(spec.Name("a")).Append(
					Child(spec.Name("y")),
				),
			},
			expect: []*Segment{
				Child(spec.Name("x"), spec.Name("a")).Append(
					Child(spec.Name("y")),
				),
			},
		},
		{
			name: "merge_same_nested_multi_select",
			children: []*Segment{
				Child(spec.Name("x"), spec.Name("y")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
					),
				),
				Child(spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
					),
				),
			},
			expect: []*Segment{
				Child(spec.Name("x"), spec.Name("y"), spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
					),
				),
			},
		},
		{
			name: "no_merge_diff_depth",
			children: []*Segment{
				Child(spec.Name("x"), spec.Name("y")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")).Append(
							Child(spec.Name("c")),
						),
					),
				),
				Child(spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
					),
				),
			},
			expect: []*Segment{
				Child(spec.Name("x"), spec.Name("y")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")).Append(
							Child(spec.Name("c")),
						),
					),
				),
				Child(spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
					),
				),
			},
		},
		{
			name: "merge_nested_selectors",
			children: []*Segment{
				Child(spec.Name("x"), spec.Name("y")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
						Child(spec.Name("c")),
					),
				),
				Child(spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b")),
						Child(spec.Name("c")),
					),
				),
			},
			expect: []*Segment{
				Child(spec.Name("x"), spec.Name("y"), spec.Name("z")).Append(
					Child(spec.Name("a")).Append(
						Child(spec.Name("b"), spec.Name("c")),
					),
				),
			},
		},
		{
			name:     "merge_descendants",
			children: []*Segment{Descendant(spec.Name("x")), Descendant(spec.Name("x"))},
			expect:   []*Segment{Descendant(spec.Name("x"))},
		},
		{
			name:     "merge_descendant_child",
			children: []*Segment{Descendant(spec.Name("x")), Child(spec.Name("x"))},
			expect:   []*Segment{Descendant(spec.Name("x"))},
		},
		{
			name:     "merge_descendant_sub_child",
			children: []*Segment{Descendant(spec.Name("x"), spec.Name("y")), Child(spec.Name("x"))},
			expect:   []*Segment{Descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			name:     "merge_child_descendant",
			children: []*Segment{Child(spec.Name("x")), Descendant(spec.Name("x"))},
			expect:   []*Segment{Descendant(spec.Name("x"))},
		},
		{
			name:     "merge_child_sub_descendant",
			children: []*Segment{Child(spec.Name("x")), Descendant(spec.Name("x"), spec.Name("y"))},
			expect:   []*Segment{Descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			name:     "merge_only_common_descendant_child_selector",
			children: []*Segment{Descendant(spec.Name("x")), Child(spec.Name("x"), spec.Name("y"))},
			expect:   []*Segment{Descendant(spec.Name("x")), Child(spec.Name("y"))},
		},
		{
			name:     "merge_only_common_prev_descendant_selector",
			children: []*Segment{Child(spec.Name("x"), spec.Name("y")), Descendant(spec.Name("x"))},
			expect:   []*Segment{Child(spec.Name("y")), Descendant(spec.Name("x"))},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seg := &Segment{children: tc.children}
			seg.deduplicate()
			a.Equal(tc.expect, seg.children)
		})
	}
}

func TestRemoveCommonSelectorsFrom(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		sel1 []spec.Selector
		sel2 []spec.Selector
		exp2 []spec.Selector
		res  bool
	}{
		{
			name: "empty",
		},
		{
			name: "empty_seg2",
			sel1: []spec.Selector{spec.Name("x")},
		},
		{
			name: "no_commonality",
			sel1: []spec.Selector{spec.Name("x")},
			sel2: []spec.Selector{spec.Name("y")},
			exp2: []spec.Selector{spec.Name("y")},
		},
		{
			name: "remove_one_all",
			sel1: []spec.Selector{spec.Name("x")},
			sel2: []spec.Selector{spec.Name("x")},
			exp2: []spec.Selector{},
			res:  true,
		},
		{
			name: "remove_one",
			sel1: []spec.Selector{spec.Name("x"), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x")},
			exp2: []spec.Selector{},
			res:  true,
		},
		{
			name: "remove_one_leave_one",
			sel1: []spec.Selector{spec.Name("x"), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Name("a")},
			exp2: []spec.Selector{spec.Name("a")},
			res:  false,
		},
		{
			name: "remove_sub_slice",
			sel1: []spec.Selector{spec.Slice(1, 3), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Slice(1, 2)},
			exp2: []spec.Selector{spec.Name("x")},
			res:  false,
		},
		{
			name: "remove_index_matching_slice",
			sel1: []spec.Selector{spec.Slice(1, 3), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Index(2)},
			exp2: []spec.Selector{spec.Name("x")},
			res:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seg1 := &Segment{selectors: tc.sel1}
			seg2 := &Segment{selectors: tc.sel2}
			a.Equal(tc.res, seg1.removeCommonSelectorsFrom(seg2))
			a.Equal(tc.exp2, seg2.selectors, "selectors 2")
		})
	}
}

func TestSameBranches(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		name string
		seg1 *Segment
		seg2 *Segment
		exp  bool
	}{
		{
			name: "empties",
			seg1: Child(),
			seg2: Child(),
			exp:  true,
		},
		{
			name: "single_child_eq_name",
			seg1: Child().Append(Child(spec.Name("x"))),
			seg2: Child().Append(Child(spec.Name("x"))),
			exp:  true,
		},
		{
			name: "single_child_ne_name",
			seg1: Child().Append(Child(spec.Name("x"))),
			seg2: Child().Append(Child(spec.Name("y"))),
			exp:  false,
		},
		{
			name: "single_child_eq_multi_select",
			seg1: Child().Append(Child(spec.Name("x"), spec.Index(0))),
			seg2: Child().Append(Child(spec.Name("x"), spec.Index(0))),
			exp:  true,
		},
		{
			name: "single_child_ne_multi_select",
			seg1: Child().Append(Child(spec.Name("x"), spec.Index(0))),
			seg2: Child().Append(Child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			name: "single_child_ne_index_slice",
			seg1: Child().Append(Child(spec.Name("x"), spec.Slice(0))),
			seg2: Child().Append(Child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			name: "single_child_ne_slices",
			seg1: Child().Append(Child(spec.Name("x"), spec.Slice(0))),
			seg2: Child().Append(Child(spec.Name("x"), spec.Slice(1))),
			exp:  false,
		},
		{
			name: "single_child_eq_filter",
			seg1: Child().Append(Child(simpleExists)),
			seg2: Child().Append(Child(simpleExists)),
			exp:  true,
		},
		{
			name: "single_child_ne_filter",
			seg1: Child().Append(Child(simpleExists)),
			seg2: Child().Append(Child(diffExists)),
			exp:  false,
		},
		{
			name: "wildcards",
			seg1: Child().Append(Child(spec.Wildcard)),
			seg2: Child().Append(Child(spec.Wildcard)),
			exp:  true,
		},
		{
			name: "wildcard_with_eq_name_is_ne",
			seg1: Child().Append(Child(spec.Wildcard, spec.Name("x"))),
			seg2: Child().Append(Child(spec.Wildcard, spec.Name("x"))),
			exp:  false,
		},
		{
			name: "diff_branches_child_count",
			seg1: Child().Append(Child(spec.Name("x"))),
			seg2: Child().Append(Child(spec.Name("x")), Child(spec.Name("x"))),
			exp:  false,
		},
		{
			name: "diff_children_child_count",
			seg1: Child().Append(Child(spec.Name("x")), Child(spec.Name("x"))),
			seg2: Child().Append(Child(spec.Name("x"))),
			exp:  false,
		},
		{
			name: "same_children",
			seg1: Child().Append(Child(spec.Name("x")), Child(spec.Name("x"))),
			seg2: Child().Append(Child(spec.Name("x")), Child(spec.Name("x"))),
			exp:  true,
		},
		{
			name: "same_nested_children",
			seg1: Child().Append(
				Child(spec.Name("x")).Append(
					Child(spec.Index(0)).Append(
						Descendant(spec.Slice(4)),
					),
				),
				Child(spec.Name("y")),
			),
			seg2: Child().Append(
				Child(spec.Name("x")).Append(
					Child(spec.Index(0)).Append(
						Descendant(spec.Slice(4)),
					),
				),
				Child(spec.Name("y")),
			),
			exp: true,
		},
		{
			name: "diff_nested_children",
			seg1: Child().Append(
				Child(spec.Name("x")).Append(
					Child(spec.Index(0)).Append(
						Descendant(spec.Slice(4)),
					),
				),
				Child(spec.Name("y")),
			),
			seg2: Child().Append(
				Child(spec.Name("x")).Append(
					Child(spec.Index(0)).Append(
						Descendant(spec.Slice(3)),
					),
				),
				Child(spec.Name("y")),
			),
			exp: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.seg1.sameBranches(tc.seg2))
		})
	}
}

func TestMergeSlices(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *Segment
		exp  *Segment
	}{
		{
			name: "Empty",
			seg:  Child(),
			exp:  Child(),
		},
		{
			name: "no_slices",
			seg:  Child(spec.Name("x"), spec.Index(0)),
			exp:  Child(spec.Name("x"), spec.Index(0)),
		},
		{
			name: "one_slice",
			seg:  Child(spec.Name("x"), spec.Slice(12, 18)),
			exp:  Child(spec.Name("x"), spec.Slice(12, 18)),
		},
		{
			name: "sub_slice",
			seg:  Child(spec.Name("x"), spec.Slice(10), spec.Slice(12, 18)),
			exp:  Child(spec.Name("x"), spec.Slice(10)),
		},
		{
			name: "slice_sub",
			seg:  Child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(10)),
			exp:  Child(spec.Name("x"), spec.Slice(10)),
		},
		{
			name: "sub_slice",
			seg:  Child(spec.Name("x"), spec.Slice(8), spec.Slice(10), spec.Slice(12, 18)),
			exp:  Child(spec.Name("x"), spec.Slice(8)),
		},
		{
			name: "multi_slice_sub",
			seg:  Child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(10)),
			exp:  Child(spec.Name("x"), spec.Slice(8)),
		},
		{
			name: "multi_overlaps",
			seg:  Child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(2, 5), spec.Slice(4, 5)),
			exp:  Child(spec.Name("x"), spec.Slice(8), spec.Slice(2, 5)),
		},
		{
			name: "multi_overlap_reverse",
			seg:  Child(spec.Name("x"), spec.Slice(4, 5), spec.Slice(2, 5), spec.Slice(8), spec.Slice(12, 18)),
			exp:  Child(spec.Name("x"), spec.Slice(2, 5), spec.Slice(8)),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.seg.mergeSlices()
			a.Equal(tc.exp.selectors, tc.seg.selectors)
		})
	}
}
