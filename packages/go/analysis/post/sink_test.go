// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package post

import (
	"context"
	"testing"

	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFilteredRelationshipSink_NoOpDoesNotBatch(t *testing.T) {
	var (
		ctx     = context.Background()
		ctrl    = gomock.NewController(t)
		mockDB  = graph_mocks.NewMockDatabase(ctrl)
		builder = NewTrackerBuilder()
	)

	builder.TrackEdge(100, 1, 2, kindA, propsEmpty)

	sink := NewFilteredRelationshipSink(ctx, "test", mockDB, builder.Build())
	require.True(t, sink.Submit(ctx, EnsureRelationshipJob{FromID: graph.ID(1), ToID: graph.ID(2), Kind: kindA}))

	sink.Done()
	require.Empty(t, sink.Stats().RelationshipsCreated)
	require.Empty(t, sink.Stats().RelationshipsDeleted)
}
