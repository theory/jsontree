RFC 9535 JSONPath Tree Queries in Go
====================================

[![⚖️ MIT]][mit] [![📚 Docs]][docs] [![🗃️ Report Card]][card] [![🛠️ Build Status]][ci] [![📊 Coverage]][cov]

The jsonpath package provides [RFC 9535 JSONPath] tree selection in Go.

## How it Works

Relies on the [github.com/theory/jsonpath] package's [Selector]s for
execution. But while JSONPath expresses segments as a path, where one segment
leads to the next, JSONTree expresses them as a tree, where one segment leads
to any number of subsequent segments. This allows multiple JSONPath
expressions to be combined into a single query that can select multiple parts
of a structured JSON value and preserve that subset of its structure.

In other words, JSONPath represents a list of selectors:

```jsonpath
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

JSONTree, on the other hand, represents as a tree of selectors:

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
*   Back at the root object, kf it contains the key "bar", pass that value to
    the next segments
*   If the value of "bar" is an object that contains the key "hi", add that
    full path to the result

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
  [github.com/theory/jsonpath]: https://pkg.go.dev/github.com/theory/jsonpath
  [Selector]: https://pkg.go.dev/github.com/theory/jsonpath/spec#Selector
