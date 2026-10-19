// Copyright 2024 Specter Ops, Inc.
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

package hybrid

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
)

func hybridTestNode(id graph.ID, kind graph.Kind, props map[string]any) *graph.Node {
	return graph.NewNode(id, graph.AsProperties(props), kind)
}

func TestCloudManagedGroupSID(t *testing.T) {
	const sid = "S-1-5-21-1111111111-2222222222-3333333333-4100"

	t.Run("cloud-managed with on-prem SID", func(t *testing.T) {
		cloudManaged, hasSID, got, err := cloudManagedGroupSID(hybridTestNode(1, azure.Group, map[string]any{
			azure.IsCloudManaged.String(): true,
			azure.OnPremID.String():       sid,
		}))
		assert.NoError(t, err)
		assert.True(t, cloudManaged)
		assert.True(t, hasSID)
		assert.Equal(t, sid, got)
	})

	t.Run("not cloud-managed is skipped", func(t *testing.T) {
		cloudManaged, _, _, err := cloudManagedGroupSID(hybridTestNode(2, azure.Group, map[string]any{
			azure.IsCloudManaged.String(): false,
			azure.OnPremID.String():       sid,
		}))
		assert.NoError(t, err)
		assert.False(t, cloudManaged)
	})

	t.Run("cloud-managed without on-prem SID has no SID", func(t *testing.T) {
		cloudManaged, hasSID, _, err := cloudManagedGroupSID(hybridTestNode(3, azure.Group, map[string]any{
			azure.IsCloudManaged.String(): true,
		}))
		assert.NoError(t, err)
		assert.True(t, cloudManaged)
		assert.False(t, hasSID)
	})

	t.Run("missing isCloudManaged property is skipped", func(t *testing.T) {
		cloudManaged, _, _, err := cloudManagedGroupSID(hybridTestNode(4, azure.Group, map[string]any{
			azure.OnPremID.String(): sid,
		}))
		assert.NoError(t, err)
		assert.False(t, cloudManaged)
	})
}

func TestIsCloudSyncExcludedGroup(t *testing.T) {
	const prefix = "S-1-5-21-1111111111-2222222222-3333333333"

	t.Run("adminCount==1 is excluded", func(t *testing.T) {
		assert.True(t, isCloudSyncExcludedGroup(hybridTestNode(1, ad.Group, map[string]any{
			ad.AdminCount.String():   true,
			common.ObjectID.String(): prefix + "-4100",
		})))
	})

	t.Run("built-in blocklist RIDs are excluded", func(t *testing.T) {
		for _, rid := range []string{"512", "513", "514", "518", "519", "520", "525", "526", "527"} {
			assert.Truef(t, isCloudSyncExcludedGroup(hybridTestNode(2, ad.Group, map[string]any{
				common.ObjectID.String(): prefix + "-" + rid,
			})), "RID %s should be excluded", rid)
		}
	})

	t.Run("ordinary group is not excluded", func(t *testing.T) {
		assert.False(t, isCloudSyncExcludedGroup(hybridTestNode(3, ad.Group, map[string]any{
			common.ObjectID.String(): prefix + "-4100",
		})))
	})

	t.Run("adminCount==false ordinary group is not excluded", func(t *testing.T) {
		assert.False(t, isCloudSyncExcludedGroup(hybridTestNode(4, ad.Group, map[string]any{
			ad.AdminCount.String():   false,
			common.ObjectID.String(): prefix + "-4100",
		})))
	})
}
