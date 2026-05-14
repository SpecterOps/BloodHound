// Copyright 2026 Specter Ops, Inc.
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

package datapipe_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// recordingPipeline is a minimal datapipe.Pipeline implementation that records
// the order in which its methods are invoked. Used to assert daemon call sequencing.
type recordingPipeline struct {
	mu    sync.Mutex
	calls []string
}

func (p *recordingPipeline) record(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, name)
}

func (p *recordingPipeline) snapshot() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, len(p.calls))
	copy(out, p.calls)
	return out
}

func (p *recordingPipeline) Start(_ context.Context) error       { p.record("Start"); return nil }
func (p *recordingPipeline) PruneData(_ context.Context) error   { p.record("PruneData"); return nil }
func (p *recordingPipeline) DeleteData(_ context.Context) error  { p.record("DeleteData"); return nil }
func (p *recordingPipeline) IngestTasks(_ context.Context) error { p.record("IngestTasks"); return nil }
func (p *recordingPipeline) Analyze(_ context.Context) error     { p.record("Analyze"); return nil }
func (p *recordingPipeline) Optimize(_ context.Context) error    { p.record("Optimize"); return nil }

func (p *recordingPipeline) IsPrimary(ctx context.Context, _ model.DatapipeStatus) (bool, context.Context) {
	return true, ctx
}

// runDaemonOnce starts the daemon, waits for at least one full operational loop
// iteration, then cancels and returns once Start has unwound.
func runDaemonOnce(t *testing.T, daemon *datapipe.Daemon) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		daemon.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("daemon.Start did not return after context cancellation")
	}
}

func TestDaemon_StartInvokesOptimizeOnBootAndAfterAnalyze(t *testing.T) {
	ctrl := gomock.NewController(t)

	db := dbmocks.NewMockDatabase(ctrl)
	db.EXPECT().SetDatapipeStatus(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	pipe := &recordingPipeline{}

	// startDelay short so the operational loop runs once during the test;
	// tickInterval long so the loop runs at most once before context expires.
	daemon := datapipe.NewDaemon(pipe, 10*time.Millisecond, time.Hour, db)

	runDaemonOnce(t, daemon)

	calls := pipe.snapshot()
	require.GreaterOrEqual(t, len(calls), 6, "expected boot sequence plus one full loop iteration, got: %v", calls)

	// Boot order: pipeline Start hook, then Optimize, before any operational work.
	assert.Equal(t, []string{"Start", "Optimize"}, calls[:2], "boot sequence")

	// First operational loop iteration: DeleteData, IngestTasks, Analyze, Optimize.
	assert.Equal(t, []string{"DeleteData", "IngestTasks", "Analyze", "Optimize"}, calls[2:6], "loop sequence")
}

func TestDaemon_StartTransitionsThroughOptimizingStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	var (
		statusMu sync.Mutex
		statuses []model.DatapipeStatus
	)

	db := dbmocks.NewMockDatabase(ctrl)
	db.EXPECT().SetDatapipeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, s model.DatapipeStatus) error {
			statusMu.Lock()
			defer statusMu.Unlock()
			statuses = append(statuses, s)
			return nil
		}).AnyTimes()

	pipe := &recordingPipeline{}

	daemon := datapipe.NewDaemon(pipe, 10*time.Millisecond, time.Hour, db)

	runDaemonOnce(t, daemon)

	statusMu.Lock()
	defer statusMu.Unlock()

	require.NotEmpty(t, statuses)

	// Boot transitions begin with Starting then Optimizing (each followed by Idle in the wrapper).
	require.GreaterOrEqual(t, len(statuses), 4, "expected at least Starting/Idle/Optimizing/Idle, got: %v", statuses)
	assert.Equal(t, model.DatapipeStatusStarting, statuses[0])
	assert.Equal(t, model.DatapipeStatusIdle, statuses[1])
	assert.Equal(t, model.DatapipeStatusOptimizing, statuses[2])
	assert.Equal(t, model.DatapipeStatusIdle, statuses[3])
}
