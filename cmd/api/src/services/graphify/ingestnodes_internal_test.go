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
	"errors"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/mocks"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNormalizeEinNodeProperties(t *testing.T) {
	var (
		nowUTC     = time.Now().UTC()
		objectID   = "objectid"
		properties = map[string]any{
			ReconcileProperty:               false,
			common.Name.String():            "name",
			common.OperatingSystem.String(): "temple",
			ad.DistinguishedName.String():   "distinguished-name",
		}
		normalizedProperties = normalizeEinNodeProperties(properties, objectID, nowUTC)
	)

	assert.Nil(t, normalizedProperties[ReconcileProperty])
	assert.NotNil(t, normalizedProperties[common.LastSeen.String()])
	assert.Equal(t, "OBJECTID", normalizedProperties[common.ObjectID.String()])
	assert.Equal(t, "NAME", normalizedProperties[common.Name.String()])
	assert.Equal(t, "DISTINGUISHED-NAME", normalizedProperties[ad.DistinguishedName.String()])
	assert.Equal(t, "TEMPLE", normalizedProperties[common.OperatingSystem.String()])
}

func TestMaybeSubmitNodeUpdate(t *testing.T) {
	t.Run("there is no changelog, submit to batch and track stats", func(t *testing.T) {
		var (
			ctx              = context.Background()
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			ingestCtx        = NewIngestContext(ctx)

			node       = graph.PrepareNode(graph.NewProperties().Set("hello", "world"), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		nodesProcessed, _, nodesWritten, _ := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), nodesProcessed)
		require.Equal(t, int64(0), nodesWritten)

		// mock expects
		mockBatchUpdater.EXPECT().UpdateNodeBy(nodeUpdate).Return(nil).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)

		// Verify stats were incremented
		nodesProcessed, _, nodesWritten, _ = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), nodesProcessed, "NodesProcessed should be incremented")
		require.Equal(t, int64(1), nodesWritten, "NodesWritten should be incremented")
	})

	t.Run("new change, submit to batch and track stats", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, WithChangeManager(mockChangeManager))

			objectID   = "1234"
			node       = graph.PrepareNode(graph.NewProperties().Set("objectid", objectID), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
			change     = changelog.NewNodeChange(objectID, node.Kinds, node.Properties)
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		nodesProcessed, _, nodesWritten, _ := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), nodesProcessed)
		require.Equal(t, int64(0), nodesWritten)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(true, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateNodeBy(nodeUpdate).Return(nil).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)

		// Verify stats were incremented
		nodesProcessed, _, nodesWritten, _ = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), nodesProcessed, "NodesProcessed should be incremented")
		require.Equal(t, int64(1), nodesWritten, "NodesWritten should be incremented")
	})

	t.Run("unmodified, submit to changelog and track processed only", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, WithChangeManager(mockChangeManager))

			objectID   = "1234"
			node       = graph.PrepareNode(graph.NewProperties().Set("objectid", objectID), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
			change     = changelog.NewNodeChange(objectID, node.Kinds, node.Properties)
		)

		// Wrap the mock with counting wrapper to track stats
		ingestCtx.BindBatchUpdater(mockBatchUpdater)

		// Verify initial stats
		nodesProcessed, _, nodesWritten, _ := ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(0), nodesProcessed)
		require.Equal(t, int64(0), nodesWritten)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(false, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Times(0)
		mockChangeManager.EXPECT().Submit(ctx, change).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)

		// Verify stats: processed incremented, written NOT incremented (deduplicated)
		nodesProcessed, _, nodesWritten, _ = ingestCtx.Stats.GetCounts()
		require.Equal(t, int64(1), nodesProcessed, "NodesProcessed should be incremented even when deduplicated")
		require.Equal(t, int64(0), nodesWritten, "NodesWritten should NOT be incremented when deduplicated")
	})
}

func TestIngestGenericData_RegisterNodeKind(t *testing.T) {
	t.Run("new unknown kind calls registrar once", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			called           []string
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				called = append(called, kind.String())
				return nil
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(1)

		err := IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}})
		require.NoError(t, err)
		assert.Equal(t, []string{"UnknownKind"}, called)
	})

	t.Run("same kind across multiple nodes in one context calls registrar once", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			called           []string
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				called = append(called, kind.String())
				return nil
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(2)

		require.NoError(t, IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}}))
		require.NoError(t, IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-2"}}}))
		assert.Equal(t, []string{"UnknownKind"}, called)
	})

	t.Run("two distinct kinds call registrar once each", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			called           []string
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				called = append(called, kind.String())
				return nil
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(1)

		err := IngestGenericData(ingestCtx, graph.StringKind("UnknownKindA"), ConvertedData{NodeProps: []ein.IngestibleNode{{
			ObjectID: "node-1",
			Labels:   []graph.Kind{graph.StringKind("UnknownKindB")},
		}}})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"UnknownKindA", "UnknownKindB"}, called)
	})

	t.Run("tag kind does not call registrar", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			called           []string
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				called = append(called, kind.String())
				return nil
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(1)

		err := IngestGenericData(ingestCtx, graph.StringKind("Tag_Custom"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}})
		require.NoError(t, err)
		assert.Empty(t, called)
	})

	t.Run("registrar error propagates and prevents node update", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			wantErr          = errors.New("register failed")
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				return wantErr
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Times(0)

		err := IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), wantErr.Error())
	})

	t.Run("registrar error does not mark kind seen", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			wantErr          = errors.New("register failed")
			called           []string
			ingestCtx        = NewIngestContext(context.Background(), WithNodeKindRegistrar(func(kind graph.Kind) error {
				called = append(called, kind.String())
				if len(called) == 1 {
					return wantErr
				}
				return nil
			}))
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(1)

		err := IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), wantErr.Error())

		err = IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-2"}}})
		require.NoError(t, err)
		assert.Equal(t, []string{"UnknownKind", "UnknownKind"}, called)
	})

	t.Run("missing registrar remains safe", func(t *testing.T) {
		var (
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			ingestCtx        = NewIngestContext(context.Background())
		)

		ingestCtx.BindBatchUpdater(mockBatchUpdater)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Return(nil).Times(1)

		err := IngestGenericData(ingestCtx, graph.StringKind("UnknownKind"), ConvertedData{NodeProps: []ein.IngestibleNode{{ObjectID: "node-1"}}})
		require.NoError(t, err)
	})
}
