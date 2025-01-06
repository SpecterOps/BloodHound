// Copyright 2024 Specter Ops, Inc.
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

package pgsql

import (
	"errors"
	"fmt"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	ErrNoAvailableArrayDataType = errors.New("data type has no direct array representation")
	ErrNonArrayDataType         = errors.New("data type is not an array type")
)

const (
	TableNode Identifier = "node"
	TableEdge Identifier = "edge"

	ColumnID         Identifier = "id"
	ColumnPath       Identifier = "path"
	ColumnProperties Identifier = "properties"
	ColumnKindIDs    Identifier = "kind_ids"
	ColumnKindID     Identifier = "kind_id"
	ColumnGraphID    Identifier = "graph_id"
	ColumnStartID    Identifier = "start_id"
	ColumnNextID     Identifier = "next_id"
	ColumnEndID      Identifier = "end_id"
)

var (
	NodeTableColumns = []Identifier{
		ColumnID,
		ColumnKindIDs,
		ColumnProperties,
	}

	EdgeTableColumns = []Identifier{
		ColumnID,
		ColumnStartID,
		ColumnEndID,
		ColumnKindID,
		ColumnProperties,
	}
)

type DataType string

func (s DataType) NodeType() string {
	return "data_type"
}

const (
	UnsetDataType            DataType = ""
	UnknownDataType          DataType = "UNKNOWN"
	Reference                DataType = "REFERENCE"
	Null                     DataType = "NULL"
	NodeComposite            DataType = "nodecomposite"
	NodeCompositeArray       DataType = "nodecomposite[]"
	EdgeComposite            DataType = "edgecomposite"
	EdgeCompositeArray       DataType = "edgecomposite[]"
	PathComposite            DataType = "pathcomposite"
	Int                      DataType = "int"
	IntArray                 DataType = "int[]"
	Int2                     DataType = "int2"
	Int2Array                DataType = "int2[]"
	Int4                     DataType = "int4"
	Int4Array                DataType = "int4[]"
	Int8                     DataType = "int8"
	Int8Array                DataType = "int8[]"
	Float4                   DataType = "float4"
	Float4Array              DataType = "float4[]"
	Float8                   DataType = "float8"
	Float8Array              DataType = "float8[]"
	Boolean                  DataType = "bool"
	Text                     DataType = "text"
	TextArray                DataType = "text[]"
	JSONB                    DataType = "jsonb"
	JSONBArray               DataType = "jsonb[]"
	Numeric                  DataType = "numeric"
	NumericArray             DataType = "numeric[]"
	Date                     DataType = "date"
	TimeWithTimeZone         DataType = "time with time zone"
	TimeWithoutTimeZone      DataType = "time without time zone"
	Interval                 DataType = "interval"
	TimestampWithTimeZone    DataType = "timestamp with time zone"
	TimestampWithoutTimeZone DataType = "timestamp without time zone"

	Scope                 DataType = "scope"
	ParameterIdentifier   DataType = "parameter_identifier"
	ExpansionPattern      DataType = "expansion_pattern"
	ExpansionPath         DataType = "expansion_path"
	ExpansionRootNode     DataType = "expansion_root_node"
	ExpansionEdge         DataType = "expansion_edge"
	ExpansionTerminalNode DataType = "expansion_terminal_node"
)

func (s DataType) IsKnown() bool {
	switch s {
	case UnsetDataType, UnknownDataType:
		return false

	default:
		return true
	}
}

// TODO: operator, while unused, is part of a refactor for this function to make it operator aware
func (s DataType) Compatible(other DataType, operator Operator) (DataType, bool) {
	if s == other {
		return s, true
	}

	if other == UnknownDataType {
		// Assume unknown data types will offload type matching to the DB
		return s, true
	}

	switch s {
	case UnknownDataType:
		// Assume unknown data types will offload type matching to the DB
		return other, true

	case Text:
		return Text, true

	case Float4:
		switch other {
		case Float8:
			return Float8, true

		case Float4Array:
			return Float4, true

		case Float8Array:
			return Float8, true

		case Text:
			return Text, true
		}

	case Float8:
		switch other {
		case Float4:
			return Float8, true

		case Float4Array, Float8Array:
			return Float8, true

		case Text:
			return Text, true
		}

	case Numeric:
		switch other {
		case Float4, Float8, Int2, Int4, Int8:
			return Numeric, true

		case Float4Array, Float8Array, NumericArray:
			return Numeric, true

		case Text:
			return Text, true
		}

	case Int2:
		switch other {
		case Int2:
			return Int2, true

		case Int4:
			return Int4, true

		case Int8:
			return Int8, true

		case Int2Array:
			return Int2, true

		case Int4Array:
			return Int4, true

		case Int8Array:
			return Int8, true

		case Text:
			return Text, true
		}

	case Int4:
		switch other {
		case Int2, Int4:
			return Int4, true

		case Int8:
			return Int8, true

		case Int2Array, Int4Array:
			return Int4, true

		case Int8Array:
			return Int8, true

		case Text:
			return Text, true
		}

	case Int8:
		switch other {
		case Int2, Int4, Int8:
			return Int8, true

		case Int2Array, Int4Array, Int8Array:
			return Int8, true

		case Text:
			return Text, true
		}

	case Int:
		switch other {
		case Int2, Int4, Int:
			return Int, true

		case Int8:
			return Int8, true

		case Text:
			return Text, true
		}

	case Int2Array:
		switch other {
		case Int2Array, Int4Array, Int8Array:
			return other, true
		}

	case Int4Array:
		switch other {
		case Int4Array, Int8Array:
			return other, true
		}

	case Float4Array:
		switch other {
		case Float4Array, Float8Array:
			return other, true
		}
	}

	return UnsetDataType, false
}

func (s DataType) TextConvertable() bool {
	switch s {
	case TimestampWithoutTimeZone, TimestampWithTimeZone, TimeWithoutTimeZone, TimeWithTimeZone, Date, Text:
		return true

	default:
		return false
	}
}

func (s DataType) MatchesOneOf(others ...DataType) bool {
	for _, other := range others {
		if s == other {
			return true
		}
	}

	return false
}

func (s DataType) IsArrayType() bool {
	switch s {
	case Int2Array, Int4Array, Int8Array, Float4Array, Float8Array, TextArray, JSONBArray, NodeCompositeArray, EdgeCompositeArray, NumericArray:
		return true
	}

	return false
}

func (s DataType) ToUpdateResultType() (DataType, error) {
	switch s {
	case NodeComposite:
		return s, nil
	case EdgeComposite:
		return s, nil
	default:
		return UnsetDataType, fmt.Errorf("data type %s has no update result representation", s)
	}
}

func (s DataType) ToArrayType() (DataType, error) {
	switch s {
	case Int2, Int2Array:
		return Int2Array, nil
	case Int4, Int4Array:
		return Int4Array, nil
	case Int8, Int8Array:
		return Int8Array, nil
	case Float4, Float4Array:
		return Float4Array, nil
	case Float8, Float8Array:
		return Float8Array, nil
	case Text, TextArray:
		return TextArray, nil
	case Numeric, NumericArray:
		return NumericArray, nil
	default:
		return UnknownDataType, ErrNoAvailableArrayDataType
	}
}

func (s DataType) ArrayBaseType() (DataType, error) {
	switch s {
	case Int2Array:
		return Int2, nil
	case Int4Array:
		return Int4, nil
	case Int8Array:
		return Int8, nil
	case Float4Array:
		return Float4, nil
	case Float8Array:
		return Float8, nil
	case TextArray:
		return Text, nil
	case NumericArray:
		return Numeric, nil
	default:
		return s, nil
	}
}

func (s DataType) String() string {
	return string(s)
}

var CompositeTypes = []DataType{NodeComposite, NodeCompositeArray, EdgeComposite, EdgeCompositeArray, PathComposite}

func NegotiateValue(value any) (any, error) {
	switch typedValue := value.(type) {
	case graph.ID:
		return typedValue.Uint64(), nil

	case []graph.ID:
		return graph.IDsToUint64Slice(typedValue), nil

	default:
		return value, nil
	}
}

func ValueToDataType(value any) (DataType, error) {
	switch typedValue := value.(type) {
	case time.Time:
		if typedValue.Location() != nil && typedValue.Location().String() != time.Local.String() {
			return TimestampWithTimeZone, nil
		}

		return TimestampWithoutTimeZone, nil

	case time.Duration:
		return Interval, nil

	// * uint8 is here since it can't fit in a signed byte and therefore must coerce into a higher sized type
	case uint8, int8, int16:
		return Int2, nil

	// * uint8 is here since it can't fit in a signed byte and therefore must coerce into a higher sized type
	case []uint8, []int8, []int16:
		return Int2Array, nil

	// * uint16 is here since it can't fit in a signed 16-bit value and therefore must coerce into a higher sized type
	case uint16, int32:
		return Int4, nil

	// * uint16 is here since it can't fit in a signed 16-bit value and therefore must coerce into a higher sized type
	case []uint16, []int32:
		return Int4Array, nil

	// * uint32 is here since it can't fit in a signed 16-bit value and therefore must coerce into a higher sized type
	// * uint is here because it is architecture dependent but expecting it to be an unsigned value between 32-bits and
	//   64-bits is fine.
	// * int is here for the same reasons as uint
	case uint32, uint, uint64, int, int64, graph.ID:
		return Int8, nil

	// * uint32 is here since it can't fit in a signed 16-bit value and therefore must coerce into a higher sized type
	// * uint is here because it is architecture dependent but expecting it to be an unsigned value between 32-bits and
	//   64-bits is fine.
	// * int is here for the same reasons as uint
	case []uint32, []uint, []uint64, []int, []int64, []graph.ID:
		return Int8Array, nil

	case float32:
		return Float4, nil

	case []float32:
		return Float4Array, nil

	case float64:
		return Float8, nil

	case []float64:
		return Float8Array, nil

	case string:
		return Text, nil

	case []string:
		return TextArray, nil

	case bool:
		return Boolean, nil

	case graph.Kind:
		return Int2, nil

	case graph.Kinds:
		return Int2Array, nil

	case []any:
		return anySliceType(typedValue)

	default:
		return UnknownDataType, fmt.Errorf("unable to map value type %T to a pgsql suitable data type", value)
	}
}

func anySliceType(slice []any) (DataType, error) {
	if len(slice) == 0 {
		return Null, nil
	}

	if expectedType, err := ValueToDataType(slice[0]); err != nil {
		return UnsetDataType, err
	} else {
		for idx, element := range slice[1:] {
			if elementType, err := ValueToDataType(element); err != nil {
				return UnsetDataType, err
			} else if expectedType != elementType {
				return UnsetDataType, fmt.Errorf("[]any slice mixes value types - expected %s but got %s for element %d", expectedType.String(), elementType.String(), idx)
			}
		}

		return expectedType.ToArrayType()
	}
}
