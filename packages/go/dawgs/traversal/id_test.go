package traversal

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIDBreadthFirstContextCancel(t *testing.T) {
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
