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
	node := IngestibleNode{
		ObjectID: data.ID,
		PropertyMap: map[string]any{
			common.Name.String():                   data.Name,
			common.LastCollected.String():          ingestTime,
			azure.TenantID.String():                strings.ToUpper(data.Properties.TenantID),
			azure.DomainName.String():              data.Properties.DomainName,
			azure.DomainConfigurationType.String(): data.Properties.DomainConfigurationType,
			azure.SyncScope.String():               data.Properties.SyncScope,
			azure.SyncApplicationID.String():       strings.ToUpper(data.Properties.SyncApplicationID),
		},
		Labels: []graph.Kind{azure.EntraDS},
	}

	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.FilteredSyncEnabled.String(), data.Properties.FilteredSync)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.NTLMV1Enabled.String(), data.Properties.DomainSecuritySettings.NTLMV1)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.TLSV1Enabled.String(), data.Properties.DomainSecuritySettings.TLSV1)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.SyncNTLMPasswordsEnabled.String(), data.Properties.DomainSecuritySettings.SyncNTLMPasswords)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.SyncKerberosPasswordsEnabled.String(), data.Properties.DomainSecuritySettings.SyncKerberosPasswords)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.SyncOnPremPasswordsEnabled.String(), data.Properties.DomainSecuritySettings.SyncOnPremPasswords)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.KerberosRC4EncryptionEnabled.String(), data.Properties.DomainSecuritySettings.KerberosRC4Encryption)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.KerberosArmoringEnabled.String(), data.Properties.DomainSecuritySettings.KerberosArmoring)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.LDAPSigningEnabled.String(), data.Properties.DomainSecuritySettings.LDAPSigning)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.ChannelBindingEnabled.String(), data.Properties.DomainSecuritySettings.ChannelBinding)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.SyncOnPremSAMAccountNameEnabled.String(), data.Properties.DomainSecuritySettings.SyncOnPremSAMAccountName)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.LDAPSEnabled.String(), data.Properties.LDAPSSettings.LDAPS)
	setAzureDomainServiceBooleanProperty(node.PropertyMap, azure.LDAPSExternalAccessEnabled.String(), data.Properties.LDAPSSettings.ExternalAccess)

	return node
}

func setAzureDomainServiceBooleanProperty(properties map[string]any, propertyName, rawValue string) {
	switch {
	case strings.EqualFold(strings.TrimSpace(rawValue), "Enabled"):
		properties[propertyName] = true
	case strings.EqualFold(strings.TrimSpace(rawValue), "Disabled"):
		properties[propertyName] = false
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
			Kind:  azure.EntraDS,
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
					Kind:  azure.EntraDS,
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
