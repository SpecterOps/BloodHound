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

func defaultMapValue(rawValue, target any) (bool, error) {
	switch typedTarget := target.(type) {
	case *uint:
		if value, err := AsNumeric[uint](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]uint:
		if value, err := AsNumericSlice[uint](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *uint8:
		if value, err := AsNumeric[uint8](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]uint8:
		if value, err := AsNumericSlice[uint8](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *uint16:
		if value, err := AsNumeric[uint16](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]uint16:
		if value, err := AsNumericSlice[uint16](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *uint32:
		if value, err := AsNumeric[uint32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]uint32:
		if value, err := AsNumericSlice[uint32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *uint64:
		if value, err := AsNumeric[uint64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]uint64:
		if value, err := AsNumericSlice[uint64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *int:
		if value, err := AsNumeric[int](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]int:
		if value, err := AsNumericSlice[int](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *int8:
		if value, err := AsNumeric[int8](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]int8:
		if value, err := AsNumericSlice[int8](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *int16:
		if value, err := AsNumeric[int16](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]int16:
		if value, err := AsNumericSlice[int16](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *int32:
		if value, err := AsNumeric[int32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]int32:
		if value, err := AsNumericSlice[int32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *int64:
		if value, err := AsNumeric[int64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]int64:
		if value, err := AsNumericSlice[int64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *ID:
		if value, err := AsNumeric[ID](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]ID:
		if value, err := AsNumericSlice[ID](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *float32:
		if value, err := AsNumeric[float32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]float32:
		if value, err := AsNumericSlice[float32](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *float64:
		if value, err := AsNumeric[float64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *[]float64:
		if value, err := AsNumericSlice[float64](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *bool:
		if value, typeOK := rawValue.(bool); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to bool", value)
		} else {
			*typedTarget = value
		}

	case *Kind:
		if strValue, typeOK := rawValue.(string); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to string", rawValue)
		} else {
			*typedTarget = StringKind(strValue)
		}

	case *string:
		if value, typeOK := rawValue.(string); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to string", rawValue)
		} else {
			*typedTarget = value
		}

	case *[]Kind:
		if kindValues, err := AsKinds(rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = kindValues
		}

	case *Kinds:
		if kindValues, err := AsKinds(rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = kindValues
		}

	case *[]string:
		if stringValues, err := SliceOf[string](rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = stringValues
		}

	case *time.Time:
		if value, err := AsTime(rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	default:
		return false, nil
	}

	return true, nil
}

type MapFunc func(rawValue, target any) (bool, error)

type valueMapper struct {
	mapperFuncs []MapFunc
	values      []any
	idx         int
}

func NewValueMapper(values []any, mappers ...MapFunc) ValueMapper {
	return &valueMapper{
		mapperFuncs: append(mappers, defaultMapValue),
		values:      values,
		idx:         0,
	}
}

func (s *valueMapper) Next() (any, error) {
	if s.idx >= len(s.values) {
		return nil, fmt.Errorf("attempting to get more values than returned - saw %d but wanted %d", len(s.values), s.idx+1)
	}

	nextValue := s.values[s.idx]
	s.idx++

	return nextValue, nil
}

func (s *valueMapper) Map(target any) error {
	if rawValue, err := s.Next(); err != nil {
		return err
	} else {
		for _, mapperFunc := range s.mapperFuncs {
			if mapped, err := mapperFunc(rawValue, target); err != nil {
				return err
			} else if mapped {
				return nil
			}
		}
	}

	return fmt.Errorf("unsupported scan type %T", target)
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

func (s *valueMapper) MapOptions(targets ...any) (any, error) {
	if rawValue, err := s.Next(); err != nil {
		return nil, err
	} else {
		for _, target := range targets {
			for _, mapperFunc := range s.mapperFuncs {
				if mapped, _ := mapperFunc(rawValue, target); mapped {
					return target, nil
				}
			}
		}

		return nil, fmt.Errorf("no matching target given for type: %T", rawValue)
	}
}

func (s *valueMapper) Scan(targets ...any) error {
	for idx, mapValue := range targets {
		if err := s.Map(mapValue); err != nil {
			return err
		} else {
			targets[idx] = mapValue
		}
	}

	return nil
}
