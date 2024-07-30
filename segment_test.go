package jsontree

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			segs: []*Segment{Child(Wildcard)},
			str:  "$\n└── [*]\n",
		},
		{
			name: "one_key",
			segs: []*Segment{Child(Name("foo"))},
			str:  "$\n└── [\"foo\"]\n",
		},
		{
			name: "two_keys",
			segs: []*Segment{Child(Name("foo")), Child(Name("bar"))},
			str:  "$\n├── [\"foo\"]\n└── [\"bar\"]\n",
		},
		{
			name: "two_keys_and_sub_keys",
			segs: []*Segment{
				Child(Name("foo")).Append(
					Child(Name("x")),
					Child(Name("y")),
					Descendant(Name("z")),
				),
				Child(Name("bar")).Append(
					Child(Name("a"), Index(42), Slice(0, 8, 2)),
					Child(Name("b")),
					Child(Name("c")),
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
				Child(Name("foo")).Append(
					Child(Name("x")),
					Child(Name("y")).Append(
						Child(Wildcard).Append(
							Child(Name("a")),
							Child(Name("b")),
						),
					),
				),
				Child(Name("bar")).Append(
					Child(Name("go")),
					Child(Name("z")).Append(
						Child(Wildcard).Append(
							Child(Name("c")),
							Child(Name("d")).Append(
								Child(Slice(2, 3)),
							),
						),
					),
					Child(Name("hi")),
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
			segs: []*Segment{Child(Wildcard)},
			str:  "$\n└── [*]\n",
		},
		{
			name: "one_index",
			segs: []*Segment{Child(Index(0))},
			str:  "$\n└── [0]\n",
		},
		{
			name: "two_indexes",
			segs: []*Segment{Child(Index(0), Index(2))},
			str:  "$\n└── [0,2]\n",
		},
		{
			name: "other_two_indexes",
			segs: []*Segment{Child(Index(0)), Child(Index(2))},
			str:  "$\n├── [0]\n└── [2]\n",
		},
		{
			name: "index_index",
			segs: []*Segment{Child(Index(0)).Append(Child(Index(2)))},
			str:  "$\n└── [0]\n    └── [2]\n",
		},
		{
			name: "two_keys_and_sub_indexes",
			segs: []*Segment{
				Child(Name("foo")).Append(
					Child(Index(0)),
					Child(Index(1)),
					Child(Index(2)),
				),
				Child(Name("bar")).Append(
					Child(Index(3)),
					Child(Index(4)),
					Child(Index(5)),
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n := JSONTree{root: &Segment{children: tc.segs}}
			a.Equal(tc.str, n.String())
		})
	}
}
