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

package graphify

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDeduplicateKinds(t *testing.T) {
	kinds := []graph.Kind{graph.StringKind("Same"), graph.StringKind("Same"), graph.StringKind("Different")}

	deduped := deduplicateKinds(kinds)
	require.Len(t, deduped, 2)
	require.Equal(t, deduped[0].String(), "Same")
	require.Equal(t, deduped[1].String(), "Different")
}

func TestMergeNodeKinds(t *testing.T) {
	kinds := []graph.Kind{
		graph.StringKind("Same"),
		graph.StringKind("Same"),
		graph.StringKind("Different"),
		graph.EmptyKind,
	}

	merged := MergeNodeKinds(graph.StringKind("Base"), kinds...)
	require.Len(t, merged, 3)
	require.Equal(t, merged[0].String(), "Base")
	require.Equal(t, merged[1].String(), "Same")
	require.Equal(t, merged[2].String(), "Different")

	merged = MergeNodeKinds(graph.EmptyKind, kinds...)
	require.Len(t, merged, 2)
	require.Equal(t, merged[0].String(), "Same")
	require.Equal(t, merged[1].String(), "Different")
}

func TestMaybeSubmitRelationshipUpdate(t *testing.T) {
	t.Run("there is no changelog, submit to batch and track stats", func(t *testing.T) {
		var (
			ctx              = context.Background()
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			ingestCtx        = NewIngestContext(ctx)

			startNode = graph.PrepareNode(graph.NewProperties().Set("objectid", "start123"), graph.StringKind("kindA"))
			endNode   = graph.PrepareNode(graph.NewProperties().Set("objectid", "end456"), graph.StringKind("kindB"))
			rel       = graph.PrepareRelationship(graph.NewProperties().Set("hello", "world"), graph.StringKind("relKind"))
			relUpdate = graph.RelationshipUpdate{
				Start:        startNode,
				End:          endNode,
				Relationship: rel,
			}
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		_, relsProcessed, _, relsWritten := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), relsProcessed)
		require.Equal(t, int64(0), relsWritten)

		// mock expects
		mockBatchUpdater.EXPECT().UpdateRelationshipBy(relUpdate).Return(nil).Times(1)

		err := maybeSubmitRelationshipUpdate(ingestCtx, relUpdate)
		require.NoError(t, err)

		// Verify stats were incremented
		_, relsProcessed, _, relsWritten = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), relsProcessed, "RelationshipsProcessed should be incremented")
		require.Equal(t, int64(1), relsWritten, "RelationshipsWritten should be incremented")
	})

	t.Run("new change, submit to batch and track stats", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, WithChangeManager(mockChangeManager))

			sourceObjectID = "source123"
			targetObjectID = "target456"
			startNode      = graph.PrepareNode(graph.NewProperties().Set("objectid", sourceObjectID), graph.StringKind("kindA"))
			endNode        = graph.PrepareNode(graph.NewProperties().Set("objectid", targetObjectID), graph.StringKind("kindB"))
			rel            = graph.PrepareRelationship(graph.NewProperties().Set("hello", "world"), graph.StringKind("relKind"))
			relUpdate      = graph.RelationshipUpdate{
				Start:        startNode,
				End:          endNode,
				Relationship: rel,
			}
			change = changelog.NewEdgeChange(sourceObjectID, targetObjectID, rel.Kind, rel.Properties)
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		_, relsProcessed, _, relsWritten := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), relsProcessed)
		require.Equal(t, int64(0), relsWritten)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(true, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateRelationshipBy(relUpdate).Return(nil).Times(1)

		err := maybeSubmitRelationshipUpdate(ingestCtx, relUpdate)
		require.NoError(t, err)

		// Verify stats were incremented
		_, relsProcessed, _, relsWritten = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), relsProcessed, "RelationshipsProcessed should be incremented")
		require.Equal(t, int64(1), relsWritten, "RelationshipsWritten should be incremented")
	})

	t.Run("unmodified, submit to changelog and track processed only", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, WithChangeManager(mockChangeManager))

			sourceObjectID = "source123"
			targetObjectID = "target456"
			startNode      = graph.PrepareNode(graph.NewProperties().Set("objectid", sourceObjectID), graph.StringKind("kindA"))
			endNode        = graph.PrepareNode(graph.NewProperties().Set("objectid", targetObjectID), graph.StringKind("kindB"))
			rel            = graph.PrepareRelationship(graph.NewProperties().Set("hello", "world"), graph.StringKind("relKind"))
			relUpdate      = graph.RelationshipUpdate{
				Start:        startNode,
				End:          endNode,
				Relationship: rel,
			}
			change = changelog.NewEdgeChange(sourceObjectID, targetObjectID, rel.Kind, rel.Properties)
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		_, relsProcessed, _, relsWritten := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), relsProcessed)
		require.Equal(t, int64(0), relsWritten)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(false, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateRelationshipBy(gomock.Any()).Times(0)
		mockChangeManager.EXPECT().Submit(ctx, change).Times(1)

		err := maybeSubmitRelationshipUpdate(ingestCtx, relUpdate)
		require.NoError(t, err)

		// Verify stats: processed incremented, written NOT incremented (deduplicated)
		_, relsProcessed, _, relsWritten = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), relsProcessed, "RelationshipsProcessed should be incremented even when deduplicated")
		require.Equal(t, int64(0), relsWritten, "RelationshipsWritten should NOT be incremented when deduplicated")
	})
}
