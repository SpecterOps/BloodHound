// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"fmt"
	"strconv"
	"time"
)

type numeric interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64 | ID
}

func SliceOf[T any](raw any) ([]T, error) {
	if slice, typeOK := raw.([]any); !typeOK {
		return nil, fmt.Errorf("expected []any slice but received %T", raw)
	} else {
		sliceCopy := make([]T, len(slice))

		for idx, sliceValue := range slice {
			if typedSliceValue, typeOK := sliceValue.(T); !typeOK {
				var empty T
				return nil, fmt.Errorf("expected type %T but received %T", empty, sliceValue)
			} else {
				sliceCopy[idx] = typedSliceValue
			}
		}

		return sliceCopy, nil
	}
}

func AsNumeric[T numeric](rawValue any) (T, error) {
	var empty T

	switch typedValue := rawValue.(type) {
	case uint:
		return T(typedValue), nil
	case uint8:
		return T(typedValue), nil
	case uint16:
		return T(typedValue), nil
	case uint32:
		return T(typedValue), nil
	case uint64:
		return T(typedValue), nil
	case int:
		return T(typedValue), nil
	case int8:
		return T(typedValue), nil
	case int16:
		return T(typedValue), nil
	case int32:
		return T(typedValue), nil
	case int64:
		return T(typedValue), nil
	case float32:
		return T(typedValue), nil
	case float64:
		return T(typedValue), nil
	case string:
		if parsedInt, err := strconv.ParseInt(typedValue, 10, 64); err != nil {
			if parsedFloat, err := strconv.ParseFloat(typedValue, 64); err != nil {
				return empty, fmt.Errorf("unable to parse numeric value from raw value %s", typedValue)
			} else {
				return T(parsedFloat), nil
			}
		} else {
			return T(parsedInt), nil
		}

	default:
		return empty, fmt.Errorf("unable to convert raw value %T as numeric", rawValue)
	}
}

func castNumericSlice[R numeric, T any](src []T) ([]R, error) {
	dst := make([]R, len(src))

	for idx, srcValue := range src {
		if numericValue, err := AsNumeric[R](srcValue); err != nil {
			return nil, err
		} else {
			dst[idx] = numericValue
		}
	}

	return dst, nil
}

func AsNumericSlice[T numeric](rawValue any) ([]T, error) {
	switch typedValue := rawValue.(type) {
	case []any:
		return castNumericSlice[T](typedValue)
	case []uint:
		return castNumericSlice[T](typedValue)
	case []uint8:
		return castNumericSlice[T](typedValue)
	case []uint16:
		return castNumericSlice[T](typedValue)
	case []uint32:
		return castNumericSlice[T](typedValue)
	case []uint64:
		return castNumericSlice[T](typedValue)
	case []int:
		return castNumericSlice[T](typedValue)
	case []int8:
		return castNumericSlice[T](typedValue)
	case []int16:
		return castNumericSlice[T](typedValue)
	case []int32:
		return castNumericSlice[T](typedValue)
	case []int64:
		return castNumericSlice[T](typedValue)
	case []float32:
		return castNumericSlice[T](typedValue)
	case []float64:
		return castNumericSlice[T](typedValue)
	default:
		return nil, fmt.Errorf("unable to convert raw value %T as a numeric slice", rawValue)
	}
}

func AsKinds(rawValue any) (Kinds, error) {
	if stringValues, err := SliceOf[string](rawValue); err != nil {
		return nil, err
	} else {
		return StringsToKinds(stringValues), nil
	}
}

func AsTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case string:
		if parsedTime, err := time.Parse(time.RFC3339Nano, typedValue); err != nil {
			return time.Time{}, err
		} else {
			return parsedTime, nil
		}

	case float64:
		return time.Unix(int64(typedValue), 0), nil

	case int64:
		return time.Unix(typedValue, 0), nil

	case time.Time:
		return typedValue, nil

	default:
		return time.Time{}, fmt.Errorf("unexecpted type %T will not negotiate to time.Time", value)
	}
}

func defaultMapValue(rawValue, target any) bool {
	switch typedTarget := target.(type) {
	case *uint:
		if value, err := AsNumeric[uint](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]uint:
		if value, err := AsNumericSlice[uint](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *uint8:
		if value, err := AsNumeric[uint8](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]uint8:
		if value, err := AsNumericSlice[uint8](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *uint16:
		if value, err := AsNumeric[uint16](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]uint16:
		if value, err := AsNumericSlice[uint16](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *uint32:
		if value, err := AsNumeric[uint32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]uint32:
		if value, err := AsNumericSlice[uint32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *uint64:
		if value, err := AsNumeric[uint64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]uint64:
		if value, err := AsNumericSlice[uint64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *int:
		if value, err := AsNumeric[int](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]int:
		if value, err := AsNumericSlice[int](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *int8:
		if value, err := AsNumeric[int8](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]int8:
		if value, err := AsNumericSlice[int8](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *int16:
		if value, err := AsNumeric[int16](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]int16:
		if value, err := AsNumericSlice[int16](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *int32:
		if value, err := AsNumeric[int32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]int32:
		if value, err := AsNumericSlice[int32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *int64:
		if value, err := AsNumeric[int64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]int64:
		if value, err := AsNumericSlice[int64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *ID:
		if value, err := AsNumeric[ID](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]ID:
		if value, err := AsNumericSlice[ID](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *float32:
		if value, err := AsNumeric[float32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]float32:
		if value, err := AsNumericSlice[float32](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *float64:
		if value, err := AsNumeric[float64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *[]float64:
		if value, err := AsNumericSlice[float64](rawValue); err == nil {
			*typedTarget = value
			return true
		}

	case *bool:
		if value, typeOK := rawValue.(bool); typeOK {
			*typedTarget = value
			return true
		}

	case *Kind:
		if strValue, typeOK := rawValue.(string); typeOK {
			*typedTarget = StringKind(strValue)
			return true
		}

	case *string:
		if value, typeOK := rawValue.(string); typeOK {
			*typedTarget = value
			return true
		}

	case *[]Kind:
		if kindValues, err := AsKinds(rawValue); err == nil {
			*typedTarget = kindValues
			return true
		}

	case *Kinds:
		if kindValues, err := AsKinds(rawValue); err == nil {
			*typedTarget = kindValues
			return true
		}

	case *[]string:
		if stringValues, err := SliceOf[string](rawValue); err == nil {
			*typedTarget = stringValues
			return true
		}

	case *time.Time:
		if value, err := AsTime(rawValue); err == nil {
			*typedTarget = value
			return true
		}
	}

	// Unsupported by this mapper
	return false
}

type MapFunc func(rawValue, target any) bool

type ValueMapper struct {
	mapperFuncs []MapFunc
}

func NewValueMapper(mappers ...MapFunc) ValueMapper {
	return ValueMapper{
		mapperFuncs: append(mappers, defaultMapValue),
	}
}

func (s ValueMapper) TryMap(value, target any) bool {
	for _, mapperFunc := range s.mapperFuncs {
		if mapperFunc(value, target) {
			return true
		}
	}

	return false
}
