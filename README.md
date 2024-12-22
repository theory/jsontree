RFC 9535 JSONPath Tree Queries in Go
====================================

[![⚖️ MIT]][mit] [![📚 Docs]][docs] [![🗃️ Report Card]][card] [![🛠️ Build Status]][ci] [![📊 Coverage]][cov]

The jsontree package provides [RFC 9535 JSONPath] tree selection in Go.

## Learn More

*   🛝 Try it out in the [Playground].
*   📦 See the [package documentation][docs] for usage and examples.
*   📚 See the [RFC 9535 JSONPath] standard for details on the JSONPath query
    syntax and examples of its usage.

## JSONTree

While [RFC 9535 JSONPath] queries select and return an array of values from
the end of a path expression, JSONTree queries merge multiple JSONPath queries
into a single query that selects values from multiple path expressions. They
return results not as an array, but as a subset of the query input, preserving
the paths for each selected value.

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

Selects these values ([playground][play1]):

``` json
[8.95, 12.99, 8.99, 22.99, 399]
```

While this JSONPath query:

``` jsonpath
$..author
```

Selects ([playground][play2]):

``` json
[
  "Nigel Rees",
  "Evelyn Waugh",
  "Herman Melville",
  "J. R. R. Tolkien"
]
```

JSONTree merges these two queries into a single query that returns the
appropriate subset of the original JSON object([playground][play3]):

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

### Array Selection Modes

JSONTree queries select array values in one of two modes: *ordered mode* or
*fixed mode*.

#### Ordered Mode

The default mode, *ordered mode*, preserves the order of items selected from
an array, but not their index positions. For example, given a JSONTree
expression that selects indexes 1, 4, and 3:

```jsonpath
$[1, 4, 3]
```

And an array with six values:

``` json
[
  "zero",
  "one",
  null,
  null,
  "four",
  "five"
]
```

The query wil return the following result ([playground][play4]):

```json
[
  "one",
  null,
  "four"
]
```

Note that the items retain the order in which the appear in the input array,
but their indexes have changed:

*   Item `"one"` moved from index 1 in the input array to index 0 in the
    result
*   The `null` selected from index 3 in the input array appears at index 1 in
    the result
*   Item `"four"` moved from index 4 in the input array to index 2 in the
    result

#### Fixed Mode

In some cases it may be preferable to preserve the index positions of selected
values. *Fixed mode* does so by offsetting selected values with `null`s. For
example, given a JSONTree expression that selects indexes 1, 4, and 3:

``` jsonpath
$[1, 4, 3]
```

And an array with six values:

``` json
[
  "zero",
  "one",
  null,
  null,
  "four",
  "five"
]
```

Fixed mode produces this result ([playground][play5]):

```json
[
  null,
  "one",
  null,
  null,
  "four"
]
```

The values from indexes 1, 3, and 4 remain at those positions in the result,
with gaps between them taken up by `null`s.

Note that the `null` at index 3 selected from the source array is
indistinguishable from the `null` for the unselected values at indexes 0 and
2. To avoid this ambiguity in fixed mode, either disallow `null` values in
inputs, or select only indexes and slices from the start of an array, with no
gaps. For example, selecting indexes 0 - 2:

```jsonpath
$[0:3]
```

Requires no `null` filler values, so we can be sure that the `null`s at
indexes 0 and 2 are from the source ([playground][play6]):

```json
[
  null,
  "one",
  null
]
```

### Use Cases

A couple of use cases drove the conception and design of JSONPath.

#### Permissions

Consider an application in which [ACL]s define permissions for groups of users
to access specific branches or fields of JSON documents. When delivering a
document, the app would:

*   Fetch the groups the user belongs to
*   Convert the permissions from each into JSONPath queries
*   Compile the JSONPath queries into an *ordered mode* JSONTree query
*   Select and return the permitted subset of the document to the user

#### Selective Indexing

Consider a searchable document storage system. For large or complex documents,
it may be infeasible or unnecessary to index the entire document for full-text
search. To index a subset of the fields or branches, one would:

*   Define JSONPaths the fields or branches to index
*   Compile the JSONPath queries into a *fixed mode* JSONTree query
*   Select and submit only the specified subset of each document to the
    indexing system

## How it Works

The jsontree package merges any number of [jsonpath] package queries into a
single tree query, and relies on the [jsonpath] package's [Selector]s for
execution. But while JSONPath expresses a *sequence* of path segments, where
one segment leads to the next, JSONTree compiles them into a *tree*, where one
segment leads to any number of segment branches. This allows multiple JSONPath
expressions to be combined into a single query that selects multiple parts of
a structured JSON value and preserves that subset of its structure.

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

JSONTree, on the other hand, represents a tree of selectors, for example
combining these JSONPaths:

```jsonpath
$.foo["x"].*["a", "b"]
$.foo["y"].*["a", "b"]
$.bar.hi
```

Into this tree structure ([playground][play7]):

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

``` json
["x", "y", "x", "x", 0, 1, 0]
```

Reduces to ([playground][play8]):

``` json
["x", "y", 0, 1]
```

Wildcards eliminate the need for any other selectors, so this JSONPath
segment:

``` jsonpath
["x", "y", 3, *]
```

Reduces to ([playground][play9]):

``` jsonpath
[*]
```

Indexes and slices encapsulated by other slices can be removed, as well. For
example, this segment:

``` jsonpath
[1, 3, 6, 0:4]
```

Reduces to ([playground][play10]):

``` jsonpath
[:4,6]
```

Slice subsets can also be pruned, e.g., this segment:

``` jsonpath
[2:4, 1:3, 0:5]
```

Reduces to ([playground][play11]):

``` jsonpath
[:5]
```

Duplicate filter selectors can also be eliminated; to whit, this segment:

``` jsonpath
[?@.price < 10, ?@.price < 10]
```

Of course reduces to ([playground][play12]):

``` jsonpath
[?@.price < 10]
```

### Segment Merging

The jsontree package also merges selectors *between* path branches where it
can, to reduce redundant selection and maximize query performance. It does so
wherever branches contain equivalent sub-segments, or when their selectors are
equivalent.

For example, given these two paths:

``` jsonpath
$.a.x
$.a.y
```

It will compile this query structure that merges the `x` and `y` segments into
a single segment ([playground][play13]):

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

It creates this tree query that merges the `b` and `x` segments
([playground][play14]):

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

It will compile this tree query ([playground][play15]):

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

Slice and index comparison is also sometimes impossible, mainly when using
negative indexes, since the results depend on the length of input.

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
  [Playground]: https://theory.github.io/jsontree/ "Go JSONTree Playground"
  [jsonpath]: https://pkg.go.dev/jsonpath
  [Selector]: https://pkg.go.dev/jsonpath/spec#Selector
  [ACL]: https://en.wikipedia.org/wiki/Access-control_list
    "Wikipedia: Access-control list"

  <!-- Playground links -->
  [play1]: https://theory.github.io/jsonpath/?p=%2524..price&j=%257B%250A%2520%2520%2522store%2522%253A%2520%257B%250A%2520%2520%2520%2520%2522book%2522%253A%2520%255B%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522reference%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Nigel%2520Rees%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sayings%2520of%2520the%2520Century%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.95%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Evelyn%2520Waugh%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sword%2520of%2520Honour%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252012.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Herman%2520Melville%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Moby%2520Dick%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-553-21311-3%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522J.%2520R.%2520R.%2520Tolkien%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522The%2520Lord%2520of%2520the%2520Rings%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-395-19395-8%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252022.99%250A%2520%2520%2520%2520%2520%2520%257D%250A%2520%2520%2520%2520%255D%252C%250A%2520%2520%2520%2520%2522bicycle%2522%253A%2520%257B%250A%2520%2520%2520%2520%2520%2520%2522color%2522%253A%2520%2522red%2522%252C%250A%2520%2520%2520%2520%2520%2520%2522price%2522%253A%2520399%250A%2520%2520%2520%2520%257D%250A%2520%2520%257D%250A%257D&o=0
  [play2]: https://theory.github.io/jsonpath/?p=%2524..author&j=%257B%250A%2520%2520%2522store%2522%253A%2520%257B%250A%2520%2520%2520%2520%2522book%2522%253A%2520%255B%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522reference%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Nigel%2520Rees%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sayings%2520of%2520the%2520Century%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.95%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Evelyn%2520Waugh%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sword%2520of%2520Honour%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252012.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Herman%2520Melville%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Moby%2520Dick%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-553-21311-3%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522J.%2520R.%2520R.%2520Tolkien%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522The%2520Lord%2520of%2520the%2520Rings%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-395-19395-8%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252022.99%250A%2520%2520%2520%2520%2520%2520%257D%250A%2520%2520%2520%2520%255D%252C%250A%2520%2520%2520%2520%2522bicycle%2522%253A%2520%257B%250A%2520%2520%2520%2520%2520%2520%2522color%2522%253A%2520%2522red%2522%252C%250A%2520%2520%2520%2520%2520%2520%2522price%2522%253A%2520399%250A%2520%2520%2520%2520%257D%250A%2520%2520%257D%250A%257D&o=0
  [play3]: https://theory.github.io/jsontree/?p=%2524..author%250A%2524..price&j=%257B%250A%2520%2520%2522store%2522%253A%2520%257B%250A%2520%2520%2520%2520%2522book%2522%253A%2520%255B%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522reference%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Nigel%2520Rees%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sayings%2520of%2520the%2520Century%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.95%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Evelyn%2520Waugh%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Sword%2520of%2520Honour%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252012.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522Herman%2520Melville%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522Moby%2520Dick%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-553-21311-3%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%25208.99%250A%2520%2520%2520%2520%2520%2520%257D%252C%250A%2520%2520%2520%2520%2520%2520%257B%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522category%2522%253A%2520%2522fiction%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522author%2522%253A%2520%2522J.%2520R.%2520R.%2520Tolkien%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522title%2522%253A%2520%2522The%2520Lord%2520of%2520the%2520Rings%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522isbn%2522%253A%2520%25220-395-19395-8%2522%252C%250A%2520%2520%2520%2520%2520%2520%2520%2520%2522price%2522%253A%252022.99%250A%2520%2520%2520%2520%2520%2520%257D%250A%2520%2520%2520%2520%255D%252C%250A%2520%2520%2520%2520%2522bicycle%2522%253A%2520%257B%250A%2520%2520%2520%2520%2520%2520%2522color%2522%253A%2520%2522red%2522%252C%250A%2520%2520%2520%2520%2520%2520%2522price%2522%253A%2520399%250A%2520%2520%2520%2520%257D%250A%2520%2520%257D%250A%257D&o=0
  [play4]: https://theory.github.io/jsontree/?p=%2524%255B1%252C%25204%252C%25203%255D&j=%255B%250A%2520%2520%2522zero%2522%252C%250A%2520%2520%2522one%2522%252C%250A%2520%2520null%252C%250A%2520%2520null%252C%250A%2520%2520%2522four%2522%252C%250A%2520%2520%2522five%2522%250A%255D&o=0
  [play5]: https://theory.github.io/jsontree/?p=%2524%255B1%252C%25204%252C%25203%255D&j=%255B%250A%2520%2520%2522zero%2522%252C%250A%2520%2520%2522one%2522%252C%250A%2520%2520null%252C%250A%2520%2520null%252C%250A%2520%2520%2522four%2522%252C%250A%2520%2520%2522five%2522%250A%255D&o=1
  [play6]: https://theory.github.io/jsontree/?p=%2524%255B0%253A3%255D&j=%255B%250A%2520%2520%2522zero%2522%252C%250A%2520%2520%2522one%2522%252C%250A%2520%2520null%252C%250A%2520%2520null%252C%250A%2520%2520%2522four%2522%252C%250A%2520%2520%2522five%2522%250A%255D&o=1
  [play7]: https://theory.github.io/jsontree/?p=%2524.foo%255B%2522x%2522%255D.*%255B%2522a%2522%252C%2520%2522b%2522%255D%250A%2524.foo%255B%2522y%2522%255D.*%255B%2522a%2522%252C%2520%2522b%2522%255D%250A%2524.bar.hi&j=%255B%250A%2520%2520%2522zero%2522%252C%250A%2520%2520%2522one%2522%252C%250A%2520%2520null%252C%250A%2520%2520null%252C%250A%2520%2520%2522four%2522%252C%250A%2520%2520%2522five%2522%250A%255D&o=2
  [play8]: https://theory.github.io/jsontree/?p=%2524%255B%2522x%2522%252C%2520%2522y%2522%252C%2520%2522x%2522%252C%2520%2522x%2522%252C%25200%252C%25201%252C%25200%255D&j=&o=2
  [play9]: https://theory.github.io/jsontree/?p=%2524%255B%2522x%2522%252C%2520%2522y%2522%252C%25203%252C%2520*%255D%255B0%255D&j=&o=2
  [play10]: https://theory.github.io/jsontree/?p=%2524%255B1%252C%25203%252C%25206%252C%25200%253A4%255D&j=&o=2
  [play11]: https://theory.github.io/jsontree/?p=%2524%255B2%253A4%252C%25201%253A3%252C%25200%253A5%255D&j=&o=2
  [play12]: https://theory.github.io/jsontree/?p=%2524%255B%253F%2540.price%2520%253C%252010%252C%2520%253F%2540.price%2520%253C%252010%255D&j=&o=2
  [play13]: https://theory.github.io/jsontree/?p=%2524.a.x%250A%2524.a.y&j=&o=2
  [play14]: https://theory.github.io/jsontree/?p=%2524.a.b.c.d%250A%2524.a.x.c.d&j=&o=2
  [play15]: https://theory.github.io/jsontree/?p=%2524.a.x.b%250A%2524.a.y.b%250A%2524.a..x.b%250A%2524.a..y.b&j=&o=2
