package jsontree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

func TestRunRoot(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		val  any
	}{
		{"string", "Hi there 😀"},
		{"bool", true},
		{"float", 98.6},
		{"float64", float64(98.6)},
		{"float32", float32(98.6)},
		{"int", 42},
		{"int64", int64(42)},
		{"int32", int32(42)},
		{"int16", int16(42)},
		{"int8", int8(42)},
		{"uint64", uint64(42)},
		{"uint32", uint32(42)},
		{"uint16", uint16(42)},
		{"uint8", uint8(42)},
		{"struct", struct{ x int }{}},
		{"nil", nil},
		{"map", map[string]any{"x": true, "y": []any{1, 2}}},
		{"slice", []any{1, 2, map[string]any{"x": true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sel := &Tree{root: &segment{}}
			a.Equal(tc.val, sel.Select(tc.val))
			switch tc.val.(type) {
			case map[string]any, []any:
				return
			default:
				// Anything other than a slice or map returns nil if
				// there are path segments.
				sel.root = &segment{children: []*segment{child(spec.Wildcard)}}
				a.Nil(sel.Select(tc.val))
			}
		})
	}
}

func TestObjectSelection(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		segs []*segment
		obj  map[string]any
		exp  map[string]any
	}{
		{
			name: "root",
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"x": true, "y": []any{1, 2}},
		},
		{
			name: "one_key_scalar",
			segs: []*segment{child(spec.Name("x"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"x": true},
		},
		{
			name: "one_key_array",
			segs: []*segment{child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"y": []any{1, 2}},
		},
		{
			name: "one_key_object",
			segs: []*segment{child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": map[string]any{"a": 1}},
			exp:  map[string]any{"y": map[string]any{"a": 1}},
		},
		{
			name: "filter_object",
			segs: []*segment{child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Comparison(
					spec.SingularQuery(false, []spec.Selector{spec.Name("a")}),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			}}))},
			obj: map[string]any{
				"kim":   map[string]any{"a": 42, "firm": "HHM"},
				"jimmy": map[string]any{"a": 41, "firm": "JMM"},
				"chuck": map[string]any{"a": 43, "firm": "on leave"},
			},
			exp: map[string]any{
				"kim":   map[string]any{"a": 42, "firm": "HHM"},
				"chuck": map[string]any{"a": 43, "firm": "on leave"},
			},
		},
		{
			name: "filter_object_key",
			segs: []*segment{child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Comparison(
					spec.SingularQuery(false, []spec.Selector{spec.Name("a")}),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			}})).Append(child(spec.Name("firm")))},
			obj: map[string]any{
				"kim":   map[string]any{"a": 42, "firm": "HHM"},
				"jimmy": map[string]any{"a": 41, "firm": "JMM"},
				"chuck": map[string]any{"a": 43, "firm": "on leave"},
			},
			exp: map[string]any{
				"kim":   map[string]any{"firm": "HHM"},
				"chuck": map[string]any{"firm": "on leave"},
			},
		},
		{
			name: "multiple_keys",
			segs: []*segment{child(spec.Name("x")), child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}, "z": "hi"},
			exp:  map[string]any{"x": true, "y": []any{1, 2}},
		},
		{
			name: "key_and_filter",
			segs: []*segment{child(spec.Name("x")), child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Comparison(
					spec.SingularQuery(false, []spec.Selector{spec.Name("z")}),
					spec.EqualTo,
					spec.Literal("hi"),
				),
			}}))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{"x": true, "y": map[string]any{"z": "hi"}},
		},
		{
			name: "key_then_filter_cur_true",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(false, []*spec.Segment{spec.Child(spec.Index(0))})),
			}})))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": []any{1}}, "z": "hi"},
			exp: map[string]any{"y": map[string]any{"z": []any{1}}},
		},
		{
			name: "key_then_filter_cur_false",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(false, []*spec.Segment{spec.Child(spec.Index(1))})),
			}})))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": []any{1}}, "z": "hi"},
			exp: map[string]any{},
		},
		{
			name: "key_then_filter_root_true",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(true, []*spec.Segment{spec.Child(spec.Name("x"))})),
			}})))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{"y": map[string]any{"z": "hi"}},
		},
		{
			name: "key_then_filter_root_false",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(true, []*spec.Segment{spec.Child(spec.Name("a"))})),
			}})))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{},
		},
		{
			name: "three_level_path",
			segs: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("i")),
					),
				),
			},
			obj: map[string]any{
				"x": map[string]any{
					"a": map[string]any{
						"i": []any{1, 2},
						"j": 42,
					},
					"b": "no",
				},
				"y": 1,
			},
			exp: map[string]any{
				"x": map[string]any{
					"a": map[string]any{
						"i": []any{1, 2},
					},
				},
			},
		},
		{
			name: "nested_multiple_keys",
			segs: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("a")),
					child(spec.Name("b")),
				),
			},
			obj: map[string]any{"x": map[string]any{"a": "go", "b": "no", "c": 1}, "y": 1},
			exp: map[string]any{"x": map[string]any{"a": "go", "b": "no"}},
		},
		{
			name: "varying_nesting_levels",
			segs: []*segment{
				child(spec.Name("x")).Append(
					child(spec.Name("a")),
					child(spec.Name("b")).Append(
						child(spec.Name("i")),
					),
				),
			},
			obj: map[string]any{
				"x": map[string]any{
					"a": "go",
					"b": map[string]any{"i": 12, "j": 1},
					"c": 1,
				},
				"y": 1,
			},
			exp: map[string]any{"x": map[string]any{"a": "go", "b": map[string]any{"i": 12}}},
		},
		{
			name: "wildcard_keys",
			segs: []*segment{
				child(spec.Wildcard).Append(
					child(spec.Name("a")),
					child(spec.Name("b")),
				),
			},
			obj: map[string]any{
				"x": map[string]any{"a": "go", "b": 2, "c": 5},
				"y": map[string]any{"a": 2, "b": 3, "d": 3},
			},
			exp: map[string]any{
				"x": map[string]any{"a": "go", "b": 2},
				"y": map[string]any{"a": 2, "b": 3},
			},
		},
		{
			name: "any_key_indexes",
			segs: []*segment{
				child(spec.Wildcard).Append(
					child(spec.Index(0)),
					child(spec.Index(1)),
				),
			},
			obj: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a", 2, "b", 3, "d", 3},
			},
			exp: map[string]any{
				"x": []any{"a", "go"},
				"y": []any{"a", 2},
			},
		},
		{
			name: "any_key_nonexistent_index",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Index(1)))},
			obj: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a"},
			},
			exp: map[string]any{"x": []any{nil, "go"}},
		},
		{
			name: "nonexistent_key",
			segs: []*segment{child(spec.Name("x"))},
			obj:  map[string]any{"y": []any{1, 2}},
			exp:  map[string]any{},
		},
		{
			name: "nonexistent_branch_key",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("z")))},
			obj:  map[string]any{"y": []any{1, 2}},
			exp:  map[string]any{},
		},
		{
			name: "wildcard_then_nonexistent_key",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("x")))},
			obj:  map[string]any{"y": map[string]any{"a": 1}},
			exp:  map[string]any{},
		},
		{
			name: "not_an_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			obj:  map[string]any{"x": true},
			exp:  map[string]any{},
		},
		{
			name: "not_an_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			obj:  map[string]any{"x": true},
			exp:  map[string]any{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tree := Tree{child().Append(tc.segs...)}
			a.Equal(tc.exp, tree.Select(tc.obj))
		})
	}

	for _, tc := range []struct {
		name string
		segs []*segment
		src  map[string]any
		dst  map[string]any
		err  string
	}{
		{
			name: "dest_not_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			src:  map[string]any{"x": map[string]any{}},
			dst:  map[string]any{"x": []any{1}},
			err:  `jsontree: expected destination object but got []interface {}`,
		},
		{
			name: "dest_not_array",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Index(1)))},
			src:  map[string]any{"x": []any{}},
			dst:  map[string]any{"x": map[string]any{"x": 1}},
			err:  `jsontree: expected destination array but got map[string]interface {}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// In general a value in dst should only be a map because we sanitize
			// the segments in advance, but this check ensures it at runtime.
			tree := &Tree{}
			a.PanicsWithValue(tc.err, func() {
				tree.selectObjectSegment(&segment{children: tc.segs}, nil, tc.src, tc.dst)
			})
		})
	}
}

func TestArraySelection(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		segs []*segment
		ary  []any
		exp  []any
	}{
		{
			name: "root",
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x", true, "y", []any{1, 2}},
		},
		{
			name: "index_zero",
			segs: []*segment{child(spec.Index(0))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x"},
		},
		{
			name: "index_one",
			segs: []*segment{child(spec.Index(1))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{nil, true},
		},
		{
			name: "index_three",
			segs: []*segment{child(spec.Index(3))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{nil, nil, nil, []any{1, 2}},
		},
		{
			name: "multiple_indexes",
			segs: []*segment{child(spec.Index(1), spec.Index(3))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{nil, true, nil, []any{1, 2}},
		},
		{
			name: "nested_indices",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Index(0)))},
			ary:  []any{[]any{1, 2}, "x", true, "y"},
			exp:  []any{[]any{1}},
		},
		{
			name: "nested_multiple_indices",
			segs: []*segment{child(spec.Index(0)).Append(
				child(spec.Index(0)), child(spec.Index(1)),
			)},
			ary: []any{[]any{1, 2}, "x", true, "y"},
			exp: []any{[]any{1, 2}},
		},
		{
			name: "nested_index_gaps",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Index(1)))},
			ary:  []any{"x", []any{1, 2}, true, "y"},
			exp:  []any{nil, []any{nil, 2}},
		},
		{
			name: "three_level_index_path",
			segs: []*segment{child(spec.Index(0)).Append(
				child(spec.Index(0)).Append(child(spec.Index(0))),
			)},
			ary: []any{[]any{[]any{42, 12}, 2}, "x", true, "y"},
			exp: []any{[]any{[]any{42}}},
		},
		{
			name: "varying_nesting_levels_mixed",
			segs: []*segment{
				child(spec.Index(0)).Append(
					child(spec.Index(0)).Append(child(spec.Index(0))),
				),
				child(spec.Index(1)),
				child(spec.Index(3)).Append(
					child(spec.Name("y")),
					child(spec.Name("z")),
				),
			},
			ary: []any{
				[]any{[]any{42, 12}, 2},
				"x",
				true,
				map[string]any{"y": "hi", "z": 1, "x": "no"},
			},
			exp: []any{
				[]any{[]any{42}},
				"x",
				nil,
				map[string]any{"y": "hi", "z": 1},
			},
		},
		{
			name: "filter_exists",
			segs: []*segment{child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Paren(spec.LogicalOr{spec.LogicalAnd{
					spec.Existence(spec.Query(true, []*spec.Segment{})),
				}}),
			}}))},
			ary: []any{1, 3},
			exp: []any{1, 3},
		},
		{
			name: "filter_compare",
			segs: []*segment{child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Comparison(
					spec.SingularQuery(false, []spec.Selector{}),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			}}))},
			ary: []any{1, 64, 42, 2},
			exp: []any{nil, 64, 42},
		},
		{
			name: "key_then_filter_cur_true",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(false, []*spec.Segment{spec.Child(spec.Index(1))})),
			}})))},
			ary: []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			exp: []any{nil, []any{nil, []any{99, 3}}},
		},
		{
			name: "key_then_filter_cur_false",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(false, []*spec.Segment{spec.Child(spec.Index(2))})),
			}})))},
			ary: []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			exp: []any{},
		},
		{
			name: "key_then_filter_root_true",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(true, []*spec.Segment{spec.Child(spec.Index(2))})),
			}})))},
			ary: []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			exp: []any{nil, []any{42, []any{99, 3}}},
		},
		{
			name: "key_then_filter_root_false",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.LogicalOr{spec.LogicalAnd{
				spec.Existence(spec.Query(true, []*spec.Segment{spec.Child(spec.Index(3))})),
			}})))},
			ary: []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			exp: []any{},
		},
		{
			name: "wildcard_indexes_index",
			segs: []*segment{
				child(spec.Wildcard).Append(
					child(spec.Index(0)),
					child(spec.Index(2)),
				),
			},
			ary: []any{[]any{1, 2, 3}, []any{3, 2, 1}, []any{4, 5, 6}},
			exp: []any{[]any{1, nil, 3}, []any{3, nil, 1}, []any{4, nil, 6}},
		},
		{
			name: "nonexistent_index",
			segs: []*segment{child(spec.Index(3))},
			ary:  []any{"y", []any{1, 2}},
			exp:  []any{},
		},
		{
			name: "nonexistent_branch_index",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Index(3)))},
			ary:  []any{[]any{0, 1, 2, 3}, []any{0, 1, 2}},
			exp:  []any{[]any{nil, nil, nil, 3}},
		},
		{
			name: "not_an_array_index_1",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "not_an_array_index_0",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "wildcard_not_an_array_index_1",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "mix_wildcard_keys",
			segs: []*segment{
				child(spec.Wildcard).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi", "y": "go"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true, "y": 21},
			},
			exp: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true},
			},
		},
		{
			name: "mix_wildcard_nonexistent_key",
			segs: []*segment{
				child(spec.Wildcard).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
			exp: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
		},
		{
			name: "mix_wildcard_index",
			segs: []*segment{
				child(spec.Wildcard).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(1))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{
				[]any{"x"},
				[]any{"x", "bo"},
				[]any{"x"},
			},
		},
		{
			name: "mix_wildcard_nonexistent_index",
			segs: []*segment{
				child(spec.Wildcard).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(3))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{
				[]any{"x"},
				[]any{"x"},
				[]any{"x"},
			},
		},
		{
			name: "wildcard_nonexistent_key",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{map[string]any{"a": 1}},
		},
		{
			name: "wildcard_nonexistent_middle_key",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{
				map[string]any{"a": 1},
				nil,
				map[string]any{"a": 5},
			},
		},
		{
			name: "wildcard_nested_nonexistent_key",
			segs: []*segment{child(spec.Wildcard).Append(
				child(spec.Wildcard).Append(child(spec.Name("a"))),
			)},
			ary: []any{
				map[string]any{
					"x": map[string]any{"a": 1},
					"y": map[string]any{"b": 1},
				},
				map[string]any{
					"y": map[string]any{"b": 1},
				},
			},
			exp: []any{map[string]any{"x": map[string]any{"a": 1}}},
		},
		{
			name: "wildcard_nested_nonexistent_index",
			segs: []*segment{child(spec.Wildcard).Append(
				child(spec.Wildcard).Append(child(spec.Index(1))),
			)},
			ary: []any{
				map[string]any{
					"x": []any{1, 2},
					"y": []any{3},
				},
				map[string]any{
					"z": []any{1},
				},
			},
			exp: []any{map[string]any{"x": []any{nil, 2}}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tree := Tree{child().Append(tc.segs...)}
			a.Equal(tc.exp, tree.Select(tc.ary))
		})
	}

	for _, tc := range []struct {
		name string
		segs []*segment
		src  []any
		dst  []any
		err  string
	}{
		{
			name: "dest_not_an_array",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Index(1)))},
			src:  []any{[]any{}},
			dst:  []any{"x", []any{1}},
			err:  `jsontree: expected destination array but got string`,
		},
		{
			name: "dest_not_an_object",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Name("x")))},
			src:  []any{map[string]any{"x": 1}},
			dst:  []any{[]any{1}},
			err:  `jsontree: expected destination object but got []interface {}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// In general a value in dst should only be a slice because we
			// sanitize the segments in advance, but this check ensures it at
			// runtime.
			tree := &Tree{}
			a.PanicsWithValue(tc.err, func() {
				tree.selectArraySegment(&segment{children: tc.segs}, nil, tc.src, tc.dst)
			})
		})
	}
}

func TestSliceSelection(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		segs []*segment
		ary  []any
		exp  []any
	}{
		{
			name: "slice_0_2",
			segs: []*segment{child(spec.Slice(0, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x", true},
		},
		{
			name: "slice_0_1",
			segs: []*segment{child(spec.Slice(0, 1))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x"},
		},
		{
			name: "slice_2_5",
			segs: []*segment{child(spec.Slice(2, 5))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{nil, nil, "y", []any{1, 2}, 42},
		},
		{
			name: "slice_2_5_over_len",
			segs: []*segment{child(spec.Slice(2, 5))},
			ary:  []any{"x", true, "y"},
			exp:  []any{nil, nil, "y"},
		},
		{
			name: "slice_defaults",
			segs: []*segment{child(spec.Slice())},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
		},
		{
			name: "default_start",
			segs: []*segment{child(spec.Slice(nil, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true},
		},
		{
			name: "default_end",
			segs: []*segment{child(spec.Slice(2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{nil, nil, "y", []any{1, 2}, 42, nil, 78},
		},
		{
			name: "step_2",
			segs: []*segment{child(spec.Slice(nil, nil, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, "y", nil, 42, nil, 78},
		},
		{
			name: "step_3",
			segs: []*segment{child(spec.Slice(nil, nil, 3))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, nil, []any{1, 2}, nil, nil, 78},
		},
		{
			name: "multiple_slices",
			segs: []*segment{child(spec.Slice(0, 1), spec.Slice(3, 4))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, nil, []any{1, 2}},
		},
		{
			name: "overlapping_slices",
			segs: []*segment{child(spec.Slice(0, 3), spec.Slice(2, 4))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true, "y", []any{1, 2}},
		},
		{
			name: "nested_slices",
			segs: []*segment{child(spec.Slice(0, 2)).Append(child(spec.Slice(1, 2)))},
			ary: []any{
				[]any{"hi", 42, true},
				[]any{"go", "on"},
				[]any{"yo", 98.6, false},
				"x", true, "y",
			},
			exp: []any{
				[]any{nil, 42},
				[]any{nil, "on"},
			},
		},
		{
			name: "nested_multiple_indices",
			segs: []*segment{child(spec.Slice(0, 2)).Append(
				child(spec.Slice(1, 2)),
				child(spec.Slice(3, 5)),
			)},
			ary: []any{
				[]any{"hi", 42, true, 64, []any{}, 7},
				[]any{"go", "on", false, 88, []any{1}, 8},
				[]any{"yo", 98.6, false, 2, []any{3, 4}, 9},
				"x", true, "y",
			},
			exp: []any{
				[]any{nil, 42, nil, 64, []any{}},
				[]any{nil, "on", nil, 88, []any{1}},
			},
		},
		{
			name: "three_level_slice_path",
			segs: []*segment{child(spec.Slice(0, 2)).Append(
				child(spec.Slice(0, 1)).Append(child(spec.Slice(0, 1))),
			)},
			ary: []any{
				[]any{[]any{42, 12}, 2},
				[]any{[]any{16, true, "x"}, 7},
				"x", true, "y",
			},
			exp: []any{
				[]any{[]any{42}},
				[]any{[]any{16}},
			},
		},
		{
			name: "varying_nesting_levels_mixed",
			segs: []*segment{
				child(spec.Slice(0, 2)).Append(
					child(spec.Slice(0, 1)).Append(child(spec.Slice(0, 1))),
				),
				child(spec.Slice(2, 3)),
				child(spec.Slice(3, 4)).Append(
					child(spec.Name("y")), child(spec.Name("z")),
				),
			},
			ary: []any{
				[]any{[]any{42, 12}, 2},
				"x",
				true,
				map[string]any{"y": "hi", "z": 1, "x": "no"},
				"go",
			},
			exp: []any{
				[]any{[]any{42}},
				nil,
				true,
				map[string]any{"y": "hi", "z": 1},
			},
		},
		{
			name: "wildcard_slices_index",
			segs: []*segment{child(spec.Wildcard).Append(
				child(spec.Slice(0, 2)),
				child(spec.Slice(3, 4)),
			)},
			ary: []any{
				[]any{1, 2, 3, 4, 5},
				[]any{3, 2, 1, 0, -1},
				[]any{4, 5, 6, 7, 8},
			},
			exp: []any{
				[]any{1, 2, nil, 4},
				[]any{3, 2, nil, 0},
				[]any{4, 5, nil, 7},
			},
		},
		{
			name: "nonexistent_slice",
			segs: []*segment{child(spec.Slice(3, 5))},
			ary:  []any{"y", []any{1, 2}},
			exp:  []any{},
		},
		{
			name: "nonexistent_branch_index",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Slice(3, 5)))},
			ary:  []any{[]any{0, 1, 2, 3, 4}, []any{0, 1, 2}},
			exp:  []any{[]any{nil, nil, nil, 3, 4}},
		},
		{
			name: "not_an_array_index_1",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "not_an_array",
			segs: []*segment{child(spec.Slice(0, 5)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "wildcard_not_an_array_index_1",
			segs: []*segment{child(spec.Wildcard).Append(child(spec.Slice(0, 5)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			name: "mix_slice_keys",
			segs: []*segment{
				child(spec.Slice(0, 5)).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi", "y": "go"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true, "y": 21},
			},
			exp: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true},
			},
		},
		{
			name: "mix_slice_nonexistent_key",
			segs: []*segment{
				child(spec.Slice(0, 5)).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
			exp: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
		},
		{
			name: "mix_slice_index",
			segs: []*segment{
				child(spec.Slice(0, 5)).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(1))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{
				[]any{"x"},
				[]any{"x", "bo"},
				[]any{"x"},
			},
		},
		{
			name: "mix_slice_nonexistent_index",
			segs: []*segment{
				child(spec.Slice(0, 5)).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(3))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{
				[]any{"x"},
				[]any{"x"},
				[]any{"x"},
			},
		},
		{
			name: "slice_nonexistent_key",
			segs: []*segment{child(spec.Slice(0, 5)).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{map[string]any{"a": 1}},
		},
		{
			name: "slice_nonexistent_middle_key",
			segs: []*segment{child(spec.Slice(0, 5)).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{
				map[string]any{"a": 1},
				nil,
				map[string]any{"a": 5},
			},
		},
		{
			name: "slice_nested_nonexistent_key",
			segs: []*segment{child(spec.Slice(0, 5)).Append(
				child(spec.Wildcard).Append(child(spec.Name("a"))),
			)},
			ary: []any{
				map[string]any{
					"x": map[string]any{"a": 1},
					"y": map[string]any{"b": 1},
				},
				map[string]any{
					"y": map[string]any{"b": 1},
				},
			},
			exp: []any{map[string]any{"x": map[string]any{"a": 1}}},
		},
		{
			name: "slice_nested_nonexistent_index",
			segs: []*segment{child(spec.Slice(0, 5)).Append(
				child(spec.Wildcard).Append(child(spec.Index(1))),
			)},
			ary: []any{
				map[string]any{
					"x": []any{1, 2},
					"y": []any{3},
				},
				map[string]any{
					"z": []any{1},
				},
			},
			exp: []any{map[string]any{"x": []any{nil, 2}}},
		},
		{
			name: "slice_neg",
			segs: []*segment{child(spec.Slice(nil, nil, -1))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x", true, "y", []any{1, 2}},
		},
		{
			name: "slice_5_0_neg2",
			segs: []*segment{child(spec.Slice(5, 0, -2))},
			ary:  []any{"x", true, "y", 8, 13, 25, 23, 78, 13},
			exp:  []any{nil, true, nil, 8, nil, 25},
		},
		{
			name: "nested_neg_slices",
			segs: []*segment{child(spec.Slice(2, nil, -1)).Append(child(spec.Slice(2, 0, -1)))},
			ary: []any{
				[]any{"hi", 42, true},
				[]any{"go", "on"},
				[]any{"yo", 98.6, false},
				"x", true, "y",
			},
			exp: []any{
				[]any{nil, 42, true},
				[]any{nil, "on"},
				[]any{nil, 98.6, false},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tree := Tree{child().Append(tc.segs...)}
			a.Equal(tc.exp, tree.Select(tc.ary))
		})
	}
}

func TestDescendants(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	json := map[string]any{
		"o": map[string]any{"j": 1, "k": 2},
		"a": []any{5, 3, []any{map[string]any{"j": 4}, map[string]any{"k": 6}}},
	}

	for _, tc := range []struct {
		name  string
		segs  []*segment
		input any
		exp   any
	}{
		{
			name:  "descendant_name",
			segs:  []*segment{descendant(spec.Name("j"))},
			input: json,
			exp: map[string]any{
				"o": map[string]any{"j": 1},
				"a": []any{nil, nil, []any{map[string]any{"j": 4}}},
			},
		},
		{
			name:  "un_descendant_name",
			segs:  []*segment{descendant(spec.Name("o"))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"j": 1, "k": 2}},
		},
		{
			name:  "nested_name",
			segs:  []*segment{child(spec.Name("o")).Append(descendant(spec.Name("k")))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"k": 2}},
		},
		{
			name:  "nested_wildcard",
			segs:  []*segment{child(spec.Name("o")).Append(descendant(spec.Wildcard))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"j": 1, "k": 2}},
		},
		{
			name:  "single_index",
			segs:  []*segment{descendant(spec.Index(0))},
			input: json,
			exp:   map[string]any{"a": []any{5, nil, []any{map[string]any{"j": 4}}}},
		},
		{
			name:  "nested_index",
			segs:  []*segment{child(spec.Name("a")).Append(descendant(spec.Index(0)))},
			input: json,
			exp:   map[string]any{"a": []any{5, nil, []any{map[string]any{"j": 4}}}},
		},
		{
			name: "multiples",
			segs: []*segment{
				child(spec.Name("profile")).Append(
					descendant(spec.Name("last")),
					descendant(spec.Name("contacts")).Append(
						child(spec.Name("primary")),
						child(spec.Name("secondary")),
					),
				),
			},
			input: map[string]any{
				"profile": map[string]any{
					"name": map[string]any{
						"first": "Barrack",
						"last":  "Obama",
					},
					"contacts": map[string]any{
						"email": map[string]any{
							"primary":   "foo@example.com",
							"secondary": "2nd@example.net",
						},
						"phones": map[string]any{
							"primary":   "123456789",
							"secondary": "987654321",
							"fax":       "1029384758",
						},
						"addresses": map[string]any{
							"primary": []any{
								"123 Main Street",
								"Whatever", "OR", "98754",
							},
							"work": []any{
								"whatever",
								"XYZ", "NY", "10093",
							},
						},
					},
				},
			},
			exp: map[string]any{
				"profile": map[string]any{
					"name": map[string]any{
						"last": "Obama",
					},
					"contacts": map[string]any{
						"email": map[string]any{
							"primary":   "foo@example.com",
							"secondary": "2nd@example.net",
						},
						"phones": map[string]any{
							"primary":   "123456789",
							"secondary": "987654321",
						},
						"addresses": map[string]any{
							"primary": []any{
								"123 Main Street",
								"Whatever", "OR", "98754",
							},
						},
					},
				},
			},
		},
		{
			name:  "do_not_include_parent_key",
			segs:  []*segment{descendant(spec.Name("o")).Append(child(spec.Name("k")))},
			input: map[string]any{"o": map[string]any{"o": "hi", "k": 2}},
			exp:   map[string]any{"o": map[string]any{"k": 2}},
		},
		{
			name:  "do_not_include_parent_index",
			segs:  []*segment{descendant(spec.Index(0)).Append(child(spec.Index(1)))},
			input: []any{[]any{42, 98}},
			exp:   []any{[]any{nil, 98}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tree := Tree{child().Append(tc.segs...)}
			a.Equal(tc.exp, tree.Select(tc.input))
		})
	}
}

func TestFilterSelection(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		path   string
		input  any
		output any
	}{
		{
			name:   "root_exists",
			path:   "$[?$]",
			input:  []any{1, 2},
			output: []any{1, 2},
		},
		{
			name:   "current_exists",
			path:   "$[?@]",
			input:  []any{1, 2},
			output: []any{1, 2},
		},
		{
			name:   "current_gt_1",
			path:   "$[? @ > 1]",
			input:  []any{nil, 2},
			output: []any{nil, 2},
		},
		{
			name:   "current_lt_2",
			path:   "$[? @ < 2]",
			input:  []any{1, 2},
			output: []any{1},
		},
		{
			name:   "current_gt_2",
			path:   "$[? @ > 2]",
			input:  []any{1, 2},
			output: []any{},
		},
		{
			name:   "obj_current_gt_1",
			path:   "$[? @ > 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"y": 2},
		},
		{
			name:   "obj_current_eq_1",
			path:   "$[? @ == 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1},
		},
		{
			name:   "obj_root_exists",
			path:   "$[? $]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1, "y": 2},
		},
		{
			name:   "obj_current_eq_1",
			path:   "$[? @ == 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1},
		},
		{
			name: "obj_current_key_gt_name",
			path: "$[? @.n > 12].name",
			input: map[string]any{
				"x": map[string]any{"n": 42, "name": "one"},
				"y": 2,
				"z": map[string]any{"n": 12, "name": "one"},
			},
			output: map[string]any{"x": map[string]any{"name": "one"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := jsonpath.MustParse(tc.path)
			segs := make([]*segment, len(path.Query().Segments()))
			for i, s := range path.Query().Segments() {
				segs[i] = child(s.Selectors()...)
				segs[i].descendant = s.IsDescendant()
				if i > 0 {
					segs[i-1].Append(segs[i])
				}
			}
			tree := Tree{child().Append(segs[0])}
			a.Equal(tc.output, tree.Select(tc.input))
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		paths []string
		exp   *Tree
	}{
		{
			name:  "root_only",
			paths: []string{"$"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "two_root_only",
			paths: []string{"$", "$"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "one_name",
			paths: []string{"$.a"},
			exp:   &Tree{root: child().Append(child(spec.Name("a")))},
		},
		{
			name:  "two_names",
			paths: []string{"$.a.b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "two_names_index",
			paths: []string{"$.a.b[1]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Index(1)),
						),
					),
				),
			},
		},
		{
			name:  "two_names_descendant",
			paths: []string{"$.a..b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						descendant(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "dup_two_names_descendant",
			paths: []string{"$.a..b", "$.a..b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						descendant(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "merge_descendant",
			paths: []string{"$.a..b", "$.a.b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						descendant(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "merge_descendant_children",
			paths: []string{"$.a..b.c", "$.a.b.c"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						descendant(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
			},
		},
		{
			name:  "two_single_key_paths",
			paths: []string{"$.a", "$.b"},
			exp: &Tree{
				root: child().Append(child(spec.Name("a"), spec.Name("b"))),
			},
		},
		{
			name:  "two_identical_paths",
			paths: []string{"$.a.b", "$.a.b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "diff_parents_same_child",
			paths: []string{"$.a.x", "$.b.x"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a"), spec.Name("b")).Append(
						child(spec.Name("x")),
					),
				),
			},
		},
		{
			name:  "diff_parents_diff_children",
			paths: []string{"$.a.x", "$.b.y"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("x")),
					),
					child(spec.Name("b")).Append(
						child(spec.Name("y")),
					),
				),
			},
		},
		{
			name:  "same_parent_different_child",
			paths: []string{"$.a.x", "$.a.y"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("x"), spec.Name("y")),
					),
				),
			},
		},
		{
			name:  "deeply_nested_same_from_diff_parent",
			paths: []string{"$.a.b.c.d", "$.a.x.c.d"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b"), spec.Name("x")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d")),
							),
						),
					),
				),
			},
		},
		{
			name:  "uneven_mixed_nested",
			paths: []string{"$.a.b.c.d", "$.a.x.c.d.e"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d")),
							),
						),
						child(spec.Name("x")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d")).Append(
									child(spec.Name("e")),
								),
							),
						),
					),
				),
			},
		},
		{
			name:  "different_leaves",
			paths: []string{"$.a.b.c.d", "$.a.x.c.e"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d")),
							),
						),
						child(spec.Name("x")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("e")),
							),
						),
					),
				),
			},
		},
		{
			name:  "split_later",
			paths: []string{"$.a.b.c.d.f", "$.a.b.c.e.g"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d")).Append(
									child(spec.Name("f")),
								),
								child(spec.Name("e")).Append(
									child(spec.Name("g")),
								),
							),
						),
					),
				),
			},
		},
		{
			name:  "four_identical_paths",
			paths: []string{"$.a.b", "$.a.b", "$.a.b", "$.a.b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
				),
			},
		},
		{
			name:  "same_diff_same",
			paths: []string{"$.a.x.b", "$.a.y.b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("x"), spec.Name("y")).Append(
							child(spec.Name("b")),
						),
					),
				),
			},
		},
		{
			name:  "same_diff_diff",
			paths: []string{"$.a.x.b", "$.a.y.c"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("x")).Append(
							child(spec.Name("b")),
						),
						child(spec.Name("y")).Append(
							child(spec.Name("c")),
						),
					),
				),
			},
		},
		{
			name:  "dupe_two_names_index",
			paths: []string{"$.a.b[1]", "$.a.b[1]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Index(1)),
						),
					),
				),
			},
		},
		{
			name:  "diff_indexes",
			paths: []string{"$[0]", "$[1]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0), spec.Index(1)),
				),
			},
		},
		{
			name:  "diff_sub_indexes",
			paths: []string{"$[0][0]", "$[0][1]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0)).Append(
						child(spec.Index(0), spec.Index(1)),
					),
				),
			},
		},
		{
			name:  "diff_index_name_key",
			paths: []string{"$[0][0].a", "$[0][1].b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0)).Append(
						child(spec.Index(0)).Append(
							child(spec.Name("a")),
						),
						child(spec.Index(1)).Append(
							child(spec.Name("b")),
						),
					),
				),
			},
		},
		{
			name:  "same_same_idx_diff_key",
			paths: []string{"$[0][0].a", "$[0][0].b"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0)).Append(
						child(spec.Index(0)).Append(
							child(spec.Name("a"), spec.Name("b")),
						),
					),
				),
			},
		},
		{
			name:  "same_diff_idx_diff_child",
			paths: []string{"$[0][0].a", "$[0][1].a"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0)).Append(
						child(spec.Index(0), spec.Index(1)).Append(
							child(spec.Name("a")),
						),
					),
				),
			},
		},
		{
			name:  "triple_same_diff_idx_diff_child",
			paths: []string{"$[0][0].a", "$[0][1].a", "$[0][3].a"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0)).Append(
						child(spec.Index(0), spec.Index(1), spec.Index(3)).Append(
							child(spec.Name("a")),
						),
					),
				),
			},
		},
		{
			name:  "wildcard",
			paths: []string{"$.*", "$.*"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "wildcard_seg",
			paths: []string{"$.*", "$[*]"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "wildcard_trumps_all",
			paths: []string{`$["x", 4, *]`, "$[*, 1]"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "wildcard_trumps_all_inverse",
			paths: []string{"$[1, *]", `$["x", 4, *]`},
			exp:   &Tree{root: child()},
		},
		{
			name:  "drop_trailing_wildcard",
			paths: []string{"$.a.*", "$.a.*"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")),
			)},
		},
		{
			name:  "drop_trailing_wildcard_diff_key",
			paths: []string{"$.a.*", "$.b.*"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a"), spec.Name("b")),
			)},
		},
		{
			name:  "wildcard_then_a",
			paths: []string{"$[1, *].a", `$["x", 4, *].a`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a")),
				),
			)},
		},
		{
			name:  "wildcard_then_a_and_b",
			paths: []string{"$[1, *].a", `$["x", 4, *].b`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a"), spec.Name("b")),
				),
			)},
		},
		{
			name:  "wildcard_then_diff_then_same",
			paths: []string{"$.*.a.c", `$.*.b.c`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a"), spec.Name("b")).Append(
						child(spec.Name("c")),
					),
				),
			)},
		},
		{
			name:  "wildcard_then_diff_then_same_deep",
			paths: []string{"$.*.a.c.d", `$.*.b.c.d`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a"), spec.Name("b")).Append(
						child(spec.Name("c")).Append(
							child(spec.Name("d")),
						),
					),
				),
			)},
		},
		{
			name:  "wildcard_then_divergent_paths",
			paths: []string{"$.*.a.b", `$.*.x.y`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")),
					),
					child(spec.Name("x")).Append(
						child(spec.Name("y")),
					),
				),
			)},
		},
		{
			name:  "wildcard_and_descendant_wildcard",
			paths: []string{"$.*", "$..*"},
			exp:   &Tree{root: child()},
		},
		{
			name:  "wildcard_and_descendant_wildcard_same_child",
			paths: []string{"$.*.a", "$..*.a"},
			exp: &Tree{root: child().Append(
				descendant(spec.Wildcard).Append(
					child(spec.Name("a")),
				),
			)},
		},
		{
			name:  "wildcard_and_descendant_wildcard_diff_child",
			paths: []string{"$.*.a", "$..*.b"},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard).Append(
					child(spec.Name("a")),
				),
				descendant(spec.Wildcard).Append(
					child(spec.Name("b")),
				),
			)},
		},
		{
			name:  "merge_complementary",
			paths: []string{"$.a.x.b", "$.a.y.c", "$.a.x.c", "$.a.y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					child(spec.Name("x"), spec.Name("y")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
				),
			)},
		},
		{
			name:  "merge_complementary_desc",
			paths: []string{"$.a..x.b", "$.a..y.c", "$.a..x.c", "$.a..y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					descendant(spec.Name("x"), spec.Name("y")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
				),
			)},
		},
		{
			name:  "merge_complementary_rev_desc",
			paths: []string{"$.a.x.b", "$.a.y.b", "$.a..x.b", "$.a..y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					descendant(spec.Name("x"), spec.Name("y")).Append(
						child(spec.Name("b")),
					),
				),
			)},
		},
		{
			name:  "do_not_merge_complementary_mixed_desc",
			paths: []string{"$.a..x.b", "$.a..y.c", "$.a..x.c", "$.a.y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					descendant(spec.Name("x")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
					descendant(spec.Name("y")).Append(
						child(spec.Name("c")),
					),
					child(spec.Name("y")).Append(
						child(spec.Name("b")),
					),
				),
			)},
		},
		{
			name:  "do_not_merge_descendant",
			paths: []string{"$.a.x.b", "$.a.y.c", "$.a.x.c", "$.a..y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					child(spec.Name("x")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
					child(spec.Name("y")).Append(
						child(spec.Name("c")),
					),
					descendant(spec.Name("y")).Append(
						child(spec.Name("b")),
					),
				),
			)},
		},
		{
			name:  "do_not_merge_top_descendant",
			paths: []string{"$..a.y.c", "$.a.y.b"},
			exp: &Tree{root: child().Append(
				descendant(spec.Name("a")).Append(
					child(spec.Name("y")).Append(
						child(spec.Name("c")),
					),
				),
				child(spec.Name("a")).Append(
					child(spec.Name("y")).Append(
						child(spec.Name("b")),
					),
				),
			)},
		},
		{
			name:  "do_not_merge_top_descendant_multi",
			paths: []string{"$.a.x.b", "$..a.y.c", "$.a.x.c", "$.a.y.b"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					child(spec.Name("x")).Append(
						child(spec.Name("b"), spec.Name("c")),
					),
					child(spec.Name("y")).Append(
						child(spec.Name("b")),
					),
				),
				descendant(spec.Name("a")).Append(
					child(spec.Name("y")).Append(
						child(spec.Name("c")),
					),
				),
			)},
		},
		{
			name:  "merge_same_branch",
			paths: []string{"$.a.b.c", "$.d", "$.a..x.c"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					child(spec.Name("b")).Append(
						child(spec.Name("c")),
					),
					descendant(spec.Name("x")).Append(
						child(spec.Name("c")),
					),
				),
				child(spec.Name("d")),
			)},
		},
		{
			name:  "skip_common_branch",
			paths: []string{"$.a.b.c", "$.d", "$.a.x.c"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")).Append(
					child(spec.Name("b"), spec.Name("x")).Append(
						child(spec.Name("c")),
					),
				),
				child(spec.Name("d")),
			)},
		},
		{
			name:  "merge_index_selectors",
			paths: []string{"$[1,2,1,2,3]"},
			exp: &Tree{root: child().Append(
				child(spec.Index(1), spec.Index(2), spec.Index(3)),
			)},
		},
		{
			name:  "merge_name_selectors",
			paths: []string{`$["x", "y", "x", "r", "y"]`},
			exp: &Tree{root: child().Append(
				child(spec.Name("x"), spec.Name("y"), spec.Name("r")),
			)},
		},
		{
			name:  "merge_slice_selector",
			paths: []string{`$["x", 1, "x", 1, 2, 2:]`},
			exp: &Tree{root: child().Append(
				child(spec.Slice(2), spec.Name("x"), spec.Index(1)),
			)},
		},
		{
			name:  "merge_mixed_multi_path",
			paths: []string{`$["x", 1, "x", 1, 2, 2:]`, `$["x", 2, "y"]`},
			exp: &Tree{root: child().Append(
				child(spec.Slice(2), spec.Name("x"), spec.Index(1), spec.Name("y")),
			)},
		},
		{
			name:  "merge_leaf_node",
			paths: []string{"$.a.b.c", "$.a.b.c.d.e"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
			},
		},
		{
			name:  "merge_leaf_wildcard_leaf_node",
			paths: []string{"$.a.b.c.*", "$.a.b.c.d.e"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
			},
		},
		{
			name:  "merge_reverse_leaf_node",
			paths: []string{"$.a.b.c.d.e", "$.a.b.c"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")),
						),
					),
				),
			},
		},
		{
			name:  "merge_deep_common_selectors",
			paths: []string{"$.a.b.c.d", `$.a.b.c["e", "f"]`},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("a")).Append(
						child(spec.Name("b")).Append(
							child(spec.Name("c")).Append(
								child(spec.Name("d"), spec.Name("e"), spec.Name("f")),
							),
						),
					),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			paths := make([]*jsonpath.Path, len(tc.paths))
			for i, p := range tc.paths {
				paths[i] = jsonpath.MustParse(p)
			}
			tree := New(paths...)
			a.Equal(tc.exp.String(), tree.String())
		})
	}
}

func TestSelectorsFor(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		seg    *spec.Segment
		expect []spec.Selector
		wild   bool
	}{
		{
			name:   "empty",
			seg:    spec.Child(),
			expect: []spec.Selector{},
			wild:   false,
		},
		{
			name:   "one_name",
			seg:    spec.Child(spec.Name("x")),
			expect: []spec.Selector{spec.Name("x")},
			wild:   false,
		},
		{
			name:   "wildcard",
			seg:    spec.Child(spec.Wildcard),
			expect: []spec.Selector{spec.Wildcard},
			wild:   true,
		},
		{
			name:   "mix_wildcard",
			seg:    spec.Child(spec.Name("x"), spec.Wildcard),
			expect: []spec.Selector{spec.Wildcard},
			wild:   true,
		},
		{
			name:   "wildcard_mix",
			seg:    spec.Child(spec.Wildcard, spec.Name("x"), spec.Index(1)),
			expect: []spec.Selector{spec.Wildcard},
			wild:   true,
		},
		{
			name: "mix_selectors_frst",
			seg: spec.Child(
				spec.Name("x"),
				spec.Index(1),
				spec.Slice(2, 6),
				mkFilter("$[?@]"),
			),
			expect: []spec.Selector{
				spec.Slice(2, 6),
				spec.Name("x"),
				spec.Index(1),
				mkFilter("$[?@]"),
			},
			wild: false,
		},
		{
			name: "slices_before_indexes",
			seg: spec.Child(
				spec.Slice(8),
				spec.Index(1),
				spec.Index(2),
				spec.Slice(5, 6),
				spec.Index(4),
				spec.Slice(6),
			),
			expect: []spec.Selector{
				spec.Slice(8),
				spec.Slice(5, 6),
				spec.Slice(6),
				spec.Index(1),
				spec.Index(2),
				spec.Index(4),
			},
			wild: false,
		},
		{
			name: "merge_indexes_into_slices",
			seg: spec.Child(
				spec.Slice(2, 3),
				spec.Index(1),
				spec.Index(2),
				spec.Slice(6, 8),
				spec.Index(7),
				spec.Index(6),
			),
			expect: []spec.Selector{
				spec.Slice(2, 3),
				spec.Slice(6, 8),
				spec.Index(1),
			},
			wild: false,
		},
		{
			name: "wildcard_trump_all",
			seg: spec.Child(
				spec.Slice(2, 3),
				spec.Index(1),
				spec.Index(2),
				spec.Slice(6, 8),
				spec.Index(7),
				spec.Index(6),
				spec.Wildcard,
			),
			expect: []spec.Selector{spec.Wildcard},
			wild:   true,
		},
		{
			name: "merge_dupes_slice_first",
			seg: spec.Child(
				spec.Name("x"),
				spec.Index(1),
				spec.Name("x"),
				spec.Index(2),
				spec.Index(3),
				spec.Slice(4),
			),
			expect: []spec.Selector{
				spec.Slice(4),
				spec.Name("x"),
				spec.Index(1),
				spec.Index(2),
				spec.Index(3),
			},
			wild: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			selectors, wild := selectorsFor(tc.seg)
			a.Equal(tc.expect, selectors)
			a.Equal(tc.wild, wild)
		})
	}
}
