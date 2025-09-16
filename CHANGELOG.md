# Changelog

All notable changes to this project will be documented in this file. It uses the
[Keep a Changelog] format, and this project adheres to [Semantic Versioning].

  [Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
  [Semantic Versioning]: https://semver.org/spec/v2.0.0.html
    "Semantic Versioning 2.0.0"

## [v0.2.1] — Unreleased

### ⬆️ Dependency Updates

*   Upgraded github.com/theory/jsonpath to v0.10.1.

### 📔 Notes

*   Upgraded to `golangci-lint` v2.4.0.
*   Fixed test name scoping issues with testify objects.

  [v0.10.1]: https://github.com/theory/jsonpath/compare/v0.2.0...v0.2.1

## [v0.2.0] — 2025-05-06

### ⬆️ Dependency Updates

*   Upgraded github.com/theory/jsonpath to v0.9.0.
*   Require Go 1.23 to because the jsonpath package does.

  [v0.2.0]: https://github.com/theory/jsontree/compare/v0.1.2...v0.2.0

## [v0.1.2] — 2024-12-22

### 📚 Documentation

*   Added playground links to the README and fixed some typos.
*   Updated the `New` and `NewFixedModeTree` examples to emit JSON instead of
    Go map stringification.

### 🪲 Bug Fixes

*   Fixed an issue where slice selectors were not properly merged when more
    than one slice encapsulated another.

  [v0.1.2]: https://github.com/theory/jsontree/compare/v0.1.1...v0.1.2

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
