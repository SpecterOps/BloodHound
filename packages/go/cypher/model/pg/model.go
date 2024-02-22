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

package pg

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/specterops/bloodhound/cypher/model"
	pgModel "github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	ErrNonArrayDataType = errors.New("data type is not an array type")
)

type DataType string

const (
	UnknownDataType          DataType = "UNKNOWN"
	Reference                DataType = "REFERENCE"
	Null                     DataType = "NULL"
	Node                     DataType = "nodeComposite"
	NodeArray                DataType = "nodeComposite[]"
	Edge                     DataType = "edgeComposite"
	EdgeArray                DataType = "edgeComposite[]"
	Path                     DataType = "pathComposite"
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
	Date                     DataType = "date"
	TimeWithTimeZone         DataType = "time with time zone"
	TimeWithoutTimeZone      DataType = "time without time zone"
	Interval                 DataType = "interval"
	TimestampWithTimeZone    DataType = "timestamp with time zone"
	TimestampWithoutTimeZone DataType = "timestamp without time zone"
)

func (s DataType) IsArrayType() bool {
	switch s {
	case Int2Array, Int4Array, Int8Array, Float4Array, Float8Array, TextArray:
		return true
	}

	return false
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
	default:
		return UnknownDataType, ErrNonArrayDataType
	}
}

func (s DataType) String() string {
	return string(s)
}

var CompositeTypes = []DataType{Node, NodeArray, Edge, EdgeArray, Path}

type AnnotatedKindMatcher struct {
	model.KindMatcher
	Type DataType
}

func NewAnnotatedKindMatcher(kindMatcher *model.KindMatcher, dataType DataType) *AnnotatedKindMatcher {
	return &AnnotatedKindMatcher{
		KindMatcher: *kindMatcher,
		Type:        dataType,
	}
}

func (s *AnnotatedKindMatcher) copy() *AnnotatedKindMatcher {
	return &AnnotatedKindMatcher{
		KindMatcher: model.KindMatcher{
			Reference: s.Reference,
			Kinds:     s.Kinds,
		},
		Type: s.Type,
	}
}

type AnnotatedParameter struct {
	model.Parameter
	Type DataType
}

func NewAnnotatedParameter(parameter *model.Parameter, dataType DataType) *AnnotatedParameter {
	return &AnnotatedParameter{
		Parameter: *parameter,
		Type:      dataType,
	}
}

type Entity struct {
	Binding *AnnotatedVariable
}

func NewEntity(variable *AnnotatedVariable) *Entity {
	return &Entity{
		Binding: variable,
	}
}

type AnnotatedVariable struct {
	model.Variable
	Type DataType
}

func NewAnnotatedVariable(variable *model.Variable, dataType DataType) *AnnotatedVariable {
	return &AnnotatedVariable{
		Variable: *variable,
		Type:     dataType,
	}
}

func (s *AnnotatedVariable) copy() *AnnotatedVariable {
	if s == nil {
		return nil
	}

	return &AnnotatedVariable{
		Variable: model.Variable{
			Symbol: s.Symbol,
		},
		Type: s.Type,
	}
}

type AnnotatedPropertyLookup struct {
	model.PropertyLookup
	Type DataType
}

func NewAnnotatedPropertyLookup(propertyLookup *model.PropertyLookup, dataType DataType) *AnnotatedPropertyLookup {
	return &AnnotatedPropertyLookup{
		PropertyLookup: *propertyLookup,
		Type:           dataType,
	}
}

type AnnotatedLiteral struct {
	model.Literal
	Type DataType
}

func NewAnnotatedLiteral(literal *model.Literal, dataType DataType) *AnnotatedLiteral {
	return &AnnotatedLiteral{
		Literal: *literal,
		Type:    dataType,
	}
}

func NewStringLiteral(value string) *AnnotatedLiteral {
	return NewAnnotatedLiteral(model.NewStringLiteral(value), Text)
}

type PropertiesReference struct {
	Reference *AnnotatedVariable
}

type Subquery struct {
	PatternElements []*model.PatternElement
	Filter          model.Expression
}

type SubQueryAnnotation struct {
	FilterExpression model.Expression
}

type SQLTypeAnnotation struct {
	Type DataType
}

func NewSQLTypeAnnotationFromExpression(expression model.Expression) (*SQLTypeAnnotation, error) {
	switch typedExpression := expression.(type) {
	case *model.Parameter:
		return NewSQLTypeAnnotationFromValue(typedExpression.Value)

	case *model.Literal:
		return NewSQLTypeAnnotationFromLiteral(typedExpression)

	case *model.ListLiteral:
		var expectedTypeAnnotation *SQLTypeAnnotation

		for _, listExpressionItem := range *typedExpression {
			if listExpressionItemLiteral, isLiteral := listExpressionItem.(*model.Literal); isLiteral {
				if literalTypeAnnotation, err := NewSQLTypeAnnotationFromLiteral(listExpressionItemLiteral); err != nil {
					return nil, err
				} else if expectedTypeAnnotation != nil && expectedTypeAnnotation.Type != literalTypeAnnotation.Type {
					return nil, fmt.Errorf("list literal contains mixed types")
				} else {
					expectedTypeAnnotation = literalTypeAnnotation
				}
			}
		}

		return expectedTypeAnnotation, nil

	default:
		return nil, fmt.Errorf("unsupported expression type %T for SQL type annotation", expression)
	}
}

func NewSQLTypeAnnotationFromLiteral(literal *model.Literal) (*SQLTypeAnnotation, error) {
	if literal.Null {
		return &SQLTypeAnnotation{
			Type: Null,
		}, nil
	}

	return NewSQLTypeAnnotationFromValue(literal.Value)
}

func NewSQLTypeAnnotationFromValue(value any) (*SQLTypeAnnotation, error) {
	if value == nil {
		return &SQLTypeAnnotation{
			Type: Null,
		}, nil
	}

	switch typedValue := value.(type) {
	case []uint16, []int16, pgtype.Int2Array:
		return &SQLTypeAnnotation{
			Type: Int2Array,
		}, nil

	case []uint32, []int32, []graph.ID, pgtype.Int4Array:
		return &SQLTypeAnnotation{
			Type: Int4Array,
		}, nil

	case []uint64, []int64, pgtype.Int8Array:
		return &SQLTypeAnnotation{
			Type: Int8Array,
		}, nil

	case uint16, int16:
		return &SQLTypeAnnotation{
			Type: Int2,
		}, nil

	case uint32, int32, graph.ID:
		return &SQLTypeAnnotation{
			Type: Int4,
		}, nil

	case uint, int, uint64, int64:
		return &SQLTypeAnnotation{
			Type: Int8,
		}, nil

	case float32:
		return &SQLTypeAnnotation{
			Type: Float4,
		}, nil

	case []float32:
		return &SQLTypeAnnotation{
			Type: Float4Array,
		}, nil

	case float64:
		return &SQLTypeAnnotation{
			Type: Float8,
		}, nil

	case []float64:
		return &SQLTypeAnnotation{
			Type: Float8Array,
		}, nil

	case bool:
		return &SQLTypeAnnotation{
			Type: Boolean,
		}, nil

	case string:
		return &SQLTypeAnnotation{
			Type: Text,
		}, nil

	case time.Time:
		return &SQLTypeAnnotation{
			Type: TimestampWithTimeZone,
		}, nil

	case pgtype.JSONB:
		return &SQLTypeAnnotation{
			Type: JSONB,
		}, nil

	case []string, pgtype.TextArray:
		return &SQLTypeAnnotation{
			Type: TextArray,
		}, nil

	case *model.ListLiteral:
		return NewSQLTypeAnnotationFromExpression(typedValue)

	default:
		return nil, fmt.Errorf("literal type %T is not supported", value)
	}
}

type NodeKindsReference struct {
	Variable model.Expression
}

func NewNodeKindsReference(ref *AnnotatedVariable) *NodeKindsReference {
	return &NodeKindsReference{
		Variable: ref,
	}
}

type EdgeKindReference struct {
	Variable model.Expression
}

func NewEdgeKindReference(ref *AnnotatedVariable) *EdgeKindReference {
	return &EdgeKindReference{
		Variable: ref,
	}
}

type Delete struct {
	Binding    *AnnotatedVariable
	NodeDelete bool
	EdgeDelete bool
}

func NewDelete() *Delete {
	return &Delete{
		NodeDelete: false,
		EdgeDelete: false,
	}
}

func (s *Delete) IsMixed() bool {
	return s.NodeDelete && s.EdgeDelete
}

func (s *Delete) Table() string {
	if s.NodeDelete {
		return pgModel.NodeTable
	}

	if s.EdgeDelete {
		return pgModel.EdgeTable
	}

	return ""
}

type PropertyMutation struct {
	Reference *PropertiesReference
	Additions *AnnotatedParameter
	Removals  *AnnotatedParameter
}

type KindMutation struct {
	Variable  *AnnotatedVariable
	Additions *AnnotatedParameter
	Removals  *AnnotatedParameter
}
