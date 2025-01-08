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
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDataType_CoerceToSupertype(t *testing.T) {
	testCases := []struct {
		LeftTypes       []DataType
		RightTypes      []DataType
		Expected        DataType
		ExpectRightType bool
	}{{
		LeftTypes:  []DataType{UnknownDataType},
		RightTypes: []DataType{Int},
		Expected:   Int,
	}, {
		LeftTypes:  []DataType{Int},
		RightTypes: []DataType{UnknownDataType},
		Expected:   Int,
	}, {
		LeftTypes:  []DataType{Int8},
		RightTypes: []DataType{Int2, Int4, Int, Int8},
		Expected:   Int8,
	}, {
		LeftTypes:  []DataType{Int4},
		RightTypes: []DataType{Int2, Int4},
		Expected:   Int4,
	}, {
		LeftTypes:  []DataType{Int4},
		RightTypes: []DataType{Int},
		Expected:   Int,
	}, {
		LeftTypes:  []DataType{Int4},
		RightTypes: []DataType{Int8},
		Expected:   Int8,
	}, {
		LeftTypes:       []DataType{Int2},
		RightTypes:      []DataType{Int2, Int4, Int, Int8},
		ExpectRightType: true,
	}, {
		LeftTypes:       []DataType{Int},
		RightTypes:      []DataType{Int, Int8},
		ExpectRightType: true,
	}, {
		LeftTypes:       []DataType{Float4},
		RightTypes:      []DataType{Float4, Float8, Numeric},
		ExpectRightType: true,
	}, {
		LeftTypes:  []DataType{Float8},
		RightTypes: []DataType{Float4},
		Expected:   Float8,
	}, {
		LeftTypes:       []DataType{Float8},
		RightTypes:      []DataType{Float8, Numeric},
		ExpectRightType: true,
	}, {
		LeftTypes:  []DataType{Numeric},
		RightTypes: []DataType{Numeric, Float8, Float4, Int8, Int, Int4, Int2},
		Expected:   Numeric,
	}}

	for _, testCase := range testCases {
		for _, leftType := range testCase.LeftTypes {
			for _, rightType := range testCase.RightTypes {
				superType, coerced := leftType.CoerceToSupertype(rightType)

				if !coerced {
					t.Fatalf("coercing left type %s to right type %s failed", leftType, rightType)
				}

				if testCase.ExpectRightType {
					require.Equalf(t, rightType, superType, "expected type %s does not match super type %s", rightType, superType)
				} else {
					require.Equalf(t, testCase.Expected, superType, "expected type %s does not match super type %s", testCase.Expected, superType)
				}

			}
		}
	}
}

func TestDataType_Comparable(t *testing.T) {
	testCases := []struct {
		LeftTypes  []DataType
		Operators  []Operator
		RightTypes []DataType
		Expected   bool
	}{
		// Supported comparisons
		{
			LeftTypes:  []DataType{Int, Int8, Int4, Int2},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{Int, Int8, Int4, Int2, Float8, Float4, Numeric},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{Float8, Float4, Numeric},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{Int, Int8, Int4, Int2, Float8, Float4, Numeric},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{NodeComposite},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{NodeComposite},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{EdgeComposite},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{EdgeComposite},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{PathComposite},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{PathComposite},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{JSONB},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{JSONB},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{AnyArray},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{AnyArray},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{Text},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{Text},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{Boolean},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{Boolean},
			Expected:   true,
		},

		// Right hand unknown types should not be comparable against any left hand int type
		{
			LeftTypes:  []DataType{Int},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{UnknownDataType},
			Expected:   false,
		},

		// Right hand unknown types should not be comparable against any left hand float type
		{
			LeftTypes:  []DataType{Float8},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{UnknownDataType},
			Expected:   false,
		},

		// Left hand unknown types should not be comparable against any right hand type
		{
			LeftTypes:  []DataType{UnknownDataType},
			Operators:  []Operator{OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorGreaterThanOrEqualTo, OperatorLessThan, OperatorLessThanOrEqualTo},
			RightTypes: []DataType{Int},
			Expected:   false,
		},

		// Validate text operations
		{
			LeftTypes:  []DataType{Text},
			Operators:  []Operator{OperatorLike, OperatorILike, OperatorSimilarTo, OperatorRegexMatch},
			RightTypes: []DataType{Text},
			Expected:   true,
		},

		// Text operations on non-text types should fail
		{
			LeftTypes:  []DataType{Int},
			Operators:  []Operator{OperatorLike, OperatorILike, OperatorSimilarTo, OperatorRegexMatch},
			RightTypes: []DataType{Int},
			Expected:   false,
		},

		// Array types may use the overlap operator but only if their base types match
		{
			LeftTypes:  []DataType{IntArray},
			Operators:  []Operator{OperatorPGArrayOverlap},
			RightTypes: []DataType{IntArray},
			Expected:   true,
		},
		{
			LeftTypes:  []DataType{IntArray},
			Operators:  []Operator{OperatorPGArrayOverlap},
			RightTypes: []DataType{Int},
			Expected:   false,
		},

		// Catch all for any unsupported operator
		{
			LeftTypes:  []DataType{Int},
			Operators:  []Operator{"Unsupported Operator Class"},
			RightTypes: []DataType{Int},
			Expected:   false,
		},
	}

	for idx, testCase := range testCases {
		for _, leftType := range testCase.LeftTypes {
			for _, operator := range testCase.Operators {
				for _, rightType := range testCase.RightTypes {
					result := leftType.IsComparable(rightType, operator)
					require.Equalf(t, testCase.Expected, result, "failed test case %d: %+v, %+v", idx, testCase.LeftTypes, testCase.RightTypes)
				}
			}
		}
	}
}

func TestValueToDataType(t *testing.T) {
	testCases := []struct {
		Value        any
		ExpectedType DataType
	}{{
		Value:        uint8(1),
		ExpectedType: Int2,
	}, {
		Value:        uint16(1),
		ExpectedType: Int4,
	}, {
		Value:        uint32(1),
		ExpectedType: Int8,
	}, {
		Value:        uint64(1),
		ExpectedType: Int8,
	}, {
		Value:        int8(1),
		ExpectedType: Int2,
	}, {
		Value:        int16(1),
		ExpectedType: Int2,
	}, {
		Value:        int32(1),
		ExpectedType: Int4,
	}, {
		Value:        int64(1),
		ExpectedType: Int8,
	}, {
		Value:        int(1),
		ExpectedType: Int8,
	}, {
		Value:        []uint8{},
		ExpectedType: Int2Array,
	}, {
		Value:        []uint16{},
		ExpectedType: Int4Array,
	}, {
		Value:        []uint32{},
		ExpectedType: Int8Array,
	}, {
		Value:        []uint64{},
		ExpectedType: Int8Array,
	}, {
		Value:        []uint{},
		ExpectedType: Int8Array,
	}, {
		Value:        float32(1),
		ExpectedType: Float4,
	}, {
		Value:        float64(1),
		ExpectedType: Float8,
	}, {
		Value:        []float32{},
		ExpectedType: Float4Array,
	}, {
		Value:        []float64{},
		ExpectedType: Float8Array,
	}, {
		Value:        "1",
		ExpectedType: Text,
	}, {
		Value:        []string{},
		ExpectedType: TextArray,
	}, {
		Value:        false,
		ExpectedType: Boolean,
	}, {
		Value:        graph.StringKind("test"),
		ExpectedType: Int2,
	}, {
		Value:        graph.Kinds{},
		ExpectedType: Int2Array,
	}, {
		Value:        []any{"1", "2"},
		ExpectedType: TextArray,
	}, {
		Value:        time.Duration(5),
		ExpectedType: Interval,
	}, {
		Value:        time.Now().UTC(),
		ExpectedType: TimestampWithTimeZone,
	}, {
		Value:        time.Now().Local(),
		ExpectedType: TimestampWithoutTimeZone,
	}}

	for _, testCase := range testCases {
		dataType, err := ValueToDataType(testCase.Value)

		require.Nil(t, err)
		require.Equal(t, testCase.ExpectedType, dataType)
	}
}
