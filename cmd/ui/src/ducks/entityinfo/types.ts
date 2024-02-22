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

import { AzureNodeKind, EntityKinds } from 'bh-shared-ui';

// --- AIACA
export interface AIACAInfo extends EntityInfo {
    props: BasicInfo & {
        certthumbprint: string;
        hascrosscertificatepair: boolean;
        description?: string;
    };
    controllables: number;
    controllers: number;
}

// --- CertTemplate
export interface CertTemplateInfo extends EntityInfo {
    props: BasicInfo & {
        applicationpolicies: string;
        authorizedsignatures: number;
        certificateapplicationpolicy: string[];
        ekus: string[];
        enrolleesuppliessubject: boolean;
        issuancepolicies: string;
        oid: string;
        renewalperiod: string;
        requiresmanagerapproval: boolean;
        schemaversion: number;
        validityperiod: string;
    };
    controllables: number;
    controllers: number;
}

// --- Computer
export interface ComputerInfoGraph extends GraphInfo {
    enabled: boolean;
    haslaps: boolean;
    lastlogontimestamp: number;
    operatingsystem: string;
    owned: boolean;
    pwdlastset: number;
    unconstraineddelegation: boolean;
}

export interface ComputerInfo extends EntityInfo {
    props: BasicInfo & {
        enabled: boolean;
        haslaps: boolean;
        lastlogontimestamp: number;
        operatingsystem: string;
        owned: boolean;
        pwdlastset: number;
        unconstraineddelegation: boolean;
        lastlogon?: number;
    };
    adminRights: number;
    adminUsers: number;
    constrainedPrivs: number;
    constrainedUsers: number;
    controllables: number;
    controllers: number;
    dcomRights: number;
    dcomUsers: number;
    groupMembership: number;
    psRemoteRights: number;
    psRemoteUsers: number;
    rdpRights: number;
    rdpUsers: number;
    sessions: number;
    sqlAdminUsers: number;
}

// --- Containers
export interface ContainersInfo extends EntityInfo {
    props: BasicInfo & {
        description?: string;
    };
    controllers: number;
}

// --- Domain
export interface DomainInfoGraph extends GraphInfo {
    functionallevel: string;
}

export interface DomainInfo extends EntityInfo {
    props: BasicInfo & {
        functionallevel: number;
        description?: string;
        'ms-ds-machineaccountquota'?: number;
    };
    foreignUsers: number;
    foreignGroups: number;
    foreignAdmins: number;
    foreignGPOControllers: number;

    inboundTrusts: number;
    outboundTrusts: number;
    controllers: number;

    dcsyncers: number;
    groups: number;
    computers: number;
    ous: number;
    gpos: number;
    users: number;
    containers: number;
}

// --- EnterpriseCA
export interface EnterpriseCAInfo extends EntityInfo {
    props: BasicInfo & {
        basicconstraintpathlength: number;
        caname: string;
        casecuritycollected: boolean;
        certchain: string[];
        certname: string;
        certthumbprint: string;
        dnshostname: string;
        enrollmentagentrestrictionscollected: boolean;
        flags: string;
        hasbasicconstraints: boolean;
        hasenrollmentagentrestrictions?: boolean;
        isuserspecifiessanenabled?: boolean;
        isuserspecifiessanenabledcollected: boolean;
        description?: string;
    };
    controllables: number;
    controllers: number;
}

// --- GPO
export interface GPOInfoGraph extends GraphInfo {
    gpcpath: string;
}

export interface GPOInfo extends EntityInfo {
    props: BasicInfo & {
        gpcpath: string;
        description?: string;
    };
    computers: number;
    controllers: number;
    ous: number;
    users: number;
    tierzero: number;
}

// --- Group
export interface GroupInfoGraph extends GraphInfo {
    admincount: false;
}

export interface GroupInfo extends EntityInfo {
    props: BasicInfo & {
        admincount: boolean;
        description?: string;
    };
    adminRights: number;
    controllables: number;
    controllers: number;
    dcomRights: number;
    members: number;
    membership: number;
    psRemoteRights: number;
    rdpRights: number;
    sessions: number;
}

// --- NTAuthStore
export interface NTAuthStoreInfo extends EntityInfo {
    props: BasicInfo & {
        certthumbprints: string[];
        description?: string;
    };
    controllables: number;
    controllers: number;
}

// --- OU
export interface OUInfoGraph extends GraphInfo {
    blocksinheritance: boolean;
}

export interface OUInfo extends EntityInfo {
    props: BasicInfo & {
        blocksinheritance: boolean;
        description?: string;
    };
    gpos: number;
    users: number;
    groups: number;
    computers: number;
}

// --- RootCA
export interface RootCAInfo extends EntityInfo {
    props: BasicInfo & {
        certthumbprint: string;
        description?: string;
    };
    controllables: number;
    controllers: number;
}

// --- User
export interface UserInfoGraph extends GraphInfo {
    admincount: boolean;
    dontreqpreauth: boolean;
    enabled: boolean;
    hasspn: boolean;
    lastlogon: number;
    lastlogontimestamp: number;
    owned: boolean;
    passwordnotreqd: boolean;
    pwdlastset: number;
    sensitive: boolean;
    unconstraineddelegation: boolean;
}

export interface UserInfo extends EntityInfo {
    props: BasicInfo & {
        admincount: boolean;
        dontreqpreauth: boolean;
        enabled: boolean;
        hasspn: boolean;
        lastlogon: number;
        lastlogontimestamp: number;
        owned: boolean;
        passwordnotreqd: boolean;
        pwdlastset: number;
        sensitive: boolean;
        unconstraineddelegation: boolean;
        displayname?: string;
        email?: string;
        title?: string;
        homedirectory?: string;
        description?: string;
        userpassword?: string;
        pwdneverexpires?: string;
        serviceprincipalnames?: string[];
        allowedtodelegate?: boolean;
        sidhistory?: string[];
    };
    adminRights: number;
    constrainedDelegation: number;
    controllables: number;
    controllers: number;
    dcomRights: number;
    groupMembership: number;
    psRemoteRights: number;
    rdpRights: number;
    sessions: number;
    sqlAdmin: number;
}

// --- Meta
export type MetaInfoGraph = GraphInfo;

// --- Azure Entities
export interface AZAppInfo extends AZEntityInfo {
    props: BasicInfo & {
        objectid: string;
        service_principal_id: string;
    };
    inbound_object_control: number;
}

export interface AZDeviceInfo extends AZEntityInfo {
    props: BasicInfo & {
        operatingsystem: string;
    };
    inboundExecutionPrivileges: number;
    inbound_object_control: number;
}

export interface AZGroupInfo extends AZEntityInfo {
    props: BasicInfo & {
        admincount: boolean;
        isroleassignable: string;
        description?: string;
    };
    outbound_object_control: number;
    inbound_object_control: number;
    group_members: number;
    group_membership: number;
    roles: number;
}

export interface AZKeyVaultInfo extends AZEntityInfo {
    props: BasicInfo;
    Readers: {
        AllReaders: number;
        KeyReaders: number;
        CertificateReaders: number;
        SecretReaders: number;
    };
    inbound_object_control: number;
}

export interface AZManagementGroupInfo extends AZEntityInfo {
    props: BasicInfo;
    descendents: {
        descendent_counts: {
            [AzureNodeKind.ManagementGroup]: number;
            [AzureNodeKind.Subscription]: number;
            [AzureNodeKind.ResourceGroup]: number;
            [AzureNodeKind.VM]: number;
            [AzureNodeKind.KeyVault]: number;
        };
    };
    inbound_object_control: number;
}

export interface AZResourceGroupInfo extends AZEntityInfo {
    props: BasicInfo;
    descendents: {
        descendent_counts: {
            [AzureNodeKind.VM]: number;
            [AzureNodeKind.KeyVault]: number;
        };
    };

    inbound_object_control: number;
}

export interface AZRoleInfo extends AZEntityInfo {
    props: BasicInfo & {
        scope: string;
        roletemplateid: string;
    };
    active_assignments: number;
    pim_assignments: number;
}

export interface AZServicePrincipalInfo extends AZEntityInfo {
    props: BasicInfo & {
        appid: string;
        serviceprincipalid: string;
    };
    outbound_object_control: number;
    inbound_object_control: number;
    roles: number;
}

export interface AZSubscriptionInfo extends AZEntityInfo {
    props: BasicInfo;
    descendents: {
        descendent_counts: {
            [AzureNodeKind.ResourceGroup]: number;
            [AzureNodeKind.VM]: number;
            [AzureNodeKind.KeyVault]: number;
        };
    };
    inbound_object_control: number;
}

export interface AZTenantInfo extends AZEntityInfo {
    props: BasicInfo & {
        license: string;
    };
    descendents: {
        descendent_counts: {
            [AzureNodeKind.User]: number;
            [AzureNodeKind.Group]: number;
            [AzureNodeKind.App]: number;
            [AzureNodeKind.ServicePrincipal]: number;
            [AzureNodeKind.Device]: number;
            [AzureNodeKind.ManagementGroup]: number;
            [AzureNodeKind.Subscription]: number;
            [AzureNodeKind.ResourceGroup]: number;
            [AzureNodeKind.VM]: number;
            [AzureNodeKind.KeyVault]: number;
        };
    };
    inbound_object_control: number;
}

export interface AZUserInfo extends AZEntityInfo {
    props: BasicInfo & {
        licenses: string;
        mfaenforced: boolean;
        userprincipalname: string;
    };
    execution_privileges: number;
    outbound_object_control: number;
    inbound_object_control: number;
    group_membership: number;
    roles: number;
}

export interface AZVMInfo extends AZEntityInfo {
    props: BasicInfo & {
        operatingsystem: string;
    };

    inboundExecutionPrivileges: number;
    inbound_object_control: number;
}

const ENTITY_INFO_OPEN = 'app/entityinfo/OPEN';
const SET_SELECTED_NODE = 'app/entityinfo/SELECTED_NODE';

export { ENTITY_INFO_OPEN, SET_SELECTED_NODE };

export interface EntityInfo {
    props: BasicInfo;
}

export interface AZEntityInfo {
    kind: EntityKinds;
    props: BasicInfo;
}

export interface BasicInfo {
    name: string;
    objectid: string;
    displayname?: string;
    description?: string;
    whencreated?: number;
    highvalue?: boolean;
    system_tags?: string;
}

export interface GraphInfo extends BasicInfo {
    category: string;
    level: number;
    nodetype: string;
    type: string;
}

// --- Entity Info Panel
export type EntityInfoState = {
    open: boolean;
    selectedNode: SelectedNode | null;
};

interface SetEntityInfoOpenAction {
    type: typeof ENTITY_INFO_OPEN;
    open: boolean;
}

export type SelectedNode = {
    id: string;
    type: EntityKinds;
    name: string;
    graphId?: string;
};

interface SetSelectedNodeAction {
    type: typeof SET_SELECTED_NODE;
    selectedNode: SelectedNode;
}

export type EntityInfoActionTypes = SetEntityInfoOpenAction | SetSelectedNodeAction;
