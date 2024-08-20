package jsonpointer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var docBytes = []byte(`{
  "foo": ["bar", "baz"],
  "zoo": {
    "too": ["bar", "baz"],
    "goo": {
      "loo": ["bar", "baz"]
    }
  },
  "": 0,
  "a/b": 1,
  "c%d": 2,
  "e^f": 3,
  "g|h": 4,
  "i\\j": 5,
  "k\"l": 6,
  " ": 7,
  "m~n": 8,
  "deep": {
    "nested": {
      "object": {
        "with": {
          "array": [1, 2, 3, 4, 5]
        }
      }
    }
  },
  "empty": {},
  "emptyArray": []
}`)

func TestParse(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		parsed string
		err    string
	}{
		{"Empty string", "", "", ""},
		{"Root", "#/", "/", ""},
		{"Simple property", "#/foo", "/foo", ""},
		{"Property with trailing slash", "#/foo/", "/foo/", ""},
		{"URL fragment", "https://example.com#/foo", "/foo", ""},
		{"Multiple segments", "#/foo/bar/baz", "/foo/bar/baz", ""},
		{"Empty segment", "#//", "//", ""},
		{"Escaped characters", "#/a~1b~1c~0d", "/a~1b~1c~0d", ""},
		{"URL encoded characters", "#/foo%20bar", "/foo bar", ""},
		{"Invalid URL", "://", "", "missing protocol scheme"},
		{"Invalid pointer", "#7", "", "non-empty references must begin with a '/' character"},
		{"Complex URL", "https://example.com/api/v1#/data/0/name", "/data/0/name", ""},
		{"Pointer with query params", "https://example.com/api?param=value#/foo", "/foo", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.raw)
			if c.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), c.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.parsed, got.String())
			}
		})
	}
}

func TestEval(t *testing.T) {
	document := map[string]any{}
	err := json.Unmarshal(docBytes, &document)
	require.NoError(t, err)

	testCases := []struct {
		name       string
		pointerStr string
		expected   any
		err        string
	}{
		{"Root document", "", document, ""},
		{"Simple property", "/foo", document["foo"], ""},
		{"Array element", "/foo/0", "bar", ""},
		{"Nested object", "/zoo/too/0", "bar", ""},
		{"Deeply nested", "/zoo/goo/loo/0", "bar", ""},
		{"Empty key", "/", float64(0), ""},
		{"Key with escaped /", "/a~1b", float64(1), ""},
		{"Key with %", "/c%d", float64(2), ""},
		{"Key with ^", "/e^f", float64(3), ""},
		{"Key with |", "/g|h", float64(4), ""},
		{"Key with \\", "/i\\j", float64(5), ""},
		{"Key with \"", "/k\"l", float64(6), ""},
		{"Key with space", "/ ", float64(7), ""},
		{"Key with escaped ~", "/m~0n", float64(8), ""},
		{"URL fragment root", "#", document, ""},
		{"URL fragment property", "#/foo", document["foo"], ""},
		{"URL fragment array", "#/foo/0", "bar", ""},
		{"URL encoded /", "#/a~1b", float64(1), ""},
		{"URL encoded %", "#/c%25d", float64(2), ""},
		{"URL encoded ^", "#/e%5Ef", float64(3), ""},
		{"URL encoded |", "#/g%7Ch", float64(4), ""},
		{"URL encoded \\", "#/i%5Cj", float64(5), ""},
		{"URL encoded \"", "#/k%22l", float64(6), ""},
		{"URL encoded space", "#/%20", float64(7), ""},
		{"URL with fragment", "https://example.com#/m~0n", float64(8), ""},
		{"Deep nested array", "/deep/nested/object/with/array/4", float64(5), ""},
		{"Empty object", "/empty", map[string]any{}, ""},
		{"Empty array", "/emptyArray", []any{}, ""},
		{"Non-existent key", "/undefined", nil, "key not found: 'undefined' in pointer: /undefined for JSON resource"},
		{"Invalid array index (non-integer)", "/foo/bar", nil, "invalid array access: token 'bar' is not a valid array index"},
		{"Array index out of bounds", "/foo/3", nil, "array index out of bounds: index 3 exceeds array length of 2"},
		{"Non-existent nested key", "/bar/baz", nil, "key not found: 'bar' in pointer: /bar/baz for JSON resource"},
		{"Negative array index", "/foo/-1", nil, "invalid array index: -1 is negative"},
		{"Attempting object access on array", "/foo/object", nil, "invalid array access: token 'object' is not a valid array index"},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			ptr, err := Parse(c.pointerStr)
			require.NoError(t, err)

			got, err := ptr.Eval(document)
			if c.err != "" {
				assert.Error(t, err)
				assert.Equal(t, c.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, got)
			}
		})
	}
}

func TestDescendant(t *testing.T) {
	cases := []struct {
		name   string
		parent string
		path   string
		parsed string
		err    string
	}{
		{"Root to child", "/", "foo", "/foo", ""},
		{"Root to nested", "/", "foo/bar", "/foo/bar", ""},
		{"Append to existing path", "/foo", "bar", "/foo/bar", ""},
		{"Append multiple segments", "/foo", "bar/baz", "/foo/bar/baz", ""},
		{"Absolute path", "/foo", "/bar/baz", "/bar/baz", ""},
		{"Empty parent", "", "foo", "/foo", ""},
		{"Empty child", "/foo", "", "/foo", ""},
		{"Both empty", "", "", "/", ""},
		{"Parent with trailing slash", "/foo/", "bar", "/foo/bar", ""},
		{"Child with leading slash", "/foo", "/bar", "/bar", ""},
		{"Escaped characters", "/a~1b", "c~0d", "/a~1b/c~0d", ""},
		{"Root to root", "/", "/", "/", ""},
		{"Non-root to root", "/foo", "/", "/", ""},
		{"Multiple slashes in path", "/foo", "bar//baz", "/foo/bar//baz", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := Parse(c.parent)
			require.NoError(t, err)

			desc, err := p.Descendant(c.path)
			if c.err != "" {
				assert.Error(t, err)
				assert.Equal(t, c.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.parsed, desc.String())
			}
		})
	}
}

func TestPointerBuilder(t *testing.T) {
	t.Run("Build simple pointer", func(t *testing.T) {
		pb := NewPointerBuilder()
		ptr := pb.AddToken("foo").AddToken("bar").Build()
		assert.Equal(t, "/foo/bar", ptr.String())
	})

	t.Run("Build pointer with escaped characters", func(t *testing.T) {
		pb := NewPointerBuilder()
		ptr := pb.AddToken("a/b").AddToken("c~d").Build()
		assert.Equal(t, "/a~1b/c~0d", ptr.String())
	})

	t.Run("Build empty pointer", func(t *testing.T) {
		pb := NewPointerBuilder()
		ptr := pb.Build()
		assert.Equal(t, "", ptr.String())
	})
}

func TestPointerEquality(t *testing.T) {
	cases := []struct {
		name     string
		pointer1 string
		pointer2 string
		equal    bool
	}{
		{"Equal pointers", "/foo/bar", "/foo/bar", true},
		{"Different pointers", "/foo/bar", "/foo/baz", false},
		{"Different lengths", "/foo", "/foo/bar", false},
		{"Empty pointers", "", "", true},
		{"Root pointers", "/", "/", true},
		{"Escaped characters", "/a~1b/c~0d", "/a~1b/c~0d", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p1, _ := Parse(c.pointer1)
			p2, _ := Parse(c.pointer2)
			assert.Equal(t, c.equal, p1.Equals(p2))
		})
	}
}

func TestPointerIsPrefix(t *testing.T) {
	cases := []struct {
		name     string
		pointer1 string
		pointer2 string
		isPrefix bool
	}{
		{"Is prefix", "/foo", "/foo/bar", true},
		{"Is not prefix", "/foo/bar", "/foo", false},
		{"Equal pointers", "/foo/bar", "/foo/bar", true},
		{"Empty is prefix of all", "", "/foo/bar", true},
		{"Root is prefix of all", "/", "/foo/bar", true},
		{"Root is prefix of root", "/", "/", true},
		{"Empty is prefix of root", "", "/", true},
		{"Root is prefix of empty", "/", "", true},
		{"Empty is prefix of empty", "", "", true},
		{"Different paths", "/foo/bar", "/baz/qux", false},
		{"Longer is not prefix", "/foo/bar/baz", "/foo/bar", false},
		{"Prefix with different end", "/foo/bar", "/foo/baz", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p1, err := Parse(c.pointer1)
			assert.NoError(t, err)
			p2, err := Parse(c.pointer2)
			assert.NoError(t, err)
			assert.Equal(t, c.isPrefix, p1.IsPrefix(p2), "Pointer1: %v, Pointer2: %v", p1, p2)
		})
	}
}

func TestJSONMarshaling(t *testing.T) {
	cases := []struct {
		name    string
		pointer string
	}{
		{"Simple pointer", "/foo/bar"},
		{"Empty pointer", ""},
		{"Root pointer", "/"},
		{"Escaped characters", "/a~1b/c~0d"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, _ := Parse(c.pointer)
			marshaled, err := json.Marshal(p)
			assert.NoError(t, err)

			var unmarshaled Pointer
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err)

			assert.Equal(t, p, unmarshaled)
		})
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid empty pointer", "", true},
		{"Valid root pointer", "/", true},
		{"Valid simple pointer", "/foo/bar", true},
		{"Valid escaped characters", "/a~1b/c~0d", true},
		{"Valid complex pointer", "/foo/bar~1baz~0qux/0", true},
		{"Invalid: no leading slash", "foo/bar", false},
		{"Invalid: unescaped ~", "/foo/bar~", false},
		{"Invalid: unescaped ~ in middle", "/foo/bar~baz", false},
		{"Invalid: incomplete escape", "/foo/bar~", false},
		{"Invalid: double slash", "/foo//bar", true},        // Note: This is actually valid per RFC 6901
		{"Invalid: non-ASCII characters", "/foo/bár", true}, // Note: This is actually valid per RFC 6901
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.valid, Validate(c.input), "Input: %s", c.input)
		})
	}
}
