// Package jsonpointer provides utilities for working with JSON Pointers as defined in RFC 6901.
// It allows parsing, evaluating, and manipulating JSON Pointers, which are strings that identify
// specific values within a JSON document.
//
// The implementation focuses on efficiency, flexibility, and ease of use. It aims to provide
// a balance between performance and maintainability, with clear error handling and
// extensibility for future enhancements.
package jsonpointer

import (
	"reflect"
)

// WalkJSON traverses a JSON-like structure and applies a visitor function to each element.
// It supports map[string]any, []any, and uses reflection for other types.
//
// The function aims to balance performance with flexibility, providing fast-paths for common JSON types.
//
// Parameters:
// - tree: The root of the JSON-like structure to traverse.
// - visit: A function to apply to each element in the structure.
//
// Returns:
// - An error if the visitor function returns an error, or nil if the traversal is successful.
//
// Usage:
//
//	data := map[string]any{
//	    "foo": []any{1, 2, 3},
//	    "bar": map[string]any{"baz": "qux"},
//	}
//	err := WalkJSON(data, func(elem any) error {
//	    fmt.Printf("Visited: %v\n", elem)
//	    return nil
//	})
//	if err != nil {
//	    // Handle error
//	}
func WalkJSON(tree any, visit func(elem any) error) error {
	if tree == nil {
		return nil
	}

	if err := visit(tree); err != nil {
		return err
	}

	switch t := tree.(type) {
	case map[string]any:
		return walkMap(t, visit)
	case []any:
		return walkSlice(t, visit)
	default:
		return walkReflect(reflect.ValueOf(tree), visit)
	}
}

// walkMap is an unexported helper function that walks through a map[string]any.
// It iterates over all values in the map and recursively calls WalkJSON on each.
//
// Parameters:
// - m: The map to walk through.
// - visit: The visitor function to apply to each element.
//
// Returns:
// - An error if any call to WalkJSON returns an error, or nil if successful.
func walkMap(m map[string]any, visit func(elem any) error) error {
	for _, val := range m {
		if err := WalkJSON(val, visit); err != nil {
			return err
		}
	}
	return nil
}

// walkSlice is an unexported helper function that walks through a []any.
// It iterates over all elements in the slice and recursively calls WalkJSON on each.
//
// Parameters:
// - s: The slice to walk through.
// - visit: The visitor function to apply to each element.
//
// Returns:
// - An error if any call to WalkJSON returns an error, or nil if successful.
func walkSlice(s []any, visit func(elem any) error) error {
	for _, val := range s {
		if err := WalkJSON(val, visit); err != nil {
			return err
		}
	}
	return nil
}

// walkReflect is an unexported helper function that uses reflection to walk through complex types.
// It handles pointers, maps, structs, slices, and arrays.
//
// This function is used as a fallback for types that don't match the more specific cases in WalkJSON.
// It uses reflection to inspect the structure of the value and traverse it appropriately.
//
// Parameters:
// - value: The reflect.Value to walk through.
// - visit: The visitor function to apply to each element.
//
// Returns:
// - An error if any recursive call returns an error, or nil if successful.
func walkReflect(value reflect.Value, visit func(elem any) error) error {
	switch value.Kind() {
	case reflect.Ptr:
		if !value.IsNil() {
			return walkReflect(value.Elem(), visit)
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if err := walkReflectValue(value.MapIndex(key), visit); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if err := walkReflectValue(value.Field(i), visit); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := walkReflectValue(value.Index(i), visit); err != nil {
				return err
			}
		}
	}
	return nil
}

// walkReflectValue is an unexported helper function that safely walks a reflect.Value.
// It checks if the value can be converted to an interface{} before calling WalkJSON.
//
// This function is used to safely handle reflect.Value instances that may or may not
// be convertible to interface{} (e.g., unexported struct fields).
//
// Parameters:
// - value: The reflect.Value to walk.
// - visit: The visitor function to apply to each element.
//
// Returns:
// - An error if WalkJSON returns an error, or nil if successful or if the value cannot be interfaced.
func walkReflectValue(value reflect.Value, visit func(elem any) error) error {
	if value.CanInterface() {
		return WalkJSON(value.Interface(), visit)
	}
	return nil
}
