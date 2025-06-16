// Copyright 2025 Specter Ops, Inc.
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

package traversal

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIDBreadthFirstContextCancel(t *testing.T) {
	t.Parallel()
	var (
		numWorkers    = 4
		mockCtrl      = gomock.NewController(t)
		mockDB        = graph_mocks.NewMockDatabase(mockCtrl)
		mockTx        = graph_mocks.NewMockTransaction(mockCtrl)
		traversalInst = NewIDTraversal(mockDB, numWorkers)
		plan          = IDPlan{
			Root: root.Node.ID,
			Delegate: func(ctx context.Context, tx graph.Transaction, segment *graph.IDSegment) ([]*graph.IDSegment, error) {
				return []*graph.IDSegment{}, nil
			},
		}
	)

	mockDB.EXPECT().ReadTransaction(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(testCtx context.Context, logic func(tx graph.Transaction) error, options ...graph.TransactionOption) error {
		return logic(mockTx)
	}).Times(numWorkers)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := traversalInst.BreadthFirst(ctx, plan)
	require.Nil(t, err)
}
