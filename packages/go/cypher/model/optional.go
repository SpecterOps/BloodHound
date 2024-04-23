package model

// Optional is a simple generic optional type.
//
// See: https://en.wikipedia.org/wiki/Option_type
type Optional[T any] struct {
	Value T
	Set   bool
}

func ValueOptional[T any](value T) Optional[T] {
	return Optional[T]{
		Value: value,
		Set:   true,
	}
}

func PointerOptional[T any](value *T) Optional[T] {
	if value == nil {
		return EmptyOptional[T]()
	}

	return Optional[T]{
		Value: *value,
		Set:   true,
	}
}

func EmptyOptional[T any]() Optional[T] {
	var emptyT T

	return Optional[T]{
		Value: emptyT,
		Set:   false,
	}
}
