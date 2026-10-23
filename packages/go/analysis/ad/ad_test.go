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

package ad_test

import (
	"context"
	"testing"

	ad2 "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDatabase is a minimal graph.Database implementation used to exercise
// code paths that do not depend on a real driver-backed transaction.
type mockDatabase struct{}

var _ graph.Database = mockDatabase{}

func (s mockDatabase) SetWriteFlushSize(int) {}
func (s mockDatabase) SetBatchWriteSize(int) {}
func (s mockDatabase) ReadTransaction(_ context.Context, txDelegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
	return txDelegate(nil)
}
func (s mockDatabase) WriteTransaction(_ context.Context, txDelegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
	return txDelegate(nil)
}
func (s mockDatabase) BatchOperation(_ context.Context, _ graph.BatchDelegate, _ ...graph.BatchOption) error {
	return nil
}
func (s mockDatabase) AssertSchema(_ context.Context, _ graph.Schema) error    { return nil }
func (s mockDatabase) SetDefaultGraph(_ context.Context, _ graph.Graph) error  { return nil }
func (s mockDatabase) Run(_ context.Context, _ string, _ map[string]any) error { return nil }
func (s mockDatabase) Close(_ context.Context) error                           { return nil }
func (s mockDatabase) FetchKinds(_ context.Context) (graph.Kinds, error)       { return graph.Kinds{}, nil }
func (s mockDatabase) RefreshKinds(_ context.Context) error                    { return nil }
func (s mockDatabase) OptimizeStorage(_ context.Context) error                 { return nil }

func TestGetEdgeCompositionPathUnknownEdgeKind(t *testing.T) {
	edge := &graph.Relationship{Kind: graph.StringKind("NotARealEdgeKind")}

	pathSet, err := ad2.GetEdgeCompositionPath(context.Background(), mockDatabase{}, edge)

	require.Error(t, err)
	assert.ErrorContains(t, err, "no edge composition handler for kind NotARealEdgeKind")
	assert.Equal(t, 0, pathSet.Len())
}
