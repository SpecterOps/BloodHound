// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package ein_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/bloodhoundad/azurehound/v2/models"
	azure2 "github.com/bloodhoundad/azurehound/v2/models/azure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAzureDomainServiceToNode(t *testing.T) {
	ingestTime := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	data := ein.AzureDomainService{
		ID:   "/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG/PROVIDERS/MICROSOFT.AAD/DOMAINSERVICES/EXAMPLE.COM",
		Name: "example.com",
		Properties: ein.AzureDomainServiceProperties{
			TenantID:                "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
			DomainName:              "example.com",
			DomainConfigurationType: "FullySynced",
			FilteredSync:            "Enabled",
			SyncScope:               "CloudOnly",
			SyncApplicationID:       "75f5e42e-3d9f-472a-9d55-c387e29eacce",
			DomainSecuritySettings: ein.AzureDomainServiceSecuritySettings{
				NTLMV1:                   "Enabled",
				TLSV1:                    "Disabled",
				SyncNTLMPasswords:        "Enabled",
				SyncKerberosPasswords:    "Enabled",
				SyncOnPremPasswords:      "Enabled",
				KerberosRC4Encryption:    "Enabled",
				KerberosArmoring:         "Disabled",
				LDAPSigning:              "Disabled",
				ChannelBinding:           "Disabled",
				SyncOnPremSAMAccountName: "Disabled",
			},
			LDAPSSettings: ein.AzureDomainServiceLDAPSSettings{
				LDAPS:          "Enabled",
				ExternalAccess: "Enabled",
			},
		},
	}

	node := ein.ConvertAzureDomainServiceToNode(data, ingestTime)

	assert.Equal(t, data.ID, node.ObjectID)
	assert.Equal(t, []graph.Kind{azure.EntraDS}, node.Labels)
	require.Len(t, node.PropertyMap, 20)
	assert.Equal(t, data.Name, node.PropertyMap[common.Name.String()])
	assert.Equal(t, ingestTime, node.PropertyMap[common.LastCollected.String()])
	assert.Equal(t, strings.ToUpper(data.Properties.TenantID), node.PropertyMap[azure.TenantID.String()])
	assert.Equal(t, data.Properties.DomainName, node.PropertyMap[azure.DomainName.String()])
	assert.Equal(t, data.Properties.DomainConfigurationType, node.PropertyMap[azure.DomainConfigurationType.String()])
	assert.Equal(t, true, node.PropertyMap[azure.FilteredSyncEnabled.String()])
	assert.Equal(t, data.Properties.SyncScope, node.PropertyMap[azure.SyncScope.String()])
	assert.Equal(t, strings.ToUpper(data.Properties.SyncApplicationID), node.PropertyMap[azure.SyncApplicationID.String()])
	assert.Equal(t, true, node.PropertyMap[azure.NTLMV1Enabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.TLSV1Enabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.SyncNTLMPasswordsEnabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.SyncKerberosPasswordsEnabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.SyncOnPremPasswordsEnabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.KerberosRC4EncryptionEnabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.KerberosArmoringEnabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.LDAPSigningEnabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.ChannelBindingEnabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.SyncOnPremSAMAccountNameEnabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.LDAPSEnabled.String()])
	assert.Equal(t, true, node.PropertyMap[azure.LDAPSExternalAccessEnabled.String()])
}

func TestConvertAzureDomainServiceToNodeOmitsUnknownBooleanSettings(t *testing.T) {
	data := ein.AzureDomainService{
		Properties: ein.AzureDomainServiceProperties{
			FilteredSync: " enabled ",
			DomainSecuritySettings: ein.AzureDomainServiceSecuritySettings{
				NTLMV1: "FutureValue",
			},
			LDAPSSettings: ein.AzureDomainServiceLDAPSSettings{
				LDAPS: "Disabled",
			},
		},
	}

	node := ein.ConvertAzureDomainServiceToNode(data, time.Time{})

	assert.Equal(t, true, node.PropertyMap[azure.FilteredSyncEnabled.String()])
	assert.Equal(t, false, node.PropertyMap[azure.LDAPSEnabled.String()])
	assert.NotContains(t, node.PropertyMap, azure.NTLMV1Enabled.String())
	assert.NotContains(t, node.PropertyMap, azure.TLSV1Enabled.String())
}

func TestConvertAzureDomainServiceToRels(t *testing.T) {
	data := ein.AzureDomainService{
		ID:              "/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG/PROVIDERS/MICROSOFT.AAD/DOMAINSERVICES/EXAMPLE.COM",
		ResourceGroupID: "/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG",
	}

	rels := ein.ConvertAzureDomainServiceToRels(data)

	require.Len(t, rels, 1)
	assert.Equal(t, data.ResourceGroupID, rels[0].Source.Value)
	assert.Equal(t, azure.ResourceGroup, rels[0].Source.Kind)
	assert.Equal(t, data.ID, rels[0].Target.Value)
	assert.Equal(t, azure.EntraDS, rels[0].Target.Kind)
	assert.Equal(t, azure.Contains, rels[0].RelType)
	assert.Empty(t, rels[0].RelProps)

	data.ResourceGroupID = ""
	assert.Empty(t, ein.ConvertAzureDomainServiceToRels(data))
}

func TestConvertAzureDomainServiceRoleAssignmentToRels(t *testing.T) {
	resourceID := "/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG/PROVIDERS/MICROSOFT.AAD/DOMAINSERVICES/EXAMPLE.COM"
	principalID := "PRINCIPAL-ID"

	testCases := []struct {
		name             string
		roleDefinitionID string
		scope            string
		expectedKind     graph.Kind
		expectedCount    int
	}{
		{name: "owner", roleDefinitionID: constants.OwnerRoleID, scope: strings.ToLower(resourceID), expectedKind: azure.Owner, expectedCount: 1},
		{name: "user access administrator", roleDefinitionID: constants.UserAccessAdminRoleID, scope: resourceID, expectedKind: azure.UserAccessAdministrator, expectedCount: 1},
		{name: "contributor", roleDefinitionID: constants.ContributorRoleID, scope: resourceID, expectedKind: azure.Contributor, expectedCount: 1},
		{name: "domain services contributor", roleDefinitionID: constants.DomainServicesContributorRoleID, scope: resourceID, expectedKind: azure.EntraDSContributor, expectedCount: 1},
		{name: "inherited role", roleDefinitionID: constants.OwnerRoleID, scope: "/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG", expectedCount: 0},
		{name: "unsupported role", roleDefinitionID: "00000000-0000-0000-0000-000000000000", scope: resourceID, expectedCount: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data := models.AzureRoleAssignments{
				ObjectId: resourceID,
				RoleAssignments: []models.AzureRoleAssignment{{
					ObjectId:         resourceID,
					RoleDefinitionId: strings.ToUpper(testCase.roleDefinitionID),
					Assignee: azure2.RoleAssignment{Properties: azure2.RoleAssignmentPropertiesWithScope{
						PrincipalId: principalID,
						Scope:       testCase.scope,
					}},
				}},
			}

			rels := ein.ConvertAzureDomainServiceRoleAssignmentToRels(data)

			require.Len(t, rels, testCase.expectedCount)
			if testCase.expectedCount == 1 {
				assert.Equal(t, principalID, rels[0].Source.Value)
				assert.Equal(t, azure.Entity, rels[0].Source.Kind)
				assert.Equal(t, resourceID, rels[0].Target.Value)
				assert.Equal(t, azure.EntraDS, rels[0].Target.Kind)
				assert.Equal(t, testCase.expectedKind, rels[0].RelType)
				assert.Empty(t, rels[0].RelProps)
			}
		})
	}
}
