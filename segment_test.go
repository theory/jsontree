package jsontree

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

func TestWriteSeg(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *segment
		str  string
	}{
		{
			name: "empty_child",
			seg:  child(),
			str:  "[]\n",
		},
		{
			name: "empty_descendant",
			seg:  descendant(),
			str:  "..[]\n",
		},
		{
			name: "wildcard",
			seg:  child(spec.Wildcard),
			str:  "[*]\n",
		},
		{
			name: "one_key",
			seg:  child(spec.Name("foo")),
			str:  "[\"foo\"]\n",
		},
		{
			name: "two_keys",
			seg:  child(spec.Name("foo"), spec.Name("bar")),
			str:  "[\"foo\",\"bar\"]\n",
		},
		{
			name: "parent_child",
			seg: child(spec.Name("foo")).Append(
				child(spec.Name("bar")),
			),
			str: "[\"foo\"]\n└── [\"bar\"]\n",
		},
		{
			name: "parent_two_child",
			seg: child(spec.Name("foo")).Append(
				child(spec.Name("bar")),
				child(spec.Name("baz")),
			),
			str: "[\"foo\"]\n├── [\"bar\"]\n└── [\"baz\"]\n",
		},
		{
			name: "two_children_and_sub_keys",
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("foo")).Append(
					child(spec.Name("x")),
					child(spec.Name("y")),
					descendant(spec.Name("z")),
				),
				child(spec.Name("bar")).Append(
					child(spec.Name("a"), spec.Index(42), spec.Slice(0, 8, 2)),
					child(spec.Name("b")),
					child(spec.Name("c")),
				),
			),
			str: `["hi"]
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
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("foo")).Append(
					child(spec.Name("x")),
					child(spec.Name("y")).Append(
						child(spec.Wildcard).Append(
							child(spec.Name("a")),
							child(spec.Name("b")),
						),
					),
				),
				child(spec.Name("bar")).Append(
					child(spec.Name("go")),
					child(spec.Name("z")).Append(
						child(spec.Wildcard).Append(
							child(spec.Name("c")),
							child(spec.Name("d")).Append(
								child(spec.Slice(2, 3)),
							),
						),
					),
					child(spec.Name("hi")),
				),
			),
			str: `["hi"]
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
			seg:  child(spec.Wildcard),
			str:  "[*]\n",
		},
		{
			name: "one_index",
			seg:  child(spec.Index(0)),
			str:  "[0]\n",
		},
		{
			name: "two_indexes",
			seg:  child(spec.Index(0), spec.Index(2)),
			str:  "[0,2]\n",
		},
		{
			name: "index_index",
			seg:  child(spec.Index(0)).Append(child(spec.Index(2))),
			str:  "[0]\n└── [2]\n",
		},
		{
			name: "two_nested_indexes",
			seg: child(spec.Index(1)).Append(
				child(spec.Index(0)), child(spec.Index(2)),
			),
			str: "[1]\n├── [0]\n└── [2]\n",
		},
		{
			name: "two_keys_and_sub_indexes",
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("foo")).Append(
					child(spec.Index(0)),
					child(spec.Index(1)),
					child(spec.Index(2)),
				),
				child(spec.Name("bar")).Append(
					child(spec.Index(3)),
					child(spec.Index(4)),
					child(spec.Index(5)),
				),
			),
			str: `["hi"]
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
			seg: child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Paren(spec.LogicalOr{spec.LogicalAnd{
					spec.Existence(spec.Query(true, []*spec.Segment{})),
				}}),
			}})),
			str: "[?($)]\n",
		},
		{
			name: "filter_and_key",
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("x")),
				child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
					spec.Paren(spec.LogicalOr{spec.LogicalAnd{
						spec.Existence(spec.Query(true, []*spec.Segment{})),
					}}),
				}})),
			),
			str: "[\"hi\"]\n├── [\"x\"]\n└── [?($)]\n",
		},
		{
			name: "filter_and_key_segment",
			seg: child(
				spec.Name("x"),
				spec.Filter(spec.LogicalOr{spec.LogicalAnd{
					spec.Paren(spec.LogicalOr{spec.LogicalAnd{
						spec.Existence(spec.Query(true, []*spec.Segment{})),
					}}),
				}}),
			),
			str: "[\"x\",?($)]\n",
		},
		{
			name: "nested_filter",
			seg: child(spec.Name("x")).Append(
				child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
					spec.Paren(spec.LogicalOr{spec.LogicalAnd{
						spec.Existence(spec.Query(true, []*spec.Segment{})),
					}}),
				}})),
			),
			str: "[\"x\"]\n└── [?($)]\n",
		},
		{
			name: "slice_def_start_stop_neg_step",
			seg: child(spec.Name("x")).Append(
				child(spec.Slice(nil, nil, -1)),
			),
			str: "[\"x\"]\n└── [::-1]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.str, tc.seg.String(), tc.name)
		})
	}
}

func TestIsWildcard(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *segment
		exp  bool
	}{
		{"empty", &segment{}, false},
		{"name", &segment{selectors: []spec.Selector{spec.Name("x")}}, false},
		{"index", &segment{selectors: []spec.Selector{spec.Index(0)}}, false},
		{"slice", &segment{selectors: []spec.Selector{spec.Slice()}}, false},
		{"filter", &segment{selectors: []spec.Selector{mkFilter("$[?@]")}}, false},
		{"wildcard", &segment{selectors: []spec.Selector{spec.Wildcard}}, true},
		{"multiples", &segment{selectors: []spec.Selector{spec.Wildcard, spec.Index(0)}}, false},
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
			seg := &segment{selectors: tc.list}
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
		seg       *segment
		selectors []spec.Selector
		exp       bool
		same      bool
		exact     bool
	}{
		{
			name:      "empty",
			seg:       child(),
			selectors: []spec.Selector{},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			name:      "a_name",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			name:      "diff_name",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("y")},
		},
		{
			name:      "diff_length",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x"), spec.Name("y")},
		},
		{
			name:      "diff_length_has_ok",
			seg:       child(spec.Name("x"), spec.Name("y")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
		},
		{
			name:      "same_not_exact",
			seg:       child(spec.Name("x"), spec.Slice(0)),
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
		{
			name:  "defaults_neg_slice_covers_all",
			list:  []spec.Selector{spec.Slice(nil, nil, -1)},
			slice: spec.Slice(0, 2),
			exp:   true,
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
		seg    *segment
		branch []*spec.Segment
		exp    bool
	}{
		{
			name:   "empty",
			seg:    &segment{},
			branch: []*spec.Segment{},
			exp:    true,
		},
		{
			name:   "empty_is_not_name",
			seg:    &segment{},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			name:   "eq_name",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    true,
		},
		{
			name:   "ne_name",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("y"))},
			exp:    false,
		},
		{
			name: "size",
			seg: &segment{children: []*segment{
				child(spec.Name("x")),
				child(spec.Name("y")),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Name("y")),
			},
			exp: false,
		},
		{
			name: "eq_branch_mixed",
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1)).Append(
						child(spec.Wildcard).Append(
							child(mkFilter("$[?@]")),
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
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1), spec.Name("x")).Append(
						child(spec.Wildcard).Append(
							child(mkFilter("$[?@]")),
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
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1)).Append(
						child(spec.Wildcard).Append(
							child(mkFilter("$[?@]")),
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
			seg: &segment{children: []*segment{
				child(spec.Name("y")).Append(
					child(spec.Name("z")),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("y")),
			},
			exp: false,
		},
		{
			name: "diff_spec_length",
			seg: &segment{children: []*segment{
				child(spec.Name("y")),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("y")),
				spec.Child(spec.Name("z")),
			},
			exp: false,
		},
		{
			name:   "ne_spec_desc",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Descendant(spec.Name("x"))},
			exp:    false,
		},
		{
			name:   "ne_child_desc",
			seg:    &segment{children: []*segment{descendant(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			name: "ne_sub_spec_desc",
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("y")),
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
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					descendant(spec.Name("y")),
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
		{
			name:      "neg_slice_index",
			selectors: []spec.Selector{spec.Slice(nil, nil, -1)},
			merge:     []spec.Selector{spec.Index(0), spec.Index(2)},
			exp:       []spec.Selector{spec.Slice(nil, nil, -1)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seg := &segment{selectors: tc.selectors}
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
		children []*segment
		expect   []*segment
	}{
		{
			name:     "empty",
			children: []*segment{},
			expect:   []*segment{},
		},
		{
			name:     "one_child",
			children: []*segment{child(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("x"))},
		},
		{
			name:     "merge_name",
			children: []*segment{child(spec.Name("x")), child(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("x"))},
		},
		{
			name:     "merge_selectors",
			children: []*segment{child(spec.Name("x")), child(spec.Name("y"))},
			expect:   []*segment{child(spec.Name("x"), spec.Name("y"))},
		},
		{
			name: "no_merge_diff_child",
			children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("y")),
				),
				child(spec.Name("a")).Append(
					child(spec.Name("z")),
				),
			},
			expect: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("y")),
				),
				child(spec.Name("a")).Append(
					child(spec.Name("z")),
				),
			},
		},
		{
			name: "merge_same_child",
			children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("y")),
				),
				child(spec.Name("a")).Append(
					child(spec.Name("y")),
				),
			},
			expect: []*segment{
				child(spec.Name("x"), spec.Name("a")).Append(
					child(spec.Name("y")),
				),
			},
		},
		{
			name: "merge_same_nested_multi_select",
			children: []*segment{
				child(spec.Name("x"), spec.Name("y")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
				child(spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
			expect: []*segment{
				child(spec.Name("x"), spec.Name("y"), spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
		},
		{
			name: "no_merge_diff_depth",
			children: []*segment{
				child(spec.Name("x"), spec.Name("y")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
				child(spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
			expect: []*segment{
				child(spec.Name("x"), spec.Name("y")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
				child(spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
		},
		{
			name: "merge_nested_selectors",
			children: []*segment{
				child(spec.Name("x"), spec.Name("y")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
						child(spec.Name("c")),
					),
				),
				child(spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
						child(spec.Name("c")),
					),
				),
			},
			expect: []*segment{
				child(spec.Name("x"), spec.Name("y"), spec.Name("z")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
				),
			},
		},
		{
			name:     "merge_descendants",
			children: []*segment{descendant(spec.Name("x")), descendant(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			name:     "merge_descendant_child",
			children: []*segment{descendant(spec.Name("x")), child(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			name:     "merge_descendant_sub_child",
			children: []*segment{descendant(spec.Name("x"), spec.Name("y")), child(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			name:     "merge_child_descendant",
			children: []*segment{child(spec.Name("x")), descendant(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			name:     "merge_child_sub_descendant",
			children: []*segment{child(spec.Name("x")), descendant(spec.Name("x"), spec.Name("y"))},
			expect:   []*segment{descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			name:     "merge_only_common_descendant_child_selector",
			children: []*segment{descendant(spec.Name("x")), child(spec.Name("x"), spec.Name("y"))},
			expect:   []*segment{descendant(spec.Name("x")), child(spec.Name("y"))},
		},
		{
			name:     "merge_only_common_prev_descendant_selector",
			children: []*segment{child(spec.Name("x"), spec.Name("y")), descendant(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("y")), descendant(spec.Name("x"))},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			seg := &segment{children: tc.children}
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
			res:  true,
		},
		{
			name: "empty_seg2",
			sel1: []spec.Selector{spec.Name("x")},
			res:  true,
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
			seg1 := &segment{selectors: tc.sel1}
			seg2 := &segment{selectors: tc.sel2}
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
		seg1 *segment
		seg2 *segment
		exp  bool
	}{
		{
			name: "empties",
			seg1: child(),
			seg2: child(),
			exp:  true,
		},
		{
			name: "single_child_eq_name",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x"))),
			exp:  true,
		},
		{
			name: "single_child_ne_name",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("y"))),
			exp:  false,
		},
		{
			name: "single_child_eq_multi_select",
			seg1: child().Append(child(spec.Name("x"), spec.Index(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(0))),
			exp:  true,
		},
		{
			name: "single_child_ne_multi_select",
			seg1: child().Append(child(spec.Name("x"), spec.Index(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			name: "single_child_ne_index_slice",
			seg1: child().Append(child(spec.Name("x"), spec.Slice(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			name: "single_child_ne_slices",
			seg1: child().Append(child(spec.Name("x"), spec.Slice(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Slice(1))),
			exp:  false,
		},
		{
			name: "single_child_eq_filter",
			seg1: child().Append(child(simpleExists)),
			seg2: child().Append(child(simpleExists)),
			exp:  true,
		},
		{
			name: "single_child_ne_filter",
			seg1: child().Append(child(simpleExists)),
			seg2: child().Append(child(diffExists)),
			exp:  false,
		},
		{
			name: "wildcards",
			seg1: child().Append(child(spec.Wildcard)),
			seg2: child().Append(child(spec.Wildcard)),
			exp:  true,
		},
		{
			name: "wildcard_with_eq_name_is_ne",
			seg1: child().Append(child(spec.Wildcard, spec.Name("x"))),
			seg2: child().Append(child(spec.Wildcard, spec.Name("x"))),
			exp:  false,
		},
		{
			name: "diff_branches_child_count",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			exp:  false,
		},
		{
			name: "diff_children_child_count",
			seg1: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x"))),
			exp:  false,
		},
		{
			name: "same_children",
			seg1: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			exp:  true,
		},
		{
			name: "same_nested_children",
			seg1: child().Append(
				child(spec.Name("x")).Append(
					child(spec.Index(0)).Append(
						descendant(spec.Slice(4)),
					),
				),
				child(spec.Name("y")),
			),
			seg2: child().Append(
				child(spec.Name("x")).Append(
					child(spec.Index(0)).Append(
						descendant(spec.Slice(4)),
					),
				),
				child(spec.Name("y")),
			),
			exp: true,
		},
		{
			name: "diff_nested_children",
			seg1: child().Append(
				child(spec.Name("x")).Append(
					child(spec.Index(0)).Append(
						descendant(spec.Slice(4)),
					),
				),
				child(spec.Name("y")),
			),
			seg2: child().Append(
				child(spec.Name("x")).Append(
					child(spec.Index(0)).Append(
						descendant(spec.Slice(3)),
					),
				),
				child(spec.Name("y")),
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
		seg  *segment
		exp  *segment
	}{
		{
			name: "Empty",
			seg:  child(),
			exp:  child(),
		},
		{
			name: "no_slices",
			seg:  child(spec.Name("x"), spec.Index(0)),
			exp:  child(spec.Name("x"), spec.Index(0)),
		},
		{
			name: "one_slice",
			seg:  child(spec.Name("x"), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(12, 18)),
		},
		{
			name: "sub_slice",
			seg:  child(spec.Name("x"), spec.Slice(10), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(10)),
		},
		{
			name: "slice_sub",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(10)),
			exp:  child(spec.Name("x"), spec.Slice(10)),
		},
		{
			name: "sub_slice_slice",
			seg:  child(spec.Name("x"), spec.Slice(8), spec.Slice(10), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(8)),
		},
		{
			name: "multi_slice_sub",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(10)),
			exp:  child(spec.Name("x"), spec.Slice(8)),
		},
		{
			name: "multi_overlaps",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(2, 5), spec.Slice(4, 5)),
			exp:  child(spec.Name("x"), spec.Slice(8), spec.Slice(2, 5)),
		},
		{
			name: "multi_overlap_reverse",
			seg:  child(spec.Name("x"), spec.Slice(4, 5), spec.Slice(2, 5), spec.Slice(8), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(2, 5), spec.Slice(8)),
		},
		{
			name: "three_in_one",
			seg:  child(spec.Slice(2, 4), spec.Slice(1, 3), spec.Slice(0, 5)),
			exp:  child(spec.Slice(0, 5)),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.seg.mergeSlices()
			a.Equal(tc.exp.selectors, tc.seg.selectors)
		})
	}
}
