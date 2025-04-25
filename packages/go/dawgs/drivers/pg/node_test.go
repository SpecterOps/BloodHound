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
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/pg/pgutil"

	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	NodeKind1 = graph.StringKind("NodeKind1")
	NodeKind2 = graph.StringKind("NodeKind2")
	EdgeKind1 = graph.StringKind("EdgeKind1")
	EdgeKind2 = graph.StringKind("EdgeKind2")
)

func newKindMapper() KindMapper {
	mapper := pgutil.NewInMemoryKindMapper()

	// This is here to make SQL output a little more predictable for test cases
	mapper.Put(NodeKind1)
	mapper.Put(NodeKind2)
	mapper.Put(EdgeKind1)
	mapper.Put(EdgeKind2)

	return mapper
}

func TestNodeQuery(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockTx        = graph_mocks.NewMockTransaction(mockCtrl)
		mockResult    = graph_mocks.NewMockResult(mockCtrl)
		kindMapper    = newKindMapper()
		nodeQueryInst = &nodeQuery{
			liveQuery: newLiveQuery(context.Background(), mockTx, kindMapper),
		}
	)

	mockTx.EXPECT().Raw("-- match (n) where n.prop = $ return n limit 1\nwith s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'prop') = @pi0::text)) select s0.n0 as n from s0 limit 1;", gomock.Any()).Return(mockResult)

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

func TestNodeQueryOrderByNodeIdWithLimit(t *testing.T) {
	var (
		mockCtrl      = gomock.NewController(t)
		mockTx        = graph_mocks.NewMockTransaction(mockCtrl)
		mockResult    = graph_mocks.NewMockResult(mockCtrl)
		kindMapper    = newKindMapper()
		nodeQueryInst = &nodeQuery{
			liveQuery: newLiveQuery(context.Background(), mockTx, kindMapper),
		}
	)

	mockTx.EXPECT().Raw("-- match (n) where n.prop = $ return n order by id(n) desc limit 2\nwith s0 as (select (n0.id, n0.kind_ids, n0.properties)::nodecomposite as n0 from node n0 where ((n0.properties ->> 'prop') = @pi0::text)) select s0.n0 as n from s0 order by (s0.n0).id desc limit 2;", gomock.Any()).Return(mockResult)

	mockResult.EXPECT().Error().Return(nil)
	mockResult.EXPECT().Next().AnyTimes()
	mockResult.EXPECT().Close().Return().Times(2)
	nodeQueryInst.Filter(
		query.Equals(query.NodeProperty("prop"), "1234"),
	)

	err := nodeQueryInst.Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		return nil
	}, query.Limit(2), query.OrderBy(query.Order(query.NodeID(), query.Descending())))
	require.Nil(t, err)
}
