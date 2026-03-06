// Copyright 2023 Specter Ops, Inc.
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

import { RequestOptions } from 'js-client-library';
import { EntityTables } from '../components';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../graphSchema';
import { apiClient } from './api';

type EntitySectionEndpointParams = {
    id: string;
    counts?: boolean;
    skip?: number;
    limit?: number;
    type?: 'graph';
};

export type EntityRelationshipQueryTypes = keyof typeof entityRelationshipEndpoints;

export interface EntityInfoDataTableProps {
    id: string;
    label: string;
    countLabel?: string;
    sections?: EntityInfoDataTableProps[];
    parentLabels?: string[];
    queryType?: EntityRelationshipQueryTypes;
}

export interface EntityInfoContentProps {
    DataTable: React.FC<EntityInfoDataTableProps>;
    id: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
    priorityTables?: EntityTables;
    additionalTables?: EntityTables;
}

let controller = new AbortController();

export const abortEntitySectionRequest = () => {
    controller.abort();
    controller = new AbortController();
};

export const MetaNodeKind = 'Meta' as const;
export const MetaDetailNodeKind = 'MetaDetail' as const;
export type EntityKinds = ActiveDirectoryNodeKind | AzureNodeKind;

export const entityInformationEndpoints: Record<EntityKinds, (id: string, options?: RequestOptions) => Promise<any>> = {
    [AzureNodeKind.Entity]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('az-base', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.App]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('applications', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.VMScaleSet]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('vm-scale-sets', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.Device]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('devices', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.FunctionApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('function-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.Group]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('groups', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.KeyVault]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('key-vaults', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.ManagementGroup]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2(
            'management-groups',
            id,
            undefined,
            false,
            undefined,
            undefined,
            undefined,
            options
        ),
    [AzureNodeKind.ResourceGroup]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('resource-groups', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.Role]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('roles', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.ServicePrincipal]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2(
            'service-principals',
            id,
            undefined,
            false,
            undefined,
            undefined,
            undefined,
            options
        ),
    [AzureNodeKind.Subscription]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('subscriptions', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.Tenant]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('tenants', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.User]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('users', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.VM]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('vms', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.ManagedCluster]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('managed-clusters', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.ContainerRegistry]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2(
            'container-registries',
            id,
            undefined,
            false,
            undefined,
            undefined,
            undefined,
            options
        ),
    [AzureNodeKind.WebApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('web-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.LogicApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('logic-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureNodeKind.AutomationAccount]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2(
            'automation-accounts',
            id,
            undefined,
            false,
            undefined,
            undefined,
            undefined,
            options
        ),
    [ActiveDirectoryNodeKind.Entity]: (id: string, options?: RequestOptions) => apiClient.getBaseV2(id, false, options),
    // LocalGroups and LocalUsers are entities that we handle directly and add the `Base` kind to so using getBaseV2 is an assumption but should work
    [ActiveDirectoryNodeKind.LocalGroup]: (id: string, options?: RequestOptions) =>
        apiClient.getBaseV2(id, false, options),
    [ActiveDirectoryNodeKind.LocalUser]: (id: string, options?: RequestOptions) =>
        apiClient.getBaseV2(id, false, options),
    [ActiveDirectoryNodeKind.AIACA]: (id: string, options?: RequestOptions) => apiClient.getAIACAV2(id, false, options),
    [ActiveDirectoryNodeKind.CertTemplate]: (id: string, options?: RequestOptions) =>
        apiClient.getCertTemplateV2(id, false, options),
    [ActiveDirectoryNodeKind.Computer]: (id: string, options?: RequestOptions) =>
        apiClient.getComputerV2(id, false, options),
    [ActiveDirectoryNodeKind.Container]: (id: string, options?: RequestOptions) =>
        apiClient.getContainerV2(id, false, options),
    [ActiveDirectoryNodeKind.Domain]: (id: string, options?: RequestOptions) =>
        apiClient.getDomainV2(id, false, options),
    [ActiveDirectoryNodeKind.EnterpriseCA]: (id: string, options?: RequestOptions) =>
        apiClient.getEnterpriseCAV2(id, false, options),
    [ActiveDirectoryNodeKind.GPO]: (id: string, options?: RequestOptions) => apiClient.getGPOV2(id, false, options),
    [ActiveDirectoryNodeKind.Group]: (id: string, options?: RequestOptions) => apiClient.getGroupV2(id, false, options),
    [ActiveDirectoryNodeKind.NTAuthStore]: (id: string, options?: RequestOptions) =>
        apiClient.getNTAuthStoreV2(id, false, options),
    [ActiveDirectoryNodeKind.OU]: (id: string, options?: RequestOptions) => apiClient.getOUV2(id, false, options),
    [ActiveDirectoryNodeKind.RootCA]: (id: string, options?: RequestOptions) =>
        apiClient.getRootCAV2(id, false, options),
    [ActiveDirectoryNodeKind.User]: (id: string, options?: RequestOptions) => apiClient.getUserV2(id, false, options),
    [ActiveDirectoryNodeKind.IssuancePolicy]: (id: string, options?: RequestOptions) =>
        apiClient.getIssuancePolicyV2(id, false, options),
};

export const allSections: Partial<Record<EntityKinds, (id: string) => EntityInfoDataTableProps[]>> = {
    [AzureNodeKind.Entity]: (id: string) => [
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'azbase-outbound_object_control',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azbase-inbound_object_control',
        },
    ],
    [AzureNodeKind.App]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azapp-inbound_object_control',
        },
    ],
    [AzureNodeKind.VMScaleSet]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azvmscaleset-inbound_object_control',
        },
    ],
    [AzureNodeKind.Device]: (id: string) => [
        {
            id,
            label: 'Local Admins',
            queryType: 'azdevice-local_admins',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azdevice-inbound_object_control',
        },
    ],
    [AzureNodeKind.FunctionApp]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azfunctionapp-inbound_object_control',
        },
    ],
    [AzureNodeKind.Group]: (id: string) => [
        {
            id,
            label: 'Members',
            queryType: 'azgroup-members',
        },
        {
            id,
            label: 'Member Of',
            queryType: 'azgroup-member_of',
        },
        {
            id,
            label: 'Roles',
            queryType: 'azgroup-roles',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azgroup-inbound_object_control',
        },
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'azgroup-outbound_object_control',
        },
    ],
    [AzureNodeKind.KeyVault]: (id: string) => [
        {
            id,
            label: 'Vault Readers',
            countLabel: 'All Readers',
            sections: [
                {
                    id,
                    label: 'Key Readers',
                    queryType: 'azkeyvault-key_readers',
                },
                {
                    id,
                    label: 'Certificate Readers',
                    queryType: 'azkeyvault-certificate_readers',
                },
                {
                    id,
                    label: 'Secret Readers',
                    queryType: 'azkeyvault-secret_readers',
                },
                {
                    id,
                    label: 'All Readers',
                    queryType: 'azkeyvault-all_readers',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azkeyvault-inbound_object_control',
        },
    ],
    [AzureNodeKind.ManagementGroup]: (id: string) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant Management Groups',
                    queryType: 'azmanagementgroup-descendant_management_groups',
                },
                {
                    id,
                    label: 'Descendant Subscriptions',
                    queryType: 'azmanagementgroup-descendant_subscriptions',
                },
                {
                    id,
                    label: 'Descendant Resource Groups',
                    queryType: 'azmanagementgroup-descendant_resource_groups',
                },
                {
                    id,
                    label: 'Descendant VMs',
                    queryType: 'azmanagementgroup-descendant_vms',
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    queryType: 'azmanagementgroup-descendant_managed_clusters',
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    queryType: 'azmanagementgroup-descendant_vm_scale_sets',
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    queryType: 'azmanagementgroup-descendant_container_registries',
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    queryType: 'azmanagementgroup-descendant_web_apps',
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    queryType: 'azmanagementgroup-descendant_automation_accounts',
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    queryType: 'azmanagementgroup-descendant_key_vaults',
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    queryType: 'azmanagementgroup-descendant_function_apps',
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    queryType: 'azmanagementgroup-descendant_logic_apps',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azmanagementgroup-inbound_object_control',
        },
    ],
    [AzureNodeKind.ResourceGroup]: (id: string) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant VMs',
                    queryType: 'azresourcegroup-descendant_vms',
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    queryType: 'azresourcegroup-descendant_managed_clusters',
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    queryType: 'azresourcegroup-descendant_vm_scale_sets',
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    queryType: 'azresourcegroup-descendant_container_registries',
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    queryType: 'azresourcegroup-descendant_automation_accounts',
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    queryType: 'azresourcegroup-descendant_key_vaults',
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    queryType: 'azresourcegroup-descendant_web_apps',
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    queryType: 'azresourcegroup-descendant_function_apps',
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    queryType: 'azresourcegroup-descendant_logic_apps',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azresourcegroup-inbound_object_control',
        },
    ],
    [AzureNodeKind.Role]: (id: string) => [
        {
            id,
            label: 'Active Assignments',
            queryType: 'azrole-active_assignments',
        },
        {
            id,
            label: 'Approvers',
            queryType: 'azrole-approvers',
        },
    ],
    [AzureNodeKind.ServicePrincipal]: (id: string) => {
        const base: EntityInfoDataTableProps[] = [
            {
                id,
                label: 'Roles',
                queryType: 'azserviceprincipal-roles',
            },
            {
                id,
                label: 'Inbound Object Control',
                queryType: 'azserviceprincipal-inbound_object_control',
            },
            {
                id,
                label: 'Outbound Object Control',
                queryType: 'azserviceprincipal-outbound_object_control',
            },
            {
                id,
                label: 'Inbound Abusable App Role Assignments',
                queryType: 'azserviceprincipal-inbound_abusable_app_role_assignments',
            },
        ];
        if (id === '00000003-0000-0000-C000-000000000000') {
            const OutboundAbusableAppRoleAssignmentsProp: EntityInfoDataTableProps = {
                id,
                label: 'Outbound Abusable App Role Assignments',
                queryType: 'azserviceprincipal-outbound_abusable_app_role_assignments',
            };

            return [...base, OutboundAbusableAppRoleAssignmentsProp];
        }

        return base;
    },
    [AzureNodeKind.Subscription]: (id: string) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant Resource Groups',
                    queryType: 'azsubscription-descendant_objects-descendant_resource_groups',
                },
                {
                    id,
                    label: 'Descendant VMs',
                    queryType: 'azsubscription-descendant_objects-descendant_vms',
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    queryType: 'azsubscription-descendant_objects-descendant_managed_clusters',
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    queryType: 'azsubscription-descendant_objects-descendant_vm_scale_sets',
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    queryType: 'azsubscription-descendant_objects-descendant_container_registries',
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    queryType: 'azsubscription-descendant_objects-descendant_automation_accounts',
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    queryType: 'azsubscription-descendant_objects-descendant_key_vaults',
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    queryType: 'azsubscription-descendant_objects-descendant_web_apps',
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    queryType: 'azsubscription-descendant_objects-descendant_function_apps',
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    queryType: 'azsubscription-descendant_objects-descendant_logic_apps',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azsubscription-inbound_object_control',
        },
    ],
    [AzureNodeKind.Tenant]: (id: string) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant Users',
                    queryType: 'aztenant-descendant_users',
                },
                {
                    id,
                    label: 'Descendant Groups',
                    queryType: 'aztenant-descendant_groups',
                },
                {
                    id,
                    label: 'Descendant Management Groups',
                    queryType: 'aztenant-descendant_management_groups',
                },
                {
                    id,
                    label: 'Descendant Subscriptions',
                    queryType: 'aztenant-descendant_subscriptions',
                },
                {
                    id,
                    label: 'Descendant Resource Groups',
                    queryType: 'aztenant-descendant_resource_groups',
                },
                {
                    id,
                    label: 'Descendant VMs',
                    queryType: 'aztenant-descendant_vms',
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    queryType: 'aztenant-descendant_managed_clusters',
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    queryType: 'aztenant-descendant_vm_scale_sets',
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    queryType: 'aztenant-descendant_container_registries',
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    queryType: 'aztenant-descendant_web_apps',
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    queryType: 'aztenant-descendant_automation_accounts',
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    queryType: 'aztenant-descendant_key_vaults',
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    queryType: 'aztenant-descendant_function_apps',
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    queryType: 'aztenant-descendant_logic_apps',
                },
                {
                    id,
                    label: 'Descendant App Registrations',
                    queryType: 'aztenant-descendant_app_registrations',
                },
                {
                    id,
                    label: 'Descendant Service Principals',
                    queryType: 'aztenant-descendant_service_principals',
                },
                {
                    id,
                    label: 'Descendant Devices',
                    queryType: 'aztenant-descendant_devices',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'aztenant-inbound_object_control',
        },
    ],
    [AzureNodeKind.User]: (id: string) => [
        {
            id,
            label: 'Member Of',
            queryType: 'azuser-member_of',
        },
        {
            id,
            label: 'Roles',
            queryType: 'azuser-roles',
        },
        {
            id,
            label: 'Execution Privileges',
            queryType: 'azuser-execution_privileges',
        },
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'azuser-outbound_object_control',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azuser-inbound_object_control',
        },
    ],
    [AzureNodeKind.VM]: (id: string) => [
        {
            id,
            label: 'Local Admins',
            queryType: 'azvm-local_admins',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azvm-inbound_object_control',
        },
    ],
    [AzureNodeKind.ManagedCluster]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azmanagedcluster-inbound_object_control',
        },
    ],
    [AzureNodeKind.ContainerRegistry]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azcontainerregistry-inbound_object_control',
        },
    ],
    [AzureNodeKind.WebApp]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azwebapp-inbound_object_control',
        },
    ],
    [AzureNodeKind.LogicApp]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azlogicapp-inbound_object_control',
        },
    ],
    [AzureNodeKind.AutomationAccount]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'azautomationaccount-inbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.Entity]: (id: string) => [
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'base-outbound_object_control',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'base-inbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.Container]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'container-inbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.AIACA]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'aiaca-inbound_object_control',
        },
        {
            id,
            label: 'PKI Hierarchy',
            queryType: 'aiaca-pki_hierarchy',
        },
    ],
    [ActiveDirectoryNodeKind.CertTemplate]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'certtemplate-inbound_object_control',
        },
        {
            id,
            label: 'Published To CAs',
            queryType: 'certtemplate-published_to_cas',
        },
    ],
    [ActiveDirectoryNodeKind.Computer]: (id: string) => [
        {
            id,
            label: 'Sessions',
            queryType: 'computer-sessions',
        },
        {
            id,
            label: 'Local Admins',
            queryType: 'computer-local_admins',
        },
        {
            id,
            label: 'Inbound Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Users',
                    queryType: 'computer-rdp_users',
                },
                {
                    id,
                    label: 'PSRemote Users',
                    queryType: 'computer-psremote_users',
                },
                {
                    id,
                    label: 'DCOM Users',
                    queryType: 'computer-dcom_users',
                },
                {
                    id,
                    label: 'SQL Admin Users',
                    queryType: 'computer-sql_admin_users',
                },
                {
                    id,
                    label: 'Constrained Delegation Users',
                    queryType: 'computer-constrained_delegation_users',
                },
            ],
        },
        {
            id,
            label: 'Member Of',
            queryType: 'computer-member_of',
        },
        {
            id,
            label: 'Local Admin Privileges',
            queryType: 'computer-local_admin_privileges',
        },
        {
            id,
            label: 'Outbound Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    queryType: 'computer-rdp_privileges',
                },
                {
                    id,
                    label: 'PSRemote Rights',
                    queryType: 'computer-psremote_rights',
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    queryType: 'computer-dcom_privileges',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'computer-inbound_object_control',
        },
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'computer-outbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.Domain]: (id: string) => [
        {
            id,
            label: 'Foreign Members',
            sections: [
                {
                    id,
                    label: 'Foreign Users',
                    queryType: 'domain-foreign_users',
                },
                {
                    id,
                    label: 'Foreign Groups',
                    queryType: 'domain-foreign_groups',
                },
                {
                    id,
                    label: 'Foreign Admins',
                    queryType: 'domain-foreign_admins',
                },
                {
                    id,
                    label: 'Foreign GPO Controllers',
                    queryType: 'domain-foreign_gpo_controllers',
                },
            ],
        },
        {
            id,
            label: 'Inbound Trusts',
            queryType: 'domain-inbound_trusts',
        },
        {
            id,
            label: 'Outbound Trusts',
            queryType: 'domain-outbound_trusts',
        },
        {
            id,
            label: 'Controllers',
            queryType: 'domain-controllers',
        },
        {
            id,
            label: 'ADCS Escalations',
            queryType: 'domain-adcs_escalations',
        },
    ],
    [ActiveDirectoryNodeKind.EnterpriseCA]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'enterpriseca-inbound_object_control',
        },
        {
            id,
            label: 'PKI Hierarchy',
            queryType: 'enterpriseca-pki_hierarchy',
        },
        {
            id,
            label: 'Published Templates',
            queryType: 'enterpriseca-published_templates',
        },
    ],
    [ActiveDirectoryNodeKind.GPO]: (id: string) => [
        {
            id,
            label: 'Affected Objects',
            sections: [
                {
                    id,
                    label: 'OUs',
                    queryType: 'gpo-ous',
                },
                {
                    id,
                    label: 'Computers',
                    queryType: 'gpo-computers',
                },
                {
                    id,
                    label: 'Users',
                    queryType: 'gpo-users',
                },
                {
                    id,
                    label: 'Tier Zero Objects',
                    queryType: 'gpo-tier_zero_objects',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'gpo-inbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.Group]: (id: string) => [
        {
            id,
            label: 'Sessions',
            queryType: 'group-sessions',
        },
        {
            id,
            label: 'Members',
            queryType: 'group-members',
        },
        {
            id,
            label: 'Member Of',
            queryType: 'group-member_of',
        },
        {
            id,
            label: 'Local Admin Privileges',
            queryType: 'group-local_admin_privileges',
        },
        {
            id,
            label: 'Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    queryType: 'group-rdp_privileges',
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    queryType: 'group-dcom_privileges',
                },
                {
                    id,
                    label: 'PSRemote Rights',
                    queryType: 'group-psremote_rights',
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'group-inbound_object_control',
        },
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'group-outbound_object_control',
        },
    ],
    [ActiveDirectoryNodeKind.NTAuthStore]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'ntauthstore-inbound_object_control',
        },
        {
            id,
            label: 'Trusted CAs',
            queryType: 'ntauthstore-trusted_cas',
        },
    ],
    [ActiveDirectoryNodeKind.OU]: (id: string) => [
        {
            id,
            label: 'Affecting GPOs',
            queryType: 'ou-affecting_gpos',
        },
        {
            id,
            label: 'Groups',
            queryType: 'ou-groups',
        },
        {
            id,
            label: 'Computers',
            queryType: 'ou-computers',
        },
        {
            id,
            label: 'Users',
            queryType: 'ou-users',
        },
    ],
    [ActiveDirectoryNodeKind.RootCA]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'rootca-inbound_object_control',
        },
        {
            id,
            label: 'PKI Hierarchy',
            queryType: 'rootca-pki_hierarchy',
        },
    ],
    [ActiveDirectoryNodeKind.IssuancePolicy]: (id: string) => [
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'issuancepolicy-inbound_object_control',
        },
        {
            id,
            label: 'Linked Certificate Templates',
            queryType: 'issuancepolicy-linked_certificate_templates',
        },
    ],
    [ActiveDirectoryNodeKind.User]: (id: string) => [
        {
            id,
            label: 'Sessions',
            queryType: 'user-sessions',
        },
        {
            id,
            label: 'Member Of',
            queryType: 'user-member_of',
        },
        {
            id,
            label: 'Local Admin Privileges',
            queryType: 'user-local_admin_privileges',
        },
        {
            id,
            label: 'Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    queryType: 'user-rdp_privileges',
                },
                {
                    id,
                    label: 'PSRemote Privileges',
                    queryType: 'user-psremote_privileges',
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    queryType: 'user-dcom_privileges',
                },
                {
                    id,
                    label: 'SQL Admin Rights',
                    queryType: 'user-sql_admin_rights',
                },
                {
                    id,
                    label: 'Constrained Delegation Privileges',
                    queryType: 'user-constrained_delegation_privileges',
                },
            ],
        },
        {
            id,
            label: 'Outbound Object Control',
            queryType: 'user-outbound_object_control',
        },
        {
            id,
            label: 'Inbound Object Control',
            queryType: 'user-inbound_object_control',
        },
    ],
};

export type EntityRelationshipEndpoint = Record<string, (params: EntitySectionEndpointParams) => Promise<any>>;

export const entityRelationshipEndpoints = {
    'azbase-outbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('az-base', id, 'outbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azbase-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('az-base', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azapp-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('applications', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azvmscaleset-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('vm-scale-sets', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azdevice-local_admins': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('devices', id, 'inbound-execution-privileges', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azdevice-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('devices', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azfunctionapp-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('function-apps', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azgroup-members': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('groups', id, 'group-members', counts, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'azgroup-member_of': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('groups', id, 'group-membership', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azgroup-roles': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('groups', id, 'roles', counts, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'azgroup-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('groups', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azgroup-outbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('groups', id, 'outbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azkeyvault-key_readers': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('key-vaults', id, 'key-readers', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => {
                if (type !== 'graph') res.data.countLabel = 'Key Readers';
                return res.data;
            }),
    'azkeyvault-certificate_readers': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('key-vaults', id, 'certificate-readers', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => {
                if (type !== 'graph') res.data.countLabel = 'Certificate Readers';
                return res.data;
            }),
    'azkeyvault-secret_readers': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('key-vaults', id, 'secret-readers', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => {
                if (type !== 'graph') res.data.countLabel = 'Secret Readers';
                return res.data;
            }),
    'azkeyvault-all_readers': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('key-vaults', id, 'all-readers', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => {
                if (type !== 'graph') res.data.countLabel = 'All Readers';
                return res.data;
            }),
    'azkeyvault-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('key-vaults', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_management_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-management-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_subscriptions': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-subscriptions', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_resource_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-resource-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_vms': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-virtual-machines', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_managed_clusters': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-managed-clusters', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_vm_scale_sets': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-vm-scale-sets', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_container_registries': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-container-registries', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_web_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-web-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_automation_accounts': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-automation-accounts', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_key_vaults': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-key-vaults', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_function_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-function-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-descendant_logic_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'descendent-logic-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azmanagementgroup-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('management-groups', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_vms': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-virtual-machines', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_managed_clusters': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-managed-clusters', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_vm_scale_sets': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-vm-scale-sets', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_container_registries': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-container-registries', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_automation_accounts': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-automation-accounts', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_key_vaults': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-key-vaults', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_web_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-web-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_function_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-function-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-descendant_logic_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'descendent-logic-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azresourcegroup-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('resource-groups', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azrole-active_assignments': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('roles', id, 'active-assignments', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azrole-approvers': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('roles', id, 'role-approvers', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azserviceprincipal-roles': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('service-principals', id, 'roles', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azserviceprincipal-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('service-principals', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azserviceprincipal-outbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('service-principals', id, 'outbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azserviceprincipal-inbound_abusable_app_role_assignments': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2(
                'service-principals',
                id,
                'inbound-abusable-app-role-assignments',
                counts,
                skip,
                limit,
                type,
                {
                    signal: controller.signal,
                }
            )
            .then((res) => res.data),
    'azserviceprincipal-outbound_abusable_app_role_assignments': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2(
                'service-principals',
                id,
                'outbound-abusable-app-role-assignments',
                counts,
                skip,
                limit,
                type,
                {
                    signal: controller.signal,
                }
            )
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_resource_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-resource-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_vms': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-virtual-machines', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_managed_clusters': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-managed-clusters', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_vm_scale_sets': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-vm-scale-sets', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_container_registries': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-container-registries', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_automation_accounts': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-automation-accounts', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_key_vaults': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-key-vaults', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_web_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-web-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_function_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-function-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-descendant_objects-descendant_logic_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'descendent-logic-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azsubscription-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('subscriptions', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_users': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-users', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_management_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-management-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_subscriptions': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-subscriptions', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_resource_groups': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-resource-groups', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_vms': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-virtual-machines', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_managed_clusters': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-managed-clusters', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_vm_scale_sets': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-vm-scale-sets', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_container_registries': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-container-registries', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_web_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-web-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_automation_accounts': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-automation-accounts', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_key_vaults': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-key-vaults', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_function_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-function-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_logic_apps': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-logic-apps', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_app_registrations': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-applications', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_service_principals': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-service-principals', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-descendant_devices': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'descendent-devices', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'aztenant-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('tenants', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azuser-member_of': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('users', id, 'group-membership', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azuser-roles': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('users', id, 'roles', counts, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'azuser-execution_privileges': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('users', id, 'outbound-execution-privileges', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azuser-outbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('users', id, 'outbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azuser-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('users', id, 'inbound-control', counts, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'azvm-local_admins': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('vms', id, 'inbound-execution-privileges', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azvm-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('vms', id, 'inbound-control', counts, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'azmanagedcluster-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('managed-clusters', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azcontainerregistry-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('container-registries', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azwebapp-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('web-apps', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azlogicapp-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('logic-apps', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'azautomationaccount-inbound_object_control': ({ id, counts, skip, limit, type }) =>
        apiClient
            .getAZEntityInfoV2('automation-accounts', id, 'inbound-control', counts, skip, limit, type, {
                signal: controller.signal,
            })
            .then((res) => res.data),
    'base-outbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getBaseControllablesV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'base-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getBaseControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'container-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getContainerControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'aiaca-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getAIACAControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'aiaca-pki_hierarchy': ({ id, skip, limit, type }) =>
        apiClient.getAIACAPKIHierarchyV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'certtemplate-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getCertTemplateControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'certtemplate-published_to_cas': ({ id, skip, limit, type }) =>
        apiClient
            .getCertTemplatePublishedToCAsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-sessions': ({ id, skip, limit, type }) =>
        apiClient.getComputerSessionsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-local_admins': ({ id, skip, limit, type }) =>
        apiClient.getComputerAdminUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-rdp_users': ({ id, skip, limit, type }) =>
        apiClient.getComputerRDPUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-psremote_users': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerPSRemoteUsersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-dcom_users': ({ id, skip, limit, type }) =>
        apiClient.getComputerDCOMUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-sql_admin_users': ({ id, skip, limit, type }) =>
        apiClient.getComputerSQLAdminsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-constrained_delegation_users': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerConstrainedDelegationRightsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-member_of': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerGroupMembershipV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-local_admin_privileges': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerAdminRightsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-rdp_privileges': ({ id, skip, limit, type }) =>
        apiClient.getComputerRDPRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-psremote_rights': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-dcom_privileges': ({ id, skip, limit, type }) =>
        apiClient.getComputerDCOMRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'computer-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'computer-outbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getComputerControllablesV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-foreign_users': ({ id, skip, limit, type }) =>
        apiClient.getDomainForeignUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'domain-foreign_groups': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainForeignGroupsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-foreign_admins': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainForeignAdminsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-foreign_gpo_controllers': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainForeignGPOControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-inbound_trusts': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainInboundTrustsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-outbound_trusts': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainOutboundTrustsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'domain-controllers': ({ id, skip, limit, type }) =>
        apiClient.getDomainControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'domain-adcs_escalations': ({ id, skip, limit, type }) =>
        apiClient
            .getDomainADCSEscalationsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'enterpriseca-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getEnterpriseCAControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'enterpriseca-pki_hierarchy': ({ id, skip, limit, type }) =>
        apiClient
            .getEnterpriseCAPKIHierarchyV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'enterpriseca-published_templates': ({ id, skip, limit, type }) =>
        apiClient
            .getEnterpriseCAPublishedTemplatesV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'gpo-ous': ({ id, skip, limit, type }) =>
        apiClient.getGPOOUsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'gpo-computers': ({ id, skip, limit, type }) =>
        apiClient.getGPOComputersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'gpo-users': ({ id, skip, limit, type }) =>
        apiClient.getGPOUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'gpo-tier_zero_objects': ({ id, skip, limit, type }) =>
        apiClient.getGPOTierZeroV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'gpo-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getGPOControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-sessions': ({ id, skip, limit, type }) =>
        apiClient.getGroupSessionsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-members': ({ id, skip, limit, type }) =>
        apiClient.getGroupMembersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-member_of': ({ id, skip, limit, type }) =>
        apiClient.getGroupMembershipsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-local_admin_privileges': ({ id, skip, limit, type }) =>
        apiClient.getGroupAdminRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-rdp_privileges': ({ id, skip, limit, type }) =>
        apiClient.getGroupRDPRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-dcom_privileges': ({ id, skip, limit, type }) =>
        apiClient.getGroupDCOMRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-psremote_rights': ({ id, skip, limit, type }) =>
        apiClient
            .getGroupPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'group-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getGroupControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'group-outbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getGroupControllablesV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'ntauthstore-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getNTAuthStoreControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'ntauthstore-trusted_cas': ({ id, skip, limit, type }) =>
        apiClient
            .getNTAuthStoreTrustedCAsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'ou-affecting_gpos': ({ id, skip, limit, type }) =>
        apiClient.getOUGPOsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'ou-groups': ({ id, skip, limit, type }) =>
        apiClient.getOUGroupsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'ou-computers': ({ id, skip, limit, type }) =>
        apiClient.getOUComputersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'ou-users': ({ id, skip, limit, type }) =>
        apiClient.getOUUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'rootca-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getRootCAControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'rootca-pki_hierarchy': ({ id, skip, limit, type }) =>
        apiClient.getRootCAPKIHierarchyV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'issuancepolicy-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient
            .getIssuancePolicyControllersV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'issuancepolicy-linked_certificate_templates': ({ id, skip, limit, type }) =>
        apiClient
            .getIssuancePolicyLinkedTemplatesV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'user-sessions': ({ id, skip, limit, type }) =>
        apiClient.getUserSessionsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-member_of': ({ id, skip, limit, type }) =>
        apiClient.getUserMembershipsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-local_admin_privileges': ({ id, skip, limit, type }) =>
        apiClient.getUserAdminRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-rdp_privileges': ({ id, skip, limit, type }) =>
        apiClient.getUserRDPRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-psremote_privileges': ({ id, skip, limit, type }) =>
        apiClient.getUserPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-dcom_privileges': ({ id, skip, limit, type }) =>
        apiClient.getUserDCOMRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-sql_admin_rights': ({ id, skip, limit, type }) =>
        apiClient.getUserSQLAdminRightsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-constrained_delegation_privileges': ({ id, skip, limit, type }) =>
        apiClient
            .getUserConstrainedDelegationRightsV2(id, skip, limit, type, { signal: controller.signal })
            .then((res) => res.data),
    'user-outbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getUserControllablesV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
    'user-inbound_object_control': ({ id, skip, limit, type }) =>
        apiClient.getUserControllersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
} as const satisfies EntityRelationshipEndpoint;
