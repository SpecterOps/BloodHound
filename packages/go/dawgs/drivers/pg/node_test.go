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

package pg

import (
	"context"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

type testKindMapper struct {
	known map[string]int16
}

func (s testKindMapper) MapKindID(kindID int16) (graph.Kind, bool) {
	panic("implement me")
}

func (s testKindMapper) MapKindIDs(kindIDs ...int16) (graph.Kinds, []int16) {
	panic("implement me")
}

func (s testKindMapper) MapKind(kind graph.Kind) (int16, bool) {
	panic("implement me")
}

func (s testKindMapper) AssertKinds(tx graph.Transaction, kinds graph.Kinds) ([]int16, error) {
	panic("implement me")
}

func (s testKindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	var (
		kindIDs      = make([]int16, 0, len(kinds))
		missingKinds = make([]graph.Kind, 0, len(kinds))
	)

	for _, kind := range kinds {
		if kindID, hasKind := s.known[kind.String()]; hasKind {
			kindIDs = append(kindIDs, kindID)
		} else {
			missingKinds = append(missingKinds, kind)
		}
	}

	return kindIDs, missingKinds
}

func TestNodeQuery(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockTx     = graph_mocks.NewMockTransaction(mockCtrl)
		mockResult = graph_mocks.NewMockResult(mockCtrl)

		kindMapper = testKindMapper{
			known: map[string]int16{
				"NodeKindA": 1,
				"NodeKindB": 2,
				"EdgeKindA": 3,
				"EdgeKindB": 4,
			},
		}

		nodeQueryInst = &nodeQuery{
			liveQuery: newLiveQuery(context.Background(), mockTx, kindMapper),
		}
	)

	mockTx.EXPECT().Raw("select (n.id, n.kind_ids, n.properties)::nodeComposite as n from node as n where (n.properties->>'prop')::text = @p0 limit 1", gomock.Any()).Return(mockResult)

	mockResult.EXPECT().Error().Return(nil)
	mockResult.EXPECT().Next().Return(true)
	mockResult.EXPECT().Close().Return()
	mockResult.EXPECT().Scan(gomock.Any()).Return(nil)

	nodeQueryInst.Filter(
		query.Equals(query.NodeProperty("prop"), "1234"),
	)

	_, err := nodeQueryInst.First()
	require.Nil(t, err)
}
