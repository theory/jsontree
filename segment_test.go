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

	for _, tc := range []struct {
		test string
		seg  *segment
		str  string
	}{
		{
			test: "empty_child",
			seg:  child(),
			str:  "[]\n",
		},
		{
			test: "empty_descendant",
			seg:  descendant(),
			str:  "..[]\n",
		},
		{
			test: "wildcard",
			seg:  child(spec.Wildcard()),
			str:  "[*]\n",
		},
		{
			test: "one_key",
			seg:  child(spec.Name("foo")),
			str:  "[\"foo\"]\n",
		},
		{
			test: "two_keys",
			seg:  child(spec.Name("foo"), spec.Name("bar")),
			str:  "[\"foo\",\"bar\"]\n",
		},
		{
			test: "parent_child",
			seg: child(spec.Name("foo")).Append(
				child(spec.Name("bar")),
			),
			str: "[\"foo\"]\n└── [\"bar\"]\n",
		},
		{
			test: "parent_two_child",
			seg: child(spec.Name("foo")).Append(
				child(spec.Name("bar")),
				child(spec.Name("baz")),
			),
			str: "[\"foo\"]\n├── [\"bar\"]\n└── [\"baz\"]\n",
		},
		{
			test: "two_children_and_sub_keys",
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
			test: "mixed_and_deep",
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("foo")).Append(
					child(spec.Name("x")),
					child(spec.Name("y")).Append(
						child(spec.Wildcard()).Append(
							child(spec.Name("a")),
							child(spec.Name("b")),
						),
					),
				),
				child(spec.Name("bar")).Append(
					child(spec.Name("go")),
					child(spec.Name("z")).Append(
						child(spec.Wildcard()).Append(
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
			test: "wildcard",
			seg:  child(spec.Wildcard()),
			str:  "[*]\n",
		},
		{
			test: "one_index",
			seg:  child(spec.Index(0)),
			str:  "[0]\n",
		},
		{
			test: "two_indexes",
			seg:  child(spec.Index(0), spec.Index(2)),
			str:  "[0,2]\n",
		},
		{
			test: "index_index",
			seg:  child(spec.Index(0)).Append(child(spec.Index(2))),
			str:  "[0]\n└── [2]\n",
		},
		{
			test: "two_nested_indexes",
			seg: child(spec.Index(1)).Append(
				child(spec.Index(0)), child(spec.Index(2)),
			),
			str: "[1]\n├── [0]\n└── [2]\n",
		},
		{
			test: "two_keys_and_sub_indexes",
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
			test: "filter",
			seg: child(spec.Filter(spec.And(
				spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
			))),
			str: "[?($)]\n",
		},
		{
			test: "filter_and_key",
			seg: child(spec.Name("hi")).Append(
				child(spec.Name("x")),
				child(spec.Filter(spec.And(
					spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
				))),
			),
			str: "[\"hi\"]\n├── [\"x\"]\n└── [?($)]\n",
		},
		{
			test: "filter_and_key_segment",
			seg: child(
				spec.Name("x"),
				spec.Filter(spec.And(
					spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
				)),
			),
			str: "[\"x\",?($)]\n",
		},
		{
			test: "nested_filter",
			seg: child(spec.Name("x")).Append(
				child(spec.Filter(spec.And(
					spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
				))),
			),
			str: "[\"x\"]\n└── [?($)]\n",
		},
		{
			test: "slice_def_start_stop_neg_step",
			seg: child(spec.Name("x")).Append(
				child(spec.Slice(nil, nil, -1)),
			),
			str: "[\"x\"]\n└── [::-1]\n",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.str, tc.seg.String(), tc.test)
		})
	}
}

func TestIsWildcard(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		seg  *segment
		exp  bool
	}{
		{"empty", &segment{}, false},
		{"name", &segment{selectors: []spec.Selector{spec.Name("x")}}, false},
		{"index", &segment{selectors: []spec.Selector{spec.Index(0)}}, false},
		{"slice", &segment{selectors: []spec.Selector{spec.Slice()}}, false},
		{"filter", &segment{selectors: []spec.Selector{mkFilter("$[?@]")}}, false},
		{"wildcard", &segment{selectors: []spec.Selector{spec.Wildcard()}}, true},
		{"multiples", &segment{selectors: []spec.Selector{spec.Wildcard(), spec.Index(0)}}, false},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, tc.seg.isWildcard())
		})
	}
}

func TestHasSelector(t *testing.T) {
	t.Parallel()

	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		test  string
		list  []spec.Selector
		sel   spec.Selector
		exp   bool
		exact bool
	}{
		{
			test: "wildcard_empty",
			list: []spec.Selector{},
			sel:  spec.Wildcard(),
		},
		{
			test: "wildcard_name",
			list: []spec.Selector{spec.Name("foo")},
			sel:  spec.Wildcard(),
		},
		{
			test: "wildcard_index",
			list: []spec.Selector{spec.Index(1)},
			sel:  spec.Wildcard(),
		},
		{
			test:  "wildcard_wildcard",
			list:  []spec.Selector{spec.Wildcard()},
			sel:   spec.Wildcard(),
			exp:   true,
			exact: true,
		},
		{
			test: "name_empty",
			list: []spec.Selector{},
			sel:  spec.Name("foo"),
		},
		{
			test:  "name_exists",
			list:  []spec.Selector{spec.Name("foo")},
			sel:   spec.Name("foo"),
			exp:   true,
			exact: true,
		},
		{
			test:  "name_exists_list",
			list:  []spec.Selector{spec.Name("foo"), spec.Name("bar"), spec.Index(0)},
			sel:   spec.Name("foo"),
			exp:   true,
			exact: true,
		},
		{
			test: "name_not_exists_list",
			list: []spec.Selector{spec.Name("foo"), spec.Name("bar"), spec.Index(0)},
			sel:  spec.Name("hello"),
		},
		{
			test: "name_wildcard",
			list: []spec.Selector{spec.Wildcard()},
			sel:  spec.Name("foo"),
			exp:  true,
		},
		{
			test: "name_index",
			list: []spec.Selector{spec.Index(0)},
			sel:  spec.Name("foo"),
		},
		{
			test: "index_empty",
			list: []spec.Selector{},
			sel:  spec.Index(1),
		},
		{
			test:  "index_exists",
			list:  []spec.Selector{spec.Index(1)},
			sel:   spec.Index(1),
			exp:   true,
			exact: true,
		},
		{
			test: "index_wildcard",
			list: []spec.Selector{spec.Wildcard()},
			sel:  spec.Index(1),
			exp:  true,
		},
		{
			test: "index_not_exists",
			list: []spec.Selector{spec.Index(2)},
			sel:  spec.Index(1),
		},
		{
			test:  "index_in_list",
			list:  []spec.Selector{spec.Name("foo"), spec.Index(0), spec.Index(1)},
			sel:   spec.Index(1),
			exp:   true,
			exact: true,
		},
		{
			test: "index_not_in_list",
			list: []spec.Selector{spec.Name("foo"), spec.Index(0), spec.Index(1)},
			sel:  spec.Index(2),
		},
		{
			test: "index_in_default_slice",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(2),
			exp:  true,
		},
		{
			test: "index_in_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(2),
			exp:  true,
		},
		{
			test: "index_in_explicit_slice_step",
			list: []spec.Selector{spec.Slice(1, 4, 2)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			test: "index_not_in_explicit_slice_step",
			list: []spec.Selector{spec.Slice(1, 4, 2)},
			sel:  spec.Index(2),
		},
		{
			test: "index_not_in_backwards_slice",
			list: []spec.Selector{spec.Slice(4, 1)},
			sel:  spec.Index(2),
		},
		{
			test: "index_start_of_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(1),
			exp:  true,
		},
		{
			test: "index_end_of_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			test: "index_gt_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(5),
		},
		{
			test: "index_lt_explicit_slice",
			list: []spec.Selector{spec.Slice(1, 4)},
			sel:  spec.Index(0),
		},
		{
			test: "index_not_in_neg_slice",
			list: []spec.Selector{spec.Slice(-4, -1)},
			sel:  spec.Index(2),
		},
		{
			test: "neg_index_in_default",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(-5),
			exp:  true,
		},
		{
			test: "neg_one_in_default",
			list: []spec.Selector{spec.Slice()},
			sel:  spec.Index(-1),
			exp:  true,
		},
		{
			test: "neg_one_in_explicit",
			list: []spec.Selector{spec.Slice(0, 5)},
			sel:  spec.Index(-1),
			exp:  true,
		},
		{
			test: "neg_just_in_explicit",
			list: []spec.Selector{spec.Slice(0, 2)},
			sel:  spec.Index(-2),
			exp:  true,
		},
		{
			test: "neg_not_in_explicit",
			list: []spec.Selector{spec.Slice(0, 2)},
			sel:  spec.Index(-3),
		},
		{
			test: "in_neg_step",
			list: []spec.Selector{spec.Slice(5, 2, -1)},
			sel:  spec.Index(3),
			exp:  true,
		},
		{
			test: "not_in_neg_two_step",
			list: []spec.Selector{spec.Slice(6, 2, -2)},
			sel:  spec.Index(4),
		},
		{
			test: "not_in_neg_three_step",
			list: []spec.Selector{spec.Slice(6, 1, -3)},
			sel:  spec.Index(2),
		},
		{
			test: "slice_empty",
			list: []spec.Selector{},
			sel:  spec.Slice(),
		},
		{
			test: "slice_step_0",
			list: []spec.Selector{},
			sel:  spec.Slice(1, 3, 0),
			exp:  true,
		},
		{
			test: "slice_start_stop_equal",
			list: []spec.Selector{},
			sel:  spec.Slice(3, 3),
			exp:  true,
		},
		{
			test:  "same_slice",
			list:  []spec.Selector{spec.Slice(1, 3)},
			sel:   spec.Slice(1, 3),
			exp:   true,
			exact: true,
		},
		{
			test: "within_slice_start",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(2, 3),
			exp:  true,
		},
		{
			test: "before_start",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(0, 2),
		},
		{
			test: "after_end",
			list: []spec.Selector{spec.Slice(1, 3)},
			sel:  spec.Slice(1, 4),
		},
		{
			test: "no_selectors",
			list: []spec.Selector{},
			sel:  simpleExists,
		},
		{
			test:  "has_filter",
			list:  []spec.Selector{simpleExists},
			sel:   simpleExists,
			exp:   true,
			exact: true,
		},
		{
			test: "not_has_filter",
			list: []spec.Selector{simpleExists},
			sel:  diffExists,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, selectorsContain(tc.list, tc.sel))
			seg := &segment{selectors: tc.list}
			a.Equal(tc.exp, seg.hasSelector(tc.sel))
			a.Equal(tc.exact, seg.hasExactSelector(tc.sel))
		})
	}
}

func TestHasSelectors(t *testing.T) {
	t.Parallel()
	// simpleExists := mkFilter("$[?@]")
	// diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		test      string
		seg       *segment
		selectors []spec.Selector
		exp       bool
		same      bool
		exact     bool
	}{
		{
			test:      "empty",
			seg:       child(),
			selectors: []spec.Selector{},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			test:      "a_name",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
			same:      true,
			exact:     true,
		},
		{
			test:      "diff_name",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("y")},
		},
		{
			test:      "diff_length",
			seg:       child(spec.Name("x")),
			selectors: []spec.Selector{spec.Name("x"), spec.Name("y")},
		},
		{
			test:      "diff_length_has_ok",
			seg:       child(spec.Name("x"), spec.Name("y")),
			selectors: []spec.Selector{spec.Name("x")},
			exp:       true,
		},
		{
			test:      "same_not_exact",
			seg:       child(spec.Name("x"), spec.Slice(0)),
			selectors: []spec.Selector{spec.Name("x"), spec.Index(0)},
			exp:       true,
			same:      true,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, tc.seg.hasSelectors(tc.selectors))
			a.Equal(tc.same, tc.seg.hasSameSelectors(tc.selectors))
			a.Equal(tc.exact, tc.seg.hasExactSelectors(tc.selectors))
		})
	}
}

func TestContainsIndex(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test  string
		slice spec.SliceSelector
		idx   int
		exp   bool
	}{
		{
			test:  "in_slice_end",
			slice: spec.Slice(nil, 6),
			idx:   5,
			exp:   true,
		},
		{
			test:  "not_in_slice_end",
			slice: spec.Slice(nil, 6),
			idx:   6,
			exp:   false,
		},
		{
			test:  "in_bounded",
			slice: spec.Slice(2, 6),
			idx:   5,
			exp:   true,
		},
		{
			test:  "not_in_bounded",
			slice: spec.Slice(2, 6),
			idx:   6,
			exp:   false,
		},
		{
			test:  "in_bounded_step_two",
			slice: spec.Slice(2, 6, 2),
			idx:   4,
			exp:   true,
		},
		{
			test:  "not_in_bounded_step_two",
			slice: spec.Slice(2, 6, 2),
			idx:   3,
			exp:   false,
		},
		{
			test:  "index_not_in_backwards_slice",
			slice: spec.Slice(4, 1),
			idx:   2,
			exp:   false,
		},
		{
			test:  "at_slice_start",
			slice: spec.Slice(1, 4),
			idx:   1,
			exp:   true,
		},
		{
			test:  "at_slice_default_start",
			slice: spec.Slice(nil, 4),
			idx:   0,
			exp:   true,
		},
		{
			test:  "before_slice_start",
			slice: spec.Slice(2, 4),
			idx:   1,
			exp:   false,
		},
		{
			test:  "neg_start",
			slice: spec.Slice(-4, 20),
			idx:   2,
			exp:   false,
		},
		{
			test:  "neg_end",
			slice: spec.Slice(0, -1),
			idx:   2,
			exp:   false,
		},
		{
			test:  "both_neg",
			slice: spec.Slice(-4, -1),
			idx:   0,
			exp:   false,
		},
		{
			test:  "end_lt_start",
			slice: spec.Slice(12, 0),
			idx:   5,
			exp:   false,
		},
		{
			test:  "in_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   3,
			exp:   true,
		},
		{
			test:  "not_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   1,
			exp:   false,
		},
		{
			test:  "exclude_end_neg_one_step",
			slice: spec.Slice(5, 2, -1),
			idx:   2,
			exp:   false,
		},
		{
			test:  "in_neg_one_step_start",
			slice: spec.Slice(5, 2, -1),
			idx:   5,
			exp:   true,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

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

	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		test   string
		list   []spec.Selector
		filter *spec.FilterSelector
		exp    bool
	}{
		{
			test:   "no_selectors",
			list:   []spec.Selector{},
			filter: simpleExists,
			exp:    false,
		},
		{
			test:   "has_filter",
			list:   []spec.Selector{simpleExists},
			filter: simpleExists,
			exp:    true,
		},
		{
			test:   "not_has_filter",
			list:   []spec.Selector{simpleExists},
			filter: diffExists,
			exp:    false,
		},
		{
			test:   "same_operands",
			list:   []spec.Selector{mkFilter("$[?@.x > @.y]")},
			filter: mkFilter("$[?@.x > @.y]"),
			exp:    true,
		},
		{
			test:   "diff_cmp_operands",
			list:   []spec.Selector{mkFilter("$[?@.a > @.y]")},
			filter: mkFilter("$[?@.x > @.y]"),
			exp:    false,
		},
		{
			test:   "diff_and_operands",
			list:   []spec.Selector{mkFilter("$[?@.a && @.y]")},
			filter: mkFilter("$[?@.x && @.y]"),
			exp:    false,
		},
		{
			test:   "diff_or_operands",
			list:   []spec.Selector{mkFilter("$[?@.a || @.y]")},
			filter: mkFilter("$[?@.x || @.y]"),
			exp:    false,
		},
		// These examples could be equivalent if spec.ComparisonExpr
		// normalized its expressions for deterministic order.
		{
			test:   "reversed_operands",
			list:   []spec.Selector{mkFilter("$[?@.x > @.y]")},
			filter: mkFilter("$[?@.y > @.x]"),
			exp:    false,
		},
		{
			test:   "reversed_eq_operands",
			list:   []spec.Selector{mkFilter("$[?@.x == @.y]")},
			filter: mkFilter("$[?@.y == @.x]"),
			exp:    false,
		},
		{
			test:   "reversed_ne_operands",
			list:   []spec.Selector{mkFilter("$[?@.x != @.y]")},
			filter: mkFilter("$[?@.y != @.x]"),
			exp:    false,
		},
		// These examples could be equal if spec.LogicalOr and spec.LogicalAnd
		// normalized adopted https://pkg.go.dev/sort#Interface.
		{
			test:   "reversed_or_operands",
			list:   []spec.Selector{mkFilter("$[?@.x || @.y]")},
			filter: mkFilter("$[?@.y || @.x]"),
			exp:    false,
		},
		{
			test:   "reversed_and_operands",
			list:   []spec.Selector{mkFilter("$[?@.x && @.y]")},
			filter: mkFilter("$[?@.y && @.x]"),
			exp:    false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, containsFilter(tc.list, tc.filter))
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

	for _, tc := range []struct {
		test  string
		list  []spec.Selector
		slice spec.SliceSelector
		exp   bool
	}{
		{
			test:  "no_selectors",
			list:  []spec.Selector{},
			slice: spec.Slice(),
			exp:   false,
		},
		{
			test:  "step_0",
			list:  []spec.Selector{},
			slice: spec.Slice(1, 3, 0),
			exp:   true,
		},
		{
			test:  "start_stop_equal",
			list:  []spec.Selector{},
			slice: spec.Slice(3, 3),
			exp:   true,
		},
		{
			test:  "same_slice",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 3),
			exp:   true,
		},
		{
			test:  "within_start",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(2, 3),
			exp:   true,
		},
		{
			test:  "within_end",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 2),
			exp:   true,
		},
		{
			test:  "before_start",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(0, 2),
			exp:   false,
		},
		{
			test:  "after_end",
			list:  []spec.Selector{spec.Slice(1, 3)},
			slice: spec.Slice(1, 4),
			exp:   false,
		},
		{
			test:  "multiple_of_step",
			list:  []spec.Selector{spec.Slice(1, 3, 2)},
			slice: spec.Slice(1, 3, 4),
			exp:   true,
		},
		{
			test:  "out_of_step",
			list:  []spec.Selector{spec.Slice(1, 3, 2)},
			slice: spec.Slice(1, 3, 3),
			exp:   false,
		},
		{
			test:  "over_step",
			list:  []spec.Selector{spec.Slice(1, 3, 4)},
			slice: spec.Slice(1, 3, 2),
			exp:   false,
		},

		{
			test:  "same_backward_slice",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(3, 1, -1),
			exp:   true,
		},
		{
			test:  "within_start_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(3, 2, -1),
			exp:   true,
		},
		{
			test:  "within_end_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(2, 1, -1),
			exp:   true,
		},
		{
			test:  "after_end_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(2, 0, -1),
			exp:   false,
		},
		{
			test:  "before_start_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(4, 1, -1),
			exp:   false,
		},
		{
			test:  "multiple_of_backward_step",
			list:  []spec.Selector{spec.Slice(3, 1, -2)},
			slice: spec.Slice(3, 1, -4),
			exp:   true,
		},
		{
			test:  "out_of_step_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -2)},
			slice: spec.Slice(3, 1, -3),
			exp:   false,
		},
		{
			test:  "over_step_backward",
			list:  []spec.Selector{spec.Slice(3, 1, -4)},
			slice: spec.Slice(3, 1, -2),
			exp:   false,
		},
		{
			test: "opposite_step_not_in_range",
			// $[1:3:1] != $[3:1:-1]
			list:  []spec.Selector{spec.Slice(3, 1, -1)},
			slice: spec.Slice(1, 3, 1),
			exp:   false,
		},
		{
			// $[2:0:-1] == $[1:3:1]
			test:  "opposites",
			list:  []spec.Selector{spec.Slice(1, 3, 1)},
			slice: spec.Slice(2, 0, -1),
			exp:   true,
		},
		{
			// $[1:3:1] == $[2:0:-1]
			test:  "inverted_opposites",
			list:  []spec.Selector{spec.Slice(2, 0, -1)},
			slice: spec.Slice(1, 3, 1),
			exp:   true,
		},
		{
			// $[2:0:-2] mod within $[1:3:1]
			test:  "opposite_mod_step",
			list:  []spec.Selector{spec.Slice(1, 3, 1)},
			slice: spec.Slice(2, 0, -2),
			exp:   true,
		},
		{
			// $[1:3:1] not mod within $[2:0:-2]
			test:  "opposite_not_mod_step",
			list:  []spec.Selector{spec.Slice(2, 0, -2)},
			slice: spec.Slice(1, 3, 1),
			exp:   false,
		},
		{
			// $[2:0:-1] within $[1:5:1]
			test:  "within_opposite",
			list:  []spec.Selector{spec.Slice(1, 5, 1)},
			slice: spec.Slice(2, 0, -1),
			exp:   true,
		},
		{
			test:  "equals_index",
			list:  []spec.Selector{spec.Index(3)},
			slice: spec.Slice(3, 4),
			exp:   true,
		},
		{
			test:  "equals_index_inverted",
			list:  []spec.Selector{spec.Index(3)},
			slice: spec.Slice(4, 3, -1),
			exp:   true,
		},
		{
			test:  "not_equals_index",
			list:  []spec.Selector{spec.Index(4)},
			slice: spec.Slice(3, 4),
			exp:   false,
		},
		{
			test:  "not_equals_index_inverted",
			list:  []spec.Selector{spec.Index(4)},
			slice: spec.Slice(4, 3, -1),
			exp:   false,
		},
		{
			// XXX Compare all indexes to slice range?
			test:  "equals_all_indexes",
			list:  []spec.Selector{spec.Index(3), spec.Index(4)},
			slice: spec.Slice(3, 5),
			exp:   false,
		},
		{
			test:  "defaults_neg_slice_covers_all",
			list:  []spec.Selector{spec.Slice(nil, nil, -1)},
			slice: spec.Slice(0, 2),
			exp:   true,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, containsSlice(tc.list, tc.slice))
		})
	}
}

func TestIsBranch(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test   string
		seg    *segment
		branch []*spec.Segment
		exp    bool
	}{
		{
			test:   "empty",
			seg:    &segment{},
			branch: []*spec.Segment{},
			exp:    true,
		},
		{
			test:   "empty_is_not_name",
			seg:    &segment{},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			test:   "eq_name",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    true,
		},
		{
			test:   "ne_name",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("y"))},
			exp:    false,
		},
		{
			test: "size",
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
			test: "eq_branch_mixed",
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1)).Append(
						child(spec.Wildcard()).Append(
							child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1)),
				spec.Child(spec.Wildcard()),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: true,
		},
		{
			test: "ne_child_selectors",
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1), spec.Name("x")).Append(
						child(spec.Wildcard()).Append(
							child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1)),
				spec.Child(spec.Wildcard()),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: false,
		},
		{
			test: "ne_spec_selectors",
			seg: &segment{children: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Index(0), spec.Slice(1)).Append(
						child(spec.Wildcard()).Append(
							child(mkFilter("$[?@]")),
						),
					),
				),
			}},
			branch: []*spec.Segment{
				spec.Child(spec.Name("x")),
				spec.Child(spec.Index(0), spec.Slice(1), spec.Name("x")),
				spec.Child(spec.Wildcard()),
				spec.Child(mkFilter("$[?@]")),
			},
			exp: false,
		},
		{
			test: "diff_child_length",
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
			test: "diff_spec_length",
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
			test:   "ne_spec_desc",
			seg:    &segment{children: []*segment{child(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Descendant(spec.Name("x"))},
			exp:    false,
		},
		{
			test:   "ne_child_desc",
			seg:    &segment{children: []*segment{descendant(spec.Name("x"))}},
			branch: []*spec.Segment{spec.Child(spec.Name("x"))},
			exp:    false,
		},
		{
			test: "ne_sub_spec_desc",
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
			test: "ne_sub_child_desc",
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, tc.seg.isBranch(tc.branch))
		})
	}
}

func TestMergeSelectors(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test      string
		selectors []spec.Selector
		merge     []spec.Selector
		exp       []spec.Selector
	}{
		{
			test: "none",
		},
		{
			test:  "name_into_empty",
			merge: []spec.Selector{spec.Name("x")},
			exp:   []spec.Selector{spec.Name("x")},
		},
		{
			test:      "name_into_existing",
			selectors: []spec.Selector{spec.Name("x")},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Name("x")},
		},
		{
			test:      "name_into_index",
			selectors: []spec.Selector{spec.Index(0)},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Index(0), spec.Name("x")},
		},
		{
			test:      "name_into_index_no_dupe",
			selectors: []spec.Selector{spec.Index(0), spec.Name("x")},
			merge:     []spec.Selector{spec.Name("x")},
			exp:       []spec.Selector{spec.Index(0), spec.Name("x")},
		},
		{
			test:      "index_slice",
			selectors: []spec.Selector{spec.Slice(2), spec.Name("x")},
			merge:     []spec.Selector{spec.Index(2), spec.Name("y")},
			exp:       []spec.Selector{spec.Slice(2), spec.Name("x"), spec.Name("y")},
		},
		{
			test:      "neg_slice_index",
			selectors: []spec.Selector{spec.Slice(nil, nil, -1)},
			merge:     []spec.Selector{spec.Index(0), spec.Index(2)},
			exp:       []spec.Selector{spec.Slice(nil, nil, -1)},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			seg := &segment{selectors: tc.selectors}
			seg.mergeSelectors(tc.merge)
			assert.Equal(t, tc.exp, seg.selectors)
		})
	}
}

func TestMergeChildren(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test     string
		children []*segment
		expect   []*segment
	}{
		{
			test:     "empty",
			children: []*segment{},
			expect:   []*segment{},
		},
		{
			test:     "one_child",
			children: []*segment{child(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("x"))},
		},
		{
			test:     "merge_name",
			children: []*segment{child(spec.Name("x")), child(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("x"))},
		},
		{
			test:     "merge_selectors",
			children: []*segment{child(spec.Name("x")), child(spec.Name("y"))},
			expect:   []*segment{child(spec.Name("x"), spec.Name("y"))},
		},
		{
			test: "no_merge_diff_child",
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
			test: "merge_same_child",
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
			test: "merge_same_nested_multi_select",
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
			test: "no_merge_diff_depth",
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
			test: "merge_nested_selectors",
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
			test:     "merge_descendants",
			children: []*segment{descendant(spec.Name("x")), descendant(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			test:     "merge_descendant_child",
			children: []*segment{descendant(spec.Name("x")), child(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			test:     "merge_descendant_sub_child",
			children: []*segment{descendant(spec.Name("x"), spec.Name("y")), child(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			test:     "merge_child_descendant",
			children: []*segment{child(spec.Name("x")), descendant(spec.Name("x"))},
			expect:   []*segment{descendant(spec.Name("x"))},
		},
		{
			test:     "merge_child_sub_descendant",
			children: []*segment{child(spec.Name("x")), descendant(spec.Name("x"), spec.Name("y"))},
			expect:   []*segment{descendant(spec.Name("x"), spec.Name("y"))},
		},
		{
			test:     "merge_only_common_descendant_child_selector",
			children: []*segment{descendant(spec.Name("x")), child(spec.Name("x"), spec.Name("y"))},
			expect:   []*segment{descendant(spec.Name("x")), child(spec.Name("y"))},
		},
		{
			test:     "merge_only_common_prev_descendant_selector",
			children: []*segment{child(spec.Name("x"), spec.Name("y")), descendant(spec.Name("x"))},
			expect:   []*segment{child(spec.Name("y")), descendant(spec.Name("x"))},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			seg := &segment{children: tc.children}
			seg.deduplicate()
			assert.Equal(t, tc.expect, seg.children)
		})
	}
}

func TestRemoveCommonSelectorsFrom(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		sel1 []spec.Selector
		sel2 []spec.Selector
		exp2 []spec.Selector
		res  bool
	}{
		{
			test: "empty",
			res:  true,
		},
		{
			test: "empty_seg2",
			sel1: []spec.Selector{spec.Name("x")},
			res:  true,
		},
		{
			test: "no_commonality",
			sel1: []spec.Selector{spec.Name("x")},
			sel2: []spec.Selector{spec.Name("y")},
			exp2: []spec.Selector{spec.Name("y")},
		},
		{
			test: "remove_one_all",
			sel1: []spec.Selector{spec.Name("x")},
			sel2: []spec.Selector{spec.Name("x")},
			exp2: []spec.Selector{},
			res:  true,
		},
		{
			test: "remove_one",
			sel1: []spec.Selector{spec.Name("x"), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x")},
			exp2: []spec.Selector{},
			res:  true,
		},
		{
			test: "remove_one_leave_one",
			sel1: []spec.Selector{spec.Name("x"), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Name("a")},
			exp2: []spec.Selector{spec.Name("a")},
			res:  false,
		},
		{
			test: "remove_sub_slice",
			sel1: []spec.Selector{spec.Slice(1, 3), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Slice(1, 2)},
			exp2: []spec.Selector{spec.Name("x")},
			res:  false,
		},
		{
			test: "remove_index_matching_slice",
			sel1: []spec.Selector{spec.Slice(1, 3), spec.Name("y")},
			sel2: []spec.Selector{spec.Name("x"), spec.Index(2)},
			exp2: []spec.Selector{spec.Name("x")},
			res:  false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			seg1 := &segment{selectors: tc.sel1}
			seg2 := &segment{selectors: tc.sel2}
			a.Equal(tc.res, seg1.removeCommonSelectorsFrom(seg2))
			a.Equal(tc.exp2, seg2.selectors, "selectors 2")
		})
	}
}

func TestSameBranches(t *testing.T) {
	t.Parallel()

	simpleExists := mkFilter("$[?@]")
	diffExists := mkFilter("$[?@.a]")

	for _, tc := range []struct {
		test string
		seg1 *segment
		seg2 *segment
		exp  bool
	}{
		{
			test: "empties",
			seg1: child(),
			seg2: child(),
			exp:  true,
		},
		{
			test: "single_child_eq_name",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x"))),
			exp:  true,
		},
		{
			test: "single_child_ne_name",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("y"))),
			exp:  false,
		},
		{
			test: "single_child_eq_multi_select",
			seg1: child().Append(child(spec.Name("x"), spec.Index(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(0))),
			exp:  true,
		},
		{
			test: "single_child_ne_multi_select",
			seg1: child().Append(child(spec.Name("x"), spec.Index(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			test: "single_child_ne_index_slice",
			seg1: child().Append(child(spec.Name("x"), spec.Slice(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Index(1))),
			exp:  false,
		},
		{
			test: "single_child_ne_slices",
			seg1: child().Append(child(spec.Name("x"), spec.Slice(0))),
			seg2: child().Append(child(spec.Name("x"), spec.Slice(1))),
			exp:  false,
		},
		{
			test: "single_child_eq_filter",
			seg1: child().Append(child(simpleExists)),
			seg2: child().Append(child(simpleExists)),
			exp:  true,
		},
		{
			test: "single_child_ne_filter",
			seg1: child().Append(child(simpleExists)),
			seg2: child().Append(child(diffExists)),
			exp:  false,
		},
		{
			test: "wildcards",
			seg1: child().Append(child(spec.Wildcard())),
			seg2: child().Append(child(spec.Wildcard())),
			exp:  true,
		},
		{
			test: "wildcard_with_eq_name_is_ne",
			seg1: child().Append(child(spec.Wildcard(), spec.Name("x"))),
			seg2: child().Append(child(spec.Wildcard(), spec.Name("x"))),
			exp:  false,
		},
		{
			test: "diff_branches_child_count",
			seg1: child().Append(child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			exp:  false,
		},
		{
			test: "diff_children_child_count",
			seg1: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x"))),
			exp:  false,
		},
		{
			test: "same_children",
			seg1: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			seg2: child().Append(child(spec.Name("x")), child(spec.Name("x"))),
			exp:  true,
		},
		{
			test: "same_nested_children",
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
			test: "diff_nested_children",
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, tc.seg1.sameBranches(tc.seg2))
		})
	}
}

func TestMergeSlices(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		seg  *segment
		exp  *segment
	}{
		{
			test: "Empty",
			seg:  child(),
			exp:  child(),
		},
		{
			test: "no_slices",
			seg:  child(spec.Name("x"), spec.Index(0)),
			exp:  child(spec.Name("x"), spec.Index(0)),
		},
		{
			test: "one_slice",
			seg:  child(spec.Name("x"), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(12, 18)),
		},
		{
			test: "sub_slice",
			seg:  child(spec.Name("x"), spec.Slice(10), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(10)),
		},
		{
			test: "slice_sub",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(10)),
			exp:  child(spec.Name("x"), spec.Slice(10)),
		},
		{
			test: "sub_slice_slice",
			seg:  child(spec.Name("x"), spec.Slice(8), spec.Slice(10), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(8)),
		},
		{
			test: "multi_slice_sub",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(10)),
			exp:  child(spec.Name("x"), spec.Slice(8)),
		},
		{
			test: "multi_overlaps",
			seg:  child(spec.Name("x"), spec.Slice(12, 18), spec.Slice(8), spec.Slice(2, 5), spec.Slice(4, 5)),
			exp:  child(spec.Name("x"), spec.Slice(8), spec.Slice(2, 5)),
		},
		{
			test: "multi_overlap_reverse",
			seg:  child(spec.Name("x"), spec.Slice(4, 5), spec.Slice(2, 5), spec.Slice(8), spec.Slice(12, 18)),
			exp:  child(spec.Name("x"), spec.Slice(2, 5), spec.Slice(8)),
		},
		{
			test: "three_in_one",
			seg:  child(spec.Slice(2, 4), spec.Slice(1, 3), spec.Slice(0, 5)),
			exp:  child(spec.Slice(0, 5)),
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			tc.seg.mergeSlices()
			assert.Equal(t, tc.exp.selectors, tc.seg.selectors)
		})
	}
}
