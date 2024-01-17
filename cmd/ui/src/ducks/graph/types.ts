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

export enum GraphNodeTypes {
    AZBase = 'AZBase',
    AZApp = 'AZApp',
    AZVMScaleSet = 'AZVMScaleSet',
    AZRole = 'AZRole',
    AZDevice = 'AZDevice',
    AZFunctionApp = 'AZFunctionApp',
    AZGroup = 'AZGroup',
    AZKeyVault = 'AZKeyVault',
    AZManagementGroup = 'AZManagementGroup',
    AZResourceGroup = 'AZResourceGroup',
    AZServicePrincipal = 'AZServicePrincipal',
    AZSubscription = 'AZSubscription',
    AZTenant = 'AZTenant',
    AZUser = 'AZUser',
    AZVM = 'AZVM',
    AZManagedCluster = 'AZManagedCluster',
    AZContainerRegistry = 'AZContainerRegistry',
    AZWebApp = 'AZWebApp',
    AZLogicApp = 'AZLogicApp',
    AZAutomationAccount = 'AZAutomationAccount',
    Base = 'Base',
    User = 'User',
    Group = 'Group',
    Computer = 'Computer',
    GPO = 'GPO',
    OU = 'OU',
    Domain = 'Domain',
    Container = 'Container',
    AIACA = 'AIACA',
    RootCA = 'RootCA',
    EnterpriseCA = 'EnterpriseCA',
    NTAuthStore = 'NTAuthStore',
    CertTemplate = 'CertTemplate',
}

export interface GraphNodeData {
    count: number;
    date: object;
    nodetype: string;
    objectid: string;
    type: string;
    name: string;
}

export interface GraphLinkData {
    source: string;
    target: string;
}

export type GraphItemData = GraphNodeData & GraphLinkData;
