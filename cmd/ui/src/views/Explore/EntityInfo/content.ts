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
import apiClient from 'src/api';
import { GraphNodeTypes } from 'src/ducks/graph/types';
import { ActiveDirectoryKind, AzureKind } from 'bh-shared-ui';
import { EntityInfoDataTableProps } from './EntityInfoDataTable';
import { controller } from 'src/views/Explore/utils';

export const entityInformationEndpoints: Record<
    GraphNodeTypes,
    (id: string, options?: RequestOptions) => Promise<any>
> = {
    [AzureKind.Entity]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('az-base', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.App]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('applications', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.VMScaleSet]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('vm-scale-sets', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.Device]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('devices', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.FunctionApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('function-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.Group]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('groups', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.KeyVault]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('key-vaults', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.ManagementGroup]: (id: string, options?: RequestOptions) =>
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
    [AzureKind.ResourceGroup]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('resource-groups', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.Role]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('roles', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.ServicePrincipal]: (id: string, options?: RequestOptions) =>
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
    [AzureKind.Subscription]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('subscriptions', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.Tenant]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('tenants', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.User]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('users', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.VM]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('vms', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.ManagedCluster]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('managed-clusters', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.ContainerRegistry]: (id: string, options?: RequestOptions) =>
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
    [AzureKind.WebApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('web-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.LogicApp]: (id: string, options?: RequestOptions) =>
        apiClient.getAZEntityInfoV2('logic-apps', id, undefined, false, undefined, undefined, undefined, options),
    [AzureKind.AutomationAccount]: (id: string, options?: RequestOptions) =>
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
    [ActiveDirectoryKind.Entity]: (id: string, options?: RequestOptions) => apiClient.getBaseV2(id, false, options),
    [ActiveDirectoryKind.Computer]: (id: string, options?: RequestOptions) =>
        apiClient.getComputerV2(id, false, options),
    [ActiveDirectoryKind.Container]: () => Promise.resolve(),
    [ActiveDirectoryKind.Domain]: (id: string, options?: RequestOptions) => apiClient.getDomainV2(id, false, options),
    [ActiveDirectoryKind.GPO]: (id: string, options?: RequestOptions) => apiClient.getGPOV2(id, false, options),
    [ActiveDirectoryKind.Group]: (id: string, options?: RequestOptions) => apiClient.getGroupV2(id, false, options),
    [ActiveDirectoryKind.OU]: (id: string, options?: RequestOptions) => apiClient.getOUV2(id, false, options),
    [ActiveDirectoryKind.User]: (id: string, options?: RequestOptions) => apiClient.getUserV2(id, false, options),
};

export const allSections: Record<GraphNodeTypes, (id: string) => EntityInfoDataTableProps[]> = {
    [AzureKind.Entity]: (id) => [
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('az-base', id, 'outbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('az-base', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.App]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('applications', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.VMScaleSet]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('vm-scale-sets', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.Device]: (id) => [
        {
            id,
            label: 'Local Admins',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('devices', id, 'inbound-execution-privileges', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('devices', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.FunctionApp]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('function-apps', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.Group]: (id) => [
        {
            id,
            label: 'Members',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('groups', id, 'group-members', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Member Of',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('groups', id, 'group-membership', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Roles',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('groups', id, 'roles', counts, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('groups', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('groups', id, 'outbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.KeyVault]: (id) => [
        {
            id,
            label: 'Vault Readers',
            sections: [
                {
                    id,
                    label: 'Key Readers',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('key-vaults', id, 'key-readers', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Certificate Readers',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('key-vaults', id, 'certificate-readers', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Secret Readers',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('key-vaults', id, 'secret-readers', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'All Readers',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('key-vaults', id, 'all-readers', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('key-vaults', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.ManagementGroup]: (id) => [
        {
            id,
            label: 'Descendant Objects',

            sections: [
                {
                    id,
                    label: 'Descendant Management Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-management-groups',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Subscriptions',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-subscriptions',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Resource Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-resource-groups',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VMs',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-virtual-machines',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-managed-clusters',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-vm-scale-sets',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-container-registries',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-web-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-automation-accounts',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-key-vaults',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-function-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'management-groups',
                                id,
                                'descendent-logic-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('management-groups', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.ResourceGroup]: (id) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant VMs',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-virtual-machines',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-managed-clusters',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-vm-scale-sets',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-container-registries',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-automation-accounts',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-key-vaults',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-web-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-function-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'resource-groups',
                                id,
                                'descendent-logic-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('resource-groups', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.Role]: (id) => [
        {
            id,
            label: 'Active Assignments',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('roles', id, 'active-assignments', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.ServicePrincipal]: (id) => {
        const BaseProps: EntityInfoDataTableProps[] = [
            {
                id,
                label: 'Roles',
                endpoint: ({ counts, skip, limit, type }) =>
                    apiClient
                        .getAZEntityInfoV2('service-principals', id, 'roles', counts, skip, limit, type, {
                            signal: controller.signal,
                        })
                        .then((res) => res.data),
            },
            {
                id,
                label: 'Inbound Object Control',
                endpoint: ({ counts, skip, limit, type }) =>
                    apiClient
                        .getAZEntityInfoV2('service-principals', id, 'inbound-control', counts, skip, limit, type, {
                            signal: controller.signal,
                        })
                        .then((res) => res.data),
            },
            {
                id,
                label: 'Outbound Object Control',
                endpoint: ({ counts, skip, limit, type }) =>
                    apiClient
                        .getAZEntityInfoV2('service-principals', id, 'outbound-control', counts, skip, limit, type, {
                            signal: controller.signal,
                        })
                        .then((res) => res.data),
            },
            {
                id,
                label: 'Inbound Abusable App Role Assignments',
                endpoint: ({ counts, skip, limit, type }) =>
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
            },
        ];

        const OutboundAbusableAppRoleAssignmentsProp: EntityInfoDataTableProps = {
            id,
            label: 'Outbound Abusable App Role Assignments',
            endpoint: ({ counts, skip, limit, type }) =>
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
        };

        return id === '00000003-0000-0000-C000-000000000000'
            ? BaseProps
            : [...BaseProps, OutboundAbusableAppRoleAssignmentsProp];
    },
    [AzureKind.Subscription]: (id) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant Resource Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-resource-groups',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VMs',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-virtual-machines',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-managed-clusters',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-vm-scale-sets',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-container-registries',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-automation-accounts',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-key-vaults',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('subscriptions', id, 'descendent-web-apps', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-function-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'subscriptions',
                                id,
                                'descendent-logic-apps',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('subscriptions', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.Tenant]: (id) => [
        {
            id,
            label: 'Descendant Objects',
            sections: [
                {
                    id,
                    label: 'Descendant Users',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-users', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-groups', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Management Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-management-groups',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Subscriptions',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-subscriptions', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Resource Groups',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-resource-groups', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VMs',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-virtual-machines',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Managed Clusters',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-managed-clusters',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant VM Scale Sets',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-vm-scale-sets', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Container Registries',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-container-registries',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Web Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-web-apps', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Automation Accounts',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-automation-accounts',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Key Vaults',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-key-vaults', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Function Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-function-apps', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Logic Apps',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-logic-apps', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant App Registrations',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-applications', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Service Principals',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2(
                                'tenants',
                                id,
                                'descendent-service-principals',
                                counts,
                                skip,
                                limit,
                                type,
                                {
                                    signal: controller.signal,
                                }
                            )
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Descendant Devices',
                    endpoint: ({ counts, skip, limit, type }) =>
                        apiClient
                            .getAZEntityInfoV2('tenants', id, 'descendent-devices', counts, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('tenants', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.User]: (id) => [
        {
            id,
            label: 'Member Of',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('users', id, 'group-membership', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Roles',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('users', id, 'roles', counts, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Execution Privileges',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('users', id, 'outbound-execution-privileges', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('users', id, 'outbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('users', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.VM]: (id) => [
        {
            id,
            label: 'Local Admins',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('vms', id, 'inbound-execution-privileges', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('vms', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.ManagedCluster]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('managed-clusters', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.ContainerRegistry]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('container-registries', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.WebApp]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('web-apps', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.LogicApp]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('logic-apps', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [AzureKind.AutomationAccount]: (id) => [
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ counts, skip, limit, type }) =>
                apiClient
                    .getAZEntityInfoV2('automation-accounts', id, 'inbound-control', counts, skip, limit, type, {
                        signal: controller.signal,
                    })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.Entity]: (id) => [
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getBaseControllablesV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getBaseControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.Computer]: (id) => [
        {
            id,
            label: 'Sessions',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerSessionsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Local Admins',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerAdminUsersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerRDPUsersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'PSRemote Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerPSRemoteUsersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'DCOM Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerDCOMUsersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'SQL Admin Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerSQLAdminsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Constrained Delegation Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerConstrainedDelegationRightsV2(id, skip, limit, type, {
                                signal: controller.signal,
                            })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Member Of',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerGroupMembershipV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Local Admin Privileges',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerAdminRightsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerRDPRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'PSRemote Rights',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getComputerDCOMRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getComputerControllablesV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.Container]: () => [],
    [ActiveDirectoryKind.Domain]: (id) => [
        {
            id,
            label: 'Foreign Members',
            sections: [
                {
                    id,
                    label: 'Foreign Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getDomainForeignUsersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Foreign Groups',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getDomainForeignGroupsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Foreign Admins',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getDomainForeignAdminsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Foreign GPO Controllers',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getDomainForeignGPOControllersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Trusts',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getDomainInboundTrustsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Trusts',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getDomainOutboundTrustsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Controllers',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getDomainControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.GPO]: (id) => [
        {
            id,
            label: 'Affected Objects',
            sections: [
                {
                    id,
                    label: 'OUs',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGPOOUsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Computers',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGPOComputersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Users',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGPOUsersV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Tier Zero Objects',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGPOTierZeroV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGPOControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.Group]: (id) => [
        {
            id,
            label: 'Sessions',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupSessionsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Members',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupMembersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Member Of',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupMembershipsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Local Admin Privileges',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupAdminRightsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGroupRDPRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGroupDCOMRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'PSRemote Rights',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getGroupPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getGroupControllablesV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.OU]: (id) => [
        {
            id,
            label: 'Affecting GPOs',
            endpoint: ({ skip, limit, type }) =>
                apiClient.getOUGPOsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
        },
        {
            id,
            label: 'Groups',
            endpoint: ({ skip, limit, type }) =>
                apiClient.getOUGroupsV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
        },
        {
            id,
            label: 'Computers',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getOUComputersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Users',
            endpoint: ({ skip, limit, type }) =>
                apiClient.getOUUsersV2(id, skip, limit, type, { signal: controller.signal }).then((res) => res.data),
        },
    ],
    [ActiveDirectoryKind.User]: (id) => [
        {
            id,
            label: 'Sessions',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getUserSessionsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Member Of',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getUserMembershipsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Local Admin Privileges',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getUserAdminRightsV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Execution Privileges',
            sections: [
                {
                    id,
                    label: 'RDP Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getUserRDPRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'PSRemote Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getUserPSRemoteRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'DCOM Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getUserDCOMRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'SQL Admin Rights',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getUserSQLAdminRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
                {
                    id,
                    label: 'Constrained Delegation Privileges',
                    endpoint: ({ skip, limit, type }) =>
                        apiClient
                            .getUserConstrainedDelegationRightsV2(id, skip, limit, type, { signal: controller.signal })
                            .then((res) => res.data),
                },
            ],
        },
        {
            id,
            label: 'Outbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getUserControllablesV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
        {
            id,
            label: 'Inbound Object Control',
            endpoint: ({ skip, limit, type }) =>
                apiClient
                    .getUserControllersV2(id, skip, limit, type, { signal: controller.signal })
                    .then((res) => res.data),
        },
    ],
};
