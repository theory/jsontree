RFC 9535 JSONPath Tree Queries in Go
====================================

[![⚖️ MIT]][mit] [![📚 Docs]][docs] [![🗃️ Report Card]][card] [![🛠️ Build Status]][ci] [![📊 Coverage]][cov]

The jsontree package provides [RFC 9535 JSONPath] tree selection in Go.

## JSONTree

While [RFC 9535 JSONPath] queries select and return an array of values from
the end of a path, JSONTree queries merge multiple JSONPath queries into a
single query object that selects values from multiple paths. They return
results not as an array, but as a subset of the query input, preserving the
paths for each selected value.

### Example

Consider this JSON:

```json
{
  "store": {
    "book": [
      {
        "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      {
        "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      {
        "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      {
        "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 399
    }
  }
}
```

This JSONPath query:

``` jsonpath
$..price
```

Selects these values:

``` json
[8.95, 12.99, 8.99, 22.99, 399]
```

While this JSONPath query:

``` jsonpath
$..author
```

Selects:

``` json
[
  "Nigel Rees",
  "Evelyn Waugh",
  "Herman Melville",
  "J. R. R. Tolkien"
]
```

JSONTree can merge these two queries into a single query that returns the
appropriate subset of the original JSON object:

``` json
{
  "store": {
    "book": [
      {
        "author": "Nigel Rees",
        "price": 8.95
      },
      {
        "author": "Evelyn Waugh",
        "price": 12.99
      },
      {
        "author": "Herman Melville",
        "price": 8.99
      },
      {
        "author": "J. R. R. Tolkien",
        "price": 22.99
      }
    ],
    "bicycle": {
      "price": 399
    }
  }
}
```

### Index Selection

The jsontree package's preservation of the input data structure in the result
it returns applies transparently to JSON objects, but requires special
handling for arrays. In order to preserve index locations exactly, any
preceding values will be `null`.

For example, given a JSONTree expression that selects indexes 2 and 4:

``` jsonpath
$[2, 4]
```

And an array with six values:

``` json
[
  "zero",
  "one",
  "two",
  null,
  "four",
  "five"
]
```

The query wil return the following result:

```json
[
  null,
  null,
  "two",
  null,
  "four"
]
```

This preserves the index position of the selected values at the cost of `null`
values for unselected values. But note that the `null` value at index three in
the source array is indistinguishable from the `null` for the "unselected"
value at index three in the result. To avoid these issues, only select indexes
and slices from the start of an array whenever possible.

A future release may add an option to drop positional preservation for array
values, in which case the result would be:

```json
[
  "two",
  "four"
]
```

### What For

What, you might wonder, is the point? A couple of use cases drove the
conception and design of the jsonpath package.

#### Permissions

Consider an application in which [ACL]s define permissions for groups of users
to access specific branches or fields of JSON documents. Whe delivering a
document, the app would:

*   Fetch the groups the user belongs to
*   Convert the permissions from each into JSONPath queries
*   Compile the permission JSONPath queries into a JSONTree query
*   Select and return the permitted subset of the document to the user

#### Selective Indexing

Consider a searchable document storage system. For large or complex documents,
it may not be feasible or required to index the entire document for searching.
To index just a subset of the fields or branches, one would:

*   Define JSONPaths the fields or branches to index
*   Compile the indexing JSONPath queries into a JSONTree query
*   Select and return only the specified subset of each document for the
    indexing system

## How it Works

The jsontree package can merge any number of [jsonpath] package queries into a
single tree query, but relies on the [jsonpath] package's [Selector]s for
execution. But while JSONPath expresses a sequence of path segments, where one
segment leads to the next, JSONTree expresses them as a tree, where one
segment leads to any number of subsequent segments. This allows multiple
JSONPath expressions to be combined into a single query that can select
multiple parts of a structured JSON value and preserve that subset of its
structure.

In other words, JSONPath represents a list of selectors, for example:

``` jsonpath
$.a.b[0].["x", "y", "z"]
```

Given an object, this JSONPath will:

*   Start at the root
*   If the root is an object and contains the key "a", pass the value of "a"
    to the next segment
*   If the value of "a" is an object that contains the key "b", pass its value
    to the next segment
*   If the value of "b" is an array with a value at index 0, pass that value
    to the next segment
*   If that segment is an object, return an array of the values under the
    subset of the keys "x", "y", and "z" that exist in the object

JSONTree, on the other hand, represents a tree of selectors, for example:

```tree
$
├── ["foo"]
│   ├── ["x"]
│   └── ["y"]
│       └── [*]
│           └── ["a", "b"]
└── ["bar"]
	└── ["hi"]
```

Given an object, this JSONTree will:

*   Start at the root
*   If the root is an object that contains the key "foo", pass that value to
    the next segments
*   If the value of "foo" is an object that contains the key "x", add that
    full path to the result
*   If the value of "foo" is an object that contains the key "y", pass that
    value to the next segments
*   If the value of "y" is an array or object, pass all of its values to the
    next segment
*   If any of those values is an object that contains the keys "a" or "b", add
    the full path to those values to the result.
*   Back at the root object, if it contains the key "bar", pass that value to
    the next segments
*   If the value of "bar" is an object that contains the key "hi", add that
    full path to the result

## From Path to Tree

The jsontree package applies a number of heuristics to compile an efficient
query tree by removing duplicate or redundant selectors and merging
overlapping path segments.

### Selector Merging

Whereas [jsonpath] returns values for each selector that matches, even
duplicate selectors, jsontree does not, since it preserves the original data
structure along the selected paths. Redundant selectors can therefore be
eliminated. For example, this JSONPath segment:

``` jsonpath
["x", "y", "x", "x", 0, 1, 0]
```

Reduces to:

```jsonpath
["x", "y", 0, 1]
```

Wildcards eliminate the need for any other selectors, so this JSONPath
segment:

``` jsonpath
["x", "y", 3, *]
```

Reduces to:

``` jsonpath
[*]
```

Indexes and slices encapsulated by other slices can be removed, as well. For
example, this segment:

``` jsonpath
[1, 3, 6, 0:4]
```

Reduces to:

``` jsonpath
[0:4, 6]
```

Slice subsets can also be pruned, e.g., this segment:

``` jsonpath
[2:4, 1:3 0:5]
```

Reduces to:

``` jsonpath
[0:5]
```

Duplicate filter selectors can also be eliminated; to whit, this segment:

``` jsonpath
[?@.price < 10, ?@.price < 10]
```

Of course reduces to:

``` jsonpath
[?@.price < 10]
```

### Segment Merging

The jsontree package also merges selectors between paths where it can, to
reduce redundant selection and maximize query performance. It does so wherever
segments at the same level of hierarchy contain equivalent sub-segments, or
when their selectors are equivalent.

For example, given these two paths:

``` jsonpath
$.a.x
$.a.y
```

It will compile this query structure that merges the `x` and `y` segments into
a single segment:

``` tree
$
└── ["a"]
    └── ["x", "y"]
```

And for these two JSONPath queries:

``` jsonpath
$.a.b.c.d
$.a.x.c.d
```

It creates this tree query that merges the `b` and `x` segments:

``` tree
$
└── ["a"]
    └── ["b", "x"]
        └── ["c"]
            └── ["d"]
```

When one segment is a descendant segment (`..[]`) and the other is not, it
discards the non-descendant segment only when both constitute the same
sub-paths. For example, given these four paths:

``` jsonpath
$.a.x.b
$.a.y.b
$.a..x.b
$.a..y.b
```

It will compile this tree query:

``` tree
$
└── ["a"]
    └── ..["x", "y"]
        └── ["b"]
```

Once the paths have been compiled into a tree, the jsontree package makes
another pass over the tree to eliminate remaining duplicates and redundancies,
notably segments with identical children, taking descendant vs regular child
selectors into account, and then merging slice selectors.

### Limitations

The selector comparisons are imperfect. For example, two filters can be
logically equivalent but have different orders of operands, so would not be
considered equivalent. This may be rectified in the future by normalizing
filter stringification.

Slice comparison is also sometimes impossible, mainly when using negative
index arguments, since the resulting depend on the length of input.

These redundancies should be acceptable, however, as relatively less common
expressions that trigger multiple selection of the same values. Their places
in the resulting data structure are unchanged, however, so the outcome will be
the same.

## Copyright

Copyright © 2024 David E. Wheeler

  [⚖️ MIT]: https://img.shields.io/badge/License-MIT-blue.svg "⚖️ MIT License"
  [mit]: https://opensource.org/license/MIT "⚖️ MIT License"
  [📚 Docs]: https://godoc.org/github.com/theory/jsontree?status.svg "📚 Documentation"
  [docs]: https://pkg.go.dev/github.com/theory/jsontree "📄 Documentation"
  [🗃️ Report Card]: https://goreportcard.com/badge/github.com/theory/jsontree
    "🗃️ Report Card"
  [card]: https://goreportcard.com/report/github.com/theory/jsontree
    "🗃️ Report Card"
  [🛠️ Build Status]: https://github.com/theory/jsontree/actions/workflows/ci.yml/badge.svg
    "🛠️ Build Status"
  [ci]: https://github.com/theory/jsontree/actions/workflows/ci.yml
    "🛠️ Build Status"
  [📊 Coverage]: https://codecov.io/gh/theory/jsontree/graph/badge.svg?token=TjLPa2bF5s
    "📊 Code Coverage"
  [cov]: https://codecov.io/gh/theory/jsontree "📊 Code Coverage"
  [RFC 9535 JSONPath]: https://www.rfc-editor.org/rfc/rfc9535.html
    "RFC 9535 JSONPath: Query Expressions for JSON"
  [RFC 9535]: https://datatracker.ietf.org/doc/rfc9535/
    "JSONPath: Query Expressions for JSON"
  [jsonpath]: https://pkg.go.dev/jsonpath
  [Selector]: https://pkg.go.dev/jsonpath/spec#Selector
  [ACL]: https://en.wikipedia.org/wiki/Access-control_list
    "Wikipedia: Access-control list"
