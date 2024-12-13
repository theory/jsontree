# Changelog

All notable changes to this project will be documented in this file. It uses the
[Keep a Changelog] format, and this project adheres to [Semantic Versioning].

  [Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
  [Semantic Versioning]: https://semver.org/spec/v2.0.0.html
    "Semantic Versioning 2.0.0"

## [v0.1.1] — 2024-12-12

### 🪲 Bug Fixes

The theme of this release is *Details, details.*

*   Fixed a bug where indexes failed to merge into slices with step -1.

  [v0.1.1]: https://github.com/theory/jsontree/compare/v0.1.0...v0.1.1

## [v0.1.0] — 2024-12-09

The theme of this release is *Standards Matter.*

### ⚡ Improvements

*   First release, everything is new!
*   Relies on [github.com/theory/jsonpath], a full [RFC 9535] JSONPath
    implementation, for path parsing and execution.
*   Selects a subtree of input for one or more path queries.
*   Returns a structure-preserving result.
*   Selects array items in *ordered* mode or *fixed* mode, preserving the
    order or index of selected items.

### 🏗️ Build Setup

*   Built with Go
*   Use `go get` to add to a project
*   The public interface is stable and unlikely to change

### 📚 Documentation

*   Comprehensive documentation of JSONTree, array selection modes, use cases,
    and operation in the [README].
*   Docs on [pkg.go.dev]

  [github.com/theory/jsonpath]: https://pkg.go.dev/github.com/theory/jsonpath
  [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
    "RFC 9535 JSONPath: Query Expressions for JSON"
  [pkg.go.dev]: https://pkg.go.dev/github.com/theory/jsontree
  [README]: https://github.com/theory/jsontree/blob/v0.1.0/README.md
