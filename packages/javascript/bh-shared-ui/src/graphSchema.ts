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
    AIACA = 'AIACA',
    RootCA = 'RootCA',
    EnterpriseCA = 'EnterpriseCA',
    NTAuthStore = 'NTAuthStore',
    CertTemplate = 'CertTemplate',
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
        case ActiveDirectoryNodeKind.AIACA:
            return 'AIACA';
        case ActiveDirectoryNodeKind.RootCA:
            return 'RootCA';
        case ActiveDirectoryNodeKind.EnterpriseCA:
            return 'EnterpriseCA';
        case ActiveDirectoryNodeKind.NTAuthStore:
            return 'NTAuthStore';
        case ActiveDirectoryNodeKind.CertTemplate:
            return 'CertTemplate';
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
    RootCAFor = 'RootCAFor',
    DCFor = 'DCFor',
    PublishedTo = 'PublishedTo',
    ManageCertificates = 'ManageCertificates',
    ManageCA = 'ManageCA',
    DelegatedEnrollmentAgent = 'DelegatedEnrollmentAgent',
    Enroll = 'Enroll',
    HostsCAService = 'HostsCAService',
    WritePKIEnrollmentFlag = 'WritePKIEnrollmentFlag',
    WritePKINameFlag = 'WritePKINameFlag',
    NTAuthStoreFor = 'NTAuthStoreFor',
    TrustedForNTAuth = 'TrustedForNTAuth',
    EnterpriseCAFor = 'EnterpriseCAFor',
    CanAbuseUPNCertMapping = 'CanAbuseUPNCertMapping',
    CanAbuseWeakCertBinding = 'CanAbuseWeakCertBinding',
    IssuedSignedBy = 'IssuedSignedBy',
    GoldenCert = 'GoldenCert',
    EnrollOnBehalfOf = 'EnrollOnBehalfOf',
    ADCSESC1 = 'ADCSESC1',
    ADCSESC3 = 'ADCSESC3',
    ADCSESC4 = 'ADCSESC4',
    ADCSESC5 = 'ADCSESC5',
    ADCSESC6a = 'ADCSESC6a',
    ADCSESC6b = 'ADCSESC6b',
    ADCSESC7 = 'ADCSESC7',
    ADCSESC9a = 'ADCSESC9a',
    ADCSESC9b = 'ADCSESC9b',
    ADCSESC10a = 'ADCSESC10a',
    ADCSESC10b = 'ADCSESC10b',
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
        case ActiveDirectoryRelationshipKind.RootCAFor:
            return 'RootCAFor';
        case ActiveDirectoryRelationshipKind.DCFor:
            return 'DCFor';
        case ActiveDirectoryRelationshipKind.PublishedTo:
            return 'PublishedTo';
        case ActiveDirectoryRelationshipKind.ManageCertificates:
            return 'ManageCertificates';
        case ActiveDirectoryRelationshipKind.ManageCA:
            return 'ManageCA';
        case ActiveDirectoryRelationshipKind.DelegatedEnrollmentAgent:
            return 'DelegatedEnrollmentAgent';
        case ActiveDirectoryRelationshipKind.Enroll:
            return 'Enroll';
        case ActiveDirectoryRelationshipKind.HostsCAService:
            return 'HostsCAService';
        case ActiveDirectoryRelationshipKind.WritePKIEnrollmentFlag:
            return 'WritePKIEnrollmentFlag';
        case ActiveDirectoryRelationshipKind.WritePKINameFlag:
            return 'WritePKINameFlag';
        case ActiveDirectoryRelationshipKind.NTAuthStoreFor:
            return 'NTAuthStoreFor';
        case ActiveDirectoryRelationshipKind.TrustedForNTAuth:
            return 'TrustedForNTAuth';
        case ActiveDirectoryRelationshipKind.EnterpriseCAFor:
            return 'EnterpriseCAFor';
        case ActiveDirectoryRelationshipKind.CanAbuseUPNCertMapping:
            return 'CanAbuseUPNCertMapping';
        case ActiveDirectoryRelationshipKind.CanAbuseWeakCertBinding:
            return 'CanAbuseWeakCertBinding';
        case ActiveDirectoryRelationshipKind.IssuedSignedBy:
            return 'IssuedSignedBy';
        case ActiveDirectoryRelationshipKind.GoldenCert:
            return 'GoldenCert';
        case ActiveDirectoryRelationshipKind.EnrollOnBehalfOf:
            return 'EnrollOnBehalfOf';
        case ActiveDirectoryRelationshipKind.ADCSESC1:
            return 'ADCSESC1';
        case ActiveDirectoryRelationshipKind.ADCSESC3:
            return 'ADCSESC3';
        case ActiveDirectoryRelationshipKind.ADCSESC4:
            return 'ADCSESC4';
        case ActiveDirectoryRelationshipKind.ADCSESC5:
            return 'ADCSESC5';
        case ActiveDirectoryRelationshipKind.ADCSESC6a:
            return 'ADCSESC6a';
        case ActiveDirectoryRelationshipKind.ADCSESC6b:
            return 'ADCSESC6b';
        case ActiveDirectoryRelationshipKind.ADCSESC7:
            return 'ADCSESC7';
        case ActiveDirectoryRelationshipKind.ADCSESC9a:
            return 'ADCSESC9a';
        case ActiveDirectoryRelationshipKind.ADCSESC9b:
            return 'ADCSESC9b';
        case ActiveDirectoryRelationshipKind.ADCSESC10a:
            return 'ADCSESC10a';
        case ActiveDirectoryRelationshipKind.ADCSESC10b:
            return 'ADCSESC10b';
        default:
            return undefined;
    }
}
export type ActiveDirectoryKind = ActiveDirectoryNodeKind | ActiveDirectoryRelationshipKind;
export const EdgeCompositionRelationships = [
    'GoldenCert',
    'ADCSESC1',
    'ADCSESC3',
    'ADCSESC6a',
    'ADCSESC6b',
    'ADCSESC9a',
    'ADCSESC9b',
    'ADCSESC10a',
    'ADCSESC10b',
];
export enum ActiveDirectoryKindProperties {
    AdminCount = 'admincount',
    CASecurityCollected = 'casecuritycollected',
    CAName = 'caname',
    CertChain = 'certchain',
    CertName = 'certname',
    CertThumbprint = 'certthumbprint',
    CertThumbprints = 'certthumbprints',
    HasEnrollmentAgentRestrictions = 'hasenrollmentagentrestrictions',
    EnrollmentAgentRestrictionsCollected = 'enrollmentagentrestrictionscollected',
    IsUserSpecifiesSanEnabled = 'isuserspecifiessanenabled',
    IsUserSpecifiesSanEnabledCollected = 'isuserspecifiessanenabledcollected',
    HasBasicConstraints = 'hasbasicconstraints',
    BasicConstraintPathLength = 'basicconstraintpathlength',
    DNSHostname = 'dnshostname',
    CrossCertificatePair = 'crosscertificatepair',
    DistinguishedName = 'distinguishedname',
    DomainFQDN = 'domain',
    DomainSID = 'domainsid',
    Sensitive = 'sensitive',
    HighValue = 'highvalue',
    BlocksInheritance = 'blocksinheritance',
    IsACL = 'isacl',
    IsACLProtected = 'isaclprotected',
    IsDeleted = 'isdeleted',
    Enforced = 'enforced',
    Department = 'department',
    HasCrossCertificatePair = 'hascrosscertificatepair',
    HasSPN = 'hasspn',
    UnconstrainedDelegation = 'unconstraineddelegation',
    LastLogon = 'lastlogon',
    LastLogonTimestamp = 'lastlogontimestamp',
    IsPrimaryGroup = 'isprimarygroup',
    HasLAPS = 'haslaps',
    DontRequirePreAuth = 'dontreqpreauth',
    LogonType = 'logontype',
    HasURA = 'hasura',
    PasswordNeverExpires = 'pwdneverexpires',
    PasswordNotRequired = 'passwordnotreqd',
    FunctionalLevel = 'functionallevel',
    TrustType = 'trusttype',
    SidFiltering = 'sidfiltering',
    TrustedToAuth = 'trustedtoauth',
    SamAccountName = 'samaccountname',
    CertificateMappingMethodsRaw = 'certificatemappingmethodsraw',
    CertificateMappingMethods = 'certificatemappingmethods',
    StrongCertificateBindingEnforcementRaw = 'strongcertificatebindingenforcementraw',
    StrongCertificateBindingEnforcement = 'strongcertificatebindingenforcement',
    EKUs = 'ekus',
    SubjectAltRequireUPN = 'subjectaltrequireupn',
    SubjectAltRequireDNS = 'subjectaltrequiredns',
    SubjectAltRequireDomainDNS = 'subjectaltrequiredomaindns',
    SubjectAltRequireEmail = 'subjectaltrequireemail',
    SubjectAltRequireSPN = 'subjectaltrequirespn',
    SubjectRequireEmail = 'subjectrequireemail',
    AuthorizedSignatures = 'authorizedsignatures',
    ApplicationPolicies = 'applicationpolicies',
    IssuancePolicies = 'issuancepolicies',
    SchemaVersion = 'schemaversion',
    RequiresManagerApproval = 'requiresmanagerapproval',
    AuthenticationEnabled = 'authenticationenabled',
    EnrolleeSuppliesSubject = 'enrolleesuppliessubject',
    CertificateApplicationPolicy = 'certificateapplicationpolicy',
    CertificateNameFlag = 'certificatenameflag',
    EffectiveEKUs = 'effectiveekus',
    EnrollmentFlag = 'enrollmentflag',
    Flags = 'flags',
    NoSecurityExtension = 'nosecurityextension',
    RenewalPeriod = 'renewalperiod',
    ValidityPeriod = 'validityperiod',
    OID = 'oid',
    HomeDirectory = 'homedirectory',
}
export function ActiveDirectoryKindPropertiesToDisplay(value: ActiveDirectoryKindProperties): string | undefined {
    switch (value) {
        case ActiveDirectoryKindProperties.AdminCount:
            return 'Admin Count';
        case ActiveDirectoryKindProperties.CASecurityCollected:
            return 'CA Security Collected';
        case ActiveDirectoryKindProperties.CAName:
            return 'CA Name';
        case ActiveDirectoryKindProperties.CertChain:
            return 'Certificate Chain';
        case ActiveDirectoryKindProperties.CertName:
            return 'Certificate Name';
        case ActiveDirectoryKindProperties.CertThumbprint:
            return 'Certificate Thumbprint';
        case ActiveDirectoryKindProperties.CertThumbprints:
            return 'Certificate Thumbprints';
        case ActiveDirectoryKindProperties.HasEnrollmentAgentRestrictions:
            return 'Has Enrollment Agent Restrictions';
        case ActiveDirectoryKindProperties.EnrollmentAgentRestrictionsCollected:
            return 'Enrollment Agent Restrictions Collected';
        case ActiveDirectoryKindProperties.IsUserSpecifiesSanEnabled:
            return 'Is User Specifies San Enabled';
        case ActiveDirectoryKindProperties.IsUserSpecifiesSanEnabledCollected:
            return 'Is User Specifies San Enabled Collected';
        case ActiveDirectoryKindProperties.HasBasicConstraints:
            return 'Has Basic Constraints';
        case ActiveDirectoryKindProperties.BasicConstraintPathLength:
            return 'Basic Constraint Path Length';
        case ActiveDirectoryKindProperties.DNSHostname:
            return 'DNS Hostname';
        case ActiveDirectoryKindProperties.CrossCertificatePair:
            return 'Cross Certificate Pair';
        case ActiveDirectoryKindProperties.DistinguishedName:
            return 'Distinguished Name';
        case ActiveDirectoryKindProperties.DomainFQDN:
            return 'Domain FQDN';
        case ActiveDirectoryKindProperties.DomainSID:
            return 'Domain SID';
        case ActiveDirectoryKindProperties.Sensitive:
            return 'Marked Sensitive';
        case ActiveDirectoryKindProperties.HighValue:
            return 'High Value';
        case ActiveDirectoryKindProperties.BlocksInheritance:
            return 'Blocks Inheritance';
        case ActiveDirectoryKindProperties.IsACL:
            return 'Is ACL';
        case ActiveDirectoryKindProperties.IsACLProtected:
            return 'ACL Inheritance Denied';
        case ActiveDirectoryKindProperties.IsDeleted:
            return 'Is Deleted';
        case ActiveDirectoryKindProperties.Enforced:
            return 'Enforced';
        case ActiveDirectoryKindProperties.Department:
            return 'Department';
        case ActiveDirectoryKindProperties.HasCrossCertificatePair:
            return 'Has Cross Certificate Pair';
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
        case ActiveDirectoryKindProperties.PasswordNeverExpires:
            return 'Password Never Expires';
        case ActiveDirectoryKindProperties.PasswordNotRequired:
            return 'Password Not Required';
        case ActiveDirectoryKindProperties.FunctionalLevel:
            return 'Functional Level';
        case ActiveDirectoryKindProperties.TrustType:
            return 'Trust Type';
        case ActiveDirectoryKindProperties.SidFiltering:
            return 'SID Filtering Enabled';
        case ActiveDirectoryKindProperties.TrustedToAuth:
            return 'Trusted For Constrained Delegation';
        case ActiveDirectoryKindProperties.SamAccountName:
            return 'SAM Account Name';
        case ActiveDirectoryKindProperties.CertificateMappingMethodsRaw:
            return 'Certificate Mapping Methods (Raw)';
        case ActiveDirectoryKindProperties.CertificateMappingMethods:
            return 'Certificate Mapping Methods';
        case ActiveDirectoryKindProperties.StrongCertificateBindingEnforcementRaw:
            return 'Strong Certificate Binding Enforcement (Raw)';
        case ActiveDirectoryKindProperties.StrongCertificateBindingEnforcement:
            return 'Strong Certificate Binding Enforcement';
        case ActiveDirectoryKindProperties.EKUs:
            return 'Enhanced Key Usage';
        case ActiveDirectoryKindProperties.SubjectAltRequireUPN:
            return 'Subject Alternative Name Require UPN';
        case ActiveDirectoryKindProperties.SubjectAltRequireDNS:
            return 'Subject Alternative Name Require DNS';
        case ActiveDirectoryKindProperties.SubjectAltRequireDomainDNS:
            return 'Subject Alternative Name Require Domain DNS';
        case ActiveDirectoryKindProperties.SubjectAltRequireEmail:
            return 'Subject Alternative Name Require Email';
        case ActiveDirectoryKindProperties.SubjectAltRequireSPN:
            return 'Subject Alternative Name Require SPN';
        case ActiveDirectoryKindProperties.SubjectRequireEmail:
            return 'Subject Require Email';
        case ActiveDirectoryKindProperties.AuthorizedSignatures:
            return 'Authorized Signatures Required';
        case ActiveDirectoryKindProperties.ApplicationPolicies:
            return 'Application Policies';
        case ActiveDirectoryKindProperties.IssuancePolicies:
            return 'Issuance Policies';
        case ActiveDirectoryKindProperties.SchemaVersion:
            return 'Schema Version';
        case ActiveDirectoryKindProperties.RequiresManagerApproval:
            return 'Requires Manager Approval';
        case ActiveDirectoryKindProperties.AuthenticationEnabled:
            return 'Authentication Enabled';
        case ActiveDirectoryKindProperties.EnrolleeSuppliesSubject:
            return 'Enrollee Supplies Subject';
        case ActiveDirectoryKindProperties.CertificateApplicationPolicy:
            return 'Certificate Application Policies';
        case ActiveDirectoryKindProperties.CertificateNameFlag:
            return 'Certificate Name Flags';
        case ActiveDirectoryKindProperties.EffectiveEKUs:
            return 'Effective EKUs';
        case ActiveDirectoryKindProperties.EnrollmentFlag:
            return 'Enrollment Flags';
        case ActiveDirectoryKindProperties.Flags:
            return 'Flags';
        case ActiveDirectoryKindProperties.NoSecurityExtension:
            return 'No Security Extension';
        case ActiveDirectoryKindProperties.RenewalPeriod:
            return 'Renewal Period';
        case ActiveDirectoryKindProperties.ValidityPeriod:
            return 'Validity Period';
        case ActiveDirectoryKindProperties.OID:
            return 'OID';
        case ActiveDirectoryKindProperties.HomeDirectory:
            return 'Home Directory';
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
        ActiveDirectoryRelationshipKind.GoldenCert,
        ActiveDirectoryRelationshipKind.ADCSESC1,
        ActiveDirectoryRelationshipKind.ADCSESC3,
        ActiveDirectoryRelationshipKind.ADCSESC4,
        ActiveDirectoryRelationshipKind.ADCSESC5,
        ActiveDirectoryRelationshipKind.ADCSESC6a,
        ActiveDirectoryRelationshipKind.ADCSESC6b,
        ActiveDirectoryRelationshipKind.ADCSESC7,
        ActiveDirectoryRelationshipKind.ADCSESC9a,
        ActiveDirectoryRelationshipKind.ADCSESC9b,
        ActiveDirectoryRelationshipKind.ADCSESC10a,
        ActiveDirectoryRelationshipKind.ADCSESC10b,
        ActiveDirectoryRelationshipKind.DCFor,
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
export type AzureKind = AzureNodeKind | AzureRelationshipKind;
export enum AzureKindProperties {
    AppOwnerOrganizationID = 'appownerorganizationid',
    AppDescription = 'appdescription',
    AppDisplayName = 'appdisplayname',
    ServicePrincipalType = 'serviceprincipaltype',
    UserType = 'usertype',
    TenantID = 'tenantid',
    ServicePrincipalID = 'service_principal_id',
    ServicePrincipalNames = 'service_principal_names',
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
        case AzureKindProperties.ServicePrincipalNames:
            return 'Service Principal Names';
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
export enum CommonNodeKind {
    MigrationData = 'MigrationData',
}
export function CommonNodeKindToDisplay(value: CommonNodeKind): string | undefined {
    switch (value) {
        case CommonNodeKind.MigrationData:
            return 'MigrationData';
        default:
            return undefined;
    }
}
export enum CommonKindProperties {
    ObjectID = 'objectid',
    Name = 'name',
    DisplayName = 'displayname',
    Description = 'description',
    OwnerObjectID = 'owner_objectid',
    Collected = 'collected',
    OperatingSystem = 'operatingsystem',
    SystemTags = 'system_tags',
    UserTags = 'user_tags',
    LastSeen = 'lastseen',
    WhenCreated = 'whencreated',
    Enabled = 'enabled',
    PasswordLastSet = 'pwdlastset',
    Title = 'title',
    Email = 'email',
    IsInherited = 'isinherited',
}
export function CommonKindPropertiesToDisplay(value: CommonKindProperties): string | undefined {
    switch (value) {
        case CommonKindProperties.ObjectID:
            return 'Object ID';
        case CommonKindProperties.Name:
            return 'Name';
        case CommonKindProperties.DisplayName:
            return 'Display Name';
        case CommonKindProperties.Description:
            return 'Description';
        case CommonKindProperties.OwnerObjectID:
            return 'Owner Object ID';
        case CommonKindProperties.Collected:
            return 'Collected';
        case CommonKindProperties.OperatingSystem:
            return 'Operating System';
        case CommonKindProperties.SystemTags:
            return 'Node System Tags';
        case CommonKindProperties.UserTags:
            return 'Node User Tags';
        case CommonKindProperties.LastSeen:
            return 'Last Collected by BloodHound';
        case CommonKindProperties.WhenCreated:
            return 'Created';
        case CommonKindProperties.Enabled:
            return 'Enabled';
        case CommonKindProperties.PasswordLastSet:
            return 'Password Last Set';
        case CommonKindProperties.Title:
            return 'Title';
        case CommonKindProperties.Email:
            return 'Email';
        case CommonKindProperties.IsInherited:
            return 'Is Inherited';
        default:
            return undefined;
    }
}
