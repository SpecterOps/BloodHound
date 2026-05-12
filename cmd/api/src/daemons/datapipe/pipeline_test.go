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

	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	driver := &optimizingGraphDB{}
	p := &BHCEPipeline{graphdb: driver}

	require.NoError(t, p.Optimize(context.Background()))
	assert.Equal(t, 1, driver.calls, "driver.Optimize should be invoked exactly once")
}

func TestBHCEPipeline_Optimize_DriverImplementsOptimizer_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	driver := &optimizingGraphDB{returns: wantErr}
	p := &BHCEPipeline{graphdb: driver}

	err := p.Optimize(context.Background())
	assert.ErrorIs(t, err, wantErr)
}

func TestBHCEPipeline_Optimize_DriverDoesNotImplementOptimizer(t *testing.T) {
	p := &BHCEPipeline{graphdb: stubGraphDB{}}

	// The type-assertion miss path must be silent and successful.
	require.NoError(t, p.Optimize(context.Background()))
}
