# jsonpointer

## Description

The `jsonpointer` package provides utilities for working with JSON Pointers as defined in [IETF RFC6901](https://tools.ietf.org/html/rfc6901). It allows parsing, evaluating, and manipulating JSON Pointers, which are strings that identify specific values within a JSON document.

Key features:

-   Parse JSON Pointers from strings, including those embedded in URLs
-   Evaluate JSON Pointers against JSON documents
-   Create and manipulate JSON Pointers programmatically
-   Validate JSON Pointer strings
-   Traverse JSON documents using JSON Pointers

## Usage

### Parsing JSON Pointers

```go
import "github.com/bloodhound/jsonpointer"

// Parse a simple JSON Pointer
ptr, err := jsonpointer.Parse("/foo/bar")
if err != nil {
    // Handle error
}

// Parse a JSON Pointer from a URL
urlPtr, err := jsonpointer.Parse("http://example.com/data#/foo/bar")
if err != nil {
    // Handle error
}
```

### Evaluating JSON Pointers

```go
doc := map[string]interface{}{
    "foo": map[string]interface{}{
        "bar": []interface{}{0, "hello!"},
    },
}

ptr, _ := jsonpointer.Parse("/foo/bar/1")
result, err := ptr.Eval(doc)
if err != nil {
    // Handle error
}
fmt.Println(result) // Output: hello!
```

### Creating Pointers Programmatically

```go
builder := jsonpointer.NewPointerBuilder()
ptr := builder.AddToken("foo").AddToken("bar").AddToken("0").Build()
fmt.Println(ptr.String()) // Output: /foo/bar/0
```

### Validating JSON Pointer Strings

```go
fmt.Println(jsonpointer.Validate("/foo/bar"))  // Output: true
fmt.Println(jsonpointer.Validate("foo/bar"))   // Output: false
fmt.Println(jsonpointer.Validate("/foo~bar"))  // Output: false
```

### Traversing JSON Documents

```go
doc := map[string]interface{}{
    "foo": []interface{}{1, 2, 3},
    "bar": map[string]interface{}{"baz": "qux"},
}

err := jsonpointer.WalkJSON(doc, func(elem interface{}) error {
    fmt.Printf("Visited: %v\n", elem)
    return nil
})
if err != nil {
    // Handle error
}
```

### Checking Pointer Relationships

```go
p1, _ := jsonpointer.Parse("/foo")
p2, _ := jsonpointer.Parse("/foo/bar")

fmt.Println(p1.IsPrefix(p2)) // Output: true

root, _ := jsonpointer.Parse("/")
fmt.Println(root.IsPrefix(p2)) // Output: true
```
