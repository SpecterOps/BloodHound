// Copyright 2025 Specter Ops, Inc.
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
    IssuancePolicy = 'IssuancePolicy',
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
        case ActiveDirectoryNodeKind.IssuancePolicy:
            return 'IssuancePolicy';
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
    CoerceToTGT = 'CoerceToTGT',
    GetChanges = 'GetChanges',
    GetChangesAll = 'GetChangesAll',
    GetChangesInFilteredSet = 'GetChangesInFilteredSet',
    CrossForestTrust = 'CrossForestTrust',
    SameForestTrust = 'SameForestTrust',
    SpoofSIDHistory = 'SpoofSIDHistory',
    AbuseTGTDelegation = 'AbuseTGTDelegation',
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
    RemoteInteractiveLogonRight = 'RemoteInteractiveLogonRight',
    SyncLAPSPassword = 'SyncLAPSPassword',
    WriteAccountRestrictions = 'WriteAccountRestrictions',
    WriteGPLink = 'WriteGPLink',
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
    IssuedSignedBy = 'IssuedSignedBy',
    GoldenCert = 'GoldenCert',
    EnrollOnBehalfOf = 'EnrollOnBehalfOf',
    OIDGroupLink = 'OIDGroupLink',
    ExtendedByPolicy = 'ExtendedByPolicy',
    ADCSESC1 = 'ADCSESC1',
    ADCSESC3 = 'ADCSESC3',
    ADCSESC4 = 'ADCSESC4',
    ADCSESC6a = 'ADCSESC6a',
    ADCSESC6b = 'ADCSESC6b',
    ADCSESC9a = 'ADCSESC9a',
    ADCSESC9b = 'ADCSESC9b',
    ADCSESC10a = 'ADCSESC10a',
    ADCSESC10b = 'ADCSESC10b',
    ADCSESC13 = 'ADCSESC13',
    SyncedToEntraUser = 'SyncedToEntraUser',
    CoerceAndRelayNTLMToSMB = 'CoerceAndRelayNTLMToSMB',
    CoerceAndRelayNTLMToADCS = 'CoerceAndRelayNTLMToADCS',
    WriteOwnerLimitedRights = 'WriteOwnerLimitedRights',
    WriteOwnerRaw = 'WriteOwnerRaw',
    OwnsLimitedRights = 'OwnsLimitedRights',
    OwnsRaw = 'OwnsRaw',
    ClaimSpecialIdentity = 'ClaimSpecialIdentity',
    CoerceAndRelayNTLMToLDAP = 'CoerceAndRelayNTLMToLDAP',
    CoerceAndRelayNTLMToLDAPS = 'CoerceAndRelayNTLMToLDAPS',
    ContainsIdentity = 'ContainsIdentity',
    PropagatesACEsTo = 'PropagatesACEsTo',
    GPOAppliesTo = 'GPOAppliesTo',
    CanApplyGPO = 'CanApplyGPO',
    HasTrustKeys = 'HasTrustKeys',
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
        case ActiveDirectoryRelationshipKind.CoerceToTGT:
            return 'CoerceToTGT';
        case ActiveDirectoryRelationshipKind.GetChanges:
            return 'GetChanges';
        case ActiveDirectoryRelationshipKind.GetChangesAll:
            return 'GetChangesAll';
        case ActiveDirectoryRelationshipKind.GetChangesInFilteredSet:
            return 'GetChangesInFilteredSet';
        case ActiveDirectoryRelationshipKind.CrossForestTrust:
            return 'CrossForestTrust';
        case ActiveDirectoryRelationshipKind.SameForestTrust:
            return 'SameForestTrust';
        case ActiveDirectoryRelationshipKind.SpoofSIDHistory:
            return 'SpoofSIDHistory';
        case ActiveDirectoryRelationshipKind.AbuseTGTDelegation:
            return 'AbuseTGTDelegation';
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
        case ActiveDirectoryRelationshipKind.RemoteInteractiveLogonRight:
            return 'RemoteInteractiveLogonRight';
        case ActiveDirectoryRelationshipKind.SyncLAPSPassword:
            return 'SyncLAPSPassword';
        case ActiveDirectoryRelationshipKind.WriteAccountRestrictions:
            return 'WriteAccountRestrictions';
        case ActiveDirectoryRelationshipKind.WriteGPLink:
            return 'WriteGPLink';
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
        case ActiveDirectoryRelationshipKind.IssuedSignedBy:
            return 'IssuedSignedBy';
        case ActiveDirectoryRelationshipKind.GoldenCert:
            return 'GoldenCert';
        case ActiveDirectoryRelationshipKind.EnrollOnBehalfOf:
            return 'EnrollOnBehalfOf';
        case ActiveDirectoryRelationshipKind.OIDGroupLink:
            return 'OIDGroupLink';
        case ActiveDirectoryRelationshipKind.ExtendedByPolicy:
            return 'ExtendedByPolicy';
        case ActiveDirectoryRelationshipKind.ADCSESC1:
            return 'ADCSESC1';
        case ActiveDirectoryRelationshipKind.ADCSESC3:
            return 'ADCSESC3';
        case ActiveDirectoryRelationshipKind.ADCSESC4:
            return 'ADCSESC4';
        case ActiveDirectoryRelationshipKind.ADCSESC6a:
            return 'ADCSESC6a';
        case ActiveDirectoryRelationshipKind.ADCSESC6b:
            return 'ADCSESC6b';
        case ActiveDirectoryRelationshipKind.ADCSESC9a:
            return 'ADCSESC9a';
        case ActiveDirectoryRelationshipKind.ADCSESC9b:
            return 'ADCSESC9b';
        case ActiveDirectoryRelationshipKind.ADCSESC10a:
            return 'ADCSESC10a';
        case ActiveDirectoryRelationshipKind.ADCSESC10b:
            return 'ADCSESC10b';
        case ActiveDirectoryRelationshipKind.ADCSESC13:
            return 'ADCSESC13';
        case ActiveDirectoryRelationshipKind.SyncedToEntraUser:
            return 'SyncedToEntraUser';
        case ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToSMB:
            return 'CoerceAndRelayNTLMToSMB';
        case ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToADCS:
            return 'CoerceAndRelayNTLMToADCS';
        case ActiveDirectoryRelationshipKind.WriteOwnerLimitedRights:
            return 'WriteOwnerLimitedRights';
        case ActiveDirectoryRelationshipKind.WriteOwnerRaw:
            return 'WriteOwnerRaw';
        case ActiveDirectoryRelationshipKind.OwnsLimitedRights:
            return 'OwnsLimitedRights';
        case ActiveDirectoryRelationshipKind.OwnsRaw:
            return 'OwnsRaw';
        case ActiveDirectoryRelationshipKind.ClaimSpecialIdentity:
            return 'ClaimSpecialIdentity';
        case ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToLDAP:
            return 'CoerceAndRelayNTLMToLDAP';
        case ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToLDAPS:
            return 'CoerceAndRelayNTLMToLDAPS';
        case ActiveDirectoryRelationshipKind.ContainsIdentity:
            return 'ContainsIdentity';
        case ActiveDirectoryRelationshipKind.PropagatesACEsTo:
            return 'PropagatesACEsTo';
        case ActiveDirectoryRelationshipKind.GPOAppliesTo:
            return 'GPOAppliesTo';
        case ActiveDirectoryRelationshipKind.CanApplyGPO:
            return 'CanApplyGPO';
        case ActiveDirectoryRelationshipKind.HasTrustKeys:
            return 'HasTrustKeys';
        default:
            return undefined;
    }
}
export type ActiveDirectoryKind = ActiveDirectoryNodeKind | ActiveDirectoryRelationshipKind;
export const EdgeCompositionRelationships = [
    'GoldenCert',
    'ADCSESC1',
    'ADCSESC3',
    'ADCSESC4',
    'ADCSESC6a',
    'ADCSESC6b',
    'ADCSESC9a',
    'ADCSESC9b',
    'ADCSESC10a',
    'ADCSESC10b',
    'ADCSESC13',
    'CoerceAndRelayNTLMToSMB',
    'CoerceAndRelayNTLMToADCS',
    'CoerceAndRelayNTLMToLDAP',
    'CoerceAndRelayNTLMToLDAPS',
    'GPOAppliesTo',
    'CanApplyGPO',
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
    RoleSeparationEnabled = 'roleseparationenabled',
    RoleSeparationEnabledCollected = 'roleseparationenabledcollected',
    HasBasicConstraints = 'hasbasicconstraints',
    BasicConstraintPathLength = 'basicconstraintpathlength',
    UnresolvedPublishedTemplates = 'unresolvedpublishedtemplates',
    DNSHostname = 'dnshostname',
    CrossCertificatePair = 'crosscertificatepair',
    DistinguishedName = 'distinguishedname',
    DomainFQDN = 'domain',
    DomainSID = 'domainsid',
    Sensitive = 'sensitive',
    BlocksInheritance = 'blocksinheritance',
    IsACL = 'isacl',
    IsACLProtected = 'isaclprotected',
    InheritanceHash = 'inheritancehash',
    InheritanceHashes = 'inheritancehashes',
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
    SpoofSIDHistoryBlocked = 'spoofsidhistoryblocked',
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
    SchannelAuthenticationEnabled = 'schannelauthenticationenabled',
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
    CertificatePolicy = 'certificatepolicy',
    CertTemplateOID = 'certtemplateoid',
    GroupLinkID = 'grouplinkid',
    ObjectGUID = 'objectguid',
    ExpirePasswordsOnSmartCardOnlyAccounts = 'expirepasswordsonsmartcardonlyaccounts',
    MachineAccountQuota = 'machineaccountquota',
    SupportedKerberosEncryptionTypes = 'supportedencryptiontypes',
    TGTDelegation = 'tgtdelegation',
    PasswordStoredUsingReversibleEncryption = 'encryptedtextpwdallowed',
    SmartcardRequired = 'smartcardrequired',
    UseDESKeyOnly = 'usedeskeyonly',
    LogonScriptEnabled = 'logonscriptenabled',
    LockedOut = 'lockedout',
    UserCannotChangePassword = 'passwordcantchange',
    PasswordExpired = 'passwordexpired',
    DSHeuristics = 'dsheuristics',
    UserAccountControl = 'useraccountcontrol',
    TrustAttributesInbound = 'trustattributesinbound',
    TrustAttributesOutbound = 'trustattributesoutbound',
    MinPwdLength = 'minpwdlength',
    PwdProperties = 'pwdproperties',
    PwdHistoryLength = 'pwdhistorylength',
    LockoutThreshold = 'lockoutthreshold',
    MinPwdAge = 'minpwdage',
    MaxPwdAge = 'maxpwdage',
    LockoutDuration = 'lockoutduration',
    LockoutObservationWindow = 'lockoutobservationwindow',
    OwnerSid = 'ownersid',
    SMBSigning = 'smbsigning',
    WebClientRunning = 'webclientrunning',
    RestrictOutboundNTLM = 'restrictoutboundntlm',
    GMSA = 'gmsa',
    MSA = 'msa',
    DoesAnyAceGrantOwnerRights = 'doesanyacegrantownerrights',
    DoesAnyInheritedAceGrantOwnerRights = 'doesanyinheritedacegrantownerrights',
    ADCSWebEnrollmentHTTP = 'adcswebenrollmenthttp',
    ADCSWebEnrollmentHTTPS = 'adcswebenrollmenthttps',
    ADCSWebEnrollmentHTTPSEPA = 'adcswebenrollmenthttpsepa',
    LDAPSigning = 'ldapsigning',
    LDAPAvailable = 'ldapavailable',
    LDAPSAvailable = 'ldapsavailable',
    LDAPSEPA = 'ldapsepa',
    IsDC = 'isdc',
    HTTPEnrollmentEndpoints = 'httpenrollmentendpoints',
    HTTPSEnrollmentEndpoints = 'httpsenrollmentendpoints',
    HasVulnerableEndpoint = 'hasvulnerableendpoint',
    RequireSecuritySignature = 'requiresecuritysignature',
    EnableSecuritySignature = 'enablesecuritysignature',
    RestrictReceivingNTLMTraffic = 'restrictreceivingntmltraffic',
    NTLMMinServerSec = 'ntlmminserversec',
    NTLMMinClientSec = 'ntlmminclientsec',
    LMCompatibilityLevel = 'lmcompatibilitylevel',
    UseMachineID = 'usemachineid',
    ClientAllowedNTLMServers = 'clientallowedntlmservers',
    Transitive = 'transitive',
    GroupScope = 'groupscope',
    NetBIOS = 'netbios',
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
        case ActiveDirectoryKindProperties.RoleSeparationEnabled:
            return 'Role Separation Enabled';
        case ActiveDirectoryKindProperties.RoleSeparationEnabledCollected:
            return 'Role Separation Enabled Collected';
        case ActiveDirectoryKindProperties.HasBasicConstraints:
            return 'Has Basic Constraints';
        case ActiveDirectoryKindProperties.BasicConstraintPathLength:
            return 'Basic Constraint Path Length';
        case ActiveDirectoryKindProperties.UnresolvedPublishedTemplates:
            return 'Unresolved Published Certificate Templates';
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
        case ActiveDirectoryKindProperties.BlocksInheritance:
            return 'Blocks GPO Inheritance';
        case ActiveDirectoryKindProperties.IsACL:
            return 'Is ACL';
        case ActiveDirectoryKindProperties.IsACLProtected:
            return 'ACL Inheritance Denied';
        case ActiveDirectoryKindProperties.InheritanceHash:
            return 'ACL Inheritance Hash';
        case ActiveDirectoryKindProperties.InheritanceHashes:
            return 'ACL Inheritance Hashes';
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
        case ActiveDirectoryKindProperties.SpoofSIDHistoryBlocked:
            return 'Spoof SID History Blocked';
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
            return 'Application Policies Required';
        case ActiveDirectoryKindProperties.IssuancePolicies:
            return 'Issuance Policies Required';
        case ActiveDirectoryKindProperties.SchemaVersion:
            return 'Schema Version';
        case ActiveDirectoryKindProperties.RequiresManagerApproval:
            return 'Requires Manager Approval';
        case ActiveDirectoryKindProperties.AuthenticationEnabled:
            return 'Authentication Enabled';
        case ActiveDirectoryKindProperties.SchannelAuthenticationEnabled:
            return 'Schannel Authentication Enabled';
        case ActiveDirectoryKindProperties.EnrolleeSuppliesSubject:
            return 'Enrollee Supplies Subject';
        case ActiveDirectoryKindProperties.CertificateApplicationPolicy:
            return 'Application Policy Extensions';
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
        case ActiveDirectoryKindProperties.CertificatePolicy:
            return 'Issuance Policy Extensions';
        case ActiveDirectoryKindProperties.CertTemplateOID:
            return 'Certificate Template OID';
        case ActiveDirectoryKindProperties.GroupLinkID:
            return 'Group Link ID';
        case ActiveDirectoryKindProperties.ObjectGUID:
            return 'Object GUID';
        case ActiveDirectoryKindProperties.ExpirePasswordsOnSmartCardOnlyAccounts:
            return 'Expire Passwords on Smart Card only Accounts';
        case ActiveDirectoryKindProperties.MachineAccountQuota:
            return 'Machine Account Quota';
        case ActiveDirectoryKindProperties.SupportedKerberosEncryptionTypes:
            return 'Supported Kerberos Encryption Types';
        case ActiveDirectoryKindProperties.TGTDelegation:
            return 'TGT Delegation';
        case ActiveDirectoryKindProperties.PasswordStoredUsingReversibleEncryption:
            return 'Password Stored Using Reversible Encryption';
        case ActiveDirectoryKindProperties.SmartcardRequired:
            return 'Smartcard Required';
        case ActiveDirectoryKindProperties.UseDESKeyOnly:
            return 'Use DES Key Only';
        case ActiveDirectoryKindProperties.LogonScriptEnabled:
            return 'Logon Script Enabled';
        case ActiveDirectoryKindProperties.LockedOut:
            return 'Locked Out';
        case ActiveDirectoryKindProperties.UserCannotChangePassword:
            return 'User Cannot Change Password';
        case ActiveDirectoryKindProperties.PasswordExpired:
            return 'Password Expired';
        case ActiveDirectoryKindProperties.DSHeuristics:
            return 'DSHeuristics';
        case ActiveDirectoryKindProperties.UserAccountControl:
            return 'User Account Control';
        case ActiveDirectoryKindProperties.TrustAttributesInbound:
            return 'Trust Attributes (Inbound)';
        case ActiveDirectoryKindProperties.TrustAttributesOutbound:
            return 'Trust Attributes (Outbound)';
        case ActiveDirectoryKindProperties.MinPwdLength:
            return 'Minimum password length';
        case ActiveDirectoryKindProperties.PwdProperties:
            return 'Password Properties';
        case ActiveDirectoryKindProperties.PwdHistoryLength:
            return 'Password History Length';
        case ActiveDirectoryKindProperties.LockoutThreshold:
            return 'Lockout Threshold';
        case ActiveDirectoryKindProperties.MinPwdAge:
            return 'Minimum Password Age';
        case ActiveDirectoryKindProperties.MaxPwdAge:
            return 'Maximum Password Age';
        case ActiveDirectoryKindProperties.LockoutDuration:
            return 'Lockout Duration';
        case ActiveDirectoryKindProperties.LockoutObservationWindow:
            return 'Lockout Observation Window';
        case ActiveDirectoryKindProperties.OwnerSid:
            return 'Owner SID';
        case ActiveDirectoryKindProperties.SMBSigning:
            return 'SMB Signing';
        case ActiveDirectoryKindProperties.WebClientRunning:
            return 'WebClient Running';
        case ActiveDirectoryKindProperties.RestrictOutboundNTLM:
            return 'Restrict Outbound NTLM';
        case ActiveDirectoryKindProperties.GMSA:
            return 'GMSA';
        case ActiveDirectoryKindProperties.MSA:
            return 'MSA';
        case ActiveDirectoryKindProperties.DoesAnyAceGrantOwnerRights:
            return 'Does Any ACE Grant Owner Rights';
        case ActiveDirectoryKindProperties.DoesAnyInheritedAceGrantOwnerRights:
            return 'Does Any Inherited ACE Grant Owner Rights';
        case ActiveDirectoryKindProperties.ADCSWebEnrollmentHTTP:
            return 'ADCS Web Enrollment HTTP';
        case ActiveDirectoryKindProperties.ADCSWebEnrollmentHTTPS:
            return 'ADCS Web Enrollment HTTPS';
        case ActiveDirectoryKindProperties.ADCSWebEnrollmentHTTPSEPA:
            return 'ADCS Web Enrollment HTTPS EPA';
        case ActiveDirectoryKindProperties.LDAPSigning:
            return 'LDAP Signing';
        case ActiveDirectoryKindProperties.LDAPAvailable:
            return 'LDAP Available';
        case ActiveDirectoryKindProperties.LDAPSAvailable:
            return 'LDAPS Available';
        case ActiveDirectoryKindProperties.LDAPSEPA:
            return 'LDAPS EPA';
        case ActiveDirectoryKindProperties.IsDC:
            return 'Is Domain Controller';
        case ActiveDirectoryKindProperties.HTTPEnrollmentEndpoints:
            return 'HTTP Enrollment Endpoints';
        case ActiveDirectoryKindProperties.HTTPSEnrollmentEndpoints:
            return 'HTTPS Enrollment Endpoints';
        case ActiveDirectoryKindProperties.HasVulnerableEndpoint:
            return 'Has Vulnerable Endpoint';
        case ActiveDirectoryKindProperties.RequireSecuritySignature:
            return 'Require Security Signature';
        case ActiveDirectoryKindProperties.EnableSecuritySignature:
            return 'Enable Security Signature';
        case ActiveDirectoryKindProperties.RestrictReceivingNTLMTraffic:
            return 'Restrict Receiving NTLM Traffic';
        case ActiveDirectoryKindProperties.NTLMMinServerSec:
            return 'NTLM Min Server Sec';
        case ActiveDirectoryKindProperties.NTLMMinClientSec:
            return 'NTLM Min Client Sec';
        case ActiveDirectoryKindProperties.LMCompatibilityLevel:
            return 'LM Compatibility Level';
        case ActiveDirectoryKindProperties.UseMachineID:
            return 'Use Machine ID';
        case ActiveDirectoryKindProperties.ClientAllowedNTLMServers:
            return 'Client Allowed NTLM Servers';
        case ActiveDirectoryKindProperties.Transitive:
            return 'Transitive';
        case ActiveDirectoryKindProperties.GroupScope:
            return 'Group Scope';
        case ActiveDirectoryKindProperties.NetBIOS:
            return 'NetBIOS';
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
        ActiveDirectoryRelationshipKind.AllowedToDelegate,
        ActiveDirectoryRelationshipKind.CoerceToTGT,
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
        ActiveDirectoryRelationshipKind.WriteGPLink,
        ActiveDirectoryRelationshipKind.GoldenCert,
        ActiveDirectoryRelationshipKind.ADCSESC1,
        ActiveDirectoryRelationshipKind.ADCSESC3,
        ActiveDirectoryRelationshipKind.ADCSESC4,
        ActiveDirectoryRelationshipKind.ADCSESC6a,
        ActiveDirectoryRelationshipKind.ADCSESC6b,
        ActiveDirectoryRelationshipKind.ADCSESC9a,
        ActiveDirectoryRelationshipKind.ADCSESC9b,
        ActiveDirectoryRelationshipKind.ADCSESC10a,
        ActiveDirectoryRelationshipKind.ADCSESC10b,
        ActiveDirectoryRelationshipKind.ADCSESC13,
        ActiveDirectoryRelationshipKind.SyncedToEntraUser,
        ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToSMB,
        ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToADCS,
        ActiveDirectoryRelationshipKind.WriteOwnerLimitedRights,
        ActiveDirectoryRelationshipKind.OwnsLimitedRights,
        ActiveDirectoryRelationshipKind.ClaimSpecialIdentity,
        ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToLDAP,
        ActiveDirectoryRelationshipKind.CoerceAndRelayNTLMToLDAPS,
        ActiveDirectoryRelationshipKind.ContainsIdentity,
        ActiveDirectoryRelationshipKind.PropagatesACEsTo,
        ActiveDirectoryRelationshipKind.GPOAppliesTo,
        ActiveDirectoryRelationshipKind.CanApplyGPO,
        ActiveDirectoryRelationshipKind.HasTrustKeys,
        ActiveDirectoryRelationshipKind.DCFor,
        ActiveDirectoryRelationshipKind.SameForestTrust,
        ActiveDirectoryRelationshipKind.SpoofSIDHistory,
        ActiveDirectoryRelationshipKind.AbuseTGTDelegation,
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
    AKSContributor = 'AZAKSContributor',
    NodeResourceGroup = 'AZNodeResourceGroup',
    WebsiteContributor = 'AZWebsiteContributor',
    LogicAppContributor = 'AZLogicAppContributor',
    AZMGAddMember = 'AZMGAddMember',
    AZMGAddOwner = 'AZMGAddOwner',
    AZMGAddSecret = 'AZMGAddSecret',
    AZMGGrantAppRoles = 'AZMGGrantAppRoles',
    AZMGGrantRole = 'AZMGGrantRole',
    SyncedToADUser = 'SyncedToADUser',
    AZRoleEligible = 'AZRoleEligible',
    AZRoleApprover = 'AZRoleApprover',
    
    // All Azure Graph API Permissions as of 
    APIConnectorsReadAll = 'AZMGAPIConnectors_Read_All',
    APIConnectorsReadWriteAll = 'AZMGAPIConnectors_ReadWrite_All',
    AccessReviewReadAll = 'AZMGAccessReview_Read_All',
    AccessReviewReadWriteAll = 'AZMGAccessReview_ReadWrite_All',
    AccessReviewReadWriteMembership = 'AZMGAccessReview_ReadWrite_Membership',
    AcronymReadAll = 'AZMGAcronym_Read_All',
    AdministrativeUnitReadAll = 'AZMGAdministrativeUnit_Read_All',
    AdministrativeUnitReadWriteAll = 'AZMGAdministrativeUnit_ReadWrite_All',
    AgreementAcceptanceReadAll = 'AZMGAgreementAcceptance_Read_All',
    AgreementReadAll = 'AZMGAgreement_Read_All',
    AgreementReadWriteAll = 'AZMGAgreement_ReadWrite_All',
    AiEnterpriseInteractionReadAll = 'AZMGAiEnterpriseInteraction_Read_All',
    AppCatalogReadAll = 'AZMGAppCatalog_Read_All',
    AppCatalogReadWriteAll = 'AZMGAppCatalog_ReadWrite_All',
    AppCertTrustConfigurationReadAll = 'AZMGAppCertTrustConfiguration_Read_All',
    AppCertTrustConfigurationReadWriteAll = 'AZMGAppCertTrustConfiguration_ReadWrite_All',
    AppRoleAssignmentReadWriteAll = 'AZMGAppRoleAssignment_ReadWrite_All',
    ApplicationReadAll = 'AZMGApplication_Read_All',
    ApplicationReadWriteAll = 'AZMGApplication_ReadWrite_All',
    ApplicationReadWriteOwnedBy = 'AZMGApplication_ReadWrite_OwnedBy',
    ApprovalSolutionReadAll = 'AZMGApprovalSolution_Read_All',
    ApprovalSolutionReadWriteAll = 'AZMGApprovalSolution_ReadWrite_All',
    AttackSimulationReadAll = 'AZMGAttackSimulation_Read_All',
    AttackSimulationReadWriteAll = 'AZMGAttackSimulation_ReadWrite_All',
    AuditLogReadAll = 'AZMGAuditLog_Read_All',
    AuditLogsQueryReadAll = 'AZMGAuditLogsQuery_Read_All',
    AuthenticationContextReadAll = 'AZMGAuthenticationContext_Read_All',
    AuthenticationContextReadWriteAll = 'AZMGAuthenticationContext_ReadWrite_All',
    BillingConfigurationReadWriteAll = 'AZMGBillingConfiguration_ReadWrite_All',
    BitlockerKeyReadAll = 'AZMGBitlockerKey_Read_All',
    BitlockerKeyReadBasicAll = 'AZMGBitlockerKey_ReadBasic_All',
    BookingsAppointmentReadWriteAll = 'AZMGBookingsAppointment_ReadWrite_All',
    BookingsManageAll = 'AZMGBookings_Manage_All',
    BookingsReadAll = 'AZMGBookings_Read_All',
    BookingsReadWriteAll = 'AZMGBookings_ReadWrite_All',
    BookmarkReadAll = 'AZMGBookmark_Read_All',
    BrowserSiteListsReadAll = 'AZMGBrowserSiteLists_Read_All',
    BrowserSiteListsReadWriteAll = 'AZMGBrowserSiteLists_ReadWrite_All',
    BusinessScenarioConfigReadAll = 'AZMGBusinessScenarioConfig_Read_All',
    BusinessScenarioConfigReadOwnedBy = 'AZMGBusinessScenarioConfig_Read_OwnedBy',
    BusinessScenarioConfigReadWriteAll = 'AZMGBusinessScenarioConfig_ReadWrite_All',
    BusinessScenarioConfigReadWriteOwnedBy = 'AZMGBusinessScenarioConfig_ReadWrite_OwnedBy',
    BusinessScenarioDataReadOwnedBy = 'AZMGBusinessScenarioData_Read_OwnedBy',
    BusinessScenarioDataReadWriteOwnedBy = 'AZMGBusinessScenarioData_ReadWrite_OwnedBy',
    CalendarsReadBasicAll = 'AZMGCalendars_ReadBasic_All',
    CalendarsReadShared = 'AZMGCalendars_Read_Shared',
    CalendarsReadWriteShared = 'AZMGCalendars_ReadWrite_Shared',
    CallDelegationReadAll = 'AZMGCallDelegation_Read_All',
    CallDelegationReadWriteAll = 'AZMGCallDelegation_ReadWrite_All',
    CallEventsReadAll = 'AZMGCallEvents_Read_All',
    CallRecordsReadAll = 'AZMGCallRecords_Read_All',
    CallsAccessMediaAll = 'AZMGCalls_AccessMedia_All',
    CallsInitiateAll = 'AZMGCalls_Initiate_All',
    CallsInitiateGroupCallAll = 'AZMGCalls_InitiateGroupCall_All',
    CallsJoinGroupCallAll = 'AZMGCalls_JoinGroupCall_All',
    CallsJoinGroupCallAsGuestAll = 'AZMGCalls_JoinGroupCallAsGuest_All',
    ChangeManagementReadAll = 'AZMGChangeManagement_Read_All',
    ChannelDeleteAll = 'AZMGChannel_Delete_All',
    ChannelMemberReadAll = 'AZMGChannelMember_Read_All',
    ChannelMemberReadWriteAll = 'AZMGChannelMember_ReadWrite_All',
    ChannelMessageReadAll = 'AZMGChannelMessage_Read_All',
    ChannelMessageUpdatePolicyViolationAll = 'AZMGChannelMessage_UpdatePolicyViolation_All',
    ChannelReadBasicAll = 'AZMGChannel_ReadBasic_All',
    ChannelSettingsReadAll = 'AZMGChannelSettings_Read_All',
    ChannelSettingsReadWriteAll = 'AZMGChannelSettings_ReadWrite_All',
    ChatManageDeletionAll = 'AZMGChat_ManageDeletion_All',
    ChatMemberReadAll = 'AZMGChatMember_Read_All',
    ChatMemberReadWhereInstalled = 'AZMGChatMember_Read_WhereInstalled',
    ChatMemberReadWriteAll = 'AZMGChatMember_ReadWrite_All',
    ChatMemberReadWriteWhereInstalled = 'AZMGChatMember_ReadWrite_WhereInstalled',
    ChatMessageReadAll = 'AZMGChatMessage_Read_All',
    ChatReadAll = 'AZMGChat_Read_All',
    ChatReadBasicAll = 'AZMGChat_ReadBasic_All',
    ChatReadBasicWhereInstalled = 'AZMGChat_ReadBasic_WhereInstalled',
    ChatReadWhereInstalled = 'AZMGChat_Read_WhereInstalled',
    ChatReadWriteAll = 'AZMGChat_ReadWrite_All',
    ChatReadWriteWhereInstalled = 'AZMGChat_ReadWrite_WhereInstalled',
    ChatUpdatePolicyViolationAll = 'AZMGChat_UpdatePolicyViolation_All',
    CloudPCReadAll = 'AZMGCloudPC_Read_All',
    CloudPCReadWriteAll = 'AZMGCloudPC_ReadWrite_All',
    CommunityReadAll = 'AZMGCommunity_Read_All',
    CommunityReadWriteAll = 'AZMGCommunity_ReadWrite_All',
    ConfigurationMonitoringReadAll = 'AZMGConfigurationMonitoring_Read_All',
    ConfigurationMonitoringReadWriteAll = 'AZMGConfigurationMonitoring_ReadWrite_All',
    ConsentRequestReadAll = 'AZMGConsentRequest_Read_All',
    ConsentRequestReadApproveAll = 'AZMGConsentRequest_ReadApprove_All',
    ConsentRequestReadWriteAll = 'AZMGConsentRequest_ReadWrite_All',
    ContactsReadShared = 'AZMGContacts_Read_Shared',
    ContactsReadWriteShared = 'AZMGContacts_ReadWrite_Shared',
    ContentProcessAll = 'AZMGContent_Process_All',
    ContentProcessUser = 'AZMGContent_Process_User',
    CrossTenantInformationReadBasicAll = 'AZMGCrossTenantInformation_ReadBasic_All',
    CrossTenantUserProfileSharingReadAll = 'AZMGCrossTenantUserProfileSharing_Read_All',
    CrossTenantUserProfileSharingReadWriteAll = 'AZMGCrossTenantUserProfileSharing_ReadWrite_All',
    CustomAuthenticationExtensionReadAll = 'AZMGCustomAuthenticationExtension_Read_All',
    CustomAuthenticationExtensionReadWriteAll = 'AZMGCustomAuthenticationExtension_ReadWrite_All',
    CustomAuthenticationExtensionReceivePayload = 'AZMGCustomAuthenticationExtension_Receive_Payload',
    CustomDetectionReadAll = 'AZMGCustomDetection_Read_All',
    CustomDetectionReadWriteAll = 'AZMGCustomDetection_ReadWrite_All',
    CustomSecAttributeAssignmentReadAll = 'AZMGCustomSecAttributeAssignment_Read_All',
    CustomSecAttributeAssignmentReadWriteAll = 'AZMGCustomSecAttributeAssignment_ReadWrite_All',
    CustomSecAttributeAuditLogsReadAll = 'AZMGCustomSecAttributeAuditLogs_Read_All',
    CustomSecAttributeDefinitionReadAll = 'AZMGCustomSecAttributeDefinition_Read_All',
    CustomSecAttributeDefinitionReadWriteAll = 'AZMGCustomSecAttributeDefinition_ReadWrite_All',
    CustomSecAttributeProvisioningReadAll = 'AZMGCustomSecAttributeProvisioning_Read_All',
    CustomSecAttributeProvisioningReadWriteAll = 'AZMGCustomSecAttributeProvisioning_ReadWrite_All',
    CustomTagsReadAll = 'AZMGCustomTags_Read_All',
    CustomTagsReadWriteAll = 'AZMGCustomTags_ReadWrite_All',
    DelegatedAdminRelationshipReadAll = 'AZMGDelegatedAdminRelationship_Read_All',
    DelegatedAdminRelationshipReadWriteAll = 'AZMGDelegatedAdminRelationship_ReadWrite_All',
    DelegatedPermissionGrantReadAll = 'AZMGDelegatedPermissionGrant_Read_All',
    DelegatedPermissionGrantReadWriteAll = 'AZMGDelegatedPermissionGrant_ReadWrite_All',
    DeviceLocalCredentialReadAll = 'AZMGDeviceLocalCredential_Read_All',
    DeviceLocalCredentialReadBasicAll = 'AZMGDeviceLocalCredential_ReadBasic_All',
    DeviceManagementAppsReadAll = 'AZMGDeviceManagementApps_Read_All',
    DeviceManagementAppsReadWriteAll = 'AZMGDeviceManagementApps_ReadWrite_All',
    DeviceManagementCloudCAReadAll = 'AZMGDeviceManagementCloudCA_Read_All',
    DeviceManagementCloudCAReadWriteAll = 'AZMGDeviceManagementCloudCA_ReadWrite_All',
    DeviceManagementConfigurationReadAll = 'AZMGDeviceManagementConfiguration_Read_All',
    DeviceManagementConfigurationReadWriteAll = 'AZMGDeviceManagementConfiguration_ReadWrite_All',
    DeviceManagementManagedDevicesPrivilegedOperationsAll = 'AZMGDeviceManagementManagedDevices_PrivilegedOperations_All',
    DeviceManagementManagedDevicesReadAll = 'AZMGDeviceManagementManagedDevices_Read_All',
    DeviceManagementManagedDevicesReadWriteAll = 'AZMGDeviceManagementManagedDevices_ReadWrite_All',
    DeviceManagementRBACReadAll = 'AZMGDeviceManagementRBAC_Read_All',
    DeviceManagementRBACReadWriteAll = 'AZMGDeviceManagementRBAC_ReadWrite_All',
    DeviceManagementScriptsReadAll = 'AZMGDeviceManagementScripts_Read_All',
    DeviceManagementScriptsReadWriteAll = 'AZMGDeviceManagementScripts_ReadWrite_All',
    DeviceManagementServiceConfigReadAll = 'AZMGDeviceManagementServiceConfig_Read_All',
    DeviceManagementServiceConfigReadWriteAll = 'AZMGDeviceManagementServiceConfig_ReadWrite_All',
    DeviceReadAll = 'AZMGDevice_Read_All',
    DeviceReadWriteAll = 'AZMGDevice_ReadWrite_All',
    DeviceTemplateReadAll = 'AZMGDeviceTemplate_Read_All',
    DeviceTemplateReadWriteAll = 'AZMGDeviceTemplate_ReadWrite_All',
    DirectoryAccessAsUserAll = 'AZMGDirectory_AccessAsUser_All',
    DirectoryReadAll = 'AZMGDirectory_Read_All',
    DirectoryReadWriteAll = 'AZMGDirectory_ReadWrite_All',
    DirectoryRecommendationsReadAll = 'AZMGDirectoryRecommendations_Read_All',
    DirectoryRecommendationsReadWriteAll = 'AZMGDirectoryRecommendations_ReadWrite_All',
    DomainReadAll = 'AZMGDomain_Read_All',
    DomainReadWriteAll = 'AZMGDomain_ReadWrite_All',
    EASAccessAsUserAll = 'AZMGEAS_AccessAsUser_All',
    EWSAccessAsUserAll = 'AZMGEWS_AccessAsUser_All',
    EduAdministrationReadAll = 'AZMGEduAdministration_Read_All',
    EduAdministrationReadWriteAll = 'AZMGEduAdministration_ReadWrite_All',
    EduAssignmentsReadAll = 'AZMGEduAssignments_Read_All',
    EduAssignmentsReadBasicAll = 'AZMGEduAssignments_ReadBasic_All',
    EduAssignmentsReadWriteAll = 'AZMGEduAssignments_ReadWrite_All',
    EduAssignmentsReadWriteBasicAll = 'AZMGEduAssignments_ReadWriteBasic_All',
    EduCurriculaReadAll = 'AZMGEduCurricula_Read_All',
    EduCurriculaReadWriteAll = 'AZMGEduCurricula_ReadWrite_All',
    EduRosterReadAll = 'AZMGEduRoster_Read_All',
    EduRosterReadBasicAll = 'AZMGEduRoster_ReadBasic_All',
    EduRosterReadWriteAll = 'AZMGEduRoster_ReadWrite_All',
    EngagementConversationMigrationAll = 'AZMGEngagementConversation_Migration_All',
    EngagementMeetingConversationReadAll = 'AZMGEngagementMeetingConversation_Read_All',
    EngagementRoleReadAll = 'AZMGEngagementRole_Read_All',
    EngagementRoleReadWriteAll = 'AZMGEngagementRole_ReadWrite_All',
    EntitlementManagementReadAll = 'AZMGEntitlementManagement_Read_All',
    EntitlementManagementReadWriteAll = 'AZMGEntitlementManagement_ReadWrite_All',
    EventListenerReadAll = 'AZMGEventListener_Read_All',
    EventListenerReadWriteAll = 'AZMGEventListener_ReadWrite_All',
    ExternalConnectionReadAll = 'AZMGExternalConnection_Read_All',
    ExternalConnectionReadWriteAll = 'AZMGExternalConnection_ReadWrite_All',
    ExternalConnectionReadWriteOwnedBy = 'AZMGExternalConnection_ReadWrite_OwnedBy',
    ExternalItemReadAll = 'AZMGExternalItem_Read_All',
    ExternalItemReadWriteAll = 'AZMGExternalItem_ReadWrite_All',
    ExternalItemReadWriteOwnedBy = 'AZMGExternalItem_ReadWrite_OwnedBy',
    ExternalUserProfileReadAll = 'AZMGExternalUserProfile_Read_All',
    ExternalUserProfileReadWriteAll = 'AZMGExternalUserProfile_ReadWrite_All',
    FileStorageContainerManageAll = 'AZMGFileStorageContainer_Manage_All',
    FilesReadAll = 'AZMGFiles_Read_All',
    FilesReadSelected = 'AZMGFiles_Read_Selected',
    FilesReadWriteAll = 'AZMGFiles_ReadWrite_All',
    FilesReadWriteAppFolder = 'AZMGFiles_ReadWrite_AppFolder',
    FilesReadWriteSelected = 'AZMGFiles_ReadWrite_Selected',
    FilesSelectedOperationsSelected = 'AZMGFiles_SelectedOperations_Selected',
    FinancialsReadWriteAll = 'AZMGFinancials_ReadWrite_All',
    GroupMemberReadAll = 'AZMGGroupMember_Read_All',
    GroupMemberReadWriteAll = 'AZMGGroupMember_ReadWrite_All',
    GroupReadAll = 'AZMGGroup_Read_All',
    GroupReadWriteAll = 'AZMGGroup_ReadWrite_All',
    GGroupSettingsReadAll = 'AZMGGroupSettings_Read_All',
    GroupSettingsReadWriteAll = 'AZMGGroupSettings_ReadWrite_All',
    HealthMonitoringAlertConfigReadAll = 'AZMGHealthMonitoringAlertConfig_Read_All',
    HealthMonitoringAlertConfigReadWriteAll = 'AZMGHealthMonitoringAlertConfig_ReadWrite_All',
    HealthMonitoringAlertReadAll = 'AZMGHealthMonitoringAlert_Read_All',
    HealthMonitoringAlertReadWriteAll = 'AZMGHealthMonitoringAlert_ReadWrite_All',
    IMAPAccessAsUserAll = 'AZMGIMAP_AccessAsUser_All',
    IdentityProviderReadAll = 'AZMGIdentityProvider_Read_All',
    IdentityProviderReadWriteAll = 'AZMGIdentityProvider_ReadWrite_All',
    IdentityRiskEventReadAll = 'AZMGIdentityRiskEvent_Read_All',
    IdentityRiskEventReadWriteAll = 'AZMGIdentityRiskEvent_ReadWrite_All',
    IdentityRiskyServicePrincipalReadAll = 'AZMGIdentityRiskyServicePrincipal_Read_All',
    IdentityRiskyServicePrincipalReadWriteAll = 'AZMGIdentityRiskyServicePrincipal_ReadWrite_All',
    IdentityRiskyUserReadAll = 'AZMGIdentityRiskyUser_Read_All',
    IdentityRiskyUserReadWriteAll = 'AZMGIdentityRiskyUser_ReadWrite_All',
    IdentityUserFlowReadAll = 'AZMGIdentityUserFlow_Read_All',
    IdentityUserFlowReadWriteAll = 'AZMGIdentityUserFlow_ReadWrite_All',
    IndustryDataReadBasicAll = 'AZMGIndustryData_ReadBasic_All',
    InformationProtectionConfigReadAll = 'AZMGInformationProtectionConfig_Read_All',
    InformationProtectionContentSignAll = 'AZMGInformationProtectionContent_Sign_All',
    InformationProtectionContentWriteAll = 'AZMGInformationProtectionContent_Write_All',
    InformationProtectionPolicyReadAll = 'AZMGInformationProtectionPolicy_Read_All',
    LearningAssignedCourseReadAll = 'AZMGLearningAssignedCourse_Read_All',
    LearningAssignedCourseReadWriteAll = 'AZMGLearningAssignedCourse_ReadWrite_All',
    LearningContentReadAll = 'AZMGLearningContent_Read_All',
    LearningContentReadWriteAll = 'AZMGLearningContent_ReadWrite_All',
    LearningSelfInitiatedCourseReadAll = 'AZMGLearningSelfInitiatedCourse_Read_All',
    LearningSelfInitiatedCourseReadWriteAll = 'AZMGLearningSelfInitiatedCourse_ReadWrite_All',
    LicenseAssignmentReadAll = 'AZMGLicenseAssignment_Read_All',
    LicenseAssignmentReadWriteAll = 'AZMGLicenseAssignment_ReadWrite_All',
    LifecycleWorkflowsReadAll = 'AZMGLifecycleWorkflows_Read_All',
    LifecycleWorkflowsReadWriteAll = 'AZMGLifecycleWorkflows_ReadWrite_All',
    ListItemsSelectedOperationsSelected = 'AZMGListItems_SelectedOperations_Selected',
    ListsSelectedOperationsSelected = 'AZMGLists_SelectedOperations_Selected',
    MailReadBasicAll = 'AZMGMail_ReadBasic_All',
    MailReadBasicShared = 'AZMGMail_ReadBasic_Shared',
    MailReadShared = 'AZMGMail_Read_Shared',
    MailReadWriteShared = 'AZMGMail_ReadWrite_Shared',
    MailSendShared = 'AZMGMail_Send_Shared',
    MailboxFolderReadAll = 'AZMGMailboxFolder_Read_All',
    MailboxFolderReadWriteAll = 'AZMGMailboxFolder_ReadWrite_All',
    MailboxItemImportExportAll = 'AZMGMailboxItem_ImportExport_All',
    MailboxItemReadAll = 'AZMGMailboxItem_Read_All',
    ManagedTenantsReadAll = 'AZMGManagedTenants_Read_All',
    ManagedTenantsReadWriteAll = 'AZMGManagedTenants_ReadWrite_All',
    MemberReadHidden = 'AZMGMember_Read_Hidden',
    MultiTenantOrganizationReadAll = 'AZMGMultiTenantOrganization_Read_All',
    MultiTenantOrganizationReadBasicAll = 'AZMGMultiTenantOrganization_ReadBasic_All',
    MultiTenantOrganizationReadWriteAll = 'AZMGMultiTenantOrganization_ReadWrite_All',
    MutualTlsOauthConfigurationReadAll = 'AZMGMutualTlsOauthConfiguration_Read_All',
    MutualTlsOauthConfigurationReadWriteAll = 'AZMGMutualTlsOauthConfiguration_ReadWrite_All',
    NetworkAccessBranchReadAll = 'AZMGNetworkAccessBranch_Read_All',
    NetworkAccessBranchReadWriteAll = 'AZMGNetworkAccessBranch_ReadWrite_All',
    NetworkAccessPolicyReadAll = 'AZMGNetworkAccessPolicy_Read_All',
    NetworkAccessPolicyReadWriteAll = 'AZMGNetworkAccessPolicy_ReadWrite_All',
    NetworkAccessReadAll = 'AZMGNetworkAccess_Read_All',
    NetworkAccessReadWriteAll = 'AZMGNetworkAccess_ReadWrite_All',
    NotesReadAll = 'AZMGNotes_Read_All',
    NotesReadWriteAll = 'AZMGNotes_ReadWrite_All',
    NotesReadWriteCreatedByApp = 'AZMGNotes_ReadWrite_CreatedByApp',
    NotificationsReadWriteCreatedByApp = 'AZMGNotifications_ReadWrite_CreatedByApp',
    OnPremDirectorySynchronizationReadAll = 'AZMGOnPremDirectorySynchronization_Read_All',
    OnPremDirectorySynchronizationReadWriteAll = 'AZMGOnPremDirectorySynchronization_ReadWrite_All',
    OnPremisesPublishingProfilesReadWriteAll = 'AZMGOnPremisesPublishingProfiles_ReadWrite_All',
    OnlineMeetingAiInsightReadAll = 'AZMGOnlineMeetingAiInsight_Read_All',
    OnlineMeetingAiInsightReadChat = 'AZMGOnlineMeetingAiInsight_Read_Chat',
    OnlineMeetingArtifactReadAll = 'AZMGOnlineMeetingArtifact_Read_All',
    OnlineMeetingRecordingReadAll = 'AZMGOnlineMeetingRecording_Read_All',
    OnlineMeetingTranscriptReadAll = 'AZMGOnlineMeetingTranscript_Read_All',
    OnlineMeetingsReadAll = 'AZMGOnlineMeetings_Read_All',
    OnlineMeetingsReadWriteAll = 'AZMGOnlineMeetings_ReadWrite_All',
    OrgContactReadAll = 'AZMGOrgContact_Read_All',
    OrganizationReadAll = 'AZMGOrganization_Read_All',
    OrganizationReadWriteAll = 'AZMGOrganization_ReadWrite_All',
    OrganizationalBrandingReadAll = 'AZMGOrganizationalBranding_Read_All',
    OrganizationalBrandingReadWriteAll = 'AZMGOrganizationalBranding_ReadWrite_All',
    POPAccessAsUserAll = 'AZMGPOP_AccessAsUser_All',
    PartnerBillingReadAll = 'AZMGPartnerBilling_Read_All',
    PartnerSecurityReadAll = 'AZMGPartnerSecurity_Read_All',
    PartnerSecurityReadWriteAll = 'AZMGPartnerSecurity_ReadWrite_All',
    PendingExternalUserProfileReadAll = 'AZMGPendingExternalUserProfile_Read_All',
    PendingExternalUserProfileReadWriteAll = 'AZMGPendingExternalUserProfile_ReadWrite_All',
    PeopleReadAll = 'AZMGPeople_Read_All',
    PeopleSettingsReadAll = 'AZMGPeopleSettings_Read_All',
    PeopleSettingsReadWriteAll = 'AZMGPeopleSettings_ReadWrite_All',
    PlaceDeviceReadAll = 'AZMGPlaceDevice_Read_All',
    PlaceDeviceReadWriteAll = 'AZMGPlaceDevice_ReadWrite_All',
    PlaceDeviceTelemetryReadWriteAll = 'AZMGPlaceDeviceTelemetry_ReadWrite_All',
    PlaceReadAll = 'AZMGPlace_Read_All',
    PlaceReadWriteAll = 'AZMGPlace_ReadWrite_All',
    PolicyReadAll = 'AZMGPolicy_Read_All',
    PolicyReadAuthenticationMethod = 'AZMGPolicy_Read_AuthenticationMethod',
    PolicyReadConditionalAccess = 'AZMGPolicy_Read_ConditionalAccess',
    PolicyReadDeviceConfiguration = 'AZMGPolicy_Read_DeviceConfiguration',
    PolicyReadIdentityProtection = 'AZMGPolicy_Read_IdentityProtection',
    PolicyReadPermissionGrant = 'AZMGPolicy_Read_PermissionGrant',
    PolicyReadWriteAccessReview = 'AZMGPolicy_ReadWrite_AccessReview',
    PolicyReadWriteApplicationConfiguration = 'AZMGPolicy_ReadWrite_ApplicationConfiguration',
    PolicyReadWriteAuthenticationFlows = 'AZMGPolicy_ReadWrite_AuthenticationFlows',
    PolicyReadWriteAuthenticationMethod = 'AZMGPolicy_ReadWrite_AuthenticationMethod',
    PolicyReadWriteAuthorization = 'AZMGPolicy_ReadWrite_Authorization',
    PolicyReadWriteConditionalAccess = 'AZMGPolicy_ReadWrite_ConditionalAccess',
    PolicyReadWriteConsentRequest = 'AZMGPolicy_ReadWrite_ConsentRequest',
    PolicyReadWriteCrossTenantAccess = 'AZMGPolicy_ReadWrite_CrossTenantAccess',
    PolicyReadWriteCrossTenantCapability = 'AZMGPolicy_ReadWrite_CrossTenantCapability',
    PolicyReadWriteDeviceConfiguration = 'AZMGPolicy_ReadWrite_DeviceConfiguration',
    PolicyReadWriteExternalIdentities = 'AZMGPolicy_ReadWrite_ExternalIdentities',
    PolicyReadWriteFeatureRollout = 'AZMGPolicy_ReadWrite_FeatureRollout',
    PolicyReadWriteFedTokenValidation = 'AZMGPolicy_ReadWrite_FedTokenValidation',
    PolicyReadWriteIdentityProtection = 'AZMGPolicy_ReadWrite_IdentityProtection',
    PolicyReadWriteMobilityManagement = 'AZMGPolicy_ReadWrite_MobilityManagement',
    PolicyReadWritePermissionGrant = 'AZMGPolicy_ReadWrite_PermissionGrant',
    PolicyReadWriteSecurityDefaults = 'AZMGPolicy_ReadWrite_SecurityDefaults',
    PolicyReadWriteTrustFramework = 'AZMGPolicy_ReadWrite_TrustFramework',
    PresenceReadAll = 'AZMGPresence_Read_All',
    PresenceReadWriteAll = 'AZMGPresence_ReadWrite_All',
    PrintConnectorReadAll = 'AZMGPrintConnector_Read_All',
    PrintConnectorReadWriteAll = 'AZMGPrintConnector_ReadWrite_All',
    PrintJobManageAll = 'AZMGPrintJob_Manage_All',
    PrintJobReadAll = 'AZMGPrintJob_Read_All',
    PrintJobReadBasicAll = 'AZMGPrintJob_ReadBasic_All',
    PrintJobReadWriteAll = 'AZMGPrintJob_ReadWrite_All',
    PrintJobReadWriteBasicAll = 'AZMGPrintJob_ReadWriteBasic_All',
    PrintSettingsReadAll = 'AZMGPrintSettings_Read_All',
    PrintSettingsReadWriteAll = 'AZMGPrintSettings_ReadWrite_All',
    PrintTaskDefinitionReadWriteAll = 'AZMGPrintTaskDefinition_ReadWrite_All',
    PrinterFullControlAll = 'AZMGPrinter_FullControl_All',
    PrinterReadAll = 'AZMGPrinter_Read_All',
    PrinterReadWriteAll = 'AZMGPrinter_ReadWrite_All',
    PrinterShareReadAll = 'AZMGPrinterShare_Read_All',
    PrinterShareReadBasicAll = 'AZMGPrinterShare_ReadBasic_All',
    PrinterShareReadWriteAll = 'AZMGPrinterShare_ReadWrite_All',
    PrivilegedAccessReadAzureAD = 'AZMGPrivilegedAccess_Read_AzureAD',
    PrivilegedAccessReadAzureADGroup = 'AZMGPrivilegedAccess_Read_AzureADGroup',
    PrivilegedAccessReadAzureResources = 'AZMGPrivilegedAccess_Read_AzureResources',
    PrivilegedAccessReadWriteAzureAD = 'AZMGPrivilegedAccess_ReadWrite_AzureAD',
    PrivilegedAccessReadWriteAzureADGroup = 'AZMGPrivilegedAccess_ReadWrite_AzureADGroup',
    PrivilegedAccessReadWriteAzureResources = 'AZMGPrivilegedAccess_ReadWrite_AzureResources',
    PrivilegedAssignmentScheduleReadAzureADGroup = 'AZMGPrivilegedAssignmentSchedule_Read_AzureADGroup',
    PrivilegedAssignmentScheduleReadWriteAzureADGroup = 'AZMGPrivilegedAssignmentSchedule_ReadWrite_AzureADGroup',
    PrivilegedAssignmentScheduleRemoveAzureADGroup = 'AZMGPrivilegedAssignmentSchedule_Remove_AzureADGroup',
    PrivilegedEligibilityScheduleReadAzureADGroup = 'AZMGPrivilegedEligibilitySchedule_Read_AzureADGroup',
    PrivilegedEligibilityScheduleReadWriteAzureADGroup = 'AZMGPrivilegedEligibilitySchedule_ReadWrite_AzureADGroup',
    PrivilegedEligibilityScheduleRemoveAzureADGroup = 'AZMGPrivilegedEligibilitySchedule_Remove_AzureADGroup',
    ProfilePhotoReadAll = 'AZMGProfilePhoto_Read_All',
    ProfilePhotoReadWriteAll = 'AZMGProfilePhoto_ReadWrite_All',
    ProgramControlReadAll = 'AZMGProgramControl_Read_All',
    ProgramControlReadWriteAll = 'AZMGProgramControl_ReadWrite_All',
    ProtectionScopesComputeAll = 'AZMGProtectionScopes_Compute_All',
    ProtectionScopesComputeUser = 'AZMGProtectionScopes_Compute_User',
    ProvisioningLogReadAll = 'AZMGProvisioningLog_Read_All',
    PublicKeyInfrastructureReadAll = 'AZMGPublicKeyInfrastructure_Read_All',
    PublicKeyInfrastructureReadWriteAll = 'AZMGPublicKeyInfrastructure_ReadWrite_All',
    QnAReadAll = 'AZMGQnA_Read_All',
    RecordsManagementReadAll = 'AZMGRecordsManagement_Read_All',
        RecordsManagementReadWriteAll = 'AZMGRecordsManagement_ReadWrite_All',
    ReportSettingsReadAll = 'AZMGReportSettings_Read_All',
    ReportSettingsReadWriteAll = 'AZMGReportSettings_ReadWrite_All',
    ReportReadAll = 'AZMGReport_Read_All',
    ReportsReadAll = 'AZMGReports_Read_All',
    ResourceSpecificPermissionGrantReadForChatAll = 'AZMGResourceSpecificPermissionGrant_ReadForChat_All',
    ResourceSpecificPermissionGrantReadForTeamAll = 'AZMGResourceSpecificPermissionGrant_ReadForTeam_All',
    ResourceSpecificPermissionGrantReadForUserAll = 'AZMGResourceSpecificPermissionGrant_ReadForUser_All',
    RiskPreventionProvidersReadAll = 'AZMGRiskPreventionProviders_Read_All',
    RiskPreventionProvidersReadWriteAll = 'AZMGRiskPreventionProviders_ReadWrite_All',
    RoleAssignmentScheduleReadDirectory = 'AZMGRoleAssignmentSchedule_Read_Directory',
    RoleAssignmentScheduleReadWriteDirectory = 'AZMGRoleAssignmentSchedule_ReadWrite_Directory',
    RoleAssignmentScheduleRemoveDirectory = 'AZMGRoleAssignmentSchedule_Remove_Directory',
    RoleEligibilityScheduleReadDirectory = 'AZMGRoleEligibilitySchedule_Read_Directory',
    RoleEligibilityScheduleReadWriteDirectory = 'AZMGRoleEligibilitySchedule_ReadWrite_Directory',
    RoleEligibilityScheduleRemoveDirectory = 'AZMGRoleEligibilitySchedule_Remove_Directory',
    RoleManagementAlertReadDirectory = 'AZMGRoleManagementAlert_Read_Directory',
    RoleManagementAlertReadWriteDirectory = 'AZMGRoleManagementAlert_ReadWrite_Directory',
    RoleManagementPolicyReadAzureADGroup = 'AZMGRoleManagementPolicy_Read_AzureADGroup',
    RoleManagementPolicyReadDirectory = 'AZMGRoleManagementPolicy_Read_Directory',
    RoleManagementPolicyReadWriteAzureADGroup = 'AZMGRoleManagementPolicy_ReadWrite_AzureADGroup',
    RoleManagementPolicyReadWriteDirectory = 'AZMGRoleManagementPolicy_ReadWrite_Directory',
    RoleManagementReadAll = 'AZMGRoleManagement_Read_All',
    RoleManagementReadCloudPC = 'AZMGRoleManagement_Read_CloudPC',
    RoleManagementReadDefender = 'AZMGRoleManagement_Read_Defender',
    RoleManagementReadDirectory = 'AZMGRoleManagement_Read_Directory',
    RoleManagementReadExchange = 'AZMGRoleManagement_Read_Exchange',
    RoleManagementReadWriteCloudPC = 'AZMGRoleManagement_ReadWrite_CloudPC',
    RoleManagementReadWriteDefender = 'AZMGRoleManagement_ReadWrite_Defender',
    RoleManagementReadWriteDirectory = 'AZMGRoleManagement_ReadWrite_Directory',
    RoleManagementReadWriteExchange = 'AZMGRoleManagement_ReadWrite_Exchange',
    SchedulePermissionsReadWriteAll = 'AZMGSchedulePermissions_ReadWrite_All',
    ScheduleReadAll = 'AZMGSchedule_Read_All',
    ScheduleReadWriteAll = 'AZMGSchedule_ReadWrite_All',
    SearchConfigurationReadAll = 'AZMGSearchConfiguration_Read_All',
    SearchConfigurationReadWriteAll = 'AZMGSearchConfiguration_ReadWrite_All',
    SecurityActionsReadAll = 'AZMGSecurityActions_Read_All',
    SecurityActionsReadWriteAll = 'AZMGSecurityActions_ReadWrite_All',
    SecurityAlertReadAll = 'AZMGSecurityAlert_Read_All',
    SecurityAlertReadWriteAll = 'AZMGSecurityAlert_ReadWrite_All',
    SecurityAnalyzedMessageReadAll = 'AZMGSecurityAnalyzedMessage_Read_All',
    SecurityAnalyzedMessageReadWriteAll = 'AZMGSecurityAnalyzedMessage_ReadWrite_All',
    SecurityCopilotWorkspacesReadAll = 'AZMGSecurityCopilotWorkspaces_Read_All',
    SecurityCopilotWorkspacesReadWriteAll = 'AZMGSecurityCopilotWorkspaces_ReadWrite_All',
    SecurityEventsReadAll = 'AZMGSecurityEvents_Read_All',
    SecurityEventsReadWriteAll = 'AZMGSecurityEvents_ReadWrite_All',
    SecurityIdentitiesAccountReadAll = 'AZMGSecurityIdentitiesAccount_Read_All',
    SecurityIdentitiesActionsReadWriteAll = 'AZMGSecurityIdentitiesActions_ReadWrite_All',
    SecurityIdentitiesHealthReadAll = 'AZMGSecurityIdentitiesHealth_Read_All',
    SecurityIdentitiesHealthReadWriteAll = 'AZMGSecurityIdentitiesHealth_ReadWrite_All',
    SecurityIdentitiesSensorsReadAll = 'AZMGSecurityIdentitiesSensors_Read_All',
    SecurityIdentitiesSensorsReadWriteAll = 'AZMGSecurityIdentitiesSensors_ReadWrite_All',
    SecurityIdentitiesUserActionsReadAll = 'AZMGSecurityIdentitiesUserActions_Read_All',
    SecurityIdentitiesUserActionsReadWriteAll = 'AZMGSecurityIdentitiesUserActions_ReadWrite_All',
    SecurityIncidentReadAll = 'AZMGSecurityIncident_Read_All',
    SecurityIncidentReadWriteAll = 'AZMGSecurityIncident_ReadWrite_All',
    SensitivityLabelEvaluateAll = 'AZMGSensitivityLabel_Evaluate_All',
    SensitivityLabelsReadAll = 'AZMGSensitivityLabels_Read_All',
    ServiceHealthReadAll = 'AZMGServiceHealth_Read_All',
    ServiceMessageReadAll = 'AZMGServiceMessage_Read_All',
    ServicePrincipalEndpointReadAll = 'AZMGServicePrincipalEndpoint_Read_All',
    ServicePrincipalEndpointReadWriteAll = 'AZMGServicePrincipalEndpoint_ReadWrite_All',
    SharePointTenantSettingsReadAll = 'AZMGSharePointTenantSettings_Read_All',
    SharePointTenantSettingsReadWriteAll = 'AZMGSharePointTenantSettings_ReadWrite_All',
    ShortNotesReadAll = 'AZMGShortNotes_Read_All',
    ShortNotesReadWriteAll = 'AZMGShortNotes_ReadWrite_All',
    SignInIdentifierReadAll = 'AZMGSignInIdentifier_Read_All',
    SignInIdentifierReadWriteAll = 'AZMGSignInIdentifier_ReadWrite_All',
    SitesArchiveAll = 'AZMGSites_Archive_All',
    SitesFullControlAll = 'AZMGSites_FullControl_All',
    SitesManageAll = 'AZMGSites_Manage_All',
    SitesReadAll = 'AZMGSites_Read_All',
    SitesReadWriteAll = 'AZMGSites_ReadWrite_All',
    SpiffeTrustDomainReadAll = 'AZMGSpiffeTrustDomain_Read_All',
    SpiffeTrustDomainReadWriteAll = 'AZMGSpiffeTrustDomain_ReadWrite_All',
    StorylineReadWriteAll = 'AZMGStoryline_ReadWrite_All',
    SubjectRightsRequestReadAll = 'AZMGSubjectRightsRequest_Read_All',
    SubjectRightsRequestReadWriteAll = 'AZMGSubjectRightsRequest_ReadWrite_All',
    SubscriptionReadAll = 'AZMGSubscription_Read_All',
    SynchronizationReadAll = 'AZMGSynchronization_Read_All',
    SynchronizationReadWriteAll = 'AZMGSynchronization_ReadWrite_All',
    TasksReadAll = 'AZMGTasks_Read_All',
    TasksReadShared = 'AZMGTasks_Read_Shared',
    TasksReadWriteAll = 'AZMGTasks_ReadWrite_All',
    TasksReadWriteShared = 'AZMGTasks_ReadWrite_Shared',
    TeamMemberReadAll = 'AZMGTeamMember_Read_All',
    TeamMemberReadWriteAll = 'AZMGTeamMember_ReadWrite_All',
    TeamMemberReadWriteNonOwnerRoleAll = 'AZMGTeamMember_ReadWriteNonOwnerRole_All',
    TeamReadBasicAll = 'AZMGTeam_ReadBasic_All',
    TeamSettingsReadAll = 'AZMGTeamSettings_Read_All',
    TeamSettingsReadWriteAll = 'AZMGTeamSettings_ReadWrite_All',
    TeamTemplatesReadAll = 'AZMGTeamTemplates_Read_All',
    TeamsActivityReadAll = 'AZMGTeamsActivity_Read_All',
    TeamsAppInstallationManageSelectedForChatAll = 'AZMGTeamsAppInstallation_ManageSelectedForChat_All',
    TeamsAppInstallationManageSelectedForTeamAll = 'AZMGTeamsAppInstallation_ManageSelectedForTeam_All',
    TeamsAppInstallationManageSelectedForUserAll = 'AZMGTeamsAppInstallation_ManageSelectedForUser_All',
    TeamsAppInstallationReadAll = 'AZMGTeamsAppInstallation_Read_All',
    TeamsAppInstallationReadForChatAll = 'AZMGTeamsAppInstallation_ReadForChat_All',
    TeamsAppInstallationReadForTeamAll = 'AZMGTeamsAppInstallation_ReadForTeam_All',
    TeamsAppInstallationReadForUserAll = 'AZMGTeamsAppInstallation_ReadForUser_All',
    TeamsAppInstallationReadSelectedForChatAll = 'AZMGTeamsAppInstallation_ReadSelectedForChat_All',
    TeamsAppInstallationReadSelectedForTeamAll = 'AZMGTeamsAppInstallation_ReadSelectedForTeam_All',
    TeamsAppInstallationReadSelectedForUserAll = 'AZMGTeamsAppInstallation_ReadSelectedForUser_All',
    TeamsAppInstallationReadWriteAndConsentForChatAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForChat_All',
    TeamsAppInstallationReadWriteAndConsentForTeamAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForTeam_All',
    TeamsAppInstallationReadWriteAndConsentForUserAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForUser_All',
    TeamsAppInstallationReadWriteAndConsentSelfForChatAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForChat_All',
    TeamsAppInstallationReadWriteAndConsentSelfForTeamAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForTeam_All',
    TeamsAppInstallationReadWriteAndConsentSelfForUserAll = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForUser_All',
    TeamsAppInstallationReadWriteForChatAll = 'AZMGTeamsAppInstallation_ReadWriteForChat_All',
    TeamsAppInstallationReadWriteForTeamAll = 'AZMGTeamsAppInstallation_ReadWriteForTeam_All',
    TeamsAppInstallationReadWriteForUserAll = 'AZMGTeamsAppInstallation_ReadWriteForUser_All',
    TeamsAppInstallationReadWriteSelectedForChatAll = 'AZMGTeamsAppInstallation_ReadWriteSelectedForChat_All',
    TeamsAppInstallationReadWriteSelectedForTeamAll = 'AZMGTeamsAppInstallation_ReadWriteSelectedForTeam_All',
    TeamsAppInstallationReadWriteSelectedForUserAll = 'AZMGTeamsAppInstallation_ReadWriteSelectedForUser_All',
    TeamsAppInstallationReadWriteSelfForChatAll = 'AZMGTeamsAppInstallation_ReadWriteSelfForChat_All',
    TeamsAppInstallationReadWriteSelfForTeamAll = 'AZMGTeamsAppInstallation_ReadWriteSelfForTeam_All',
    TeamsAppInstallationReadWriteSelfForUserAll = 'AZMGTeamsAppInstallation_ReadWriteSelfForUser_All',
    TeamsPolicyUserAssignReadWriteAll = 'AZMGTeamsPolicyUserAssign_ReadWrite_All',
    TeamsResourceAccountReadAll = 'AZMGTeamsResourceAccount_Read_All',
    TeamsTabReadAll = 'AZMGTeamsTab_Read_All',
    TeamsTabReadWriteAll = 'AZMGTeamsTab_ReadWrite_All',
    TeamsTabReadWriteForChatAll = 'AZMGTeamsTab_ReadWriteForChat_All',
    TeamsTabReadWriteForTeamAll = 'AZMGTeamsTab_ReadWriteForTeam_All',
    TeamsTabReadWriteForUserAll = 'AZMGTeamsTab_ReadWriteForUser_All',
    TeamsTabReadWriteSelfForChatAll = 'AZMGTeamsTab_ReadWriteSelfForChat_All',
    TeamsTabReadWriteSelfForTeamAll = 'AZMGTeamsTab_ReadWriteSelfForTeam_All',
    TeamsTabReadWriteSelfForUserAll = 'AZMGTeamsTab_ReadWriteSelfForUser_All',
    TeamsTelephoneNumberReadAll = 'AZMGTeamsTelephoneNumber_Read_All',
    TeamsTelephoneNumberReadWriteAll = 'AZMGTeamsTelephoneNumber_ReadWrite_All',
    TeamsUserConfigurationReadAll = 'AZMGTeamsUserConfiguration_Read_All',
    TeamworkAppSettingsReadAll = 'AZMGTeamworkAppSettings_Read_All',
    TeamworkAppSettingsReadWriteAll = 'AZMGTeamworkAppSettings_ReadWrite_All',
    TeamworkDeviceReadAll = 'AZMGTeamworkDevice_Read_All',
    TeamworkDeviceReadWriteAll = 'AZMGTeamworkDevice_ReadWrite_All',
    TeamworkMigrateAll = 'AZMGTeamwork_Migrate_All',
    TeamworkReadAll = 'AZMGTeamwork_Read_All',
    TeamworkTagReadAll = 'AZMGTeamworkTag_Read_All',
    TeamworkTagReadWriteAll = 'AZMGTeamworkTag_ReadWrite_All',
    TeamworkUserInteractionReadAll = 'AZMGTeamworkUserInteraction_Read_All',
    TermStoreReadAll = 'AZMGTermStore_Read_All',
    TermStoreReadWriteAll = 'AZMGTermStore_ReadWrite_All',
    ThreatAssessmentReadAll = 'AZMGThreatAssessment_Read_All',
    ThreatAssessmentReadWriteAll = 'AZMGThreatAssessment_ReadWrite_All',
    ThreatHuntingReadAll = 'AZMGThreatHunting_Read_All',
    ThreatIndicatorsReadAll = 'AZMGThreatIndicators_Read_All',
    ThreatIndicatorsReadWriteOwnedBy = 'AZMGThreatIndicators_ReadWrite_OwnedBy',
    ThreatIntelligenceReadAll = 'AZMGThreatIntelligence_Read_All',
    ThreatSubmissionPolicyReadWriteAll = 'AZMGThreatSubmissionPolicy_ReadWrite_All',
    ThreatSubmissionReadAll = 'AZMGThreatSubmission_Read_All',
    ThreatSubmissionReadWriteAll = 'AZMGThreatSubmission_ReadWrite_All',
    TopicReadAll = 'AZMGTopic_Read_All',
    TrustFrameworkKeySetReadAll = 'AZMGTrustFrameworkKeySet_Read_All',
    TrustFrameworkKeySetReadWriteAll = 'AZMGTrustFrameworkKeySet_ReadWrite_All',
    UnifiedGroupMemberReadAsGuest = 'AZMGUnifiedGroupMember_Read_AsGuest',
    UserActivityReadWriteCreatedByApp = 'AZMGUserActivity_ReadWrite_CreatedByApp',
    UserAuthenticationMethodReadAll = 'AZMGUserAuthenticationMethod_Read_All',
    UserAuthenticationMethodReadWriteAll = 'AZMGUserAuthenticationMethod_ReadWrite_All',
    UserDeleteRestoreAll = 'AZMGUser_DeleteRestore_All',
    UserEnableDisableAccountAll = 'AZMGUser_EnableDisableAccount_All',
    UserExportAll = 'AZMGUser_Export_All',
    UserInviteAll = 'AZMGUser_Invite_All',
    UserManageIdentitiesAll = 'AZMGUser_ManageIdentities_All',
    UserNotificationReadWriteCreatedByApp = 'AZMGUserNotification_ReadWrite_CreatedByApp',
    UserReadAll = 'AZMGUser_Read_All',
    UserReadBasicAll = 'AZMGUser_ReadBasic_All',
    UserReadWriteAll = 'AZMGUser_ReadWrite_All',
    UserReadWriteCrossCloud = 'AZMGUser_ReadWrite_CrossCloud',
    UserRevokeSessionsAll = 'AZMGUser_RevokeSessions_All',
    UserShiftPreferencesReadAll = 'AZMGUserShiftPreferences_Read_All',
    UserShiftPreferencesReadWriteAll = 'AZMGUserShiftPreferences_ReadWrite_All',
    UserStateReadWriteAll = 'AZMGUserState_ReadWrite_All',
    UserTeamworkReadAll = 'AZMGUserTeamwork_Read_All',
    UserTimelineActivityWriteCreatedByApp = 'AZMGUserTimelineActivity_Write_CreatedByApp',
    UserWindowsSettingsReadAll = 'AZMGUserWindowsSettings_Read_All',
    UserWindowsSettingsReadWriteAll = 'AZMGUserWindowsSettings_ReadWrite_All',
    VirtualAppointmentReadAll = 'AZMGVirtualAppointment_Read_All',
    VirtualAppointmentReadWriteAll = 'AZMGVirtualAppointment_ReadWrite_All',
    VirtualEventReadAll = 'AZMGVirtualEvent_Read_All',
    WindowsUpdatesReadWriteAll = 'AZMGWindowsUpdates_ReadWrite_All',
    WorkforceIntegrationReadAll = 'AZMGWorkforceIntegration_Read_All',
    WorkforceIntegrationReadWriteAll = 'AZMGWorkforceIntegration_ReadWrite_All',
    AgentApplicationCreate = 'AZMGAgentApplication_Create',
AgentIdentityCreate = 'AZMGAgentIdentity_Create',
AgreementAcceptanceRead = 'AZMGAgreementAcceptance_Read',
AiEnterpriseInteractionRead = 'AZMGAiEnterpriseInteraction_Read',
AnalyticsRead = 'AZMGAnalytics_Read',
AppCatalogSubmit = 'AZMGAppCatalog_Submit',
ApprovalSolutionRead = 'AZMGApprovalSolution_Read',
ApprovalSolutionReadWrite = 'AZMGApprovalSolution_ReadWrite',
ApprovalSolutionResponseReadWrite = 'AZMGApprovalSolutionResponse_ReadWrite',
AuditActivityRead = 'AZMGAuditActivity_Read',
AuditActivityWrite = 'AZMGAuditActivity_Write',
CalendarsRead = 'AZMGCalendars_Read',
CalendarsReadBasic = 'AZMGCalendars_ReadBasic',
CalendarsReadWrite = 'AZMGCalendars_ReadWrite',
CallDelegationRead = 'AZMGCallDelegation_Read',
CallDelegationReadWrite = 'AZMGCallDelegation_ReadWrite',
CallEventsRead = 'AZMGCallEvents_Read',
ChannelCreate = 'AZMGChannel_Create',
ChannelMessageEdit = 'AZMGChannelMessage_Edit',
ChannelMessageReadWrite = 'AZMGChannelMessage_ReadWrite',
ChannelMessageSend = 'AZMGChannelMessage_Send',
ChatCreate = 'AZMGChat_Create',
ChatMemberRead = 'AZMGChatMember_Read',
ChatMemberReadWrite = 'AZMGChatMember_ReadWrite',
ChatMessageRead = 'AZMGChatMessage_Read',
ChatMessageSend = 'AZMGChatMessage_Send',
ChatRead = 'AZMGChat_Read',
ChatReadBasic = 'AZMGChat_ReadBasic',
ChatReadWrite = 'AZMGChat_ReadWrite',
ConsentRequestCreate = 'AZMGConsentRequest_Create',
ConsentRequestRead = 'AZMGConsentRequest_Read',
ContactsRead = 'AZMGContacts_Read',
ContactsReadWrite = 'AZMGContacts_ReadWrite',
ContentActivityRead = 'AZMGContentActivity_Read',
ContentActivityWrite = 'AZMGContentActivity_Write',
CrossTenantUserProfileSharingRead = 'AZMGCrossTenantUserProfileSharing_Read',
CrossTenantUserProfileSharingReadWrite = 'AZMGCrossTenantUserProfileSharing_ReadWrite',
DeviceCommand = 'AZMGDevice_Command',
DeviceCreateFromOwnedTemplate = 'AZMGDevice_CreateFromOwnedTemplate',
DeviceRead = 'AZMGDevice_Read',
DeviceTemplateCreate = 'AZMGDeviceTemplate_Create',
EduAdministrationRead = 'AZMGEduAdministration_Read',
EduAdministrationReadWrite = 'AZMGEduAdministration_ReadWrite',
EduAssignmentsRead = 'AZMGEduAssignments_Read',
EduAssignmentsReadBasic = 'AZMGEduAssignments_ReadBasic',
EduAssignmentsReadWrite = 'AZMGEduAssignments_ReadWrite',
EduAssignmentsReadWriteBasic = 'AZMGEduAssignments_ReadWriteBasic',
EduCurriculaRead = 'AZMGEduCurricula_Read',
EduCurriculaReadWrite = 'AZMGEduCurricula_ReadWrite',
EduRosterRead = 'AZMGEduRoster_Read',
EduRosterReadBasic = 'AZMGEduRoster_ReadBasic',
EduRosterReadWrite = 'AZMGEduRoster_ReadWrite',
EngagementRoleRead = 'AZMGEngagementRole_Read',
FamilyRead = 'AZMGFamily_Read',
FileIngestionHybridOnboardingManage = 'AZMGFileIngestionHybridOnboarding_Manage',
FileIngestionIngest = 'AZMGFileIngestion_Ingest',
FileStorageContainerSelected = 'AZMGFileStorageContainer_Selected',
FileStorageContainerTypeRegSelected = 'AZMGFileStorageContainerTypeReg_Selected',
FilesRead = 'AZMGFiles_Read',
FilesReadWrite = 'AZMGFiles_ReadWrite',
GroupCreate = 'AZMGGroup_Create',
GroupConversationReadWriteAll = 'AZMGGroupConversation_ReadWrite_All',
InformationProtectionConfigRead = 'AZMGInformationProtectionConfig_Read',
InformationProtectionPolicyRead = 'AZMGInformationProtectionPolicy_Read',
LearningAssignedCourseRead = 'AZMGLearningAssignedCourse_Read',
LearningProviderRead = 'AZMGLearningProvider_Read',
LearningProviderReadWrite = 'AZMGLearningProvider_ReadWrite',
LearningSelfInitiatedCourseRead = 'AZMGLearningSelfInitiatedCourse_Read',
MailRead = 'AZMGMail_Read',
MailReadBasic = 'AZMGMail_ReadBasic',
MailReadWrite = 'AZMGMail_ReadWrite',
MailSend = 'AZMGMail_Send',
MailboxFolderRead = 'AZMGMailboxFolder_Read',
MailboxFolderReadWrite = 'AZMGMailboxFolder_ReadWrite',
MailboxItemImportExport = 'AZMGMailboxItem_ImportExport',
MailboxItemRead = 'AZMGMailboxItem_Read',
MailboxSettingsRead = 'AZMGMailboxSettings_Read',
MailboxSettingsReadWrite = 'AZMGMailboxSettings_ReadWrite',
NotesCreate = 'AZMGNotes_Create',
NotesRead = 'AZMGNotes_Read',
NotesReadWrite = 'AZMGNotes_ReadWrite',
OnlineMeetingsRead = 'AZMGOnlineMeetings_Read',
OnlineMeetingsReadWrite = 'AZMGOnlineMeetings_ReadWrite',
PeopleRead = 'AZMGPeople_Read',
PresenceRead = 'AZMGPresence_Read',
PresenceReadWrite = 'AZMGPresence_ReadWrite',
PrintJobCreate = 'AZMGPrintJob_Create',
PrintJobRead = 'AZMGPrintJob_Read',
PrintJobReadBasic = 'AZMGPrintJob_ReadBasic',
PrintJobReadWrite = 'AZMGPrintJob_ReadWrite',
PrintJobReadWriteBasic = 'AZMGPrintJob_ReadWriteBasic',
PrinterCreate = 'AZMGPrinter_Create',
ResourceSpecificPermissionGrantReadForChat = 'AZMGResourceSpecificPermissionGrant_ReadForChat',
ResourceSpecificPermissionGrantReadForTeam = 'AZMGResourceSpecificPermissionGrant_ReadForTeam',
ResourceSpecificPermissionGrantReadForUser = 'AZMGResourceSpecificPermissionGrant_ReadForUser',
SMTPSend = 'AZMGSMTP_Send',
SensitivityLabelEvaluate = 'AZMGSensitivityLabel_Evaluate',
SensitivityLabelRead = 'AZMGSensitivityLabel_Read',
ServiceMessageViewpointWrite = 'AZMGServiceMessageViewpoint_Write',
ShortNotesRead = 'AZMGShortNotes_Read',
ShortNotesReadWrite = 'AZMGShortNotes_ReadWrite',
SitesSelected = 'AZMGSites_Selected',
TasksRead = 'AZMGTasks_Read',
TasksReadWrite = 'AZMGTasks_ReadWrite',
TeamCreate = 'AZMGTeam_Create',
TeamTemplatesRead = 'AZMGTeamTemplates_Read',
TeamsActivityRead = 'AZMGTeamsActivity_Read',
TeamsActivitySend = 'AZMGTeamsActivity_Send',
TeamsAppInstallationManageSelectedForChat = 'AZMGTeamsAppInstallation_ManageSelectedForChat',
TeamsAppInstallationManageSelectedForTeam = 'AZMGTeamsAppInstallation_ManageSelectedForTeam',
TeamsAppInstallationManageSelectedForUser = 'AZMGTeamsAppInstallation_ManageSelectedForUser',
TeamsAppInstallationReadForChat = 'AZMGTeamsAppInstallation_ReadForChat',
TeamsAppInstallationReadForTeam = 'AZMGTeamsAppInstallation_ReadForTeam',
TeamsAppInstallationReadForUser = 'AZMGTeamsAppInstallation_ReadForUser',
TeamsAppInstallationReadSelectedForChat = 'AZMGTeamsAppInstallation_ReadSelectedForChat',
TeamsAppInstallationReadSelectedForTeam = 'AZMGTeamsAppInstallation_ReadSelectedForTeam',
TeamsAppInstallationReadSelectedForUser = 'AZMGTeamsAppInstallation_ReadSelectedForUser',
TeamsAppInstallationReadWriteAndConsentForChat = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForChat',
TeamsAppInstallationReadWriteAndConsentForTeam = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForTeam',
TeamsAppInstallationReadWriteAndConsentForUser = 'AZMGTeamsAppInstallation_ReadWriteAndConsentForUser',
TeamsAppInstallationReadWriteAndConsentSelfForChat = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForChat',
TeamsAppInstallationReadWriteAndConsentSelfForTeam = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForTeam',
TeamsAppInstallationReadWriteAndConsentSelfForUser = 'AZMGTeamsAppInstallation_ReadWriteAndConsentSelfForUser',
TeamsAppInstallationReadWriteForChat = 'AZMGTeamsAppInstallation_ReadWriteForChat',
TeamsAppInstallationReadWriteForTeam = 'AZMGTeamsAppInstallation_ReadWriteForTeam',
TeamsAppInstallationReadWriteForUser = 'AZMGTeamsAppInstallation_ReadWriteForUser',
TeamsAppInstallationReadWriteSelectedForChat = 'AZMGTeamsAppInstallation_ReadWriteSelectedForChat',
TeamsAppInstallationReadWriteSelectedForTeam = 'AZMGTeamsAppInstallation_ReadWriteSelectedForTeam',
TeamsAppInstallationReadWriteSelectedForUser = 'AZMGTeamsAppInstallation_ReadWriteSelectedForUser',
TeamsAppInstallationReadWriteSelfForChat = 'AZMGTeamsAppInstallation_ReadWriteSelfForChat',
TeamsAppInstallationReadWriteSelfForTeam = 'AZMGTeamsAppInstallation_ReadWriteSelfForTeam',
TeamsAppInstallationReadWriteSelfForUser = 'AZMGTeamsAppInstallation_ReadWriteSelfForUser',
TeamsTabCreate = 'AZMGTeamsTab_Create',
TeamsTabReadWriteForChat = 'AZMGTeamsTab_ReadWriteForChat',
TeamsTabReadWriteForTeam = 'AZMGTeamsTab_ReadWriteForTeam',
TeamsTabReadWriteForUser = 'AZMGTeamsTab_ReadWriteForUser',
TeamsTabReadWriteSelfForChat = 'AZMGTeamsTab_ReadWriteSelfForChat',
TeamsTabReadWriteSelfForTeam = 'AZMGTeamsTab_ReadWriteSelfForTeam',
TeamsTabReadWriteSelfForUser = 'AZMGTeamsTab_ReadWriteSelfForUser',
TeamworkTagRead = 'AZMGTeamworkTag_Read',
TeamworkTagReadWrite = 'AZMGTeamworkTag_ReadWrite',
ThreatSubmissionRead = 'AZMGThreatSubmission_Read',
ThreatSubmissionReadWrite = 'AZMGThreatSubmission_ReadWrite',
UserAuthenticationMethodRead = 'AZMGUserAuthenticationMethod_Read',
UserAuthenticationMethodReadWrite = 'AZMGUserAuthenticationMethod_ReadWrite',
UserCloudClipboardRead = 'AZMGUserCloudClipboard_Read',
UserRead = 'AZMGUser_Read',
UserReadWrite = 'AZMGUser_ReadWrite',
UserTeamworkRead = 'AZMGUserTeamwork_Read',
VirtualAppointmentNotificationSend = 'AZMGVirtualAppointmentNotification_Send',
VirtualAppointmentRead = 'AZMGVirtualAppointment_Read',
VirtualAppointmentReadWrite = 'AZMGVirtualAppointment_ReadWrite',
VirtualEventRead = 'AZMGVirtualEvent_Read',
VirtualEventReadWrite = 'AZMGVirtualEvent_ReadWrite',

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
        case AzureRelationshipKind.SyncedToADUser:
            return 'SyncedToADUser';
        case AzureRelationshipKind.AZRoleEligible:
            return 'AZRoleEligible';
        case AzureRelationshipKind.AZRoleApprover:
            return 'AZRoleApprover';
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
    LoginURL = 'loginurl',
    MFAEnforced = 'mfaenforced',
    UserPrincipalName = 'userprincipalname',
    IsAssignableToRole = 'isassignabletorole',
    PublisherDomain = 'publisherdomain',
    SignInAudience = 'signinaudience',
    RoleTemplateID = 'templateid',
    RoleDefinitionId = 'roledefinitionid',
    EndUserAssignmentRequiresApproval = 'enduserassignmentrequiresapproval',
    EndUserAssignmentRequiresCAPAuthenticationContext = 'enduserassignmentrequirescapauthenticationcontext',
    EndUserAssignmentUserApprovers = 'enduserassignmentuserapprovers',
    EndUserAssignmentGroupApprovers = 'enduserassignmentgroupapprovers',
    EndUserAssignmentRequiresMFA = 'enduserassignmentrequiresmfa',
    EndUserAssignmentRequiresJustification = 'enduserassignmentrequiresjustification',
    EndUserAssignmentRequiresTicketInformation = 'enduserassignmentrequiresticketinformation',
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
        case AzureKindProperties.LoginURL:
            return 'Login URL';
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
        case AzureKindProperties.RoleDefinitionId:
            return 'Role Definition Id';
        case AzureKindProperties.EndUserAssignmentRequiresApproval:
            return 'End User Assignment Requires Approval';
        case AzureKindProperties.EndUserAssignmentRequiresCAPAuthenticationContext:
            return 'End User Assignment Requires CAP Authentication Context';
        case AzureKindProperties.EndUserAssignmentUserApprovers:
            return 'End User Assignment User Approvers';
        case AzureKindProperties.EndUserAssignmentGroupApprovers:
            return 'End User Assignment Group Approvers';
        case AzureKindProperties.EndUserAssignmentRequiresMFA:
            return 'End User Assignment Requires MFA';
        case AzureKindProperties.EndUserAssignmentRequiresJustification:
            return 'End User Assignment Requires Justification';
        case AzureKindProperties.EndUserAssignmentRequiresTicketInformation:
            return 'End User Assignment Requires Ticket Information';
        default:
            return undefined;
    }
}
export function AzurePathfindingEdges(): AzureRelationshipKind[] {
    return [
        AzureRelationshipKind.AvereContributor,
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
        AzureRelationshipKind.SyncedToADUser,
        AzureRelationshipKind.AZRoleEligible,
        AzureRelationshipKind.AZRoleApprover,
        AzureRelationshipKind.Contains,
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
    LastCollected = 'lastcollected',
    WhenCreated = 'whencreated',
    Enabled = 'enabled',
    PasswordLastSet = 'pwdlastset',
    Title = 'title',
    Email = 'email',
    IsInherited = 'isinherited',
    CompositionID = 'compositionid',
    PrimaryKind = 'primarykind',
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
            return 'Last Seen by BloodHound';
        case CommonKindProperties.LastCollected:
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
        case CommonKindProperties.CompositionID:
            return 'Composition ID';
        case CommonKindProperties.PrimaryKind:
            return 'Primary Kind';
        default:
            return undefined;
    }
}
