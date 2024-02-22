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

package pgtransition_test

import (
	"testing"

	"github.com/specterops/bloodhound/cypher/backend/pgsql/pgtransition"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test"
	"github.com/stretchr/testify/require"
)

type kindMapper struct{}

func (k kindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	return make([]int16, len(kinds)), nil
}

func TestTranslateAllShortestPaths(t *testing.T) {
	builder := query.NewBuilder(&query.Cache{})
	builder.Apply(query.Where(
		query.And(
			query.And(query.Equals(query.StartID(), graph.ID(1)), query.Equals(query.EndProperty("name"), "1")),
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor),
			query.Equals(query.EndID(), graph.ID(5)),
		),
	))

	aspArguments, err := pgtransition.TranslateAllShortestPaths(builder.RegularQuery(), kindMapper{})
	test.RequireNilErr(t, err)

	require.Equal(t, "s.id = 1", aspArguments.RootCriteria, "Root Criteria")
	require.Equal(t, "(r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]) or r.kind_id = any(array[0]::int2[]))", aspArguments.TraversalCriteria, "Traversal Criteria")
	require.Equal(t, "e.properties->'name' = '1' and e.id = 5", aspArguments.TerminalCriteria, "Terminal Criteria")
}
