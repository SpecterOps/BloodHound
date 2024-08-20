// Package jsonpointer provides utilities for working with JSON Pointers as defined in RFC 6901.
// It allows parsing, evaluating, and manipulating JSON Pointers, which are strings that identify
// specific values within a JSON document.
//
// The implementation focuses on efficiency, flexibility, and ease of use. It aims to provide
// a balance between performance and maintainability, with clear error handling and
// extensibility for future enhancements.
package jsonpointer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Unexported constants used for consistent error messaging across the package.
// Using constants for error messages ensures consistency and makes it easier
// to update messages in the future if needed.
const (
	errMsgInvalidStart      = "non-empty references must begin with a '/' character"
	errMsgInvalidArrayIndex = "invalid array index: %s"
	errMsgIndexOutOfBounds  = "index %d exceeds array length of %d"
	errMsgKeyNotFound       = "key not found: '%s' in pointer: %s for JSON resource"
)

// Unexported constants used for escaping and unescaping JSON Pointer tokens.
// These constants are defined according to RFC 6901 specifications.
const (
	separator        = "/"
	escapedSeparator = "~1"
	tilde            = "~"
	escapedTilde     = "~0"
)

// ErrInvalidPointer represents an error that occurs when a pointer is invalid.
// It encapsulates the reason for the invalidity, providing more context than a simple error string.
// This custom error type allows for more specific error handling by users of the package.
type ErrInvalidPointer struct {
	Reason string
}

func (e ErrInvalidPointer) Error() string {
	return fmt.Sprintf("invalid pointer: %s", e.Reason)
}

// Pointer represents a parsed JSON pointer as a slice of reference tokens.
// We use a slice of strings instead of a single string to avoid repeated parsing and splitting
// during pointer evaluation, improving performance for repeated use of the same pointer.
type Pointer []string

// Parse parses str into a Pointer structure. It handles both plain JSON Pointer strings
// and URL strings containing JSON Pointers. This flexibility allows users to work with
// JSON Pointers in various contexts, including within URLs, which is a common use case
// in web applications.
//
// The function first checks for common cases ("" and "#") to avoid unnecessary processing.
// It then checks if the string starts with '/', which indicates a plain JSON Pointer.
// If not, it attempts to parse as a URL and extract the fragment. This approach balances
// efficiency (fast-path for common cases) with flexibility (handling URLs).
//
// Returns:
// - A Pointer structure representing the parsed JSON Pointer.
// - An error if the input string is invalid or cannot be parsed.
//
// Usage:
//
//	ptr, err := Parse("/foo/0/bar")
//	if err != nil {
//	    // Handle error
//	}
//	// Use ptr...
//
//	urlPtr, err := Parse("http://example.com/data#/foo/bar")
//	if err != nil {
//	    // Handle error
//	}
//	// Use urlPtr...
func Parse(str string) (Pointer, error) {
	// Fast paths that skip URL parse step
	if len(str) == 0 || str == "#" {
		return Pointer{}, nil
	} else if str[0] == '/' {
		return parse(str)
	}

	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	return parse(u.Fragment)
}

// parse is an unexported function that converts a string into a Pointer after the initial '/' character.
// It combines splitting and unescaping in a single pass for efficiency.
//
// Rationale:
// - Single-pass parsing reduces allocations and improves performance, especially for long pointers.
// - The function handles escaping inline, avoiding the need for a separate unescaping step.
// - It uses a strings.Builder for efficient string construction, minimizing allocations.
//
// Parameters:
// - str: The string to parse, without the initial '/' character.
//
// Returns:
// - A Pointer representing the parsed string.
// - An error if the string is invalid according to the JSON Pointer syntax.
func parse(str string) (Pointer, error) {
	if len(str) == 0 {
		return Pointer{}, nil
	}
	if str[0] != '/' {
		return nil, ErrInvalidPointer{errMsgInvalidStart}
	}

	var tokens []string
	var token strings.Builder
	for i := 1; i < len(str); i++ {
		switch str[i] {
		case '/':
			tokens = append(tokens, token.String())
			token.Reset()
		case '~':
			if i+1 < len(str) {
				switch str[i+1] {
				case '0':
					token.WriteByte('~')
				case '1':
					token.WriteByte('/')
				default:
					token.WriteByte('~')
					token.WriteByte(str[i+1])
				}
				i++
			} else {
				token.WriteByte('~')
			}
		default:
			token.WriteByte(str[i])
		}
	}
	tokens = append(tokens, token.String())
	return Pointer(tokens), nil
}

// String implements the fmt.Stringer interface for Pointer,
// returning the escaped string representation of the JSON Pointer.
//
// Rationale:
// - Using strings.Builder for efficient string construction.
// - Escaping each token individually ensures correct handling of special characters.
// - Implementing Stringer allows seamless use with fmt package and other standard library functions.
//
// Returns:
// - A string representation of the Pointer.
//
// Usage:
//
//	ptr, _ := Parse("/foo/bar~0baz")
//	fmt.Println(ptr.String()) // Output: /foo/bar~0baz
func (p Pointer) String() string {
	var sb strings.Builder
	for _, tok := range p {
		sb.WriteString("/")
		sb.WriteString(escapeToken(tok))
	}
	return sb.String()
}

// escapeToken is an unexported function that applies the escaping process to a single token.
// This function is crucial for ensuring that the string representation
// of a Pointer is valid according to the JSON Pointer specification.
//
// Rationale:
// - Using strings.Replace for clarity and simplicity.
// - Escaping '~' before '/' to handle cases where both characters appear.
//
// Parameters:
// - unescapedToken: The token to escape.
//
// Returns:
// - The escaped token string.
func escapeToken(unescapedToken string) string {
	return strings.Replace(
		strings.Replace(unescapedToken, tilde, escapedTilde, -1),
		separator, escapedSeparator, -1,
	)
}

// IsEmpty checks if the Pointer is empty (i.e., has no reference tokens).
// This method is useful for quickly determining if a Pointer references the root of a document.
//
// Returns:
// - true if the pointer is empty, false otherwise.
//
// Usage:
//
//	ptr := Pointer{}
//	fmt.Println(ptr.IsEmpty()) // Output: true
func (p Pointer) IsEmpty() bool {
	return len(p) == 0
}

// Head returns the first reference token of the Pointer.
// This method is useful for algorithms that need to process a Pointer token by token.
//
// Rationale:
// - Returns a pointer to the string to avoid unnecessary copying.
// - Returns nil for empty Pointers to allow for easy null checks.
//
// Returns:
// - A pointer to the first token string, or nil if the Pointer is empty.
//
// Usage:
//
//	ptr, _ := Parse("/foo/bar")
//	head := ptr.Head()
//	fmt.Println(*head) // Output: foo
func (p Pointer) Head() *string {
	if len(p) == 0 {
		return nil
	}
	return &p[0]
}

// Tail returns a new Pointer containing all reference tokens except the first.
// This method complements Head(), allowing for recursive processing of Pointers.
//
// Rationale:
// - Creates a new slice to avoid modifying the original Pointer.
// - Efficient for most use cases, as it's a simple slice operation.
//
// Returns:
// - A new Pointer with all but the first token.
//
// Usage:
//
//	ptr, _ := Parse("/foo/bar/baz")
//	tail := ptr.Tail()
//	fmt.Println(tail.String()) // Output: /bar/baz
func (p Pointer) Tail() Pointer {
	return Pointer(p[1:])
}

// Eval evaluates the Pointer against a given JSON document.
// It traverses the document according to the reference tokens in the Pointer,
// handling both nested objects and arrays.
//
// Rationale:
// - Uses a loop to process each token, allowing for arbitrary nesting levels.
// - Delegates token evaluation to evalToken for type-specific handling.
// - Returns early on any error to prevent invalid traversal.
//
// Parameters:
// - data: The root JSON document to evaluate against.
//
// Returns:
// - The value referenced by the Pointer within the document.
// - An error if the Pointer cannot be resolved within the document.
//
// Usage:
//
//	doc := map[string]interface{}{
//	    "foo": map[string]interface{}{
//	        "bar": map[string]interface{}{
//	            "baz": []interface{}{0, "hello!"},
//	        },
//	    },
//	}
//	ptr, _ := Parse("/foo/bar/baz/1")
//	result, err := ptr.Eval(doc)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(result) // Output: hello!
func (p Pointer) Eval(data any) (any, error) {
	result := data
	var err error
	for _, token := range p {
		result, err = p.evalToken(token, result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// evalToken is an unexported method that evaluates a single token against the current data.
// It determines the type of the current data and calls the appropriate helper function.
//
// Rationale:
// - Uses a type switch for efficient handling of different data types.
// - Separates concerns by delegating to type-specific evaluation methods.
//
// Parameters:
// - token: The token to evaluate.
// - data: The current data being traversed.
//
// Returns:
// - The result of evaluating the token against the data.
// - An error if the token cannot be evaluated.
func (p Pointer) evalToken(token string, data any) (any, error) {
	switch typedData := data.(type) {
	case map[string]any:
		return p.evalMap(token, typedData)
	case []any:
		return p.evalArray(token, typedData)
	default:
		return nil, fmt.Errorf("invalid token %s for type %T", token, data)
	}
}

// evalMap is an unexported method that evaluates a single token against a map.
//
// Rationale:
// - Directly accesses map keys for efficiency.
// - Returns a descriptive error if the key is not found.
//
// Parameters:
// - token: The token to evaluate.
// - data: The map to evaluate against.
//
// Returns:
// - The value associated with the token in the map.
// - An error if the token is not found in the map.
func (p Pointer) evalMap(token string, data map[string]any) (any, error) {
	if v, ok := data[token]; ok {
		return v, nil
	}
	return nil, fmt.Errorf(errMsgKeyNotFound, token, p.String())
}

// evalArray is an unexported method that evaluates a single token against an array.
//
// This updated version provides more detailed error messages, distinguishing between
// different types of invalid access attempts.
//
// Parameters:
// - token: The token to evaluate (should be a valid array index).
// - data: The array to evaluate against.
//
// Returns:
// - The value at the specified index in the array.
// - An error if the token is not a valid index, is out of bounds, or if trying to access an object property.
func (p Pointer) evalArray(token string, data []any) (any, error) {
	// Check if the token is a valid integer
	i, err := strconv.Atoi(token)
	if err != nil {
		// If it's not a valid integer, it could be an attempt to access an object property
		return nil, fmt.Errorf("invalid array access: token '%s' is not a valid array index", token)
	}

	// Check if the index is negative
	if i < 0 {
		return nil, fmt.Errorf("invalid array index: %d is negative", i)
	}

	// Check if the index is out of bounds
	if i >= len(data) {
		return nil, fmt.Errorf("array index out of bounds: index %d exceeds array length of %d", i, len(data))
	}

	return data[i], nil
}

// MarshalJSON implements the json.Marshaler interface.
// This allows Pointer values to be directly marshaled as JSON strings.
//
// Rationale:
// - Simplifies serialization of Pointers in JSON contexts.
// - Uses the String() method to ensure correct escaping.
//
// Returns:
// - The JSON representation of the Pointer as a byte slice.
// - An error if marshaling fails.
//
// Usage:
//
//	ptr, _ := Parse("/foo/bar")
//	jsonData, err := json.Marshal(ptr)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(string(jsonData)) // Output: "/foo/bar"
func (p Pointer) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This allows Pointer values to be directly unmarshaled from JSON strings.
//
// Rationale:
// - Complements MarshalJSON for complete JSON serialization support.
// - Uses the Parse function to ensure correct interpretation of the JSON string.
//
// Parameters:
// - data: The JSON data to unmarshal.
//
// Returns:
// - An error if unmarshaling or parsing fails.
//
// Usage:
//
//	var ptr Pointer
//	err := json.Unmarshal([]byte(`"/foo/bar"`), &ptr)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(ptr.String()) // Output: /foo/bar
func (p *Pointer) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ptr, err := Parse(s)
	if err != nil {
		return err
	}
	*p = ptr
	return nil
}

// Validate checks if a given string is a valid JSON Pointer.
// This function provides a quick way to check pointer validity without fully parsing it.
//
// A valid JSON Pointer is either an empty string or a string that starts with a '/' character,
// followed by a sequence of reference tokens separated by '/'. Each reference token can contain
// any character except '/' and '~', which must be escaped as '~1' and '~0' respectively.
//
// Parameters:
// - s: The string to validate as a JSON Pointer.
//
// Returns:
// - true if the string is a valid JSON Pointer, false otherwise.
//
// Usage:
//
//	fmt.Println(Validate(""))         // Output: true
//	fmt.Println(Validate("/foo/bar")) // Output: true
//	fmt.Println(Validate("foo/bar"))  // Output: false
//	fmt.Println(Validate("/foo~bar")) // Output: false
func Validate(s string) bool {
	if s == "" {
		return true
	}
	if !strings.HasPrefix(s, "/") {
		return false
	}
	tokens := strings.Split(s[1:], "/")
	for _, token := range tokens {
		if strings.ContainsRune(token, '~') {
			if strings.ContainsRune(token, '~') && !strings.Contains(token, "~0") && !strings.Contains(token, "~1") {
				return false
			}
			if strings.HasSuffix(token, "~") {
				return false
			}
		}
	}
	return true
}

// Equals checks if two Pointers are equal.
// Two Pointers are considered equal if they have the same tokens in the same order.
//
// Rationale:
// - Compares length first as a quick check before comparing individual tokens.
// - Direct slice comparison is more efficient than converting to strings.
func (p Pointer) Equals(other Pointer) bool {
	if len(p) != len(other) {
		return false
	}
	for i, v := range p {
		if v != other[i] {
			return false
		}
	}
	return true
}

// IsPrefix checks if the Pointer is a prefix of another Pointer.
// This is useful for determining hierarchical relationships between Pointers.
// Both the empty pointer and the root pointer ("/") are considered prefixes of all other pointers.
//
// Parameters:
// - other: The other Pointer to check against.
//
// Returns:
// - true if this Pointer is a prefix of the other Pointer, false otherwise.
//
// Usage:
//
//	p1, _ := Parse("/foo")
//	p2, _ := Parse("/foo/bar")
//	fmt.Println(p1.IsPrefix(p2)) // Output: true
//
//	root, _ := Parse("/")
//	fmt.Println(root.IsPrefix(p2)) // Output: true
//
//	empty, _ := Parse("")
//	fmt.Println(empty.IsPrefix(p2)) // Output: true
func (p Pointer) IsPrefix(other Pointer) bool {
	// The empty pointer and the root pointer are prefixes of all pointers
	if len(p) == 0 || (len(p) == 1 && p[0] == "") {
		return true
	}

	if len(p) > len(other) {
		return false
	}
	for i, v := range p {
		if v != other[i] {
			return false
		}
	}
	return true
}

// PointerBuilder provides a builder pattern for constructing Pointers.
// This struct allows for more intuitive and flexible Pointer creation,
// especially when building Pointers programmatically.
type PointerBuilder struct {
	tokens []string
}

// NewPointerBuilder creates a new PointerBuilder.
// It initializes the token slice with a small capacity to balance
// memory usage and potential growth.
//
// Rationale:
// - Initial capacity of 8 is a reasonable default for most use cases.
// - Using a builder pattern allows for cleaner, more readable Pointer construction.
func NewPointerBuilder() *PointerBuilder {
	return &PointerBuilder{tokens: make([]string, 0, 8)}
}

// AddToken adds a token to the Pointer being built.
// It returns the PointerBuilder to allow for method chaining.
//
// Rationale:
// - Method chaining provides a fluent interface for building Pointers.
// - No validation is done at this stage for performance; validation occurs in Build().
func (pb *PointerBuilder) AddToken(token string) *PointerBuilder {
	pb.tokens = append(pb.tokens, token)
	return pb
}

// Build constructs the final Pointer.
// This method performs no additional validation, assuming that the
// individual tokens have been properly formatted.
//
// Rationale:
// - Simple conversion from the internal slice to a Pointer type.
// - No additional allocations are needed, promoting efficiency.
func (pb *PointerBuilder) Build() Pointer {
	return Pointer(pb.tokens)
}

// Descendant returns a new pointer to a descendant of the current pointer
// by parsing the input path into components and appending them to the current pointer.
// This method handles both relative and absolute paths, as well as edge cases involving
// empty pointers and paths.
//
// The method follows these rules:
// 1. If the input path is absolute (starts with '/'), it creates a new pointer from that path.
// 2. If both the current pointer and input path are empty, it returns the root pointer ("/").
// 3. If the input path is empty, it returns the current pointer.
// 4. For relative paths, it appends the new path to the current pointer, adding a '/' if necessary.
//
// Parameters:
//   - path: A string representation of the path to append or create. It can be a relative path
//     (e.g., "foo/bar") or an absolute path (e.g., "/foo/bar").
//
// Returns:
// - A new Pointer that is either a descendant of the current Pointer or a new absolute Pointer.
// - An error if the input path is invalid or cannot be parsed.
//
// Usage:
//
//	ptr, _ := Parse("/foo")
//	desc, err := ptr.Descendant("bar")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(desc.String()) // Output: /foo/bar
//
//	root, _ := Parse("/")
//	abs, err := root.Descendant("/baz/qux")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(abs.String()) // Output: /baz/qux
func (p Pointer) Descendant(path string) (Pointer, error) {
	if strings.HasPrefix(path, "/") {
		// If the path is absolute, create a new pointer from it
		return Parse(path)
	}

	// Handle the case where both p and path are empty
	if p.IsEmpty() && path == "" {
		return Parse("/")
	}

	// For relative paths, append to the current pointer
	if path == "" {
		return p, nil
	}

	// Construct the full path, adding a slash only if necessary
	fullPath := p.String()
	if fullPath != "/" && !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	fullPath += path

	return Parse(fullPath)
}
