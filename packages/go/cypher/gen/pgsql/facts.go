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

package pgsql

import "errors"

const (
	cypherCountFunction         = "count"
	cypherDateFunction          = "date"
	cypherTimeFunction          = "time"
	cypherLocalTimeFunction     = "localtime"
	cypherDateTimeFunction      = "datetime"
	cypherLocalDateTimeFunction = "localdatetime"
	cypherDurationFunction      = "duration"
	cypherIdentityFunction      = "id"
	cypherToLowerFunction       = "toLower"
	cypherNodeLabelsFunction    = "labels"
	cypherEdgeTypeFunction      = "type"

	pgsqlAnyFunction     = "any"
	pgsqlToLowerFunction = "lower"
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
	Edge                     DataType = "edgeComposite"
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

var CompositeTypes = []DataType{Node, Edge}
