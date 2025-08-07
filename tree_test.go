package jsontree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

func TestRunRoot(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			sel := &Tree{root: &segment{}}
			a.Equal(tc.val, sel.Select(tc.val))

			switch tc.val.(type) {
			case map[string]any, []any:
				return
			default:
				// Anything other than a slice or map returns nil if
				// there are path segments.
				sel.root = &segment{children: []*segment{child(spec.Wildcard())}}
				a.Nil(sel.Select(tc.val))
			}
		})
	}
}

func TestObjectSelection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		segs []*segment
		obj  map[string]any
		exp  map[string]any
	}{
		{
			test: "root",
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"x": true, "y": []any{1, 2}},
		},
		{
			test: "one_key_scalar",
			segs: []*segment{child(spec.Name("x"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"x": true},
		},
		{
			test: "one_key_array",
			segs: []*segment{child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}},
			exp:  map[string]any{"y": []any{1, 2}},
		},
		{
			test: "one_key_object",
			segs: []*segment{child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": map[string]any{"a": 1}},
			exp:  map[string]any{"y": map[string]any{"a": 1}},
		},
		{
			test: "filter_object",
			segs: []*segment{child(spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(false, spec.Name("a")),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			)))},
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
			test: "filter_object_key",
			segs: []*segment{child(spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(false, spec.Name("a")),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			))).Append(child(spec.Name("firm")))},
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
			test: "multiple_keys",
			segs: []*segment{child(spec.Name("x")), child(spec.Name("y"))},
			obj:  map[string]any{"x": true, "y": []any{1, 2}, "z": "hi"},
			exp:  map[string]any{"x": true, "y": []any{1, 2}},
		},
		{
			test: "key_and_filter",
			segs: []*segment{child(spec.Name("x")), child(spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(false, spec.Name("z")),
					spec.EqualTo,
					spec.Literal("hi"),
				),
			)))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{"x": true, "y": map[string]any{"z": "hi"}},
		},
		{
			test: "key_then_filter_cur_true",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(false, spec.Child(spec.Index(0)))),
			))))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": []any{1}}, "z": "hi"},
			exp: map[string]any{"y": map[string]any{"z": []any{1}}},
		},
		{
			test: "key_then_filter_cur_false",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(false, spec.Child(spec.Index(1)))),
			))))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": []any{1}}, "z": "hi"},
			exp: map[string]any{},
		},
		{
			test: "key_then_filter_root_true",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(true, spec.Child(spec.Name("x")))),
			))))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{"y": map[string]any{"z": "hi"}},
		},
		{
			test: "key_then_filter_root_false",
			segs: []*segment{child(spec.Name("y")).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(true, spec.Child(spec.Name("a")))),
			))))},
			obj: map[string]any{"x": true, "y": map[string]any{"z": "hi"}, "z": "hi"},
			exp: map[string]any{},
		},
		{
			test: "three_level_path",
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
			test: "nested_multiple_keys",
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
			test: "varying_nesting_levels",
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
			test: "wildcard_keys",
			segs: []*segment{
				child(spec.Wildcard()).Append(
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
			test: "any_key_indexes",
			segs: []*segment{
				child(spec.Wildcard()).Append(
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
			test: "any_key_nonexistent_index",
			segs: []*segment{child(spec.Wildcard()).Append(child(spec.Index(1)))},
			obj: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a"},
			},
			exp: map[string]any{"x": []any{nil, "go"}},
		},
		{
			test: "nonexistent_key",
			segs: []*segment{child(spec.Name("x"))},
			obj:  map[string]any{"y": []any{1, 2}},
			exp:  map[string]any{},
		},
		{
			test: "nonexistent_branch_key",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("z")))},
			obj:  map[string]any{"y": []any{1, 2}},
			exp:  map[string]any{},
		},
		{
			test: "wildcard_then_nonexistent_key",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("x")))},
			obj:  map[string]any{"y": map[string]any{"a": 1}},
			exp:  map[string]any{},
		},
		{
			test: "not_an_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			obj:  map[string]any{"x": true},
			exp:  map[string]any{},
		},
		{
			test: "not_an_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			obj:  map[string]any{"x": true},
			exp:  map[string]any{},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			tree := Tree{child().Append(tc.segs...), true}
			a.Equal(tc.exp, tree.Select(tc.obj))
			// Test non-indexing tree.
			if tc.test == "any_key_nonexistent_index" {
				tc.exp = map[string]any{"x": []any{"go"}}
			}

			tree = Tree{child().Append(tc.segs...), false}
			a.Equal(tc.exp, tree.Select(tc.obj))
		})
	}

	for _, tc := range []struct {
		test string
		segs []*segment
		src  map[string]any
		dst  map[string]any
		err  string
	}{
		{
			test: "dest_not_object",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Name("y")))},
			src:  map[string]any{"x": map[string]any{}},
			dst:  map[string]any{"x": []any{1}},
			err:  `jsontree: expected destination object but got []interface {}`,
		},
		{
			test: "dest_not_array",
			segs: []*segment{child(spec.Name("x")).Append(child(spec.Index(1)))},
			src:  map[string]any{"x": []any{}},
			dst:  map[string]any{"x": map[string]any{"x": 1}},
			err:  `jsontree: expected destination array but got map[string]interface {}`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			// In general a value in dst should only be a map because we sanitize
			// the segments in advance, but this check ensures it at runtime.
			tree := &Tree{}

			assert.PanicsWithValue(t, tc.err, func() {
				tree.selectObjectSegment(&segment{children: tc.segs}, nil, tc.src, tc.dst)
			})
		})
	}
}

func TestArraySelection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test     string
		segs     []*segment
		ary      []any
		indexed  []any
		appended []any
	}{
		{
			test:    "root",
			ary:     []any{"x", true, "y", []any{1, 2}},
			indexed: []any{"x", true, "y", []any{1, 2}},
		},
		{
			test:    "index_zero",
			segs:    []*segment{child(spec.Index(0))},
			ary:     []any{"x", true, "y", []any{1, 2}},
			indexed: []any{"x"},
		},
		{
			test:    "index_zero_null",
			segs:    []*segment{child(spec.Index(0))},
			ary:     []any{nil, true, "y", []any{1, 2}},
			indexed: []any{nil},
		},
		{
			test:     "index_one",
			segs:     []*segment{child(spec.Index(1))},
			ary:      []any{"x", true, "y", []any{1, 2}},
			indexed:  []any{nil, true},
			appended: []any{true},
		},
		{
			test:     "index_one_null",
			segs:     []*segment{child(spec.Index(1))},
			ary:      []any{"x", nil, "y", []any{1, 2}},
			indexed:  []any{nil, nil},
			appended: []any{nil},
		},
		{
			test:     "index_three",
			segs:     []*segment{child(spec.Index(3))},
			ary:      []any{"x", true, "y", []any{1, 2}},
			indexed:  []any{nil, nil, nil, []any{1, 2}},
			appended: []any{[]any{1, 2}},
		},
		{
			test:     "index_three_null",
			segs:     []*segment{child(spec.Index(3))},
			ary:      []any{"x", true, "y", nil},
			indexed:  []any{nil, nil, nil, nil},
			appended: []any{nil},
		},
		{
			test:     "multiple_indexes",
			segs:     []*segment{child(spec.Index(1), spec.Index(3))},
			ary:      []any{"x", true, "y", []any{1, 2}},
			indexed:  []any{nil, true, nil, []any{1, 2}},
			appended: []any{true, []any{1, 2}},
		},
		{
			test:     "multiple_null_indexes",
			segs:     []*segment{child(spec.Index(1), spec.Index(3))},
			ary:      []any{"x", nil, "y", nil},
			indexed:  []any{nil, nil, nil, nil},
			appended: []any{nil, nil},
		},
		{
			test:    "nested_index",
			segs:    []*segment{child(spec.Index(0)).Append(child(spec.Index(0)))},
			ary:     []any{[]any{1, 2}, "x", true, "y"},
			indexed: []any{[]any{1}},
		},
		{
			test:    "nested_index_null",
			segs:    []*segment{child(spec.Index(0)).Append(child(spec.Index(0)))},
			ary:     []any{[]any{nil, 2}, "x", true, "y"},
			indexed: []any{[]any{nil}},
		},
		{
			test: "nested_multiple_indices",
			segs: []*segment{child(spec.Index(0)).Append(
				child(spec.Index(0)), child(spec.Index(1)),
			)},
			ary:     []any{[]any{1, 2}, "x", true, "y"},
			indexed: []any{[]any{1, 2}},
		},
		{
			test: "nested_multiple_indices_null",
			segs: []*segment{child(spec.Index(0)).Append(
				child(spec.Index(0)), child(spec.Index(1)),
			)},
			ary:     []any{[]any{1, nil}, "x", true, "y"},
			indexed: []any{[]any{1, nil}},
		},
		{
			test:     "nested_index_gaps",
			segs:     []*segment{child(spec.Index(1)).Append(child(spec.Index(1)))},
			ary:      []any{"x", []any{1, 2}, true, "y"},
			indexed:  []any{nil, []any{nil, 2}},
			appended: []any{[]any{2}},
		},
		{
			test:     "nested_index_gaps_with_null",
			segs:     []*segment{child(spec.Index(1)).Append(child(spec.Index(1)))},
			ary:      []any{"x", []any{nil, nil}, true, "y"},
			indexed:  []any{nil, []any{nil, nil}},
			appended: []any{[]any{nil}},
		},
		{
			test: "three_level_index_path",
			segs: []*segment{child(spec.Index(0)).Append(
				child(spec.Index(0)).Append(child(spec.Index(0))),
			)},
			ary:     []any{[]any{[]any{42, 12}, 2}, "x", true, "y"},
			indexed: []any{[]any{[]any{42}}},
		},
		{
			test: "varying_nesting_levels_mixed",
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
			indexed: []any{
				[]any{[]any{42}},
				"x",
				nil,
				map[string]any{"y": "hi", "z": 1},
			},
			appended: []any{
				[]any{[]any{42}},
				"x",
				map[string]any{"y": "hi", "z": 1},
			},
		},
		{
			test: "filter_exists",
			segs: []*segment{child(spec.Filter(spec.And(
				spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
			)))},
			ary:     []any{1, 3},
			indexed: []any{1, 3},
		},
		{
			test: "filter_compare",
			segs: []*segment{child(spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(false),
					spec.GreaterThanEqualTo,
					spec.Literal(int64(42)),
				),
			)))},
			ary:      []any{1, 64, 42, 2},
			indexed:  []any{nil, 64, 42},
			appended: []any{64, 42},
		},
		{
			test: "key_then_filter_cur_true",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(false, spec.Child(spec.Index(1)))),
			))))},
			ary:      []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			indexed:  []any{nil, []any{nil, []any{99, 3}}},
			appended: []any{[]any{[]any{99, 3}}},
		},
		{
			test: "key_then_filter_cur_false",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(false, spec.Child(spec.Index(2)))),
			))))},
			ary:     []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			indexed: []any{},
		},
		{
			test: "key_then_filter_root_true",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(true, spec.Child(spec.Index(2)))),
			))))},
			ary:      []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			indexed:  []any{nil, []any{42, []any{99, 3}}},
			appended: []any{[]any{42, []any{99, 3}}},
		},
		{
			test: "key_then_filter_root_false",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Filter(spec.And(
				spec.Existence(spec.Query(true, spec.Child(spec.Index(3)))),
			))))},
			ary:     []any{[]any{1, 2}, []any{42, []any{99, 3}}, []any{4, 5}},
			indexed: []any{},
		},
		{
			test: "wildcard_indexes_index",
			segs: []*segment{
				child(spec.Wildcard()).Append(
					child(spec.Index(0)),
					child(spec.Index(2)),
				),
			},
			ary:      []any{[]any{1, 2, 3}, []any{3, 2, 1}, []any{4, 5, 6}},
			indexed:  []any{[]any{1, nil, 3}, []any{3, nil, 1}, []any{4, nil, 6}},
			appended: []any{[]any{1, 3}, []any{3, 1}, []any{4, 6}},
		},
		{
			test:    "nonexistent_index",
			segs:    []*segment{child(spec.Index(3))},
			ary:     []any{"y", []any{1, 2}},
			indexed: []any{},
		},
		{
			test:     "nonexistent_branch_index",
			segs:     []*segment{child(spec.Wildcard()).Append(child(spec.Index(3)))},
			ary:      []any{[]any{0, 1, 2, 3}, []any{0, 1, 2}},
			indexed:  []any{[]any{nil, nil, nil, 3}},
			appended: []any{[]any{3}},
		},
		{
			test:    "not_an_array_index_1",
			segs:    []*segment{child(spec.Index(1)).Append(child(spec.Index(0)))},
			ary:     []any{"x", true},
			indexed: []any{},
		},
		{
			test:    "not_an_array_index_0",
			segs:    []*segment{child(spec.Index(0)).Append(child(spec.Index(0)))},
			ary:     []any{"x", true},
			indexed: []any{},
		},
		{
			test:    "wildcard_not_an_array_index_1",
			segs:    []*segment{child(spec.Wildcard()).Append(child(spec.Index(0)))},
			ary:     []any{"x", true},
			indexed: []any{},
		},
		{
			test: "mix_wildcard_keys",
			segs: []*segment{
				child(spec.Wildcard()).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi", "y": "go"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true, "y": 21},
			},
			indexed: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true},
			},
		},
		{
			test: "mix_wildcard_nonexistent_key",
			segs: []*segment{
				child(spec.Wildcard()).Append(child(spec.Name("x"))),
				child(spec.Index(1)).Append(child(spec.Name("y"))),
			},
			ary: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
			indexed: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
		},
		{
			test: "mix_wildcard_index",
			segs: []*segment{
				child(spec.Wildcard()).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(1))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			indexed: []any{
				[]any{"x"},
				[]any{"x", "bo"},
				[]any{"x"},
			},
		},
		{
			test: "mix_wildcard_nonexistent_index",
			segs: []*segment{
				child(spec.Wildcard()).Append(child(spec.Index(0))),
				child(spec.Index(1)).Append(child(spec.Index(3))),
			},
			ary: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			indexed: []any{
				[]any{"x"},
				[]any{"x"},
				[]any{"x"},
			},
		},
		{
			test: "wildcard_nonexistent_key",
			segs: []*segment{child(spec.Wildcard()).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			indexed: []any{map[string]any{"a": 1}},
		},
		{
			test: "wildcard_nonexistent_middle_key",
			segs: []*segment{child(spec.Wildcard()).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			indexed: []any{
				map[string]any{"a": 1},
				nil,
				map[string]any{"a": 5},
			},
			appended: []any{
				map[string]any{"a": 1},
				map[string]any{"a": 5},
			},
		},
		{
			test: "wildcard_nested_nonexistent_key",
			segs: []*segment{child(spec.Wildcard()).Append(
				child(spec.Wildcard()).Append(child(spec.Name("a"))),
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
			indexed: []any{map[string]any{"x": map[string]any{"a": 1}}},
		},
		{
			test: "wildcard_nested_nonexistent_index",
			segs: []*segment{child(spec.Wildcard()).Append(
				child(spec.Wildcard()).Append(child(spec.Index(1))),
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
			indexed:  []any{map[string]any{"x": []any{nil, 2}}},
			appended: []any{map[string]any{"x": []any{2}}},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			tree := Tree{child().Append(tc.segs...), true}
			a.Equal(tc.indexed, tree.Select(tc.ary))
			tree.index = false

			if tc.appended == nil {
				tc.appended = tc.indexed
			}

			a.Equal(tc.appended, tree.Select(tc.ary))
		})
	}

	for _, tc := range []struct {
		test string
		segs []*segment
		src  []any
		dst  []any
		err  string
	}{
		{
			test: "dest_not_an_array",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Index(1)))},
			src:  []any{[]any{}},
			dst:  []any{"x", []any{1}},
			err:  `jsontree: expected destination array but got string`,
		},
		{
			test: "dest_not_an_object",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Name("x")))},
			src:  []any{map[string]any{"x": 1}},
			dst:  []any{[]any{1}},
			err:  `jsontree: expected destination object but got []interface {}`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			// In general a value in dst should only be a slice because we
			// sanitize the segments in advance, but this check ensures it at
			// runtime.
			tree := &Tree{}

			assert.PanicsWithValue(t, tc.err, func() {
				tree.selectArraySegment(&segment{children: tc.segs}, nil, tc.src, tc.dst)
			})
		})
	}
}

func TestSliceSelection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		segs []*segment
		ary  []any
		exp  []any
	}{
		{
			test: "slice_0_2",
			segs: []*segment{child(spec.Slice(0, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x", true},
		},
		{
			test: "slice_0_1",
			segs: []*segment{child(spec.Slice(0, 1))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x"},
		},
		{
			test: "slice_2_5",
			segs: []*segment{child(spec.Slice(2, 5))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{nil, nil, "y", []any{1, 2}, 42},
		},
		{
			test: "slice_2_5_over_len",
			segs: []*segment{child(spec.Slice(2, 5))},
			ary:  []any{"x", true, "y"},
			exp:  []any{nil, nil, "y"},
		},
		{
			test: "slice_defaults",
			segs: []*segment{child(spec.Slice())},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
		},
		{
			test: "default_start",
			segs: []*segment{child(spec.Slice(nil, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true},
		},
		{
			test: "default_end",
			segs: []*segment{child(spec.Slice(2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{nil, nil, "y", []any{1, 2}, 42, nil, 78},
		},
		{
			test: "step_2",
			segs: []*segment{child(spec.Slice(nil, nil, 2))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, "y", nil, 42, nil, 78},
		},
		{
			test: "step_3",
			segs: []*segment{child(spec.Slice(nil, nil, 3))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, nil, []any{1, 2}, nil, nil, 78},
		},
		{
			test: "multiple_slices",
			segs: []*segment{child(spec.Slice(0, 1), spec.Slice(3, 4))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", nil, nil, []any{1, 2}},
		},
		{
			test: "overlapping_slices",
			segs: []*segment{child(spec.Slice(0, 3), spec.Slice(2, 4))},
			ary:  []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:  []any{"x", true, "y", []any{1, 2}},
		},
		{
			test: "nested_slices",
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
			test: "nested_multiple_indices",
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
			test: "three_level_slice_path",
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
			test: "varying_nesting_levels_mixed",
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
			test: "wildcard_slices_index",
			segs: []*segment{child(spec.Wildcard()).Append(
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
			test: "nonexistent_slice",
			segs: []*segment{child(spec.Slice(3, 5))},
			ary:  []any{"y", []any{1, 2}},
			exp:  []any{},
		},
		{
			test: "nonexistent_branch_index",
			segs: []*segment{child(spec.Wildcard()).Append(child(spec.Slice(3, 5)))},
			ary:  []any{[]any{0, 1, 2, 3, 4}, []any{0, 1, 2}},
			exp:  []any{[]any{nil, nil, nil, 3, 4}},
		},
		{
			test: "not_an_array_index_1",
			segs: []*segment{child(spec.Index(1)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			test: "not_an_array",
			segs: []*segment{child(spec.Slice(0, 5)).Append(child(spec.Index(0)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			test: "wildcard_not_an_array_index_1",
			segs: []*segment{child(spec.Wildcard()).Append(child(spec.Slice(0, 5)))},
			ary:  []any{"x", true},
			exp:  []any{},
		},
		{
			test: "mix_slice_keys",
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
			test: "mix_slice_nonexistent_key",
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
			test: "mix_slice_index",
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
			test: "mix_slice_nonexistent_index",
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
			test: "slice_nonexistent_key",
			segs: []*segment{child(spec.Slice(0, 5)).Append(child(spec.Name("a")))},
			ary: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{map[string]any{"a": 1}},
		},
		{
			test: "slice_nonexistent_middle_key",
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
			test: "slice_nested_nonexistent_key",
			segs: []*segment{child(spec.Slice(0, 5)).Append(
				child(spec.Wildcard()).Append(child(spec.Name("a"))),
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
			test: "slice_nested_nonexistent_index",
			segs: []*segment{child(spec.Slice(0, 5)).Append(
				child(spec.Wildcard()).Append(child(spec.Index(1))),
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
			test: "slice_neg",
			segs: []*segment{child(spec.Slice(nil, nil, -1))},
			ary:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{"x", true, "y", []any{1, 2}},
		},
		{
			test: "slice_5_0_neg2",
			segs: []*segment{child(spec.Slice(5, 0, -2))},
			ary:  []any{"x", true, "y", 8, 13, 25, 23, 78, 13},
			exp:  []any{nil, true, nil, 8, nil, 25},
		},
		{
			test: "nested_neg_slices",
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			tree := Tree{child().Append(tc.segs...), true}
			assert.Equal(t, tc.exp, tree.Select(tc.ary))
		})
	}
}

func TestDescendants(t *testing.T) {
	t.Parallel()

	json := map[string]any{
		"o": map[string]any{"j": 1, "k": 2},
		"a": []any{5, 3, []any{map[string]any{"j": 4}, map[string]any{"k": 6}}},
	}

	for _, tc := range []struct {
		test  string
		segs  []*segment
		input any
		exp   any
	}{
		{
			test:  "descendant_name",
			segs:  []*segment{descendant(spec.Name("j"))},
			input: json,
			exp: map[string]any{
				"o": map[string]any{"j": 1},
				"a": []any{nil, nil, []any{map[string]any{"j": 4}}},
			},
		},
		{
			test:  "un_descendant_name",
			segs:  []*segment{descendant(spec.Name("o"))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"j": 1, "k": 2}},
		},
		{
			test:  "nested_name",
			segs:  []*segment{child(spec.Name("o")).Append(descendant(spec.Name("k")))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"k": 2}},
		},
		{
			test:  "nested_wildcard",
			segs:  []*segment{child(spec.Name("o")).Append(descendant(spec.Wildcard()))},
			input: json,
			exp:   map[string]any{"o": map[string]any{"j": 1, "k": 2}},
		},
		{
			test:  "single_index",
			segs:  []*segment{descendant(spec.Index(0))},
			input: json,
			exp:   map[string]any{"a": []any{5, nil, []any{map[string]any{"j": 4}}}},
		},
		{
			test:  "nested_index",
			segs:  []*segment{child(spec.Name("a")).Append(descendant(spec.Index(0)))},
			input: json,
			exp:   map[string]any{"a": []any{5, nil, []any{map[string]any{"j": 4}}}},
		},
		{
			test: "multiples",
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
			test:  "do_not_include_parent_key",
			segs:  []*segment{descendant(spec.Name("o")).Append(child(spec.Name("k")))},
			input: map[string]any{"o": map[string]any{"o": "hi", "k": 2}},
			exp:   map[string]any{"o": map[string]any{"k": 2}},
		},
		{
			test:  "do_not_include_parent_index",
			segs:  []*segment{descendant(spec.Index(0)).Append(child(spec.Index(1)))},
			input: []any{[]any{42, 98}},
			exp:   []any{[]any{nil, 98}},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			tree := Tree{child().Append(tc.segs...), true}
			assert.Equal(t, tc.exp, tree.Select(tc.input))
		})
	}
}

func TestFilterSelection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test   string
		path   string
		input  any
		output any
	}{
		{
			test:   "root_exists",
			path:   "$[?$]",
			input:  []any{1, 2},
			output: []any{1, 2},
		},
		{
			test:   "current_exists",
			path:   "$[?@]",
			input:  []any{1, 2},
			output: []any{1, 2},
		},
		{
			test:   "current_gt_1",
			path:   "$[? @ > 1]",
			input:  []any{nil, 2},
			output: []any{nil, 2},
		},
		{
			test:   "current_lt_2",
			path:   "$[? @ < 2]",
			input:  []any{1, 2},
			output: []any{1},
		},
		{
			test:   "current_gt_2",
			path:   "$[? @ > 2]",
			input:  []any{1, 2},
			output: []any{},
		},
		{
			test:   "obj_current_gt_1",
			path:   "$[? @ > 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"y": 2},
		},
		{
			test:   "obj_current_eq_1",
			path:   "$[? @ == 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1},
		},
		{
			test:   "obj_root_exists",
			path:   "$[? $]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1, "y": 2},
		},
		{
			test:   "obj_current_eq_1",
			path:   "$[? @ == 1]",
			input:  map[string]any{"x": 1, "y": 2},
			output: map[string]any{"x": 1},
		},
		{
			test: "obj_current_key_gt_name",
			path: "$[? @.n > 12].name",
			input: map[string]any{
				"x": map[string]any{"n": 42, "name": "one"},
				"y": 2,
				"z": map[string]any{"n": 12, "name": "one"},
			},
			output: map[string]any{"x": map[string]any{"name": "one"}},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
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

			tree := Tree{child().Append(segs[0]), true}
			assert.Equal(t, tc.output, tree.Select(tc.input))
		})
	}
}

func TestTreeString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		segs []*segment
		str  string
	}{
		{
			test: "root_only",
			str:  "$\n",
		},
		{
			test: "wildcard",
			segs: []*segment{child(spec.Wildcard())},
			str:  "$\n└── [*]\n",
		},
		{
			test: "one_key",
			segs: []*segment{child(spec.Name("foo"))},
			str:  "$\n└── [\"foo\"]\n",
		},
		{
			test: "two_keys",
			segs: []*segment{child(spec.Name("foo"), spec.Name("bar"))},
			str:  "$\n└── [\"foo\",\"bar\"]\n",
		},
		{
			test: "two_segments",
			segs: []*segment{child(spec.Name("foo")), child(spec.Name("bar"))},
			str:  "$\n├── [\"foo\"]\n└── [\"bar\"]\n",
		},
		{
			test: "two_keys_and_sub_keys",
			segs: []*segment{
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
			test: "mixed_and_deep",
			segs: []*segment{
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
			test: "wildcard",
			segs: []*segment{child(spec.Wildcard())},
			str:  "$\n└── [*]\n",
		},
		{
			test: "one_index",
			segs: []*segment{child(spec.Index(0))},
			str:  "$\n└── [0]\n",
		},
		{
			test: "two_indexes",
			segs: []*segment{child(spec.Index(0), spec.Index(2))},
			str:  "$\n└── [0,2]\n",
		},
		{
			test: "other_two_indexes",
			segs: []*segment{child(spec.Index(0)), child(spec.Index(2))},
			str:  "$\n├── [0]\n└── [2]\n",
		},
		{
			test: "index_index",
			segs: []*segment{child(spec.Index(0)).Append(child(spec.Index(2)))},
			str:  "$\n└── [0]\n    └── [2]\n",
		},
		{
			test: "two_keys_and_sub_indexes",
			segs: []*segment{
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
			test: "filter",
			segs: []*segment{child(spec.Filter(spec.And(
				spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
			)))},
			str: "$\n└── [?($)]\n",
		},
		{
			test: "filter_and_key",
			segs: []*segment{
				child(spec.Name("x")),
				child(spec.Filter(spec.And(
					spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
				))),
			},
			str: "$\n├── [\"x\"]\n└── [?($)]\n",
		},
		{
			test: "filter_and_key_segment",
			segs: []*segment{
				child(
					spec.Name("x"),
					spec.Filter(spec.And(
						spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
					)),
				),
			},
			str: "$\n└── [\"x\",?($)]\n",
		},
		{
			test: "nested_filter",
			segs: []*segment{child(spec.Name("x")).Append(
				child(spec.Filter(spec.And(
					spec.Paren(spec.And(spec.Existence(spec.Query(true)))),
				))),
			)},
			str: "$\n└── [\"x\"]\n    └── [?($)]\n",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			n := Tree{root: &segment{children: tc.segs}}
			assert.Equal(t, tc.str, n.String(), tc.test)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test  string
		paths []string
		exp   *Tree
	}{
		{
			test:  "root_only",
			paths: []string{"$"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "two_root_only",
			paths: []string{"$", "$"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "one_name",
			paths: []string{"$.a"},
			exp:   &Tree{root: child().Append(child(spec.Name("a")))},
		},
		{
			test:  "two_names",
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
			test:  "two_names_index",
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
			test:  "two_names_descendant",
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
			test:  "dup_two_names_descendant",
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
			test:  "merge_descendant",
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
			test:  "merge_descendant_children",
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
			test:  "two_single_key_paths",
			paths: []string{"$.a", "$.b"},
			exp: &Tree{
				root: child().Append(child(spec.Name("a"), spec.Name("b"))),
			},
		},
		{
			test:  "two_identical_paths",
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
			test:  "diff_parents_same_child",
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
			test:  "diff_parents_diff_children",
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
			test:  "same_parent_different_child",
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
			test:  "deeply_nested_same_from_diff_parent",
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
			test:  "uneven_mixed_nested",
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
			test:  "different_leaves",
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
			test:  "split_later",
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
			test:  "four_identical_paths",
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
			test:  "same_diff_same",
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
			test:  "same_diff_diff",
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
			test:  "dupe_two_names_index",
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
			test:  "diff_indexes",
			paths: []string{"$[0]", "$[1]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Index(0), spec.Index(1)),
				),
			},
		},
		{
			test:  "diff_sub_indexes",
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
			test:  "diff_index_name_key",
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
			test:  "same_same_idx_diff_key",
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
			test:  "same_diff_idx_diff_child",
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
			test:  "triple_same_diff_idx_diff_child",
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
			test:  "wildcard",
			paths: []string{"$.*", "$.*"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "wildcard_seg",
			paths: []string{"$.*", "$[*]"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "wildcard_trumps_all",
			paths: []string{`$["x", 4, *]`, "$[*, 1]"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "wildcard_trumps_all_inverse",
			paths: []string{"$[1, *]", `$["x", 4, *]`},
			exp:   &Tree{root: child()},
		},
		{
			test:  "drop_trailing_wildcard",
			paths: []string{"$.a.*", "$.a.*"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a")),
			)},
		},
		{
			test:  "drop_trailing_wildcard_diff_key",
			paths: []string{"$.a.*", "$.b.*"},
			exp: &Tree{root: child().Append(
				child(spec.Name("a"), spec.Name("b")),
			)},
		},
		{
			test:  "wildcard_then_a",
			paths: []string{"$[1, *].a", `$["x", 4, *].a`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
					child(spec.Name("a")),
				),
			)},
		},
		{
			test:  "wildcard_then_a_and_b",
			paths: []string{"$[1, *].a", `$["x", 4, *].b`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
					child(spec.Name("a"), spec.Name("b")),
				),
			)},
		},
		{
			test:  "wildcard_then_diff_then_same",
			paths: []string{"$.*.a.c", `$.*.b.c`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
					child(spec.Name("a"), spec.Name("b")).Append(
						child(spec.Name("c")),
					),
				),
			)},
		},
		{
			test:  "wildcard_then_diff_then_same_deep",
			paths: []string{"$.*.a.c.d", `$.*.b.c.d`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
					child(spec.Name("a"), spec.Name("b")).Append(
						child(spec.Name("c")).Append(
							child(spec.Name("d")),
						),
					),
				),
			)},
		},
		{
			test:  "wildcard_then_divergent_paths",
			paths: []string{"$.*.a.b", `$.*.x.y`},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
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
			test:  "wildcard_and_descendant_wildcard",
			paths: []string{"$.*", "$..*"},
			exp:   &Tree{root: child()},
		},
		{
			test:  "wildcard_and_descendant_wildcard_same_child",
			paths: []string{"$.*.a", "$..*.a"},
			exp: &Tree{root: child().Append(
				descendant(spec.Wildcard()).Append(
					child(spec.Name("a")),
				),
			)},
		},
		{
			test:  "wildcard_and_descendant_wildcard_diff_child",
			paths: []string{"$.*.a", "$..*.b"},
			exp: &Tree{root: child().Append(
				child(spec.Wildcard()).Append(
					child(spec.Name("a")),
				),
				descendant(spec.Wildcard()).Append(
					child(spec.Name("b")),
				),
			)},
		},
		{
			test:  "merge_complementary",
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
			test:  "merge_complementary_desc",
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
			test:  "merge_complementary_rev_desc",
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
			test:  "do_not_merge_complementary_mixed_desc",
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
			test:  "do_not_merge_descendant",
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
			test:  "do_not_merge_top_descendant",
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
			test:  "do_not_merge_top_descendant_multi",
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
			test:  "merge_same_branch",
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
			test:  "skip_common_branch",
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
			test:  "merge_index_selectors",
			paths: []string{"$[1,2,1,2,3]"},
			exp: &Tree{root: child().Append(
				child(spec.Index(1), spec.Index(2), spec.Index(3)),
			)},
		},
		{
			test:  "merge_name_selectors",
			paths: []string{`$["x", "y", "x", "r", "y"]`},
			exp: &Tree{root: child().Append(
				child(spec.Name("x"), spec.Name("y"), spec.Name("r")),
			)},
		},
		{
			test:  "merge_slice_selector",
			paths: []string{`$["x", 1, "x", 1, 2, 2:]`},
			exp: &Tree{root: child().Append(
				child(spec.Slice(2), spec.Name("x"), spec.Index(1)),
			)},
		},
		{
			test:  "merge_mixed_multi_path",
			paths: []string{`$["x", 1, "x", 1, 2, 2:]`, `$["x", 2, "y"]`},
			exp: &Tree{root: child().Append(
				child(spec.Slice(2), spec.Name("x"), spec.Index(1), spec.Name("y")),
			)},
		},
		{
			test:  "merge_leaf_node",
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
			test:  "merge_leaf_wildcard_leaf_node",
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
			test:  "merge_reverse_leaf_node",
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
			test:  "merge_deep_common_selectors",
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
		{
			test:  "merge_slice_neg_step",
			paths: []string{"$.store.book[::-1]", "$.store.book[0, 2]"},
			exp: &Tree{
				root: child().Append(
					child(spec.Name("store")).Append(
						child(spec.Name("book")).Append(
							child(spec.Slice(nil, nil, -1)),
						),
					),
				),
			},
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			paths := make([]*jsonpath.Path, len(tc.paths))
			for i, p := range tc.paths {
				paths[i] = jsonpath.MustParse(p)
			}

			a.Equal(tc.exp, New(paths...))
			tc.exp.index = true
			a.Equal(tc.exp, NewFixedModeTree(paths...))
		})
	}
}

func TestSelectorsFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test   string
		seg    *spec.Segment
		expect []spec.Selector
		wild   bool
	}{
		{
			test: "empty",
			seg:  spec.Child(),
			wild: false,
		},
		{
			test:   "one_name",
			seg:    spec.Child(spec.Name("x")),
			expect: []spec.Selector{spec.Name("x")},
			wild:   false,
		},
		{
			test:   "wildcard",
			seg:    spec.Child(spec.Wildcard()),
			expect: []spec.Selector{spec.Wildcard()},
			wild:   true,
		},
		{
			test:   "mix_wildcard",
			seg:    spec.Child(spec.Name("x"), spec.Wildcard()),
			expect: []spec.Selector{spec.Wildcard()},
			wild:   true,
		},
		{
			test:   "wildcard_mix",
			seg:    spec.Child(spec.Wildcard(), spec.Name("x"), spec.Index(1)),
			expect: []spec.Selector{spec.Wildcard()},
			wild:   true,
		},
		{
			test: "mix_selectors_first",
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
			test: "slices_before_indexes",
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
			test: "merge_indexes_into_slices",
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
			test: "wildcard_trump_all",
			seg: spec.Child(
				spec.Slice(2, 3),
				spec.Index(1),
				spec.Index(2),
				spec.Slice(6, 8),
				spec.Index(7),
				spec.Index(6),
				spec.Wildcard(),
			),
			expect: []spec.Selector{spec.Wildcard()},
			wild:   true,
		},
		{
			test: "merge_dupes_slice_first",
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			selectors, wild := selectorsFor(tc.seg)
			a.Equal(tc.expect, selectors)
			a.Equal(tc.wild, wild)
		})
	}
}
