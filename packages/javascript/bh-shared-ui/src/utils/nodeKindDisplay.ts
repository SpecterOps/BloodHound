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

import { ActiveDirectoryNodeKind, AzureNodeKind } from '../graphSchema';

const NODE_KIND_DISPLAY_LABELS: Record<string, string> = {
    [ActiveDirectoryNodeKind.Entity]: 'Entity',
    [ActiveDirectoryNodeKind.User]: 'User',
    [ActiveDirectoryNodeKind.Computer]: 'Computer',
    [ActiveDirectoryNodeKind.Group]: 'Group',
    [ActiveDirectoryNodeKind.GPO]: 'Group Policy Object',
    [ActiveDirectoryNodeKind.OU]: 'Organizational Unit',
    [ActiveDirectoryNodeKind.Container]: 'Container',
    [ActiveDirectoryNodeKind.Domain]: 'Domain',
    [ActiveDirectoryNodeKind.LocalGroup]: 'AD Local Group',
    [ActiveDirectoryNodeKind.LocalUser]: 'AD Local User',
    [ActiveDirectoryNodeKind.AIACA]: 'AIA Certificate Authority',
    [ActiveDirectoryNodeKind.RootCA]: 'Root Certificate Authority',
    [ActiveDirectoryNodeKind.EnterpriseCA]: 'Enterprise Certificate Authority',
    [ActiveDirectoryNodeKind.NTAuthStore]: 'NTAuth Store',
    [ActiveDirectoryNodeKind.CertTemplate]: 'Certificate Template',
    [ActiveDirectoryNodeKind.IssuancePolicy]: 'Issuance Policy',
    [AzureNodeKind.Entity]: 'Azure Entity',
    [AzureNodeKind.VMScaleSet]: 'Azure VM Scale Set',
    [AzureNodeKind.App]: 'Azure Application',
    [AzureNodeKind.Role]: 'Azure Role',
    [AzureNodeKind.Device]: 'Azure Device',
    [AzureNodeKind.FunctionApp]: 'Azure Function App',
    [AzureNodeKind.Group]: 'Azure Group',
    [AzureNodeKind.KeyVault]: 'Azure Key Vault',
    [AzureNodeKind.ManagementGroup]: 'Azure Management Group',
    [AzureNodeKind.ResourceGroup]: 'Azure Resource Group',
    [AzureNodeKind.ServicePrincipal]: 'Azure Service Principal',
    [AzureNodeKind.Subscription]: 'Azure Subscription',
    [AzureNodeKind.Tenant]: 'Azure Tenant',
    [AzureNodeKind.User]: 'Azure User',
    [AzureNodeKind.VM]: 'Azure Virtual Machine',
    [AzureNodeKind.ManagedCluster]: 'Azure Managed Cluster',
    [AzureNodeKind.ContainerRegistry]: 'Azure Container Registry',
    [AzureNodeKind.WebApp]: 'Azure Web App',
    [AzureNodeKind.LogicApp]: 'Azure Logic App',
    [AzureNodeKind.AutomationAccount]: 'Azure Automation Account',
    [AzureNodeKind.FederatedIdentityCredential]: 'Azure Federated Identity Credential',
};

export const getNodeKindDisplayLabel = (kind: string): string => NODE_KIND_DISPLAY_LABELS[kind] ?? kind;
