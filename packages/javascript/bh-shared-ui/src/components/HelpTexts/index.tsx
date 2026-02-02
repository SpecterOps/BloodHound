// Copyright 2024 Specter Ops, Inc.
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

import { NormalizedNodeItem } from '../VirtualizedNodeList';
import ADCSESC1 from './ADCSESC1/ADCSESC1';
import ADCSESC10a from './ADCSESC10a/ADCSESC10a';
import ADCSESC10b from './ADCSESC10b/ADCSESC10b';
import ADCSESC13 from './ADCSESC13/ADCSESC13';
import ADCSESC3 from './ADCSESC3/ADCSESC3';
import ADCSESC4 from './ADCSESC4/ADCSESC4';
import ADCSESC6a from './ADCSESC6a/ADCSESC6a';
import ADCSESC6b from './ADCSESC6b/ADCSESC6b';
import ADCSESC9a from './ADCSESC9a/ADCSESC9a';
import ADCSESC9b from './ADCSESC9b/ADCSESC9b';
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
import AZOwner from './AZOwner/AZOwner';
import AZOwns from './AZOwns/AZOwns';
import AZPrivilegedAuthAdmin from './AZPrivilegedAuthAdmin/AZPrivilegedAuthAdmin';
import AZPrivilegedRoleAdmin from './AZPrivilegedRoleAdmin/AZPrivilegedRoleAdmin';
import AZResetPassword from './AZResetPassword/AZResetPassword';
import AZRoleApprover from './AZRoleApprover/AZRoleApprover';
import AZRoleEligible from './AZRoleEligible/AZRoleEligible';
import AZRunsAs from './AZRunsAs/AZRunsAs';
import AZUserAccessAdministrator from './AZUserAccessAdministrator/AZUserAccessAdministrator';
import AZVMAdminLogin from './AZVMAdminLogin/AZVMAdminLogin';
import AZVMContributor from './AZVMContributor/AZVMContributor';
import AZWebsiteContributor from './AZWebsiteContributor/AZWebsiteContributor';
import AbuseTGTDelegation from './AbuseTGTDelegation/AbuseTGTDelegation';
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
import ClaimSpecialIdentity from './ClaimSpecialIdentity/ClaimSpecialIdentity';
import CoerceAndRelayNTLMToADCS from './CoerceAndRelayNTLMToADCS/CoerceAndRelayNTLMToADCS';
import CoerceAndRelayNTLMToLDAP from './CoerceAndRelayNTLMToLDAP/CoerceAndRelayNTLMToLDAP';
import CoerceAndRelayNTLMToLDAPS from './CoerceAndRelayNTLMToLDAPS/CoerceAndRelayNTLMToLDAPS';
import CoerceAndRelayNTLMToSMB from './CoerceAndRelayNTLMToSMB/CoerceAndRelayNTLMToSMB';
import CoerceToTGT from './CoerceToTGT/CoerceToTGT';
import Contains from './Contains/Contains';
import CrossForestTrust from './CrossForestTrust/CrossForestTrust';
import DCFor from './DCFor/DCFor';
import DCSync from './DCSync/DCSync';
import DelegatedEnrollmentAgent from './DelegatedEnrollmentAgent/DelegatedEnrollmentAgent';
import DumpSMSAPassword from './DumpSMSAPassword/DumpSMSAPassword';
import Enroll from './Enroll/Enroll';
import EnrollOnBehalfOf from './EnrollOnBehalfOf/EnrollOnBehalfOf';
import EnterpriseCAFor from './EnterpriseCAFor/EnterpriseCAFor';
import ExecuteDCOM from './ExecuteDCOM/ExecuteDCOM';
import ExtendedByPolicy from './ExtendedByPolicy/ExtendedByPolicy';
import ForceChangePassword from './ForceChangePassword/ForceChangePassword';
import GPLink from './GPLink/GPLink';
import GenericAll from './GenericAll/GenericAll';
import GenericWrite from './GenericWrite/GenericWrite';
import GetChanges from './GetChanges/GetChanges';
import GetChangesAll from './GetChangesAll/GetChangesAll';
import GoldenCert from './GoldenCert/GoldenCert';
import HasSIDHistory from './HasSIDHistory/HasSIDHistory';
import HasSession from './HasSession/HasSession';
import HasTrustKeys from './HasTrustKeys/HasTrustKeys';
import HostsCAService from './HostsCAService/HostsCAService';
import IssuedSignedBy from './IssuedSignedBy/IssuedSignedBy';
import ManageCA from './ManageCA/ManageCA';
import ManageCertificates from './ManageCertificates/ManageCertificates';
import MemberOf from './MemberOf/MemberOf';
import NTAuthStoreFor from './NTAuthStoreFor/NTAuthStoreFor';
import OIDGroupLink from './OIDGroupLink/OIDGroupLink';
import Owns from './Owns/Owns';
import OwnsLimitedRights from './OwnsLimitedRights/OwnsLimitedRights';
import OwnsRaw from './OwnsRaw/OwnsRaw';
import ProtectAdminGroups from './ProtectAdminGroups/ProtectAdminGroups';
import PublishedTo from './PublishedTo/PublishedTo';
import ReadGMSAPassword from './ReadGMSAPassword/ReadGMSAPassword';
import ReadLAPSPassword from './ReadLAPSPassword/ReadLAPSPassword';
import RootCAFor from './RootCAFor/RootCAFor';
import SQLAdmin from './SQLAdmin/SQLAdmin';
import SameForestTrust from './SameForestTrust/SameForestTrust';
import SpoofSIDHistory from './SpoofSIDHistory/SpoofSIDHistory';
import SyncLAPSPassword from './SyncLAPSPassword/SyncLAPSPassword';
import SyncedToADUser from './SyncedToADUser/SyncedToADUser';
import SyncedToEntraUser from './SyncedToEntraUser/SyncedToEntraUser';
import TrustedForNTAuth from './TrustedForNTAuth/TrustedForNTAuth';
import WriteAccountRestrictions from './WriteAccountRestrictions/WriteAccountRestrictions';
import WriteDacl from './WriteDacl/WriteDacl';
import WriteGPLink from './WriteGPLink/WriteGPLink';
import WriteOwner from './WriteOwner/WriteOwner';
import WriteOwnerLimitedRights from './WriteOwnerLimitedRights/WriteOwnerLimitedRights';
import WriteOwnerRaw from './WriteOwnerRaw/WriteOwnerRaw';
import WritePKIEnrollmentFlag from './WritePKIEnrollmentFlag/WritePKIEnrollmentFlag';
import WritePKINameFlag from './WritePKINameFlag/WritePKINameFlag';
import WriteSPN from './WriteSPN/WriteSPN';

export type EdgeInfoProps = {
    edgeName?: string;
    sourceDBId?: number;
    sourceName?: string;
    sourceType?: string;
    targetDBId?: number;
    targetName?: string;
    targetType?: string;
    onNodeClick?: (selectedNode: NormalizedNodeItem) => void;
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
    OwnsLimitedRights: OwnsLimitedRights,
    OwnsRaw: OwnsRaw,
    WriteDacl: WriteDacl,
    WriteOwner: WriteOwner,
    WriteOwnerLimitedRights: WriteOwnerLimitedRights,
    WriteOwnerRaw: WriteOwnerRaw,
    CanRDP: CanRDP,
    ExecuteDCOM: ExecuteDCOM,
    AllowedToDelegate: AllowedToDelegate,
    CoerceToTGT: CoerceToTGT,
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
    CrossForestTrust: CrossForestTrust,
    SameForestTrust: SameForestTrust,
    SpoofSIDHistory: SpoofSIDHistory,
    AbuseTGTDelegation: AbuseTGTDelegation,
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
    AZOwner: AZOwner,
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
    WriteGPLink: WriteGPLink,
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
    AZRoleApprover: AZRoleApprover,
    AZRoleEligible: AZRoleEligible,
    Enroll: Enroll,
    EnterpriseCAFor: EnterpriseCAFor,
    RootCAFor: RootCAFor,
    PublishedTo: PublishedTo,
    NTAuthStoreFor: NTAuthStoreFor,
    IssuedSignedBy: IssuedSignedBy,
    TrustedForNTAuth: TrustedForNTAuth,
    HostsCAService: HostsCAService,
    DelegatedEnrollmentAgent: DelegatedEnrollmentAgent,
    EnrollOnBehalfOf: EnrollOnBehalfOf,
    GoldenCert: GoldenCert,
    ADCSESC1: ADCSESC1,
    ADCSESC4: ADCSESC4,
    ADCSESC3: ADCSESC3,
    ADCSESC6a: ADCSESC6a,
    ADCSESC6b: ADCSESC6b,
    ADCSESC9a: ADCSESC9a,
    ADCSESC9b: ADCSESC9b,
    ADCSESC10a: ADCSESC10a,
    ADCSESC10b: ADCSESC10b,
    ADCSESC13: ADCSESC13,
    ManageCA: ManageCA,
    ManageCertificates: ManageCertificates,
    WritePKIEnrollmentFlag: WritePKIEnrollmentFlag,
    WritePKINameFlag: WritePKINameFlag,
    DCFor: DCFor,
    OIDGroupLink: OIDGroupLink,
    ExtendedByPolicy: ExtendedByPolicy,
    SyncedToADUser: SyncedToADUser,
    SyncedToEntraUser: SyncedToEntraUser,
    CoerceAndRelayNTLMToSMB: CoerceAndRelayNTLMToSMB,
    CoerceAndRelayNTLMToLDAP: CoerceAndRelayNTLMToLDAP,
    CoerceAndRelayNTLMToLDAPS: CoerceAndRelayNTLMToLDAPS,
    CoerceAndRelayNTLMToADCS: CoerceAndRelayNTLMToADCS,
    ProtectAdminGroups: ProtectAdminGroups,
    ClaimSpecialIdentity: ClaimSpecialIdentity,
    HasTrustKeys: HasTrustKeys,
};

export default EdgeInfoComponents;
