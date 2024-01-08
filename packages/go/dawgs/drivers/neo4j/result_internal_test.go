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
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func mapTestCase[T, V any](t *testing.T, source T, expected V) {
	var (
		mapper = graph.NewValueMapper([]any{source}, mapValue)
		value  V
	)

	require.Nil(t, mapper.Map(&value))
	require.Equalf(t, expected, value, "Mapping case for type %T to %T failed. Value is: %v", source, &value, value)
}

func Test_mapValue(t *testing.T) {
	var (
		utcNow         = time.Now().UTC()
		anyStringSlice = []any{"a", "b", "c"}
		stringSlice    = []string{"a", "b", "c"}
		kindSlice      = []graph.Kind{graph.StringKind("a"), graph.StringKind("b"), graph.StringKind("c")}
		kinds          = graph.Kinds{graph.StringKind("a"), graph.StringKind("b"), graph.StringKind("c")}
	)

	mapTestCase[uint, uint](t, 0, 0)
	mapTestCase[uint8, uint8](t, 0, 0)
	mapTestCase[uint16, uint16](t, 0, 0)
	mapTestCase[uint32, uint32](t, 0, 0)
	mapTestCase[uint64, uint64](t, 0, 0)

	mapTestCase[int, int](t, 0, 0)
	mapTestCase[int8, int8](t, 0, 0)
	mapTestCase[int16, int16](t, 0, 0)
	mapTestCase[int32, int32](t, 0, 0)
	mapTestCase[int64, int64](t, 0, 0)
	mapTestCase[int64, graph.ID](t, 0, 0)

	mapTestCase[float32, float32](t, 1.5, 1.5)
	mapTestCase[float64, float64](t, 1.5, 1.5)

	mapTestCase[bool, bool](t, true, true)
	mapTestCase[string, string](t, "test", "test")

	mapTestCase[time.Time, time.Time](t, utcNow, utcNow)
	mapTestCase[string, time.Time](t, utcNow.Format(time.RFC3339Nano), utcNow)
	mapTestCase[int64, time.Time](t, utcNow.Unix(), time.Unix(utcNow.Unix(), 0))
	mapTestCase[dbtype.Time, time.Time](t, dbtype.Time(utcNow), utcNow)
	mapTestCase[dbtype.LocalTime, time.Time](t, dbtype.LocalTime(utcNow), utcNow)
	mapTestCase[dbtype.Date, time.Time](t, dbtype.Date(utcNow), utcNow)
	mapTestCase[dbtype.LocalDateTime, time.Time](t, dbtype.LocalDateTime(utcNow), utcNow)

	mapTestCase[[]any, []string](t, anyStringSlice, stringSlice)
	mapTestCase[[]any, []graph.Kind](t, anyStringSlice, kindSlice)
	mapTestCase[[]any, graph.Kinds](t, anyStringSlice, kinds)
}
