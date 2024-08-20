package jsonpointer

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type A struct {
	Foo string `json:"$foo,omitempty"`
	Boo bool   `json:"$boo"`
	Bar *B     // Change this to a pointer to avoid potential cycles
}

type B struct {
	Baz int
	C   C
	D   D
}

type C struct {
	Nope string
}

func (c C) JSONProps() map[string]any {
	return map[string]any{
		"bat":   "book",
		"stuff": false,
		"other": nil,
	}
}

type D []string

type (
	CustomMap   map[string]any
	CustomSlice []any
)

func TestWalkJSON(t *testing.T) {
	tests := []struct {
		name             string
		data             any
		expectedElements int
		expectedError    error
	}{
		{
			name: "Simple map",
			data: map[string]any{
				"a": 1,
				"b": "two",
				"c": true,
			},
			expectedElements: 4,
		},
		{
			name: "Nested map",
			data: map[string]any{
				"a": 1,
				"b": map[string]any{
					"c": "nested",
					"d": 2,
				},
			},
			expectedElements: 5,
		},
		{
			name: "Map with slice",
			data: map[string]any{
				"a": []any{1, 2, 3},
				"b": "string",
			},
			expectedElements: 6,
		},
		{
			name:             "Empty map",
			data:             map[string]any{},
			expectedElements: 1,
		},
		{
			name:             "Empty slice",
			data:             []any{},
			expectedElements: 1,
		},
		{
			name: "Complex nested structure",
			data: map[string]any{
				"a": 1,
				"b": []any{
					"string",
					map[string]any{"c": true},
					[]any{1, 2, 3},
				},
				"d": map[string]any{
					"e": "nested",
					"f": []any{4, 5, 6},
				},
			},
			expectedElements: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualElements := 0
			visitedElements := make([]string, 0)
			err := WalkJSON(tt.data, func(elem any) error {
				actualElements++
				visitedElements = append(visitedElements, fmt.Sprintf("%T: %v", elem, elem))
				return nil
			})

			if ok := os.Getenv("BLOODHOUND_JSONPOINTER_TEST_DEBUG"); ok != "" {
				// Print visited elements for debugging
				t.Logf("Visited elements (%d):", actualElements)
				for i, elem := range visitedElements {
					t.Logf("%d: %s", i+1, elem)
				}
			}

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedElements, actualElements, "Mismatch in number of elements visited")
		})
	}
}

func TestWalkJSONWithError(t *testing.T) {
	data := getTestStruct()
	expectedError := errors.New("test error")

	err := WalkJSON(data, func(elem any) error {
		return expectedError
	})

	assert.Equal(t, expectedError, err)
}

func TestWalkJSONWithPanic(t *testing.T) {
	data := getTestStruct()

	assert.Panics(t, func() {
		_ = WalkJSON(data, func(elem any) error {
			panic("unexpected panic")
		})
	})
}

func getTestStruct() *A {
	return &A{
		Foo: "fooval",
		Boo: true,
		Bar: &B{
			Baz: 1,
			C: C{
				Nope: "won't register",
			},
			D: D{"hello", "world"},
		},
	}
}

func getComplexNestedStructure() map[string]any {
	return map[string]any{
		"a": 1,
		"b": []any{
			"string",
			map[string]any{
				"c": true,
				"d": []any{1, 2, 3},
			},
		},
		"e": struct {
			F string
			G int
		}{
			F: "hello",
			G: 42,
		},
	}
}

func getCyclicStructure() *Node {
	cyclic := &Node{Value: "first"}
	cyclic.Next = &Node{Value: "second", Next: cyclic}
	return cyclic
}

func getPointerToPrimitive() *int {
	value := 42
	return &value
}

func getStructWithUnexportedFields() struct {
	exported   int
	unexported string
} {
	return struct {
		exported   int
		unexported string
	}{
		exported:   1,
		unexported: "hidden",
	}
}

func getLargeNestedStructure() map[string]any {
	result := make(map[string]any)
	current := result
	for i := 0; i < 1000; i++ {
		next := make(map[string]any)
		current["next"] = next
		current = next
	}
	return result
}

type Node struct {
	Value any
	Next  *Node
}
