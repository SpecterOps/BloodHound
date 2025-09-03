package changelog

import (
	"context"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
	mockGraph "github.com/specterops/dawgs/graph/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFeatureFlagCacheSizing2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDawgsDB := mockGraph.NewMockDatabase(ctrl)
	mockBloodhoundDB := mocks.NewMockDatabase(ctrl)

	// Counter to simulate flag flip after N polls
	callCount := 0
	mockBloodhoundDB.EXPECT().
		GetFlagByKey(gomock.Any(), appcfg.FeatureChangelog).
		DoAndReturn(func(_ context.Context, _ string) (appcfg.FeatureFlag, error) {
			callCount++
			if callCount < 3 {
				// first 2 polls -> disabled
				return appcfg.FeatureFlag{Enabled: false}, nil
			}
			// from 3rd poll onwards -> enabled
			return appcfg.FeatureFlag{Enabled: true}, nil
		}).
		AnyTimes()

	// Mock DB size calculation
	size := 0
	mockDawgsDB.EXPECT().
		ReadTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(graph.Transaction) error, _ ...any) error {
			// simulate "counting" entities
			size = 42
			return nil
		}).
		AnyTimes()

	cl := NewChangelog(mockDawgsDB, mockBloodhoundDB, DefaultOptions())
	cl.options.PollInterval = 20 * time.Millisecond

	go cl.Start(ctx)

	if cl.cache != nil {
		t.Fatalf("expected cache to be nil when feature flag is off")
	}

	// Wait for a few polls
	time.Sleep(100 * time.Millisecond)

	// At this point the flag should have flipped to enabled
	if cl.cache == nil {
		t.Fatalf("expected cache to be allocated when feature flag flips on")
	}

	require.Equal(t, 42, size)
}
