Go JSONTree Playground
======================

The source for the [Go JSONTree Playground], a stateless single-page web site
for experimenting with the Go [jsontree] package. Compiled via [TinyGo] into a
ca. 712 K (267 K compressed) [Wasm] file and loaded directly into the page.
All functionality implemented in JavaScript and Go, heavily borrowed from the
[Go JSONPath Playground], [Goldmark Playground] and [serde_json_path Sandbox].

Usage
-----

On load, the form will be filled with sample JSON and 2-4 randomly-selected
example queries. Hit the "Run Query" button to see the values the queries
select from the JSON appear in the "Query Output" field.

To try your own, paste the JSON to query into the "JSON" field and input the
one or more JSONPath queries in the input field, one per line, and hit
the "Run Query" button.

That's it.

Read on for details and additional features.

### Docs

The two buttons in the top-right corner provide documentation and links.

*   Hit the button with the circled question mark in the top right corner to
    reveal a table summarizing the JSONPath syntax.

*   Hit the button with the circled i for information about the JSONTree
    playground.

### Options

Select options for execution and the display of results:

*   **Debug**: Don't execute the query, but print its compiled form as a tree
    diagram. Useful for validating that the compiler properly merged selectors
    and segments.

*   **Fixed Mode**: Preserve the index position of items selected from arrays,
    filling gaps with `null`s. This contrasts with the default "ordered mode",
    which preserves the order but not index position of array items.

### Permalink

Hit the "Permalink" button instead of "Run Query" to reload the page with a
URL that contains the contents the JSONTree, JSON, and options. Copy the URL
to use it for sharing.

Note that the Playground is stateless; no data is stored except in the
Permalink URL itself (and whatever data collection GitHub injects; see its
[privacy statement] for details).

### Paths

Input the JSONPath queries to compile and execute, one per line, into this
field. On load, the app will pre-load 2-4 example queries. See [RFC 9535] for
details on the JSONPath language.

### JSON Input

Input the JSON against which to execute the JSONTree query. May be any kind of
JSON value, including objects, arrays, and scalar values. On load, the field
will contain the JSON object used in examples from [RFC 9535].

## Copyright and License

Copyright (c) 2024 David E. Wheeler. Distributed under the [MIT License].

Based on [Goldmark Playground] the [serde_json_path Sandbox], with icons from
[Boxicons], all distributed under the [MIT License].

  [Go JSONTree Playground]: https://theory.github.io/jsontree/playground
  [jsontree]: https://pkg.go.dev/github.com/theory/jsontree
    "pkg.go.dev: github.com/theory/jsontree"
  [Wasm]: https://webassembly.org "WebAssembly"
  [TinyGo]: https://tinygo.org
  [Go JSONPath Playground]: https://theory.github.io/jsonpath/playground
  [Goldmark Playground]: https://yuin.github.io/goldmark/playground
  [serde_json_path Sandbox]: https://serdejsonpath.live
  [privacy statement]: https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement
  [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
  [MIT License]: https://opensource.org/license/mit
  [Boxicons]: https://boxicons.com
