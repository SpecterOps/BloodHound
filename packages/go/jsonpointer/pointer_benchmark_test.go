package jsonpointer

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
)

// generateNestedJSON creates a nested JSON structure with the specified depth and breadth
func generateNestedJSON(depth, breadth int) map[string]interface{} {
	if depth == 0 {
		return map[string]interface{}{"value": rand.Intn(100)}
	}
	result := make(map[string]interface{})
	for i := 0; i < breadth; i++ {
		key := fmt.Sprintf("key%d", i)
		result[key] = generateNestedJSON(depth-1, breadth)
	}
	return result
}

// generateLargeJSON creates a large JSON structure with many top-level keys
func generateLargeJSON(size int) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key%d", i)
		result[key] = rand.Intn(100)
	}
	return result
}

// generateLargeArray creates a large JSON array
func generateLargeArray(size int) []interface{} {
	result := make([]interface{}, size)
	for i := 0; i < size; i++ {
		result[i] = rand.Intn(100)
	}
	return result
}

// BenchmarkParse benchmarks the Parse function with different input sizes
func BenchmarkParse(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"Short", "/foo/bar"},
		{"Medium", "/a/b/c/d/e/f/g/h/i/j"},
		{"Long", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"},
		{"ComplexEscaping", "/a~1b~1c~0d~0e~1f/g~0h~0i~1j~1k/l~0m~1n~0o~1p"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Parse(bm.input)
			}
		})
	}
}

// BenchmarkEval benchmarks the Eval function with different JSON structures
func BenchmarkEval(b *testing.B) {
	benchmarks := []struct {
		name     string
		json     interface{}
		pointers []string
	}{
		{
			name: "Shallow",
			json: map[string]interface{}{
				"foo": "bar",
				"baz": 42,
				"qux": []interface{}{1, 2, 3},
			},
			pointers: []string{"/foo", "/baz", "/qux/2"},
		},
		{
			name:     "Deep",
			json:     generateNestedJSON(10, 2),
			pointers: []string{"/key0/key1/key0/key1/key0/key1/key0/key1/key0/key1/value"},
		},
		{
			name:     "Wide",
			json:     generateLargeJSON(1000),
			pointers: []string{"/key0", "/key500", "/key999"},
		},
		{
			name:     "LargeArray",
			json:     map[string]interface{}{"array": generateLargeArray(10000)},
			pointers: []string{"/array/0", "/array/5000", "/array/9999"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			jsonBytes, _ := json.Marshal(bm.json)
			var doc interface{}
			_ = json.Unmarshal(jsonBytes, &doc)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, ptr := range bm.pointers {
					p, _ := Parse(ptr)
					_, _ = p.Eval(doc)
				}
			}
		})
	}
}

// BenchmarkDescendant benchmarks the Descendant function
func BenchmarkDescendant(b *testing.B) {
	benchmarks := []struct {
		name   string
		parent string
		path   string
	}{
		{"ShortToShort", "/foo", "bar"},
		{"ShortToLong", "/foo", "bar/baz/qux/quux/corge/grault/garply"},
		{"LongToShort", "/foo/bar/baz/qux/quux/corge/grault", "garply"},
		{"LongToLong", "/a/b/c/d/e/f/g", "h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			parent, _ := Parse(bm.parent)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = parent.Descendant(bm.path)
			}
		})
	}
}

// BenchmarkPointerBuilder benchmarks the PointerBuilder
func BenchmarkPointerBuilder(b *testing.B) {
	benchmarks := []struct {
		name   string
		tokens []string
	}{
		{"Few", []string{"foo", "bar", "baz"}},
		{"Many", []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				pb := NewPointerBuilder()
				for _, token := range bm.tokens {
					pb.AddToken(token)
				}
				_ = pb.Build()
			}
		})
	}
}

// BenchmarkWalkJSON benchmarks the WalkJSON function with different JSON structures
func BenchmarkWalkJSON(b *testing.B) {
	benchmarks := []struct {
		name string
		json interface{}
	}{
		{"Shallow", generateLargeJSON(100)},
		{"Deep", generateNestedJSON(10, 3)},
		{"Wide", generateLargeJSON(1000)},
		{"LargeArray", generateLargeArray(10000)},
		{"Complex", map[string]interface{}{
			"nested":     generateNestedJSON(5, 3),
			"large":      generateLargeJSON(100),
			"largeArray": generateLargeArray(1000),
		}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			jsonBytes, _ := json.Marshal(bm.json)
			var doc interface{}
			_ = json.Unmarshal(jsonBytes, &doc)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = WalkJSON(doc, func(elem interface{}) error {
					return nil
				})
			}
		})
	}
}
