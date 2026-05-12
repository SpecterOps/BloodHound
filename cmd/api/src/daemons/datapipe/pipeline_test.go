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

package datapipe

import (
	"context"
	"errors"
	"testing"
	"time"

	dbmocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// stubGraphDB is a no-op graph.Database used purely to satisfy the field type
// on BHCEPipeline. None of the unrelated methods are exercised by the tests
// in this file; if a future test reaches one, the panic surfaces the bug.
type stubGraphDB struct{}

func (stubGraphDB) SetWriteFlushSize(int)                                       {}
func (stubGraphDB) SetBatchWriteSize(int)                                       {}
func (stubGraphDB) ReadTransaction(context.Context, graph.TransactionDelegate, ...graph.TransactionOption) error {
	panic("not implemented")
}
func (stubGraphDB) WriteTransaction(context.Context, graph.TransactionDelegate, ...graph.TransactionOption) error {
	panic("not implemented")
}
func (stubGraphDB) BatchOperation(context.Context, graph.BatchDelegate, ...graph.BatchOption) error {
	panic("not implemented")
}
func (stubGraphDB) AssertSchema(context.Context, graph.Schema) error  { panic("not implemented") }
func (stubGraphDB) SetDefaultGraph(context.Context, graph.Graph) error { panic("not implemented") }
func (stubGraphDB) Run(context.Context, string, map[string]any) error { panic("not implemented") }
func (stubGraphDB) Close(context.Context) error                       { return nil }
func (stubGraphDB) FetchKinds(context.Context) (graph.Kinds, error)   { return nil, nil }
func (stubGraphDB) RefreshKinds(context.Context) error                { return nil }

// optimizingGraphDB embeds the no-op stub and additionally implements
// graph.Optimizer, recording invocations and returning a configurable error.
type optimizingGraphDB struct {
	stubGraphDB
	calls   int
	returns error
}

func (o *optimizingGraphDB) Optimize(_ context.Context) error {
	o.calls++
	return o.returns
}

func TestBHCEPipeline_Optimize_DriverImplementsOptimizer(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)
	mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{}, nil)
	mockDB.EXPECT().UpdateLastOptimizationCompleteTime(gomock.Any()).Return(nil)

	driver := &optimizingGraphDB{}
	p := &BHCEPipeline{db: mockDB, graphdb: driver}

	require.NoError(t, p.Optimize(context.Background()))
	assert.Equal(t, 1, driver.calls, "driver.Optimize should be invoked exactly once")
}

func TestBHCEPipeline_Optimize_DriverImplementsOptimizer_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)
	mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{}, nil)

	wantErr := errors.New("boom")
	driver := &optimizingGraphDB{returns: wantErr}
	p := &BHCEPipeline{db: mockDB, graphdb: driver}

	err := p.Optimize(context.Background())
	assert.ErrorIs(t, err, wantErr)
}

func TestBHCEPipeline_Optimize_DriverDoesNotImplementOptimizer(t *testing.T) {
	// The type-assertion miss path must be silent and successful, and must
	// not even read datapipe status (no mock expectations registered).
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)

	p := &BHCEPipeline{db: mockDB, graphdb: stubGraphDB{}}

	require.NoError(t, p.Optimize(context.Background()))
}

func TestBHCEPipeline_Optimize_SkipsWithinCooldown(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)
	mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{
		LastCompleteOptimizationAt: time.Now().Add(-time.Hour),
	}, nil)

	driver := &optimizingGraphDB{}
	p := &BHCEPipeline{db: mockDB, graphdb: driver}

	require.NoError(t, p.Optimize(context.Background()))
	assert.Equal(t, 0, driver.calls, "Optimize must not run while inside the cooldown window")
}

func TestBHCEPipeline_Optimize_RunsAfterCooldown(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)
	mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{
		LastCompleteOptimizationAt: time.Now().Add(-25 * time.Hour),
	}, nil)
	mockDB.EXPECT().UpdateLastOptimizationCompleteTime(gomock.Any()).Return(nil)

	driver := &optimizingGraphDB{}
	p := &BHCEPipeline{db: mockDB, graphdb: driver}

	require.NoError(t, p.Optimize(context.Background()))
	assert.Equal(t, 1, driver.calls)
}

func TestBHCEPipeline_Optimize_StatusReadFailureProceeds(t *testing.T) {
	// Fail-open: a status read failure must not block optimization, since the
	// alternative would silently mask configuration drift.
	ctrl := gomock.NewController(t)
	mockDB := dbmocks.NewMockDatabase(ctrl)
	mockDB.EXPECT().GetDatapipeStatus(gomock.Any()).Return(model.DatapipeStatusWrapper{}, errors.New("transient db error"))
	mockDB.EXPECT().UpdateLastOptimizationCompleteTime(gomock.Any()).Return(nil)

	driver := &optimizingGraphDB{}
	p := &BHCEPipeline{db: mockDB, graphdb: driver}

	require.NoError(t, p.Optimize(context.Background()))
	assert.Equal(t, 1, driver.calls)
}
