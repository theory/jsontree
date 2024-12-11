// package main provides the Wasm app.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/theory/jsonpath"
	"github.com/theory/jsontree"
)

const (
	optFixed int = 1 << iota
	optDebug
)

func main() {
	stream := make(chan struct{})

	js.Global().Set("query", js.FuncOf(query))
	js.Global().Set("optFixed", js.ValueOf(optFixed))
	js.Global().Set("optDebug", js.ValueOf(optDebug))

	<-stream
}

func query(_ js.Value, args []js.Value) any {
	queries := args[0].String()
	target := args[1].String()
	opts := args[2].Int()

	return execute(queries, target, opts)
}

func execute(queries, target string, opts int) string {
	// Parse the JSON.
	var value any
	if err := json.Unmarshal([]byte(target), &value); err != nil {
		return fmt.Sprintf("Error parsing JSON: %v", err)
	}

	// Parse the JSONPath queries
	paths := []*jsonpath.Path{}
	for lineNo, line := range strings.Split(queries, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p, err := jsonpath.Parse(line)
		if err != nil {
			return fmt.Sprintf("Error parsing line %v: %v", lineNo, err)
		}
		paths = append(paths, p)
	}

	// Compile the JSONTree.
	var tree *jsontree.Tree
	if opts&optFixed == optFixed {
		tree = jsontree.NewFixedModeTree(paths...)
	} else {
		tree = jsontree.New(paths...)
	}

	// Just output the string representation of the tree.
	if opts&optDebug == optDebug {
		return tree.String()
	}

	// Execute the query against the JSON.
	res := tree.Select(value)

	// Serialize the result
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return fmt.Sprintf("Error parsing results: %v", err)
	}

	return buf.String()
}
