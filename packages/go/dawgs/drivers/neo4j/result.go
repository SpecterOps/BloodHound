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

package neo4j

import (
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func asUint8(value any) (uint8, error) {
	switch typedValue := value.(type) {
	case uint8:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint8", value)
	}
}

func asUint16(value any) (uint16, error) {
	switch typedValue := value.(type) {
	case uint8:
		return uint16(typedValue), nil
	case uint16:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint16", value)
	}
}

func asUint32(value any) (uint32, error) {
	switch typedValue := value.(type) {
	case uint8:
		return uint32(typedValue), nil
	case uint16:
		return uint32(typedValue), nil
	case uint32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint32", value)
	}
}

func asUint64(value any) (uint64, error) {
	switch typedValue := value.(type) {
	case uint:
		return uint64(typedValue), nil
	case uint8:
		return uint64(typedValue), nil
	case uint16:
		return uint64(typedValue), nil
	case uint32:
		return uint64(typedValue), nil
	case uint64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint64", value)
	}
}

func asUint(value any) (uint, error) {
	switch typedValue := value.(type) {
	case uint:
		return typedValue, nil
	case uint8:
		return uint(typedValue), nil
	case uint16:
		return uint(typedValue), nil
	case uint32:
		return uint(typedValue), nil
	case uint64:
		return uint(typedValue), nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to uint", value)
	}
}

func asInt8(value any) (int8, error) {
	switch typedValue := value.(type) {
	case int8:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int8", value)
	}
}

func asInt16(value any) (int16, error) {
	switch typedValue := value.(type) {
	case int8:
		return int16(typedValue), nil
	case int16:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int16", value)
	}
}

func asInt32(value any) (int32, error) {
	switch typedValue := value.(type) {
	case int8:
		return int32(typedValue), nil
	case int16:
		return int32(typedValue), nil
	case int32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int32", value)
	}
}

func asInt64(value any) (int64, error) {
	switch typedValue := value.(type) {
	case graph.ID:
		return int64(typedValue), nil
	case int:
		return int64(typedValue), nil
	case int8:
		return int64(typedValue), nil
	case int16:
		return int64(typedValue), nil
	case int32:
		return int64(typedValue), nil
	case int64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func asInt(value any) (int, error) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, nil
	case int8:
		return int(typedValue), nil
	case int16:
		return int(typedValue), nil
	case int32:
		return int(typedValue), nil
	case int64:
		return int(typedValue), nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int", value)
	}
}

func asFloat32(value any) (float32, error) {
	switch typedValue := value.(type) {
	case float32:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func asFloat64(value any) (float64, error) {
	switch typedValue := value.(type) {
	case float32:
		return float64(typedValue), nil
	case float64:
		return typedValue, nil
	default:
		return 0, fmt.Errorf("unexecpted type %T will not negotiate to int64", value)
	}
}

func asTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case string:
		if parsedTime, err := time.Parse(time.RFC3339Nano, typedValue); err != nil {
			return time.Time{}, err
		} else {
			return parsedTime, nil
		}

	case dbtype.Time:
		return typedValue.Time(), nil

	case dbtype.LocalTime:
		return typedValue.Time(), nil

	case dbtype.Date:
		return typedValue.Time(), nil

	case dbtype.LocalDateTime:
		return typedValue.Time(), nil

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

func mapValue(target, rawValue any) error {
	switch typedTarget := target.(type) {
	case *uint:
		if value, err := asUint(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint8:
		if value, err := asUint8(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint16:
		if value, err := asUint16(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint32:
		if value, err := asUint32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *uint64:
		if value, err := asUint64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int:
		if value, err := asInt(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int8:
		if value, err := asInt8(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int16:
		if value, err := asInt16(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int32:
		if value, err := asInt32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *int64:
		if value, err := asInt64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *graph.ID:
		if value, err := asInt64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = graph.ID(value)
		}

	case *float32:
		if value, err := asFloat32(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *float64:
		if value, err := asFloat64(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *bool:
		if value, typeOK := rawValue.(bool); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to bool", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Kind:
		if strValue, typeOK := rawValue.(string); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to string", rawValue)
		} else {
			*typedTarget = graph.StringKind(strValue)
		}

	case *string:
		if value, typeOK := rawValue.(string); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to string", rawValue)
		} else {
			*typedTarget = value
		}

	case *[]graph.Kind:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if kindValues, err := anySliceToKinds(rawValues); err != nil {
			return err
		} else {
			*typedTarget = kindValues
		}

	case *graph.Kinds:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if kindValues, err := anySliceToKinds(rawValues); err != nil {
			return err
		} else {
			*typedTarget = kindValues
		}

	case *[]string:
		if rawValues, typeOK := rawValue.([]any); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to []any", rawValue)
		} else if stringValues, err := anySliceToStringSlice(rawValues); err != nil {
			return err
		} else {
			*typedTarget = stringValues
		}

	case *time.Time:
		if value, err := asTime(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *dbtype.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = *newRelationship(value)
		}

	case *dbtype.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = *newNode(value)
		}

	case *graph.Path:
		if value, typeOK := rawValue.(dbtype.Path); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Path", rawValue)
		} else {
			*typedTarget = newPath(value)
		}

	default:
		return fmt.Errorf("unsupported scan type %T", target)
	}

	return nil
}

type ValueMapper struct {
	values []any
	idx    int
}

func NewValueMapper(values []any) *ValueMapper {
	return &ValueMapper{
		values: values,
		idx:    0,
	}
}

func (s *ValueMapper) Next() (any, error) {
	if s.idx >= len(s.values) {
		return nil, fmt.Errorf("attempting to get more values than returned - saw %d but wanted %d", len(s.values), s.idx+1)
	}

	nextValue := s.values[s.idx]
	s.idx++

	return nextValue, nil
}

func (s *ValueMapper) Map(target any) error {
	if rawValue, err := s.Next(); err != nil {
		return err
	} else {
		return mapValue(target, rawValue)
	}
}

func (s *ValueMapper) MapOptions(targets ...any) (any, error) {
	if rawValue, err := s.Next(); err != nil {
		return nil, err
	} else {
		for _, target := range targets {
			if mapValue(target, rawValue) == nil {
				return target, nil
			}
		}

		return nil, fmt.Errorf("no matching target given for type: %T", rawValue)
	}
}

func (s *ValueMapper) Scan(targets ...any) error {
	for idx, mapValue := range targets {
		if err := s.Map(mapValue); err != nil {
			return err
		} else {
			targets[idx] = mapValue
		}
	}

	return nil
}

type internalResult struct {
	query        string
	err          error
	driverResult neo4j.Result
}

func NewResult(query string, err error, driverResult neo4j.Result) graph.Result {
	return &internalResult{
		query:        query,
		err:          err,
		driverResult: driverResult,
	}
}

func anySliceToStringSlice(rawValues []any) ([]string, error) {
	strings := make([]string, len(rawValues))

	for idx, rawValue := range rawValues {
		switch typedValue := rawValue.(type) {
		case string:
			strings[idx] = typedValue
		default:
			return nil, fmt.Errorf("unexpected type %T will not negotiate to string", rawValue)
		}
	}

	return strings, nil
}

func anySliceToKinds(rawValues []any) (graph.Kinds, error) {
	if stringValues, err := anySliceToStringSlice(rawValues); err != nil {
		return nil, err
	} else {
		return graph.StringsToKinds(stringValues), nil
	}
}

func (s *internalResult) Values() graph.ValueMapper {
	return NewValueMapper(s.driverResult.Record().Values)
}

func (s *internalResult) Scan(targets ...any) error {
	return s.Values().Scan(targets...)
}

func (s *internalResult) Next() bool {
	return s.driverResult.Next()
}

func (s *internalResult) Error() error {
	if s.err != nil {
		return s.err
	}

	if s.driverResult != nil && s.driverResult.Err() != nil {
		return graph.NewError(s.query, s.driverResult.Err())
	}

	return nil
}

func (s *internalResult) Close() {
	// Ignore the results of this call. This is called only as a best-effort attempt at a close
	s.driverResult.Consume()
}
