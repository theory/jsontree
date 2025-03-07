package jsontree_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/theory/jsonpath"
	"github.com/theory/jsontree"
)

// Given a user profile as a JSON object, execute a JSONTree query that
// creates a copy of the object that contains only fields named "last" and all
// "primary" contacts of any type.
func Example() {
	// User profile fetched from storage. Contains more fields than we need.
	src := []byte(`{
      "meta": {
        "id": "0c2d9747-c323-4f68-96d0-6c187a1826dc"
      },
      "profile": {
        "name": {
          "first": "Barrack",
          "last": "Obama"
        },
        "contacts": {
          "email": {
            "primary": "foo@example.com",
            "secondary": "2nd@example.net"
          },
          "phones": {
            "primary": "+1-234-567-8901",
            "secondary": "+1-987-654-3210",
            "fax": "+1-293-847-5829"
          },
          "addresses": {
            "primary": [
              "123 Main Street",
              "Chicago", "IL", "90210"
            ],
            "work": [
              "8080 Localhost Drive",
              "Armonk", "NY", "10093"
            ]
          }
        }
      }
    }`)

	// Parse the JSON.
	var value any
	if err := json.Unmarshal(src, &value); err != nil {
		log.Fatal(err)
	}

	// Create a JSONTree query for multiple JSONPaths.
	tree := jsontree.New(
		// Select any field under "profile" named "last".
		jsonpath.MustParse("$.profile..last"),
		// Select the "primary" field from any object under "contacts".
		jsonpath.MustParse("$.profile..contacts.primary"),
	)

	// Select a new object from the original.
	newValue := tree.Select(value)

	// Print the results.
	js, err := json.MarshalIndent(newValue, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(js))
	// Output: {
	//   "profile": {
	//     "contacts": {
	//       "addresses": {
	//         "primary": [
	//           "123 Main Street",
	//           "Chicago",
	//           "IL",
	//           "90210"
	//         ]
	//       },
	//       "email": {
	//         "primary": "foo@example.com"
	//       },
	//       "phones": {
	//         "primary": "+1-234-567-8901"
	//       }
	//     },
	//     "name": {
	//       "last": "Obama"
	//     }
	//   }
	// }
}

func ExampleTree_String() {
	tree := jsontree.New(
		jsonpath.MustParse("$.profile..last"),
		jsonpath.MustParse("$.profile..contacts.primary"),
		jsonpath.MustParse(`$.preferences[0, 2]["type", "value"]`),
		jsonpath.MustParse(`$.preferences[1]["type", "value"]`),
	)
	fmt.Printf("%v\n", tree)
	// Output:
	// $
	// ├── ["profile"]
	// │   ├── ..["last"]
	// │   └── ..["contacts"]
	// │       └── ["primary"]
	// └── ["preferences"]
	//     └── [0,2,1]
	//         └── ["type","value"]
}

// Merge two path queries into an ordered mode JSONTree query. Note that the
// second path selects the second item from the emails array, but the
// [Tree.Select] returns it as the first item.
func ExampleNew() {
	value := map[string]any{
		"name":  "Barrack Obama",
		"years": "2009-2017",
		"emails": []any{
			"potus@example.com",
			"barrack@example.net",
		},
	}

	tree := jsontree.New(
		jsonpath.MustParse("$.name"),
		jsonpath.MustParse("$.emails[1]"),
	)

	js, err := json.Marshal(tree.Select(value))
	if err != nil {
		log.Fatal(err)
	}

	// $.emails[1] appears at $emails[0].
	fmt.Println(string(js))
	// Output:
	// {"emails":["barrack@example.net"],"name":"Barrack Obama"}
}

// Merge two path queries into a fixed mode JSONTree query. Note that the
// second path selects the second item from the emails array and preserves its
// index in the array returned by [Tree.Select] by offsetting the unselected
// first item with nil.
func ExampleNewFixedModeTree() {
	value := map[string]any{
		"name":  "Barrack Obama",
		"years": "2009-2017",
		"emails": []any{
			"potus@example.com",
			"barrack@example.net",
		},
	}

	tree := jsontree.NewFixedModeTree(
		jsonpath.MustParse("$.name"),
		jsonpath.MustParse("$.emails[1]"),
	)

	js, err := json.Marshal(tree.Select(value))
	if err != nil {
		log.Fatal(err)
	}

	// $.emails[1] remains at index 1, offset by nil.
	fmt.Println(string(js))
	// Output:
	// {"emails":[null,"barrack@example.net"],"name":"Barrack Obama"}
}
