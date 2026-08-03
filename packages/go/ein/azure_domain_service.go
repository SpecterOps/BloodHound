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

package ein

import (
	"slices"
	"strings"
	"time"

	"github.com/bloodhoundad/azurehound/v2/constants"
	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

type AzureDomainServiceLDAPSSettings struct {
	LDAPS          string `json:"ldaps"`
	ExternalAccess string `json:"externalAccess"`
}

type AzureDomainServiceSecuritySettings struct {
	NTLMV1                   string `json:"ntlmV1"`
	TLSV1                    string `json:"tlsV1"`
	SyncNTLMPasswords        string `json:"syncNtlmPasswords"`
	SyncKerberosPasswords    string `json:"syncKerberosPasswords"`
	SyncOnPremPasswords      string `json:"syncOnPremPasswords"`
	KerberosRC4Encryption    string `json:"kerberosRc4Encryption"`
	KerberosArmoring         string `json:"kerberosArmoring"`
	LDAPSigning              string `json:"ldapSigning"`
	ChannelBinding           string `json:"channelBinding"`
	SyncOnPremSAMAccountName string `json:"syncOnPremSamAccountName"`
}

type AzureDomainServiceProperties struct {
	TenantID                string                             `json:"tenantId"`
	DomainName              string                             `json:"domainName"`
	DomainConfigurationType string                             `json:"domainConfigurationType"`
	FilteredSync            string                             `json:"filteredSync"`
	SyncScope               string                             `json:"syncScope"`
	SyncApplicationID       string                             `json:"syncApplicationId"`
	DomainSecuritySettings  AzureDomainServiceSecuritySettings `json:"domainSecuritySettings"`
	LDAPSSettings           AzureDomainServiceLDAPSSettings    `json:"ldapsSettings"`
}

type AzureDomainService struct {
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	ResourceGroupID   string                       `json:"resourceGroupId"`
	ResourceGroupName string                       `json:"resourceGroupName"`
	Properties        AzureDomainServiceProperties `json:"properties"`
}

func ConvertAzureDomainServiceToNode(data AzureDomainService, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		ObjectID: data.ID,
		PropertyMap: map[string]any{
			common.Name.String():                    data.Name,
			common.LastCollected.String():           ingestTime,
			azure.TenantID.String():                 strings.ToUpper(data.Properties.TenantID),
			azure.DomainName.String():               data.Properties.DomainName,
			azure.DomainConfigurationType.String():  data.Properties.DomainConfigurationType,
			azure.FilteredSync.String():             data.Properties.FilteredSync,
			azure.SyncScope.String():                data.Properties.SyncScope,
			azure.SyncApplicationID.String():        strings.ToUpper(data.Properties.SyncApplicationID),
			azure.NTLMV1.String():                   data.Properties.DomainSecuritySettings.NTLMV1,
			azure.TLSV1.String():                    data.Properties.DomainSecuritySettings.TLSV1,
			azure.SyncNTLMPasswords.String():        data.Properties.DomainSecuritySettings.SyncNTLMPasswords,
			azure.SyncKerberosPasswords.String():    data.Properties.DomainSecuritySettings.SyncKerberosPasswords,
			azure.SyncOnPremPasswords.String():      data.Properties.DomainSecuritySettings.SyncOnPremPasswords,
			azure.KerberosRC4Encryption.String():    data.Properties.DomainSecuritySettings.KerberosRC4Encryption,
			azure.KerberosArmoring.String():         data.Properties.DomainSecuritySettings.KerberosArmoring,
			azure.LDAPSigning.String():              data.Properties.DomainSecuritySettings.LDAPSigning,
			azure.ChannelBinding.String():           data.Properties.DomainSecuritySettings.ChannelBinding,
			azure.SyncOnPremSAMAccountName.String(): data.Properties.DomainSecuritySettings.SyncOnPremSAMAccountName,
			azure.LDAPS.String():                    data.Properties.LDAPSSettings.LDAPS,
			azure.LDAPSExternalAccess.String():      data.Properties.LDAPSSettings.ExternalAccess,
		},
		Labels: []graph.Kind{azure.DomainService},
	}
}

func ConvertAzureDomainServiceToRels(data AzureDomainService) []IngestibleRelationship {
	if data.ResourceGroupID == "" {
		return nil
	}

	return []IngestibleRelationship{NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: data.ResourceGroupID,
			Kind:  azure.ResourceGroup,
		},
		IngestibleEndpoint{
			Value: data.ID,
			Kind:  azure.DomainService,
		},
		IngestibleRel{
			RelProps: map[string]any{},
			RelType:  azure.Contains,
		},
	)}
}

func ConvertAzureDomainServiceRoleAssignmentToRels(data models.AzureRoleAssignments) []IngestibleRelationship {
	var relationships []IngestibleRelationship
	allowedRoleIDs := []string{
		constants.OwnerRoleID,
		constants.UserAccessAdminRoleID,
		constants.ContributorRoleID,
		constants.DomainServicesContributorRoleID,
	}

	for _, roleAssignment := range data.RoleAssignments {
		roleID := strings.ToLower(roleAssignment.RoleDefinitionId)
		if strings.EqualFold(roleAssignment.Assignee.Properties.Scope, roleAssignment.ObjectId) && slices.Contains(allowedRoleIDs, roleID) {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: roleAssignment.Assignee.GetPrincipalId(),
					Kind:  azure.Entity,
				},
				IngestibleEndpoint{
					Value: data.ObjectId,
					Kind:  azure.DomainService,
				},
				IngestibleRel{
					RelProps: map[string]any{},
					RelType:  KindFromRoleId(roleID),
				},
			))
		}
	}

	return relationships
}
