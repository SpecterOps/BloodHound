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

import AZAKSContributor from './AZAKSContributor/AZAKSContributor';
import AZAddMembers from './AZAddMembers/AZAddMembers';
import AZAddOwner from './AZAddOwner/AZAddOwner';
import AZAddSecret from './AZAddSecret/AZAddSecret';
import AZAppAdmin from './AZAppAdmin/AZAppAdmin';
import AZAutomationContributor from './AZAutomationContributor/AZAutomationContributor';
import AZAvereContributor from './AZAvereContributor/AZAvereContributor';
import AZCloudAppAdmin from './AZCloudAppAdmin/AZCloudAppAdmin';
import AZContains from './AZContains/AZContains';
import AZContributor from './AZContributor/AZContributor';
import AZExecuteCommand from './AZExecuteCommand/AZExecuteCommand';
import AZGetCertificates from './AZGetCertificates/AZGetCertificates';
import AZGetKeys from './AZGetKeys/AZGetKeys';
import AZGetSecrets from './AZGetSecrets/AZGetSecrets';
import AZGlobalAdmin from './AZGlobalAdmin/AZGlobalAdmin';
import AZHasRole from './AZHasRole/AZHasRole';
import AZKeyVaultKVContributor from './AZKeyVaultKVContributor/AZKeyVaultKVContributor';
import AZLogicAppContributor from './AZLogicAppContributor/AZLogicAppContributor';
import AZMGAddMember from './AZMGAddMember/AZMGAddMember';
import AZMGAddOwner from './AZMGAddOwner/AZMGAddOwner';
import AZMGAddSecret from './AZMGAddSecret/AZMGAddSecret';
import AZMGAppRoleAssignment_ReadWrite_All from './AZMGAppRoleAssignment_ReadWrite_All/AZMGAppRoleAssignment_ReadWrite_All';
import AZMGApplication_ReadWrite_All from './AZMGApplication_ReadWrite_All/AZMGApplication_ReadWrite_All';
import AZMGDirectory_ReadWrite_All from './AZMGDirectory_ReadWrite_All/AZMGDirectory_ReadWrite_All';
import AZMGGrantAppRoles from './AZMGGrantAppRoles/AZMGGrantAppRoles';
import AZMGGrantRole from './AZMGGrantRole/AZMGGrantRole';
import AZMGGroupMember_ReadWrite_All from './AZMGGroupMember_ReadWrite_All/AZMGGroupMember_ReadWrite_All';
import AZMGGroup_ReadWrite_All from './AZMGGroup_ReadWrite_All/AZMGGroup_ReadWrite_All';
import AZMGRoleManagement_ReadWrite_Directory from './AZMGRoleManagement_ReadWrite_Directory/AZMGRoleManagement_ReadWrite_Directory';
import AZMGServicePrincipalEndpoint_ReadWrite_All from './AZMGServicePrincipalEndpoint_ReadWrite_All/AZMGServicePrincipalEndpoint_ReadWrite_All';
import AZManagedIdentity from './AZManagedIdentity/AZManagedIdentity';
import AZMemberOf from './AZMemberOf/AZMemberOf';
import AZNodeResourceGroup from './AZNodeResourceGroup/AZNodeResourceGroup';
import AZOwns from './AZOwns/AZOwns';
import AZPrivilegedAuthAdmin from './AZPrivilegedAuthAdmin/AZPrivilegedAuthAdmin';
import AZPrivilegedRoleAdmin from './AZPrivilegedRoleAdmin/AZPrivilegedRoleAdmin';
import AZResetPassword from './AZResetPassword/AZResetPassword';
import AZRunsAs from './AZRunsAs/AZRunsAs';
import AZUserAccessAdministrator from './AZUserAccessAdministrator/AZUserAccessAdministrator';
import AZVMAdminLogin from './AZVMAdminLogin/AZVMAdminLogin';
import AZVMContributor from './AZVMContributor/AZVMContributor';
import AZWebsiteContributor from './AZWebsiteContributor/AZWebsiteContributor';
import AddAllowedToAct from './AddAllowedToAct/AddAllowedToAct';
import AddKeyCredentialLink from './AddKeyCredentialLink/AddKeyCredentialLink';
import AddMember from './AddMember/AddMember';
import AddSelf from './AddSelf/AddSelf';
import AdminTo from './AdminTo/AdminTo';
import AllExtendedRights from './AllExtendedRights/AllExtendedRights';
import AllowedToAct from './AllowedToAct/AllowedToAct';
import AllowedToDelegate from './AllowedToDelegate/AllowedToDelegate';
import CanPSRemote from './CanPSRemote/CanPSRemote';
import CanRDP from './CanRDP/CanRDP';
import Contains from './Contains/Contains';
import DCSync from './DCSync/DCSync';
import DelegatedEnrollmentAgent from './DelegatedEnrollmentAgent/DelegatedEnrollmentAgent';
import DumpSMSAPassword from './DumpSMSAPassword/DumpSMSAPassword';
import Enroll from './Enroll/Enroll';
import EnterpriseCAFor from './EnterpriseCAFor/EnterpriseCAFor';
import ExecuteDCOM from './ExecuteDCOM/ExecuteDCOM';
import ForceChangePassword from './ForceChangePassword/ForceChangePassword';
import GPLink from './GPLink/GPLink';
import GenericAll from './GenericAll/GenericAll';
import GenericWrite from './GenericWrite/GenericWrite';
import GetChanges from './GetChanges/GetChanges';
import GetChangesAll from './GetChangesAll/GetChangesAll';
import HasSIDHistory from './HasSIDHistory/HasSIDHistory';
import HasSession from './HasSession/HasSession';
import HostsCAService from './HostsCAService/HostsCAService';
import IssuedSignedBy from './IssuedSignedBy/IssuedSignedBy';
import MemberOf from './MemberOf/MemberOf';
import NTAuthStoreFor from './NTAuthStoreFor/NTAuthStoreFor';
import Owns from './Owns/Owns';
import PublishedTo from './PublishedTo/PublishedTo';
import ReadGMSAPassword from './ReadGMSAPassword/ReadGMSAPassword';
import ReadLAPSPassword from './ReadLAPSPassword/ReadLAPSPassword';
import RootCAFor from './RootCAFor/RootCAFor';
import SQLAdmin from './SQLAdmin/SQLAdmin';
import SyncLAPSPassword from './SyncLAPSPassword/SyncLAPSPassword';
import TrustedBy from './TrustedBy/TrustedBy';
import TrustedForNTAuth from './TrustedForNTAuth/TrustedForNTAuth';
import WriteAccountRestrictions from './WriteAccountRestrictions/WriteAccountRestrictions';
import WriteDacl from './WriteDacl/WriteDacl';
import WriteOwner from './WriteOwner/WriteOwner';
import WriteSPN from './WriteSPN/WriteSPN';

export type EdgeInfoProps = {
    sourceName?: string;
    sourceType?: string;
    targetName?: string;
    targetType?: string;
};

const EdgeInfoComponents = {
    GenericAll: GenericAll,
    MemberOf: MemberOf,
    AllExtendedRights: AllExtendedRights,
    AdminTo: AdminTo,
    HasSession: HasSession,
    AddMember: AddMember,
    ForceChangePassword: ForceChangePassword,
    GenericWrite: GenericWrite,
    Owns: Owns,
    WriteDacl: WriteDacl,
    WriteOwner: WriteOwner,
    CanRDP: CanRDP,
    ExecuteDCOM: ExecuteDCOM,
    AllowedToDelegate: AllowedToDelegate,
    GetChanges: GetChanges,
    GetChangesAll: GetChangesAll,
    ReadLAPSPassword: ReadLAPSPassword,
    Contains: Contains,
    GPLink: GPLink,
    AddAllowedToAct: AddAllowedToAct,
    AllowedToAct: AllowedToAct,
    SQLAdmin: SQLAdmin,
    ReadGMSAPassword: ReadGMSAPassword,
    HasSIDHistory: HasSIDHistory,
    TrustedBy: TrustedBy,
    CanPSRemote: CanPSRemote,
    AZAddMembers: AZAddMembers,
    AZAddSecret: AZAddSecret,
    AZAvereContributor: AZAvereContributor,
    AZContains: AZContains,
    AZContributor: AZContributor,
    AZExecuteCommand: AZExecuteCommand,
    AZGetCertificates: AZGetCertificates,
    AZGetKeys: AZGetKeys,
    AZGetSecrets: AZGetSecrets,
    AZHasRole: AZHasRole,
    AZManagedIdentity: AZManagedIdentity,
    AZMemberOf: AZMemberOf,
    AZOwns: AZOwns,
    AZPrivilegedAuthAdmin: AZPrivilegedAuthAdmin,
    AZPrivilegedRoleAdmin: AZPrivilegedRoleAdmin,
    AZResetPassword: AZResetPassword,
    AZUserAccessAdministrator: AZUserAccessAdministrator,
    AZGlobalAdmin: AZGlobalAdmin,
    AZAppAdmin: AZAppAdmin,
    AZCloudAppAdmin: AZCloudAppAdmin,
    AZRunsAs: AZRunsAs,
    AZVMAdminLogin: AZVMAdminLogin,
    AZVMContributor: AZVMContributor,
    WriteSPN: WriteSPN,
    AddSelf: AddSelf,
    AddKeyCredentialLink: AddKeyCredentialLink,
    DCSync: DCSync,
    SyncLAPSPassword: SyncLAPSPassword,
    WriteAccountRestrictions: WriteAccountRestrictions,
    DumpSMSAPassword: DumpSMSAPassword,
    AZMGAddMember: AZMGAddMember,
    AZMGAddOwner: AZMGAddOwner,
    AZMGAddSecret: AZMGAddSecret,
    AZMGGrantAppRoles: AZMGGrantAppRoles,
    AZMGGrantRole: AZMGGrantRole,
    AZMGAppRoleAssignment_ReadWrite_All: AZMGAppRoleAssignment_ReadWrite_All,
    AZMGApplication_ReadWrite_All: AZMGApplication_ReadWrite_All,
    AZMGDirectory_ReadWrite_All: AZMGDirectory_ReadWrite_All,
    AZMGGroupMember_ReadWrite_All: AZMGGroupMember_ReadWrite_All,
    AZMGGroup_ReadWrite_All: AZMGGroup_ReadWrite_All,
    AZMGRoleManagement_ReadWrite_Directory: AZMGRoleManagement_ReadWrite_Directory,
    AZMGServicePrincipalEndpoint_ReadWrite_All: AZMGServicePrincipalEndpoint_ReadWrite_All,
    AZWebsiteContributor: AZWebsiteContributor,
    AZAddOwner: AZAddOwner,
    AZAKSContributor: AZAKSContributor,
    AZAutomationContributor: AZAutomationContributor,
    AZKeyVaultKVContributor: AZKeyVaultKVContributor,
    AZLogicAppContributor: AZLogicAppContributor,
    AZNodeResourceGroup: AZNodeResourceGroup,
    Enroll: Enroll,
    EnterpriseCAFor: EnterpriseCAFor,
    RootCAFor: RootCAFor,
    PublishedTo: PublishedTo,
    NTAuthStoreFor: NTAuthStoreFor,
    IssuedSignedBy: IssuedSignedBy,
    TrustedForNTAuth: TrustedForNTAuth,
    HostsCAService: HostsCAService,
    DelegatedEnrollmentAgent: DelegatedEnrollmentAgent,
};

export default EdgeInfoComponents;
