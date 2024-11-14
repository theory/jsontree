package jsontree

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
