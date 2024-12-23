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
	// UnsetDataType represents a DataType that has not been visited by any logic. It is the default, zero-value for
	// the DataType type.
	UnsetDataType DataType = ""

	// UnknownDataType represents a DataType that has been visited by type inference logic but remains unknowable.
	UnknownDataType DataType = "unknown"

	Null          DataType = "null"
	Any           DataType = "any"
	NodeComposite DataType = "nodecomposite"
	EdgeComposite DataType = "edgecomposite"
	PathComposite DataType = "pathcomposite"
	Int           DataType = "int"
	Int2          DataType = "int2"
	Int4          DataType = "int4"
	Int8          DataType = "int8"
	Float4        DataType = "float4"
	Float8        DataType = "float8"
	Boolean       DataType = "bool"
	Text          DataType = "text"
	JSONB         DataType = "jsonb"
	Numeric       DataType = "numeric"

	AnyArray           DataType = "any[]"
	NodeCompositeArray DataType = "nodecomposite[]"
	EdgeCompositeArray DataType = "edgecomposite[]"
	IntArray           DataType = "int[]"
	Int2Array          DataType = "int2[]"
	Int4Array          DataType = "int4[]"
	Int8Array          DataType = "int8[]"
	Float4Array        DataType = "float4[]"
	Float8Array        DataType = "float8[]"
	TextArray          DataType = "text[]"
	JSONBArray         DataType = "jsonb[]"
	NumericArray       DataType = "numeric[]"

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

func (s DataType) IsComparable(other DataType, operator Operator) bool {
	switch operator {
	case OperatorPGArrayOverlap, OperatorArrayOverlap:
		if !s.IsArrayType() || !other.IsArrayType() {
			return false
		}

		return s == other

	case OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo:
		switch s {
		case NodeComposite, EdgeComposite, PathComposite, JSONB, AnyArray, Text, Boolean,
			IntArray, Int8Array, Int4Array, Int2Array, Float8Array, Float4Array, NumericArray,
			Date, TimeWithTimeZone, TimeWithoutTimeZone, Interval, TimestampWithTimeZone, TimestampWithoutTimeZone:
			return other == s

		case Int, Int8, Int4, Int2:
			switch other {
			case Int, Int8, Int4, Int2, Float8, Float4, Numeric:
				return true

			default:
				return false
			}

		case Float8, Float4, Numeric:
			switch other {
			case Int, Int8, Int4, Int2, Float8, Float4, Numeric:
				return true

			default:
				return false
			}

		default:
			return false
		}

	case OperatorLike, OperatorILike, OperatorSimilarTo, OperatorRegexMatch:
		switch s {
		case Text:
			return other == s
		default:
			return false
		}

	default:
		return false
	}
}

// CoerceToSupertype attempts to take the super of the type s and the type other
func (s DataType) CoerceToSupertype(other DataType) (DataType, bool) {
	switch s {
	case Int2:
		switch other {
		case Int, Int8, Int4, Int2:
			return other, true
		}

	case Int4:
		switch other {
		case Int2:
			return s, true

		case Int, Int8, Int4:
			return other, true
		}

	case Int8:
		switch other {
		case Int2, Int4:
			return s, true

		case Int, Int8:
			return other, true
		}

	case Int:
		switch other {
		case Int, Int8:
			return other, true
		}

	case Float4:
		switch other {
		case Float4, Float8, Numeric:
			return other, true
		}

	case Float8:
		switch other {
		case Float4:
			return s, true

		case Float8, Numeric:
			return other, true
		}

	case Numeric:
		switch other {
		case Float4, Float8, Int8, Int4, Int2:
			return s, true

		case Numeric:
			return other, true
		}
	}

	return UnknownDataType, false
}

func (s DataType) OperatorResultType(other DataType, operator Operator) (DataType, bool) {
	if OperatorIsComparator(operator) && s.IsComparable(other, operator) {
		return Boolean, true
	}

	// Validate all other supported operators for result type inference
	switch operator {
	case OperatorAnd, OperatorOr:
		return Boolean, true

	case OperatorAdd, OperatorSubtract, OperatorMultiply, OperatorDivide:
		if s == other {
			return s, true
		}

		if supertype, validSupertype := s.CoerceToSupertype(other); validSupertype {
			return supertype, true
		}

	case OperatorConcatenate:
		// Array types may only concatenate if their base types match
		if s.IsArrayType() {
			return s, s == other || s.ArrayBaseType() == other
		}

		if other.IsArrayType() {
			return other, s == other || s == other.ArrayBaseType()
		}

		switch s {
		case UnknownDataType:
			// Overwrite the unknown data type here and assume that it will resolve correctly
			return other, true

		case Text:
			switch other {
			case UnknownDataType:
				// Overwrite the unknown data type here and assume that it will resolve to text
				return s, true

			default:
				return s, s == other
			}

		default:
			return UnknownDataType, false
		}
	}

	return UnknownDataType, false
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
	case Int2Array, Int4Array, Int8Array, IntArray, Float4Array, Float8Array, TextArray, JSONBArray,
		NodeCompositeArray, EdgeCompositeArray, NumericArray:
		return true
	}

	return false
}

func (s DataType) ToArrayType() (DataType, error) {
	switch s {
	case Int2, Int2Array:
		return Int2Array, nil
	case Int4, Int4Array:
		return Int4Array, nil
	case Int8, Int8Array:
		return Int8Array, nil
	case Int, IntArray:
		return IntArray, nil
	case Any, AnyArray:
		return AnyArray, nil
	case JSONB, JSONBArray:
		return JSONBArray, nil
	case NodeComposite, NodeCompositeArray:
		return NodeCompositeArray, nil
	case EdgeComposite, EdgeCompositeArray:
		return EdgeCompositeArray, nil
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

func (s DataType) ArrayBaseType() DataType {
	switch s {
	case Int2Array:
		return Int2
	case Int4Array:
		return Int4
	case Int8Array:
		return Int8
	case Float4Array:
		return Float4
	case Float8Array:
		return Float8
	case TextArray:
		return Text
	case NumericArray:
		return Numeric
	case JSONBArray:
		return JSONB
	case AnyArray:
		return Any
	case NodeCompositeArray:
		return NodeComposite
	case EdgeCompositeArray:
		return EdgeComposite
	default:
		return s
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
