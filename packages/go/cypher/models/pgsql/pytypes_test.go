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
