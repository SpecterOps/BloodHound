package pgsql

import "fmt"

func SliceAs[T any, TS []T, F any, FS []F](fs FS) (TS, error) {
	ts := make(TS, len(fs))

	for idx := 0; idx < len(fs); idx++ {
		if tTyped, isTType := any(fs[idx]).(T); !isTType {
			var emptyT T
			return nil, fmt.Errorf("slice type %T does not convert to %T", fs[idx], emptyT)
		} else {
			ts[idx] = tTyped
		}
	}

	return ts, nil
}

func MustSliceAs[T any, TS []T, F any, FS []F](fs FS) TS {
	if ts, err := SliceAs[T](fs); err != nil {
		panic(err.Error())
	} else {
		return ts
	}
}

func MustSliceAllAs[T any, TS []T, F any](fs ...F) TS {
	if ts, err := SliceAs[T](fs); err != nil {
		panic(err.Error())
	} else {
		return ts
	}
}
