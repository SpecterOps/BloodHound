-- Copyright 2026 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

-- +goose Up
UPDATE schema_node_kinds AS snk
SET display_name = display_names.display_name,
    updated_at = NOW()
FROM kind AS k,
    (VALUES
        ('GPO', 'Group Policy Object'),
        ('Base', 'Entity'),
        ('OU', 'Organizational Unit'),
        ('ADLocalGroup', 'AD Local Group'),
        ('ADLocalUser', 'AD Local User'),
        ('AIACA', 'AIA Certificate Authority'),
        ('RootCA', 'Root Certificate Authority'),
        ('EnterpriseCA', 'Enterprise Certificate Authority'),
        ('NTAuthStore', 'NTAuth Store'),
        ('CertTemplate', 'Certificate Template'),
        ('IssuancePolicy', 'Issuance Policy'),
        ('AZBase', 'Azure Entity'),
        ('AZVMScaleSet', 'Azure VM Scale Set'),
        ('AZApp', 'Azure Application'),
        ('AZRole', 'Azure Role'),
        ('AZDevice', 'Azure Device'),
        ('AZFunctionApp', 'Azure Function App'),
        ('AZGroup', 'Azure Group'),
        ('AZKeyVault', 'Azure Key Vault'),
        ('AZManagementGroup', 'Azure Management Group'),
        ('AZResourceGroup', 'Azure Resource Group'),
        ('AZServicePrincipal', 'Azure Service Principal'),
        ('AZSubscription', 'Azure Subscription'),
        ('AZTenant', 'Azure Tenant'),
        ('AZUser', 'Azure User'),
        ('AZVM', 'Azure Virtual Machine'),
        ('AZManagedCluster', 'Azure Managed Cluster'),
        ('AZContainerRegistry', 'Azure Container Registry'),
        ('AZWebApp', 'Azure Web App'),
        ('AZLogicApp', 'Azure Logic App'),
        ('AZAutomationAccount', 'Azure Automation Account'),
        ('AZFederatedIdentityCredential', 'Azure Federated Identity Credential')
    ) AS display_names(name, display_name)
WHERE snk.kind_id = k.id
  AND k.name = display_names.name;

-- +goose Down
UPDATE schema_node_kinds AS snk
SET display_name = raw_names.name,
    updated_at = NOW()
FROM kind AS k,
    (VALUES
        ('GPO'),
        ('Base'),
        ('OU'),
        ('ADLocalGroup'),
        ('ADLocalUser'),
        ('AIACA'),
        ('RootCA'),
        ('EnterpriseCA'),
        ('NTAuthStore'),
        ('CertTemplate'),
        ('IssuancePolicy'),
        ('AZBase'),
        ('AZVMScaleSet'),
        ('AZApp'),
        ('AZRole'),
        ('AZDevice'),
        ('AZFunctionApp'),
        ('AZGroup'),
        ('AZKeyVault'),
        ('AZManagementGroup'),
        ('AZResourceGroup'),
        ('AZServicePrincipal'),
        ('AZSubscription'),
        ('AZTenant'),
        ('AZUser'),
        ('AZVM'),
        ('AZManagedCluster'),
        ('AZContainerRegistry'),
        ('AZWebApp'),
        ('AZLogicApp'),
        ('AZAutomationAccount'),
        ('AZFederatedIdentityCredential')
    ) AS raw_names(name)
WHERE snk.kind_id = k.id
  AND k.name = raw_names.name;
