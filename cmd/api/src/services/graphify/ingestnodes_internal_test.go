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
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/mocks"
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
	t.Run("there is no changelog, submit to batch", func(t *testing.T) {
		var (
			ctx              = context.Background()
			ctrl             = gomock.NewController(t)
			mockBatchUpdater = mocks.NewMockBatchUpdater(ctrl)
			ingestCtx        = NewIngestContext(ctx, mockBatchUpdater)

			node       = graph.PrepareNode(graph.NewProperties().Set("hello", "world"), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
		)

		// mock expects
		mockBatchUpdater.EXPECT().UpdateNodeBy(nodeUpdate).Return(nil).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)
	})

	t.Run("new change, submit to batch", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, mockBatchUpdater, WithChangeManager(mockChangeManager))

			objectID   = "1234"
			node       = graph.PrepareNode(graph.NewProperties().Set("objectid", objectID), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
			change     = changelog.NewNodeChange(objectID, node.Kinds, node.Properties)
		)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(true, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateNodeBy(nodeUpdate).Return(nil).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)
	})

	t.Run("unmodified, submit to changelog", func(t *testing.T) {
		var (
			ctx               = context.Background()
			ctrl              = gomock.NewController(t)
			mockBatchUpdater  = mocks.NewMockBatchUpdater(ctrl)
			mockChangeManager = mocks.NewMockChangeManager(ctrl)
			ingestCtx         = NewIngestContext(ctx, mockBatchUpdater, WithChangeManager(mockChangeManager))

			objectID   = "1234"
			node       = graph.PrepareNode(graph.NewProperties().Set("objectid", objectID), graph.StringKind("kindA"))
			nodeUpdate = graph.NodeUpdate{Node: node}
			change     = changelog.NewNodeChange(objectID, node.Kinds, node.Properties)
		)

		// mock expects
		mockChangeManager.EXPECT().ResolveChange(change).Return(false, nil).Times(1)
		mockBatchUpdater.EXPECT().UpdateNodeBy(gomock.Any()).Times(0)
		mockChangeManager.EXPECT().Submit(ctx, change).Times(1)

		err := maybeSubmitNodeUpdate(ingestCtx, nodeUpdate)
		require.NoError(t, err)
	})
}
