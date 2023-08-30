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

export enum ActiveDirectoryNodeKind {
    Entity = 'Base',
    User = 'User',
    Computer = 'Computer',
    Group = 'Group',
    GPO = 'GPO',
    OU = 'OU',
    Container = 'Container',
    Domain = 'Domain',
    LocalGroup = 'ADLocalGroup',
    LocalUser = 'ADLocalUser',
}
export function ActiveDirectoryNodeKindToDisplay(value: ActiveDirectoryNodeKind): string | undefined {
    switch (value) {
        case ActiveDirectoryNodeKind.Entity:
            return 'Entity';
        case ActiveDirectoryNodeKind.User:
            return 'User';
        case ActiveDirectoryNodeKind.Computer:
            return 'Computer';
        case ActiveDirectoryNodeKind.Group:
            return 'Group';
        case ActiveDirectoryNodeKind.GPO:
            return 'GPO';
        case ActiveDirectoryNodeKind.OU:
            return 'OU';
        case ActiveDirectoryNodeKind.Container:
            return 'Container';
        case ActiveDirectoryNodeKind.Domain:
            return 'Domain';
        case ActiveDirectoryNodeKind.LocalGroup:
            return 'LocalGroup';
        case ActiveDirectoryNodeKind.LocalUser:
            return 'LocalUser';
        default:
            return undefined;
    }
}
export enum ActiveDirectoryRelationshipKind {
    Owns = 'Owns',
    GenericAll = 'GenericAll',
    GenericWrite = 'GenericWrite',
    WriteOwner = 'WriteOwner',
    WriteDACL = 'WriteDacl',
    MemberOf = 'MemberOf',
    ForceChangePassword = 'ForceChangePassword',
    AllExtendedRights = 'AllExtendedRights',
    AddMember = 'AddMember',
    HasSession = 'HasSession',
    Contains = 'Contains',
    GPLink = 'GPLink',
    AllowedToDelegate = 'AllowedToDelegate',
    GetChanges = 'GetChanges',
    GetChangesAll = 'GetChangesAll',
    GetChangesInFilteredSet = 'GetChangesInFilteredSet',
    TrustedBy = 'TrustedBy',
    AllowedToAct = 'AllowedToAct',
    AdminTo = 'AdminTo',
    CanPSRemote = 'CanPSRemote',
    CanRDP = 'CanRDP',
    ExecuteDCOM = 'ExecuteDCOM',
    HasSIDHistory = 'HasSIDHistory',
    AddSelf = 'AddSelf',
    DCSync = 'DCSync',
    ReadLAPSPassword = 'ReadLAPSPassword',
    ReadGMSAPassword = 'ReadGMSAPassword',
    DumpSMSAPassword = 'DumpSMSAPassword',
    SQLAdmin = 'SQLAdmin',
    AddAllowedToAct = 'AddAllowedToAct',
    WriteSPN = 'WriteSPN',
    AddKeyCredentialLink = 'AddKeyCredentialLink',
    LocalToComputer = 'LocalToComputer',
    MemberOfLocalGroup = 'MemberOfLocalGroup',
    RemoteInteractiveLogonPrivilege = 'RemoteInteractiveLogonPrivilege',
    SyncLAPSPassword = 'SyncLAPSPassword',
    WriteAccountRestrictions = 'WriteAccountRestrictions',
}
export function ActiveDirectoryRelationshipKindToDisplay(value: ActiveDirectoryRelationshipKind): string | undefined {
    switch (value) {
        case ActiveDirectoryRelationshipKind.Owns:
            return 'Owns';
        case ActiveDirectoryRelationshipKind.GenericAll:
            return 'GenericAll';
        case ActiveDirectoryRelationshipKind.GenericWrite:
            return 'GenericWrite';
        case ActiveDirectoryRelationshipKind.WriteOwner:
            return 'WriteOwner';
        case ActiveDirectoryRelationshipKind.WriteDACL:
            return 'WriteDACL';
        case ActiveDirectoryRelationshipKind.MemberOf:
            return 'MemberOf';
        case ActiveDirectoryRelationshipKind.ForceChangePassword:
            return 'ForceChangePassword';
        case ActiveDirectoryRelationshipKind.AllExtendedRights:
            return 'AllExtendedRights';
        case ActiveDirectoryRelationshipKind.AddMember:
            return 'AddMember';
        case ActiveDirectoryRelationshipKind.HasSession:
            return 'HasSession';
        case ActiveDirectoryRelationshipKind.Contains:
            return 'Contains';
        case ActiveDirectoryRelationshipKind.GPLink:
            return 'GPLink';
        case ActiveDirectoryRelationshipKind.AllowedToDelegate:
            return 'AllowedToDelegate';
        case ActiveDirectoryRelationshipKind.GetChanges:
            return 'GetChanges';
        case ActiveDirectoryRelationshipKind.GetChangesAll:
            return 'GetChangesAll';
        case ActiveDirectoryRelationshipKind.GetChangesInFilteredSet:
            return 'GetChangesInFilteredSet';
        case ActiveDirectoryRelationshipKind.TrustedBy:
            return 'TrustedBy';
        case ActiveDirectoryRelationshipKind.AllowedToAct:
            return 'AllowedToAct';
        case ActiveDirectoryRelationshipKind.AdminTo:
            return 'AdminTo';
        case ActiveDirectoryRelationshipKind.CanPSRemote:
            return 'CanPSRemote';
        case ActiveDirectoryRelationshipKind.CanRDP:
            return 'CanRDP';
        case ActiveDirectoryRelationshipKind.ExecuteDCOM:
            return 'ExecuteDCOM';
        case ActiveDirectoryRelationshipKind.HasSIDHistory:
            return 'HasSIDHistory';
        case ActiveDirectoryRelationshipKind.AddSelf:
            return 'AddSelf';
        case ActiveDirectoryRelationshipKind.DCSync:
            return 'DCSync';
        case ActiveDirectoryRelationshipKind.ReadLAPSPassword:
            return 'ReadLAPSPassword';
        case ActiveDirectoryRelationshipKind.ReadGMSAPassword:
            return 'ReadGMSAPassword';
        case ActiveDirectoryRelationshipKind.DumpSMSAPassword:
            return 'DumpSMSAPassword';
        case ActiveDirectoryRelationshipKind.SQLAdmin:
            return 'SQLAdmin';
        case ActiveDirectoryRelationshipKind.AddAllowedToAct:
            return 'AddAllowedToAct';
        case ActiveDirectoryRelationshipKind.WriteSPN:
            return 'WriteSPN';
        case ActiveDirectoryRelationshipKind.AddKeyCredentialLink:
            return 'AddKeyCredentialLink';
        case ActiveDirectoryRelationshipKind.LocalToComputer:
            return 'LocalToComputer';
        case ActiveDirectoryRelationshipKind.MemberOfLocalGroup:
            return 'MemberOfLocalGroup';
        case ActiveDirectoryRelationshipKind.RemoteInteractiveLogonPrivilege:
            return 'RemoteInteractiveLogonPrivilege';
        case ActiveDirectoryRelationshipKind.SyncLAPSPassword:
            return 'SyncLAPSPassword';
        case ActiveDirectoryRelationshipKind.WriteAccountRestrictions:
            return 'WriteAccountRestrictions';
        default:
            return undefined;
    }
}
export enum ActiveDirectoryKind {
    Entity = 'Base',
    User = 'User',
    Computer = 'Computer',
    Group = 'Group',
    GPO = 'GPO',
    OU = 'OU',
    Container = 'Container',
    Domain = 'Domain',
    LocalGroup = 'ADLocalGroup',
    LocalUser = 'ADLocalUser',
    Owns = 'Owns',
    GenericAll = 'GenericAll',
    GenericWrite = 'GenericWrite',
    WriteOwner = 'WriteOwner',
    WriteDACL = 'WriteDacl',
    MemberOf = 'MemberOf',
    ForceChangePassword = 'ForceChangePassword',
    AllExtendedRights = 'AllExtendedRights',
    AddMember = 'AddMember',
    HasSession = 'HasSession',
    Contains = 'Contains',
    GPLink = 'GPLink',
    AllowedToDelegate = 'AllowedToDelegate',
    GetChanges = 'GetChanges',
    GetChangesAll = 'GetChangesAll',
    GetChangesInFilteredSet = 'GetChangesInFilteredSet',
    TrustedBy = 'TrustedBy',
    AllowedToAct = 'AllowedToAct',
    AdminTo = 'AdminTo',
    CanPSRemote = 'CanPSRemote',
    CanRDP = 'CanRDP',
    ExecuteDCOM = 'ExecuteDCOM',
    HasSIDHistory = 'HasSIDHistory',
    AddSelf = 'AddSelf',
    DCSync = 'DCSync',
    ReadLAPSPassword = 'ReadLAPSPassword',
    ReadGMSAPassword = 'ReadGMSAPassword',
    DumpSMSAPassword = 'DumpSMSAPassword',
    SQLAdmin = 'SQLAdmin',
    AddAllowedToAct = 'AddAllowedToAct',
    WriteSPN = 'WriteSPN',
    AddKeyCredentialLink = 'AddKeyCredentialLink',
    LocalToComputer = 'LocalToComputer',
    MemberOfLocalGroup = 'MemberOfLocalGroup',
    RemoteInteractiveLogonPrivilege = 'RemoteInteractiveLogonPrivilege',
    SyncLAPSPassword = 'SyncLAPSPassword',
    WriteAccountRestrictions = 'WriteAccountRestrictions',
}
export function ActiveDirectoryKindToDisplay(value: ActiveDirectoryKind): string | undefined {
    switch (value) {
        case ActiveDirectoryKind.Entity:
            return 'Entity';
        case ActiveDirectoryKind.User:
            return 'User';
        case ActiveDirectoryKind.Computer:
            return 'Computer';
        case ActiveDirectoryKind.Group:
            return 'Group';
        case ActiveDirectoryKind.GPO:
            return 'GPO';
        case ActiveDirectoryKind.OU:
            return 'OU';
        case ActiveDirectoryKind.Container:
            return 'Container';
        case ActiveDirectoryKind.Domain:
            return 'Domain';
        case ActiveDirectoryKind.LocalGroup:
            return 'LocalGroup';
        case ActiveDirectoryKind.LocalUser:
            return 'LocalUser';
        case ActiveDirectoryKind.Owns:
            return 'Owns';
        case ActiveDirectoryKind.GenericAll:
            return 'GenericAll';
        case ActiveDirectoryKind.GenericWrite:
            return 'GenericWrite';
        case ActiveDirectoryKind.WriteOwner:
            return 'WriteOwner';
        case ActiveDirectoryKind.WriteDACL:
            return 'WriteDACL';
        case ActiveDirectoryKind.MemberOf:
            return 'MemberOf';
        case ActiveDirectoryKind.ForceChangePassword:
            return 'ForceChangePassword';
        case ActiveDirectoryKind.AllExtendedRights:
            return 'AllExtendedRights';
        case ActiveDirectoryKind.AddMember:
            return 'AddMember';
        case ActiveDirectoryKind.HasSession:
            return 'HasSession';
        case ActiveDirectoryKind.Contains:
            return 'Contains';
        case ActiveDirectoryKind.GPLink:
            return 'GPLink';
        case ActiveDirectoryKind.AllowedToDelegate:
            return 'AllowedToDelegate';
        case ActiveDirectoryKind.GetChanges:
            return 'GetChanges';
        case ActiveDirectoryKind.GetChangesAll:
            return 'GetChangesAll';
        case ActiveDirectoryKind.GetChangesInFilteredSet:
            return 'GetChangesInFilteredSet';
        case ActiveDirectoryKind.TrustedBy:
            return 'TrustedBy';
        case ActiveDirectoryKind.AllowedToAct:
            return 'AllowedToAct';
        case ActiveDirectoryKind.AdminTo:
            return 'AdminTo';
        case ActiveDirectoryKind.CanPSRemote:
            return 'CanPSRemote';
        case ActiveDirectoryKind.CanRDP:
            return 'CanRDP';
        case ActiveDirectoryKind.ExecuteDCOM:
            return 'ExecuteDCOM';
        case ActiveDirectoryKind.HasSIDHistory:
            return 'HasSIDHistory';
        case ActiveDirectoryKind.AddSelf:
            return 'AddSelf';
        case ActiveDirectoryKind.DCSync:
            return 'DCSync';
        case ActiveDirectoryKind.ReadLAPSPassword:
            return 'ReadLAPSPassword';
        case ActiveDirectoryKind.ReadGMSAPassword:
            return 'ReadGMSAPassword';
        case ActiveDirectoryKind.DumpSMSAPassword:
            return 'DumpSMSAPassword';
        case ActiveDirectoryKind.SQLAdmin:
            return 'SQLAdmin';
        case ActiveDirectoryKind.AddAllowedToAct:
            return 'AddAllowedToAct';
        case ActiveDirectoryKind.WriteSPN:
            return 'WriteSPN';
        case ActiveDirectoryKind.AddKeyCredentialLink:
            return 'AddKeyCredentialLink';
        case ActiveDirectoryKind.LocalToComputer:
            return 'LocalToComputer';
        case ActiveDirectoryKind.MemberOfLocalGroup:
            return 'MemberOfLocalGroup';
        case ActiveDirectoryKind.RemoteInteractiveLogonPrivilege:
            return 'RemoteInteractiveLogonPrivilege';
        case ActiveDirectoryKind.SyncLAPSPassword:
            return 'SyncLAPSPassword';
        case ActiveDirectoryKind.WriteAccountRestrictions:
            return 'WriteAccountRestrictions';
        default:
            return undefined;
    }
}
export enum ActiveDirectoryKindProperties {
    AdminCount = 'admincount',
    DistinguishedName = 'distinguishedname',
    DomainFQDN = 'domain',
    DomainSID = 'domainsid',
    Sensitive = 'sensitive',
    HighValue = 'highvalue',
    BlocksInheritance = 'blocksinheritance',
    IsACL = 'isacl',
    IsACLProtected = 'isaclprotected',
    Enforced = 'enforced',
    Department = 'department',
    HasSPN = 'hasspn',
    UnconstrainedDelegation = 'unconstraineddelegation',
    LastLogon = 'lastlogon',
    LastLogonTimestamp = 'lastlogontimestamp',
    IsPrimaryGroup = 'isprimarygroup',
    HasLAPS = 'haslaps',
    DontRequirePreAuth = 'dontreqpreauth',
    LogonType = 'logontype',
    HasURA = 'hasura',
}
export function ActiveDirectoryKindPropertiesToDisplay(value: ActiveDirectoryKindProperties): string | undefined {
    switch (value) {
        case ActiveDirectoryKindProperties.AdminCount:
            return 'Admin Count';
        case ActiveDirectoryKindProperties.DistinguishedName:
            return 'Distinguished Name';
        case ActiveDirectoryKindProperties.DomainFQDN:
            return 'Domain FQDN';
        case ActiveDirectoryKindProperties.DomainSID:
            return 'Domain SID';
        case ActiveDirectoryKindProperties.Sensitive:
            return 'Sensitive';
        case ActiveDirectoryKindProperties.HighValue:
            return 'High Value';
        case ActiveDirectoryKindProperties.BlocksInheritance:
            return 'Blocks Inheritance';
        case ActiveDirectoryKindProperties.IsACL:
            return 'Is ACL';
        case ActiveDirectoryKindProperties.IsACLProtected:
            return 'ACL Inheritance Denied';
        case ActiveDirectoryKindProperties.Enforced:
            return 'Enforced';
        case ActiveDirectoryKindProperties.Department:
            return 'Department';
        case ActiveDirectoryKindProperties.HasSPN:
            return 'Has SPN';
        case ActiveDirectoryKindProperties.UnconstrainedDelegation:
            return 'Allows Unconstrained Delegation';
        case ActiveDirectoryKindProperties.LastLogon:
            return 'Last Logon';
        case ActiveDirectoryKindProperties.LastLogonTimestamp:
            return 'Last Logon (Replicated)';
        case ActiveDirectoryKindProperties.IsPrimaryGroup:
            return 'Is Primary Group';
        case ActiveDirectoryKindProperties.HasLAPS:
            return 'LAPS Enabled';
        case ActiveDirectoryKindProperties.DontRequirePreAuth:
            return 'Do Not Require Pre-Authentication';
        case ActiveDirectoryKindProperties.LogonType:
            return 'Logon Type';
        case ActiveDirectoryKindProperties.HasURA:
            return 'Has User Rights Assignment Collection';
        default:
            return undefined;
    }
}
export function ActiveDirectoryPathfindingEdges(): ActiveDirectoryRelationshipKind[] {
    return [
        ActiveDirectoryRelationshipKind.Owns,
        ActiveDirectoryRelationshipKind.GenericAll,
        ActiveDirectoryRelationshipKind.GenericWrite,
        ActiveDirectoryRelationshipKind.WriteOwner,
        ActiveDirectoryRelationshipKind.WriteDACL,
        ActiveDirectoryRelationshipKind.MemberOf,
        ActiveDirectoryRelationshipKind.ForceChangePassword,
        ActiveDirectoryRelationshipKind.AllExtendedRights,
        ActiveDirectoryRelationshipKind.AddMember,
        ActiveDirectoryRelationshipKind.HasSession,
        ActiveDirectoryRelationshipKind.Contains,
        ActiveDirectoryRelationshipKind.GPLink,
        ActiveDirectoryRelationshipKind.AllowedToDelegate,
        ActiveDirectoryRelationshipKind.TrustedBy,
        ActiveDirectoryRelationshipKind.AllowedToAct,
        ActiveDirectoryRelationshipKind.AdminTo,
        ActiveDirectoryRelationshipKind.CanPSRemote,
        ActiveDirectoryRelationshipKind.CanRDP,
        ActiveDirectoryRelationshipKind.ExecuteDCOM,
        ActiveDirectoryRelationshipKind.HasSIDHistory,
        ActiveDirectoryRelationshipKind.AddSelf,
        ActiveDirectoryRelationshipKind.DCSync,
        ActiveDirectoryRelationshipKind.ReadLAPSPassword,
        ActiveDirectoryRelationshipKind.ReadGMSAPassword,
        ActiveDirectoryRelationshipKind.DumpSMSAPassword,
        ActiveDirectoryRelationshipKind.SQLAdmin,
        ActiveDirectoryRelationshipKind.AddAllowedToAct,
        ActiveDirectoryRelationshipKind.WriteSPN,
        ActiveDirectoryRelationshipKind.AddKeyCredentialLink,
        ActiveDirectoryRelationshipKind.SyncLAPSPassword,
        ActiveDirectoryRelationshipKind.WriteAccountRestrictions,
    ];
}
export enum AzureNodeKind {
    Entity = 'AZBase',
    VMScaleSet = 'AZVMScaleSet',
    App = 'AZApp',
    Role = 'AZRole',
    Device = 'AZDevice',
    FunctionApp = 'AZFunctionApp',
    Group = 'AZGroup',
    KeyVault = 'AZKeyVault',
    ManagementGroup = 'AZManagementGroup',
    ResourceGroup = 'AZResourceGroup',
    ServicePrincipal = 'AZServicePrincipal',
    Subscription = 'AZSubscription',
    Tenant = 'AZTenant',
    User = 'AZUser',
    VM = 'AZVM',
    ManagedCluster = 'AZManagedCluster',
    ContainerRegistry = 'AZContainerRegistry',
    WebApp = 'AZWebApp',
    LogicApp = 'AZLogicApp',
    AutomationAccount = 'AZAutomationAccount',
}
export function AzureNodeKindToDisplay(value: AzureNodeKind): string | undefined {
    switch (value) {
        case AzureNodeKind.Entity:
            return 'Entity';
        case AzureNodeKind.VMScaleSet:
            return 'VMScaleSet';
        case AzureNodeKind.App:
            return 'App';
        case AzureNodeKind.Role:
            return 'Role';
        case AzureNodeKind.Device:
            return 'Device';
        case AzureNodeKind.FunctionApp:
            return 'FunctionApp';
        case AzureNodeKind.Group:
            return 'Group';
        case AzureNodeKind.KeyVault:
            return 'KeyVault';
        case AzureNodeKind.ManagementGroup:
            return 'ManagementGroup';
        case AzureNodeKind.ResourceGroup:
            return 'ResourceGroup';
        case AzureNodeKind.ServicePrincipal:
            return 'ServicePrincipal';
        case AzureNodeKind.Subscription:
            return 'Subscription';
        case AzureNodeKind.Tenant:
            return 'Tenant';
        case AzureNodeKind.User:
            return 'User';
        case AzureNodeKind.VM:
            return 'VM';
        case AzureNodeKind.ManagedCluster:
            return 'ManagedCluster';
        case AzureNodeKind.ContainerRegistry:
            return 'ContainerRegistry';
        case AzureNodeKind.WebApp:
            return 'WebApp';
        case AzureNodeKind.LogicApp:
            return 'LogicApp';
        case AzureNodeKind.AutomationAccount:
            return 'AutomationAccount';
        default:
            return undefined;
    }
}
export enum AzureRelationshipKind {
    AvereContributor = 'AZAvereContributor',
    Contains = 'AZContains',
    Contributor = 'AZContributor',
    GetCertificates = 'AZGetCertificates',
    GetKeys = 'AZGetKeys',
    GetSecrets = 'AZGetSecrets',
    HasRole = 'AZHasRole',
    MemberOf = 'AZMemberOf',
    Owner = 'AZOwner',
    RunsAs = 'AZRunsAs',
    VMContributor = 'AZVMContributor',
    AutomationContributor = 'AZAutomationContributor',
    KeyVaultContributor = 'AZKeyVaultContributor',
    VMAdminLogin = 'AZVMAdminLogin',
    AddMembers = 'AZAddMembers',
    AddSecret = 'AZAddSecret',
    ExecuteCommand = 'AZExecuteCommand',
    GlobalAdmin = 'AZGlobalAdmin',
    PrivilegedAuthAdmin = 'AZPrivilegedAuthAdmin',
    Grant = 'AZGrant',
    GrantSelf = 'AZGrantSelf',
    PrivilegedRoleAdmin = 'AZPrivilegedRoleAdmin',
    ResetPassword = 'AZResetPassword',
    UserAccessAdministrator = 'AZUserAccessAdministrator',
    Owns = 'AZOwns',
    ScopedTo = 'AZScopedTo',
    CloudAppAdmin = 'AZCloudAppAdmin',
    AppAdmin = 'AZAppAdmin',
    AddOwner = 'AZAddOwner',
    ManagedIdentity = 'AZManagedIdentity',
    ApplicationReadWriteAll = 'AZMGApplication_ReadWrite_All',
    AppRoleAssignmentReadWriteAll = 'AZMGAppRoleAssignment_ReadWrite_All',
    DirectoryReadWriteAll = 'AZMGDirectory_ReadWrite_All',
    GroupReadWriteAll = 'AZMGGroup_ReadWrite_All',
    GroupMemberReadWriteAll = 'AZMGGroupMember_ReadWrite_All',
    RoleManagementReadWriteDirectory = 'AZMGRoleManagement_ReadWrite_Directory',
    ServicePrincipalEndpointReadWriteAll = 'AZMGServicePrincipalEndpoint_ReadWrite_All',
    AKSContributor = 'AZAKSContributor',
    NodeResourceGroup = 'AZNodeResourceGroup',
    WebsiteContributor = 'AZWebsiteContributor',
    LogicAppContributor = 'AZLogicAppContributor',
    AZMGAddMember = 'AZMGAddMember',
    AZMGAddOwner = 'AZMGAddOwner',
    AZMGAddSecret = 'AZMGAddSecret',
    AZMGGrantAppRoles = 'AZMGGrantAppRoles',
    AZMGGrantRole = 'AZMGGrantRole',
}
export function AzureRelationshipKindToDisplay(value: AzureRelationshipKind): string | undefined {
    switch (value) {
        case AzureRelationshipKind.AvereContributor:
            return 'AvereContributor';
        case AzureRelationshipKind.Contains:
            return 'Contains';
        case AzureRelationshipKind.Contributor:
            return 'Contributor';
        case AzureRelationshipKind.GetCertificates:
            return 'GetCertificates';
        case AzureRelationshipKind.GetKeys:
            return 'GetKeys';
        case AzureRelationshipKind.GetSecrets:
            return 'GetSecrets';
        case AzureRelationshipKind.HasRole:
            return 'HasRole';
        case AzureRelationshipKind.MemberOf:
            return 'MemberOf';
        case AzureRelationshipKind.Owner:
            return 'Owner';
        case AzureRelationshipKind.RunsAs:
            return 'RunsAs';
        case AzureRelationshipKind.VMContributor:
            return 'VMContributor';
        case AzureRelationshipKind.AutomationContributor:
            return 'AutomationContributor';
        case AzureRelationshipKind.KeyVaultContributor:
            return 'KeyVaultContributor';
        case AzureRelationshipKind.VMAdminLogin:
            return 'VMAdminLogin';
        case AzureRelationshipKind.AddMembers:
            return 'AddMembers';
        case AzureRelationshipKind.AddSecret:
            return 'AddSecret';
        case AzureRelationshipKind.ExecuteCommand:
            return 'ExecuteCommand';
        case AzureRelationshipKind.GlobalAdmin:
            return 'GlobalAdmin';
        case AzureRelationshipKind.PrivilegedAuthAdmin:
            return 'PrivilegedAuthAdmin';
        case AzureRelationshipKind.Grant:
            return 'Grant';
        case AzureRelationshipKind.GrantSelf:
            return 'GrantSelf';
        case AzureRelationshipKind.PrivilegedRoleAdmin:
            return 'PrivilegedRoleAdmin';
        case AzureRelationshipKind.ResetPassword:
            return 'ResetPassword';
        case AzureRelationshipKind.UserAccessAdministrator:
            return 'UserAccessAdministrator';
        case AzureRelationshipKind.Owns:
            return 'Owns';
        case AzureRelationshipKind.ScopedTo:
            return 'ScopedTo';
        case AzureRelationshipKind.CloudAppAdmin:
            return 'CloudAppAdmin';
        case AzureRelationshipKind.AppAdmin:
            return 'AppAdmin';
        case AzureRelationshipKind.AddOwner:
            return 'AddOwner';
        case AzureRelationshipKind.ManagedIdentity:
            return 'ManagedIdentity';
        case AzureRelationshipKind.ApplicationReadWriteAll:
            return 'ApplicationReadWriteAll';
        case AzureRelationshipKind.AppRoleAssignmentReadWriteAll:
            return 'AppRoleAssignmentReadWriteAll';
        case AzureRelationshipKind.DirectoryReadWriteAll:
            return 'DirectoryReadWriteAll';
        case AzureRelationshipKind.GroupReadWriteAll:
            return 'GroupReadWriteAll';
        case AzureRelationshipKind.GroupMemberReadWriteAll:
            return 'GroupMemberReadWriteAll';
        case AzureRelationshipKind.RoleManagementReadWriteDirectory:
            return 'RoleManagementReadWriteDirectory';
        case AzureRelationshipKind.ServicePrincipalEndpointReadWriteAll:
            return 'ServicePrincipalEndpointReadWriteAll';
        case AzureRelationshipKind.AKSContributor:
            return 'AKSContributor';
        case AzureRelationshipKind.NodeResourceGroup:
            return 'NodeResourceGroup';
        case AzureRelationshipKind.WebsiteContributor:
            return 'WebsiteContributor';
        case AzureRelationshipKind.LogicAppContributor:
            return 'LogicAppContributor';
        case AzureRelationshipKind.AZMGAddMember:
            return 'AZMGAddMember';
        case AzureRelationshipKind.AZMGAddOwner:
            return 'AZMGAddOwner';
        case AzureRelationshipKind.AZMGAddSecret:
            return 'AZMGAddSecret';
        case AzureRelationshipKind.AZMGGrantAppRoles:
            return 'AZMGGrantAppRoles';
        case AzureRelationshipKind.AZMGGrantRole:
            return 'AZMGGrantRole';
        default:
            return undefined;
    }
}
export enum AzureKind {
    Entity = 'AZBase',
    VMScaleSet = 'AZVMScaleSet',
    App = 'AZApp',
    Role = 'AZRole',
    Device = 'AZDevice',
    FunctionApp = 'AZFunctionApp',
    Group = 'AZGroup',
    KeyVault = 'AZKeyVault',
    ManagementGroup = 'AZManagementGroup',
    ResourceGroup = 'AZResourceGroup',
    ServicePrincipal = 'AZServicePrincipal',
    Subscription = 'AZSubscription',
    Tenant = 'AZTenant',
    User = 'AZUser',
    VM = 'AZVM',
    ManagedCluster = 'AZManagedCluster',
    ContainerRegistry = 'AZContainerRegistry',
    WebApp = 'AZWebApp',
    LogicApp = 'AZLogicApp',
    AutomationAccount = 'AZAutomationAccount',
    AvereContributor = 'AZAvereContributor',
    Contains = 'AZContains',
    Contributor = 'AZContributor',
    GetCertificates = 'AZGetCertificates',
    GetKeys = 'AZGetKeys',
    GetSecrets = 'AZGetSecrets',
    HasRole = 'AZHasRole',
    MemberOf = 'AZMemberOf',
    Owner = 'AZOwner',
    RunsAs = 'AZRunsAs',
    VMContributor = 'AZVMContributor',
    AutomationContributor = 'AZAutomationContributor',
    KeyVaultContributor = 'AZKeyVaultContributor',
    VMAdminLogin = 'AZVMAdminLogin',
    AddMembers = 'AZAddMembers',
    AddSecret = 'AZAddSecret',
    ExecuteCommand = 'AZExecuteCommand',
    GlobalAdmin = 'AZGlobalAdmin',
    PrivilegedAuthAdmin = 'AZPrivilegedAuthAdmin',
    Grant = 'AZGrant',
    GrantSelf = 'AZGrantSelf',
    PrivilegedRoleAdmin = 'AZPrivilegedRoleAdmin',
    ResetPassword = 'AZResetPassword',
    UserAccessAdministrator = 'AZUserAccessAdministrator',
    Owns = 'AZOwns',
    ScopedTo = 'AZScopedTo',
    CloudAppAdmin = 'AZCloudAppAdmin',
    AppAdmin = 'AZAppAdmin',
    AddOwner = 'AZAddOwner',
    ManagedIdentity = 'AZManagedIdentity',
    ApplicationReadWriteAll = 'AZMGApplication_ReadWrite_All',
    AppRoleAssignmentReadWriteAll = 'AZMGAppRoleAssignment_ReadWrite_All',
    DirectoryReadWriteAll = 'AZMGDirectory_ReadWrite_All',
    GroupReadWriteAll = 'AZMGGroup_ReadWrite_All',
    GroupMemberReadWriteAll = 'AZMGGroupMember_ReadWrite_All',
    RoleManagementReadWriteDirectory = 'AZMGRoleManagement_ReadWrite_Directory',
    ServicePrincipalEndpointReadWriteAll = 'AZMGServicePrincipalEndpoint_ReadWrite_All',
    AKSContributor = 'AZAKSContributor',
    NodeResourceGroup = 'AZNodeResourceGroup',
    WebsiteContributor = 'AZWebsiteContributor',
    LogicAppContributor = 'AZLogicAppContributor',
    AZMGAddMember = 'AZMGAddMember',
    AZMGAddOwner = 'AZMGAddOwner',
    AZMGAddSecret = 'AZMGAddSecret',
    AZMGGrantAppRoles = 'AZMGGrantAppRoles',
    AZMGGrantRole = 'AZMGGrantRole',
}
export function AzureKindToDisplay(value: AzureKind): string | undefined {
    switch (value) {
        case AzureKind.Entity:
            return 'Entity';
        case AzureKind.VMScaleSet:
            return 'VMScaleSet';
        case AzureKind.App:
            return 'App';
        case AzureKind.Role:
            return 'Role';
        case AzureKind.Device:
            return 'Device';
        case AzureKind.FunctionApp:
            return 'FunctionApp';
        case AzureKind.Group:
            return 'Group';
        case AzureKind.KeyVault:
            return 'KeyVault';
        case AzureKind.ManagementGroup:
            return 'ManagementGroup';
        case AzureKind.ResourceGroup:
            return 'ResourceGroup';
        case AzureKind.ServicePrincipal:
            return 'ServicePrincipal';
        case AzureKind.Subscription:
            return 'Subscription';
        case AzureKind.Tenant:
            return 'Tenant';
        case AzureKind.User:
            return 'User';
        case AzureKind.VM:
            return 'VM';
        case AzureKind.ManagedCluster:
            return 'ManagedCluster';
        case AzureKind.ContainerRegistry:
            return 'ContainerRegistry';
        case AzureKind.WebApp:
            return 'WebApp';
        case AzureKind.LogicApp:
            return 'LogicApp';
        case AzureKind.AutomationAccount:
            return 'AutomationAccount';
        case AzureKind.AvereContributor:
            return 'AvereContributor';
        case AzureKind.Contains:
            return 'Contains';
        case AzureKind.Contributor:
            return 'Contributor';
        case AzureKind.GetCertificates:
            return 'GetCertificates';
        case AzureKind.GetKeys:
            return 'GetKeys';
        case AzureKind.GetSecrets:
            return 'GetSecrets';
        case AzureKind.HasRole:
            return 'HasRole';
        case AzureKind.MemberOf:
            return 'MemberOf';
        case AzureKind.Owner:
            return 'Owner';
        case AzureKind.RunsAs:
            return 'RunsAs';
        case AzureKind.VMContributor:
            return 'VMContributor';
        case AzureKind.AutomationContributor:
            return 'AutomationContributor';
        case AzureKind.KeyVaultContributor:
            return 'KeyVaultContributor';
        case AzureKind.VMAdminLogin:
            return 'VMAdminLogin';
        case AzureKind.AddMembers:
            return 'AddMembers';
        case AzureKind.AddSecret:
            return 'AddSecret';
        case AzureKind.ExecuteCommand:
            return 'ExecuteCommand';
        case AzureKind.GlobalAdmin:
            return 'GlobalAdmin';
        case AzureKind.PrivilegedAuthAdmin:
            return 'PrivilegedAuthAdmin';
        case AzureKind.Grant:
            return 'Grant';
        case AzureKind.GrantSelf:
            return 'GrantSelf';
        case AzureKind.PrivilegedRoleAdmin:
            return 'PrivilegedRoleAdmin';
        case AzureKind.ResetPassword:
            return 'ResetPassword';
        case AzureKind.UserAccessAdministrator:
            return 'UserAccessAdministrator';
        case AzureKind.Owns:
            return 'Owns';
        case AzureKind.ScopedTo:
            return 'ScopedTo';
        case AzureKind.CloudAppAdmin:
            return 'CloudAppAdmin';
        case AzureKind.AppAdmin:
            return 'AppAdmin';
        case AzureKind.AddOwner:
            return 'AddOwner';
        case AzureKind.ManagedIdentity:
            return 'ManagedIdentity';
        case AzureKind.ApplicationReadWriteAll:
            return 'ApplicationReadWriteAll';
        case AzureKind.AppRoleAssignmentReadWriteAll:
            return 'AppRoleAssignmentReadWriteAll';
        case AzureKind.DirectoryReadWriteAll:
            return 'DirectoryReadWriteAll';
        case AzureKind.GroupReadWriteAll:
            return 'GroupReadWriteAll';
        case AzureKind.GroupMemberReadWriteAll:
            return 'GroupMemberReadWriteAll';
        case AzureKind.RoleManagementReadWriteDirectory:
            return 'RoleManagementReadWriteDirectory';
        case AzureKind.ServicePrincipalEndpointReadWriteAll:
            return 'ServicePrincipalEndpointReadWriteAll';
        case AzureKind.AKSContributor:
            return 'AKSContributor';
        case AzureKind.NodeResourceGroup:
            return 'NodeResourceGroup';
        case AzureKind.WebsiteContributor:
            return 'WebsiteContributor';
        case AzureKind.LogicAppContributor:
            return 'LogicAppContributor';
        case AzureKind.AZMGAddMember:
            return 'AZMGAddMember';
        case AzureKind.AZMGAddOwner:
            return 'AZMGAddOwner';
        case AzureKind.AZMGAddSecret:
            return 'AZMGAddSecret';
        case AzureKind.AZMGGrantAppRoles:
            return 'AZMGGrantAppRoles';
        case AzureKind.AZMGGrantRole:
            return 'AZMGGrantRole';
        default:
            return undefined;
    }
}
export enum AzureKindProperties {
    AppOwnerOrganizationID = 'appownerorganizationid',
    AppDescription = 'appdescription',
    AppDisplayName = 'appdisplayname',
    ServicePrincipalType = 'serviceprincipaltype',
    UserType = 'usertype',
    TenantID = 'tenantid',
    ServicePrincipalID = 'service_principal_id',
    OperatingSystemVersion = 'operatingsystemversion',
    TrustType = 'trustype',
    IsBuiltIn = 'isbuiltin',
    AppID = 'appid',
    AppRoleID = 'approleid',
    DeviceID = 'deviceid',
    NodeResourceGroupID = 'noderesourcegroupid',
    OnPremID = 'onpremid',
    OnPremSyncEnabled = 'onpremsyncenabled',
    SecurityEnabled = 'securityenabled',
    SecurityIdentifier = 'securityidentifier',
    EnableRBACAuthorization = 'enablerbacauthorization',
    Scope = 'scope',
    Offer = 'offer',
    MFAEnabled = 'mfaenabled',
    License = 'license',
    Licenses = 'licenses',
    MFAEnforced = 'mfaenforced',
    UserPrincipalName = 'userprincipalname',
    IsAssignableToRole = 'isassignabletorole',
    PublisherDomain = 'publisherdomain',
    SignInAudience = 'signinaudience',
    RoleTemplateID = 'templateid',
}
export function AzureKindPropertiesToDisplay(value: AzureKindProperties): string | undefined {
    switch (value) {
        case AzureKindProperties.AppOwnerOrganizationID:
            return 'App Owner Organization ID';
        case AzureKindProperties.AppDescription:
            return 'App Description';
        case AzureKindProperties.AppDisplayName:
            return 'App Display Name';
        case AzureKindProperties.ServicePrincipalType:
            return 'Service Principal Type';
        case AzureKindProperties.UserType:
            return 'User Type';
        case AzureKindProperties.TenantID:
            return 'Tenant ID';
        case AzureKindProperties.ServicePrincipalID:
            return 'Service Principal ID';
        case AzureKindProperties.OperatingSystemVersion:
            return 'Operating System Version';
        case AzureKindProperties.TrustType:
            return 'Trust Type';
        case AzureKindProperties.IsBuiltIn:
            return 'Is Built In';
        case AzureKindProperties.AppID:
            return 'App ID';
        case AzureKindProperties.AppRoleID:
            return 'App Role ID';
        case AzureKindProperties.DeviceID:
            return 'Device ID';
        case AzureKindProperties.NodeResourceGroupID:
            return 'Node Resource Group ID';
        case AzureKindProperties.OnPremID:
            return 'On Prem ID';
        case AzureKindProperties.OnPremSyncEnabled:
            return 'On Prem Sync Enabled';
        case AzureKindProperties.SecurityEnabled:
            return 'Security Enabled';
        case AzureKindProperties.SecurityIdentifier:
            return 'Security Identifier';
        case AzureKindProperties.EnableRBACAuthorization:
            return 'RBAC Authorization Enabled';
        case AzureKindProperties.Scope:
            return 'Scope';
        case AzureKindProperties.Offer:
            return 'Offer';
        case AzureKindProperties.MFAEnabled:
            return 'MFA Enabled';
        case AzureKindProperties.License:
            return 'License';
        case AzureKindProperties.Licenses:
            return 'Licenses';
        case AzureKindProperties.MFAEnforced:
            return 'MFA Enforced';
        case AzureKindProperties.UserPrincipalName:
            return 'User Principal Name';
        case AzureKindProperties.IsAssignableToRole:
            return 'Is Role Assignable';
        case AzureKindProperties.PublisherDomain:
            return 'Publisher Domain';
        case AzureKindProperties.SignInAudience:
            return 'Sign In Audience';
        case AzureKindProperties.RoleTemplateID:
            return 'Role Template ID';
        default:
            return undefined;
    }
}
export function AzurePathfindingEdges(): AzureRelationshipKind[] {
    return [
        AzureRelationshipKind.AvereContributor,
        AzureRelationshipKind.Contains,
        AzureRelationshipKind.Contributor,
        AzureRelationshipKind.GetCertificates,
        AzureRelationshipKind.GetKeys,
        AzureRelationshipKind.GetSecrets,
        AzureRelationshipKind.HasRole,
        AzureRelationshipKind.MemberOf,
        AzureRelationshipKind.Owner,
        AzureRelationshipKind.RunsAs,
        AzureRelationshipKind.VMContributor,
        AzureRelationshipKind.AutomationContributor,
        AzureRelationshipKind.KeyVaultContributor,
        AzureRelationshipKind.VMAdminLogin,
        AzureRelationshipKind.AddMembers,
        AzureRelationshipKind.AddSecret,
        AzureRelationshipKind.ExecuteCommand,
        AzureRelationshipKind.GlobalAdmin,
        AzureRelationshipKind.PrivilegedAuthAdmin,
        AzureRelationshipKind.Grant,
        AzureRelationshipKind.GrantSelf,
        AzureRelationshipKind.PrivilegedRoleAdmin,
        AzureRelationshipKind.ResetPassword,
        AzureRelationshipKind.UserAccessAdministrator,
        AzureRelationshipKind.Owns,
        AzureRelationshipKind.CloudAppAdmin,
        AzureRelationshipKind.AppAdmin,
        AzureRelationshipKind.AddOwner,
        AzureRelationshipKind.ManagedIdentity,
        AzureRelationshipKind.AKSContributor,
        AzureRelationshipKind.NodeResourceGroup,
        AzureRelationshipKind.WebsiteContributor,
        AzureRelationshipKind.LogicAppContributor,
        AzureRelationshipKind.AZMGAddMember,
        AzureRelationshipKind.AZMGAddOwner,
        AzureRelationshipKind.AZMGAddSecret,
        AzureRelationshipKind.AZMGGrantAppRoles,
        AzureRelationshipKind.AZMGGrantRole,
    ];
}
