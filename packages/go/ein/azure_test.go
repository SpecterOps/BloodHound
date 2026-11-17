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

package ein_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bloodhoundad/azurehound/v2/models"
	azure2 "github.com/bloodhoundad/azurehound/v2/models/azure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAzureRoleEligibilityScheduleInstanceToRel(t *testing.T) {
	testData := models.RoleEligibilityScheduleInstance{
		Id:               "lAPpYvVpN0KRkAEhdxReELKn6QMIlSROgkgWZy9fE3c-1-e",
		RoleDefinitionId: "62e90394-69f5-4237-9190-012177145e10",
		PrincipalId:      "03e9a7b2-9508-4e24-8248-16672f5f1377",
		DirectoryScopeId: "/",
		StartDateTime:    "2024-01-04T01:22:36.867Z",
		TenantId:         "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
	}

	expectedRels := ein.ConvertAzureRoleEligibilityScheduleInstanceToRel(testData)
	require.Len(t, expectedRels, 1)
	expectedRel := expectedRels[0]
	require.Equal(t, expectedRel.Target.Value, fmt.Sprintf("%s@%s", testData.RoleDefinitionId, testData.TenantId))
	require.Equal(t, expectedRel.RelType, azure.AZRoleEligible)
	require.Equal(t, expectedRel.Source.Value, testData.PrincipalId)

	testData = models.RoleEligibilityScheduleInstance{
		Id:               "lAPpYvVpN0KRkAEhdxReELKn6QMIlSROgkgWZy9fE3c-1-e",
		RoleDefinitionId: "62e90394-69f5-4237-9190-012177145e10",
		PrincipalId:      "03e9a7b2-9508-4e24-8248-16672f5f1377",
		DirectoryScopeId: "abc123",
		StartDateTime:    "2024-01-04T01:22:36.867Z",
		TenantId:         "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
	}

	expectedRels = ein.ConvertAzureRoleEligibilityScheduleInstanceToRel(testData)
	require.Len(t, expectedRels, 0)
}

func Test_ConvertAppFederatedIdentityCredential(t *testing.T) {
	testCases := []struct {
		name       string
		testData   models.FICData
		appID      string
		tenantName string
		tenantID   string
	}{
		{
			name: "populates node properties and relationship endpoints",
			testData: models.FICData{
				Audiences:   []string{"api://AzureADTokenExchange"},
				ID:          "bbb6ccd9-91ee-4b64-9f74-865a2f5d55e4",
				Issuer:      "https://token.actions.githubusercontent.com",
				Name:        "GitHubActionsProductionEnv",
				Subject:     "repo:SpecterTst/oidc-actions-test-2:environment:Production",
				Description: "abc123",
			},
			appID:      "a7b9f1c5-1e4b-48d9-b71a-1444fc64cddc",
			tenantName: "SPECTERDEV.ONMICROSOFT1",
			tenantID:   "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
		},
		{
			name: "uppercases object identifiers while preserving empty optional description",
			testData: models.FICData{
				Audiences: []string{
					"api://AzureADTokenExchange",
					"customAudience",
				},
				ID:      "f40abc17-c0ba-4d1d-a4ac-c5025aae55d6",
				Issuer:  "https://token.actions.githubusercontent.com",
				Name:    "ThisIsABadRule",
				Subject: "repo:SpecterTst/oidc-actions-test-1:*",
			},
			appID:      "75414635-8836-4d31-82c5-1f783e520873",
			tenantName: "specterdev.onmicrosoft1",
			tenantID:   "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			node, rel := ein.ConvertAppFederatedIdentityCredential(testCase.testData, testCase.appID, testCase.tenantName, testCase.tenantID)

			require.True(t, node.IsValid())
			require.True(t, rel.IsValid())

			assert.Equal(t, testCase.testData.ID, node.ObjectID)
			require.Len(t, node.Labels, 2)
			assert.Equal(t, azure.FederatedIdentityCredential, node.Labels[0])
			assert.Equal(t, azure.Entity, node.Labels[1])

			assert.Equal(t, testCase.testData.Description, node.PropertyMap[common.Description.String()])
			assert.Equal(t, fmt.Sprintf("%s@%s", testCase.testData.Name, testCase.tenantName), node.PropertyMap[common.Name.String()])
			assert.Equal(t, testCase.testData.Issuer, node.PropertyMap[azure.Issuer.String()])
			assert.Equal(t, testCase.testData.Audiences, node.PropertyMap[azure.Audiences.String()])
			assert.Equal(t, testCase.testData.Subject, node.PropertyMap[azure.Subject.String()])
			assert.Equal(t, strings.ToUpper(testCase.tenantID), node.PropertyMap[azure.TenantID.String()])

			assert.Equal(t, azure.AZAuthenticatesTo, rel.RelType)
			assert.Equal(t, azure.FederatedIdentityCredential, rel.Source.Kind)
			assert.Equal(t, testCase.testData.ID, rel.Source.Value)
			assert.Equal(t, azure.App, rel.Target.Kind)
			assert.Equal(t, testCase.appID, rel.Target.Value)
			assert.Empty(t, rel.RelProps)
		})
	}
}

func TestConvertAzureManagementGroupContributorToRels(t *testing.T) {
	managementGroupId := "/providers/Microsoft.Management/managementGroups/mg-group-1"
	principalId := "abcde-12345-fghij-67890"

	t.Run("creates contributor relationship when scope matches", func(t *testing.T) {
		data := models.ManagementGroupContributors{
			ManagementGroupId: managementGroupId,
			Contributors: []models.ManagementGroupContributor{
				{
					ManagementGroupId: managementGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       managementGroupId,
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureManagementGroupContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, principalId, rels[0].Source.Value)
		assert.Equal(t, azure.Entity, rels[0].Source.Kind)
		assert.Equal(t, managementGroupId, rels[0].Target.Value)
		assert.Equal(t, azure.ManagementGroup, rels[0].Target.Kind)
		assert.Equal(t, azure.Contributor, rels[0].RelType)
	})

	t.Run("skips relationship when scope does not match", func(t *testing.T) {
		data := models.ManagementGroupContributors{
			ManagementGroupId: managementGroupId,
			Contributors: []models.ManagementGroupContributor{
				{
					ManagementGroupId: managementGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       "/providers/Microsoft.Management/managementGroups/different-mg",
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureManagementGroupContributorToRels(data)

		require.Len(t, rels, 0)
	})

	t.Run("creates contributor relationship when scope differs only by case", func(t *testing.T) {
		data := models.ManagementGroupContributors{
			ManagementGroupId: managementGroupId,
			Contributors: []models.ManagementGroupContributor{
				{
					ManagementGroupId: managementGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       strings.ToUpper(managementGroupId),
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureManagementGroupContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, managementGroupId, rels[0].Target.Value)
	})
}

func TestConvertAzureResourceGroupContributorToRels(t *testing.T) {
	resourceGroupId := "/subscriptions/sub-1/resourceGroups/rg-1"
	principalId := "abcde-12345-fghij-67890"

	t.Run("creates contributor relationship when scope matches", func(t *testing.T) {
		data := models.ResourceGroupContributors{
			ResourceGroupId: resourceGroupId,
			Contributors: []models.ResourceGroupContributor{
				{
					ResourceGroupId: resourceGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       resourceGroupId,
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureResourceGroupContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, principalId, rels[0].Source.Value)
		assert.Equal(t, azure.Entity, rels[0].Source.Kind)
		assert.Equal(t, resourceGroupId, rels[0].Target.Value)
		assert.Equal(t, azure.ResourceGroup, rels[0].Target.Kind)
		assert.Equal(t, azure.Contributor, rels[0].RelType)
	})

	t.Run("skips relationship when scope does not match", func(t *testing.T) {
		data := models.ResourceGroupContributors{
			ResourceGroupId: resourceGroupId,
			Contributors: []models.ResourceGroupContributor{
				{
					ResourceGroupId: resourceGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       "/subscriptions/sub-1/resourceGroups/different-rg",
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureResourceGroupContributorToRels(data)

		require.Len(t, rels, 0)
	})

	t.Run("creates contributor relationship when scope differs only by case", func(t *testing.T) {
		data := models.ResourceGroupContributors{
			ResourceGroupId: resourceGroupId,
			Contributors: []models.ResourceGroupContributor{
				{
					ResourceGroupId: resourceGroupId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       strings.ToUpper(resourceGroupId),
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureResourceGroupContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, resourceGroupId, rels[0].Target.Value)
	})
}

func TestConvertAzureSubscriptionContributorToRels(t *testing.T) {
	subscriptionId := "/subscriptions/sub-1"
	principalId := "abcde-12345-fghij-67890"

	t.Run("creates contributor relationship when scope matches", func(t *testing.T) {
		data := models.SubscriptionContributors{
			SubscriptionId: subscriptionId,
			Contributors: []models.SubscriptionContributor{
				{
					SubscriptionId: subscriptionId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       subscriptionId,
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureSubscriptionContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, principalId, rels[0].Source.Value)
		assert.Equal(t, azure.Entity, rels[0].Source.Kind)
		assert.Equal(t, subscriptionId, rels[0].Target.Value)
		assert.Equal(t, azure.Subscription, rels[0].Target.Kind)
		assert.Equal(t, azure.Contributor, rels[0].RelType)
	})

	t.Run("skips relationship when scope does not match", func(t *testing.T) {
		data := models.SubscriptionContributors{
			SubscriptionId: subscriptionId,
			Contributors: []models.SubscriptionContributor{
				{
					SubscriptionId: subscriptionId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       "/subscriptions/different-sub",
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureSubscriptionContributorToRels(data)

		require.Len(t, rels, 0)
	})

	t.Run("creates contributor relationship when scope differs only by case", func(t *testing.T) {
		data := models.SubscriptionContributors{
			SubscriptionId: subscriptionId,
			Contributors: []models.SubscriptionContributor{
				{
					SubscriptionId: subscriptionId,
					Contributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       strings.ToUpper(subscriptionId),
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureSubscriptionContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, subscriptionId, rels[0].Target.Value)
	})
}

func Test_ConvertAzureRoleManagementPolicyAssignment(t *testing.T) {
	for _, testCase := range []struct {
		name             string
		requiresApproval bool
		userApprovers    []string
		groupApprovers   []string
	}{
		{name: "approval disabled", userApprovers: []string{"user-approver"}, groupApprovers: []string{"group-approver"}},
		{name: "user approvers", requiresApproval: true, userApprovers: []string{"user-approver"}},
		{name: "group approvers", requiresApproval: true, groupApprovers: []string{"group-approver"}},
		{name: "user and group approvers", requiresApproval: true, userApprovers: []string{"user-approver"}, groupApprovers: []string{"group-approver"}},
		{name: "missing approvers", requiresApproval: true},
		{name: "empty approvers", requiresApproval: true, userApprovers: []string{}, groupApprovers: []string{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				policyAssignment = models.RoleManagementPolicyAssignment{
					RoleDefinitionId:                                  "role-1234",
					TenantId:                                          "tenant-1234",
					EndUserAssignmentRequiresApproval:                 testCase.requiresApproval,
					EndUserAssignmentUserApprovers:                    testCase.userApprovers,
					EndUserAssignmentGroupApprovers:                   testCase.groupApprovers,
					EndUserAssignmentRequiresCAPAuthenticationContext: true,
					EndUserAssignmentRequiresMFA:                      true,
					EndUserAssignmentRequiresJustification:            true,
					EndUserAssignmentRequiresTicketInformation:        true,
				}
				node, relationships = ein.ConvertAzureRoleManagementPolicyAssignment(policyAssignment)
			)

			assert.Equal(t, "role-1234@tenant-1234", node.ObjectID)
			require.Len(t, node.Labels, 1)
			assert.Equal(t, azure.Role, node.Labels[0])
			assert.Equal(t, map[string]any{
				azure.RoleDefinitionId.String():                                  "ROLE-1234",
				azure.TenantID.String():                                          "TENANT-1234",
				azure.EndUserAssignmentRequiresApproval.String():                 testCase.requiresApproval,
				azure.EndUserAssignmentUserApprovers.String():                    testCase.userApprovers,
				azure.EndUserAssignmentGroupApprovers.String():                   testCase.groupApprovers,
				azure.EndUserAssignmentRequiresCAPAuthenticationContext.String(): true,
				azure.EndUserAssignmentRequiresMFA.String():                      true,
				azure.EndUserAssignmentRequiresJustification.String():            true,
				azure.EndUserAssignmentRequiresTicketInformation.String():        true,
			}, node.PropertyMap)

			var expectedApprovers = append(append([]string{}, testCase.userApprovers...), testCase.groupApprovers...)
			if !testCase.requiresApproval {
				expectedApprovers = nil
			} else if len(expectedApprovers) == 0 {
				expectedApprovers = []string{
					strings.ToUpper(azure.PrivilegedRoleAdministratorRole) + "@tenant-1234",
					strings.ToUpper(azure.CompanyAdministratorRole) + "@tenant-1234",
				}
			}
			require.Len(t, relationships, len(expectedApprovers))
			for index, relationship := range relationships {
				assert.Equal(t, expectedApprovers[index], relationship.Source.Value)
				assert.Equal(t, "role-1234@tenant-1234", relationship.Target.Value)
				assert.Equal(t, azure.Role, relationship.Target.Kind)
				assert.Equal(t, azure.AZRoleApprover, relationship.RelType)
				if len(testCase.userApprovers)+len(testCase.groupApprovers) == 0 {
					assert.Equal(t, azure.Role, relationship.Source.Kind)
				} else if index < len(testCase.userApprovers) {
					assert.Equal(t, azure.User, relationship.Source.Kind)
				} else {
					assert.Equal(t, azure.Group, relationship.Source.Kind)
				}
			}
		})
	}
}

func TestConvertAzureVirtualMachineAvereContributorToRels(t *testing.T) {
	virtualMachineId := "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Compute/virtualMachines/vm-1"
	principalId := "abcde-12345-fghij-67890"

	t.Run("creates avere contributor relationship when scope matches", func(t *testing.T) {
		data := models.VirtualMachineAvereContributors{
			VirtualMachineId: virtualMachineId,
			AvereContributors: []models.VirtualMachineAvereContributor{
				{
					VirtualMachineId: virtualMachineId,
					AvereContributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       virtualMachineId,
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureVirtualMachineAvereContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, principalId, rels[0].Source.Value)
		assert.Equal(t, azure.Entity, rels[0].Source.Kind)
		assert.Equal(t, virtualMachineId, rels[0].Target.Value)
		assert.Equal(t, azure.VM, rels[0].Target.Kind)
		assert.Equal(t, azure.AvereContributor, rels[0].RelType)
	})

	t.Run("creates avere contributor relationship when scope differs only by case", func(t *testing.T) {
		data := models.VirtualMachineAvereContributors{
			VirtualMachineId: virtualMachineId,
			AvereContributors: []models.VirtualMachineAvereContributor{
				{
					VirtualMachineId: virtualMachineId,
					AvereContributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       strings.ToUpper(virtualMachineId),
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureVirtualMachineAvereContributorToRels(data)

		require.Len(t, rels, 1)
		assert.Equal(t, virtualMachineId, rels[0].Target.Value)
	})

	t.Run("skips relationship when scope does not match", func(t *testing.T) {
		data := models.VirtualMachineAvereContributors{
			VirtualMachineId: virtualMachineId,
			AvereContributors: []models.VirtualMachineAvereContributor{
				{
					VirtualMachineId: virtualMachineId,
					AvereContributor: azure2.RoleAssignment{
						Properties: azure2.RoleAssignmentPropertiesWithScope{
							PrincipalId: principalId,
							Scope:       "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Compute/virtualMachines/different-vm",
						},
					},
				},
			},
		}

		rels := ein.ConvertAzureVirtualMachineAvereContributorToRels(data)

		require.Len(t, rels, 0)
	})
}
