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

package graphify

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertGenericNode(t *testing.T) {
	t.Run("flag off: objectid is uppercased", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, false)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "OBJECTID", converted.NodeProps[0].ObjectID)
	})

	t.Run("flag on: objectid preserves original case", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "ObjectId",
				Kinds: []string{"SomeKind"},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, true)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "ObjectId", converted.NodeProps[0].ObjectID)
	})

	t.Run("flag off: environmentid is uppercased", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
				Properties: map[string]any{
					graphschema.EnvironmentIDKey: "my-github-org",
				},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, false)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "MY-GITHUB-ORG", converted.NodeProps[0].PropertyMap[graphschema.EnvironmentIDKey])
	})

	t.Run("flag on: environmentid preserves original case", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
				Properties: map[string]any{
					graphschema.EnvironmentIDKey: "my-github-org",
				},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, true)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "my-github-org", converted.NodeProps[0].PropertyMap[graphschema.EnvironmentIDKey])
	})

	t.Run("flag on: domainsid is still uppercased for consistency with sharphound ingestion", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
				Properties: map[string]any{
					ad.DomainSID.String(): "s-1-5-21-abc",
				},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, true)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "S-1-5-21-ABC", converted.NodeProps[0].PropertyMap[ad.DomainSID.String()])
	})

	t.Run("flag on: tenantid is still uppercased for consistency with azurehound ingestion", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
				Properties: map[string]any{
					azure.TenantID.String(): "tenant-abc",
				},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, true)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "TENANT-ABC", converted.NodeProps[0].PropertyMap[azure.TenantID.String()])
	})
}

func TestConvertAzureManagedCluster_NodeResourceGroupID(t *testing.T) {
	// The collector emits subscriptionId already prefixed with "/subscriptions/".
	// The converter must not re-prepend it (which produced a doubled
	// "/subscriptions//subscriptions/" node resource group id and a stub node).
	t.Run("collector subscriptionId with /subscriptions/ prefix is not doubled", func(t *testing.T) {
		raw := json.RawMessage(`{
			"id": "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/RG-1234/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/AKS-1234",
			"name": "aks-1234",
			"subscriptionId": "/subscriptions/sub-1234",
			"resourceGroupId": "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/RG-1234",
			"tenantId": "tenant-1234",
			"properties": {"nodeResourceGroup": "MC_rg-1234_aks-1234_centralus"}
		}`)

		converted := &ConvertedAzureData{}
		convertAzureManagedCluster(raw, converted, time.Now())

		require.Len(t, converted.NodeProps, 1)
		nrg, ok := converted.NodeProps[0].PropertyMap[azure.NodeResourceGroupID.String()].(string)
		require.True(t, ok, "noderesourcegroupid property must be a string")
		assert.NotContains(t, nrg, "/subscriptions//subscriptions/", "must not double the subscriptions prefix")
		assert.Equal(t, 1, strings.Count(strings.ToLower(nrg), "/subscriptions/"), "exactly one /subscriptions/ segment")
		assert.Equal(t, "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/MC_RG-1234_AKS-1234_CENTRALUS", nrg)
	})

	t.Run("bare-guid subscriptionId still gets a single /subscriptions/ prefix", func(t *testing.T) {
		raw := json.RawMessage(`{
			"id": "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/RG-1234/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/AKS-1234",
			"name": "aks-1234",
			"subscriptionId": "sub-1234",
			"resourceGroupId": "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/RG-1234",
			"tenantId": "tenant-1234",
			"properties": {"nodeResourceGroup": "MC_rg-1234_aks-1234_centralus"}
		}`)

		converted := &ConvertedAzureData{}
		convertAzureManagedCluster(raw, converted, time.Now())

		require.Len(t, converted.NodeProps, 1)
		nrg, ok := converted.NodeProps[0].PropertyMap[azure.NodeResourceGroupID.String()].(string)
		require.True(t, ok)
		assert.NotContains(t, nrg, "/subscriptions//subscriptions/")
		assert.Equal(t, 1, strings.Count(strings.ToLower(nrg), "/subscriptions/"))
		assert.Equal(t, "/SUBSCRIPTIONS/SUB-1234/RESOURCEGROUPS/MC_RG-1234_AKS-1234_CENTRALUS", nrg)
	})
}
