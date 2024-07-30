# JSONTree

**The jsontree Go package provides [RFC 9535]-inspired tree queries for JSON.**

---

## How it Works

The design is based on the [RFC 9535] JSONPath standard, where "Segments"
contain "Selectors" that select one or more "children" or "descendants" of a
JSON object or array. There are four types of selectors:

*   "Name" selectors select the named child of an object.
*   "Index" selectors select an indexed child of an array.
*   "Wildcard" selectors select all children of a node; descendent wildcard
*   "Slice" selectors select a series of elements from an array, giving a
    selectors select *all* descendants of a node. start position, an end
    position, and an optional step value that moves the position from the
    start to the end.

["Filter" selectors are not yet supported.]

While JSONPath expresses segments as a path, where one segment leads to the
next, JSONTree expresses them as a tree, where one segment leads to any number
of subsequent segments. This allows multiple JSONPath expressions to be
combined into a single query that can select multiple parts of a structured
JSON value and preserve that subset of its structure.

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

  [RFC 9535]: https://datatracker.ietf.org/doc/rfc9535/
    "JSONPath: Query Expressions for JSON"
