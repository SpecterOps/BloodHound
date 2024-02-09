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

package ad

import "pkg.specterops.io/schemas/bh/types:types"

// Exported requirements
Properties: [...types.#StringEnum]
NodeKinds: [...types.#Kind]
RelationshipKinds: [...types.#Kind]
ACLRelationships: [...types.#Kind]
PathfindingRelationships: [...types.#Kind]
EdgeCompositionRelationships: [...types.#Kind]

// Property name enumerations

CertChain: types.#StringEnum & {
	symbol: 		"CertChain"
	schema: 		"ad"
	name:           "Certificate Chain"
	representation: "certchain"
}

CertName: types.#StringEnum & {
	symbol: 		"CertName"
	schema: 		"ad"
	name:           "Certificate Name"
	representation: "certname"
}

CertThumbprint: types.#StringEnum & {
	symbol: 		"CertThumbprint"
	schema: 		"ad"
	name:           "Certificate Thumbprint"
	representation: "certthumbprint"
}

CertThumbprints: types.#StringEnum & {
	symbol: 		"CertThumbprints"
	schema: 		"ad"
	name:           "Certificate Thumbprints"
	representation: "certthumbprints"
}

CAName: types.#StringEnum & {
	symbol: 		"CAName"
	schema: 		"ad"
	name:           "CA Name"
	representation: "caname"
}

CASecurityCollected: types.#StringEnum & {
	symbol: 		"CASecurityCollected"
	schema: 		"ad"
	name:           "CA Security Collected"
	representation: "casecuritycollected"
}

HasEnrollmentAgentRestrictions: types.#StringEnum & {
	symbol: 		"HasEnrollmentAgentRestrictions"
	schema: 		"ad"
	name:           "Has Enrollment Agent Restrictions"
	representation: "hasenrollmentagentrestrictions"
}

EnrollmentAgentRestrictionsCollected: types.#StringEnum & {
	symbol: 		"EnrollmentAgentRestrictionsCollected"
	schema: 		"ad"
	name:           "Enrollment Agent Restrictions Collected"
	representation: "enrollmentagentrestrictionscollected"
}

IsUserSpecifiesSanEnabled: types.#StringEnum & {
	symbol: 		"IsUserSpecifiesSanEnabled"
	schema: 		"ad"
	name:           "Is User Specifies San Enabled"
	representation: "isuserspecifiessanenabled"
}

IsUserSpecifiesSanEnabledCollected: types.#StringEnum & {
	symbol: 		"IsUserSpecifiesSanEnabledCollected"
	schema: 		"ad"
	name:           "Is User Specifies San Enabled Collected"
	representation: "isuserspecifiessanenabledcollected"
}

HasBasicConstraints: types.#StringEnum & {
	symbol: 		"HasBasicConstraints"
	schema: 		"ad"
	name:           "Has Basic Constraints"
	representation: "hasbasicconstraints"
}

BasicConstraintPathLength: types.#StringEnum & {
	symbol: 		"BasicConstraintPathLength"
	schema: 		"ad"
	name:           "Basic Constraint Path Length"
	representation: "basicconstraintpathlength"
}

DNSHostname: types.#StringEnum & {
	symbol: 		"DNSHostname"
	schema: 		"ad"
	name:           "DNS Hostname"
	representation: "dnshostname"
}

DistinguishedName: types.#StringEnum & {
	symbol:         "DistinguishedName"
	schema:         "ad"
	name:           "Distinguished Name"
	representation: "distinguishedname"
}

DomainFQDN: types.#StringEnum & {
	symbol:         "DomainFQDN"
	schema:         "ad"
	name:           "Domain FQDN"
	representation: "domain"
}

DomainSID: types.#StringEnum & {
	symbol:         "DomainSID"
	schema:         "ad"
	name:           "Domain SID"
	representation: "domainsid"
}

Sensitive: types.#StringEnum & {
	symbol:         "Sensitive"
	schema:         "ad"
	name:           "Marked Sensitive"
	representation: "sensitive"
}

HighValue: types.#StringEnum & {
	symbol:         "HighValue"
	schema:         "ad"
	name:           "High Value"
	representation: "highvalue"
}

BlocksInheritance: types.#StringEnum & {
	symbol:         "BlocksInheritance"
	schema:         "ad"
	name:           "Blocks Inheritance"
	representation: "blocksinheritance"
}

IsACL: types.#StringEnum & {
	symbol:         "IsACL"
	schema:         "ad"
	name:           "Is ACL"
	representation: "isacl"
}

IsACLProtected: types.#StringEnum & {
	symbol:         "IsACLProtected"
	schema:         "ad"
	name:           "ACL Inheritance Denied"
	representation: "isaclprotected"
}

IsDeleted: types.#StringEnum & {
	symbol:         "IsDeleted"
	schema:         "ad"
	name:           "Is Deleted"
	representation: "isdeleted"
}

Enforced: types.#StringEnum & {
	symbol:         "Enforced"
	schema:         "ad"
	name:           "Enforced"
	representation: "enforced"
}

LogonType: types.#StringEnum & {
	symbol:         "LogonType"
	schema:         "ad"
	name:           "Logon Type"
	representation: "logontype"
}

Department: types.#StringEnum & {
	symbol:         "Department"
	schema:         "ad"
	name:           "Department"
	representation: "department"
}

HasCrossCertificatePair: types.#StringEnum & {
	symbol:         "HasCrossCertificatePair"
	schema:         "ad"
	name:           "Has Cross Certificate Pair"
	representation: "hascrosscertificatepair"
}

HasSPN: types.#StringEnum & {
	symbol:         "HasSPN"
	schema:         "ad"
	name:           "Has SPN"
	representation: "hasspn"
}

HasLAPS: types.#StringEnum & {
	symbol: "HasLAPS"
	schema: "ad"
	name: "LAPS Enabled"
	representation: "haslaps"
}

UnconstrainedDelegation: types.#StringEnum & {
	symbol:         "UnconstrainedDelegation"
	schema:         "ad"
	name:           "Allows Unconstrained Delegation"
	representation: "unconstraineddelegation"
}

LastLogon: types.#StringEnum & {
	symbol:         "LastLogon"
	schema:         "ad"
	name:           "Last Logon"
	representation: "lastlogon"
}

LastLogonTimestamp: types.#StringEnum & {
	symbol:         "LastLogonTimestamp"
	schema:         "ad"
	name:           "Last Logon (Replicated)"
	representation: "lastlogontimestamp"
}

IsPrimaryGroup: types.#StringEnum & {
	symbol:         "IsPrimaryGroup"
	schema:         "ad"
	name:           "Is Primary Group"
	representation: "isprimarygroup"
}

AdminCount: types.#StringEnum & {
	symbol:         "AdminCount"
	schema:         "ad"
	name:           "Admin Count"
	representation: "admincount"
}

DontRequirePreAuth: types.#StringEnum & {
	symbol: "DontRequirePreAuth"
	schema: "ad"
	name: "Do Not Require Pre-Authentication"
	representation: "dontreqpreauth"
}

HasURA: types.#StringEnum & {
	symbol: "HasURA"
	schema: "ad"
	name: "Has User Rights Assignment Collection"
	representation: "hasura"
}

PasswordNeverExpires: types.#StringEnum & {
	symbol: "PasswordNeverExpires"
	schema: "ad"
	name: "Password Never Expires"
	representation: "pwdneverexpires"
}

PasswordNotRequired: types.#StringEnum & {
	symbol: "PasswordNotRequired"
	schema: "ad"
	name: "Password Not Required"
	representation: "passwordnotreqd"
}

FunctionalLevel: types.#StringEnum & {
	symbol: "FunctionalLevel"
	schema: "ad"
	name: "Functional Level"
	representation: "functionallevel"
}

TrustType: types.#StringEnum & {
	symbol: "TrustType"
	schema: "ad"
	name: "Trust Type"
	representation: "trusttype"
}

SidFiltering: types.#StringEnum & {
	symbol: "SidFiltering"
	schema: "ad"
	name: "SID Filtering Enabled"
	representation: "sidfiltering"
}

TrustedToAuth: types.#StringEnum & {
	symbol: "TrustedToAuth"
	schema: "ad"
	name: "Trusted For Constrained Delegation"
	representation: "trustedtoauth"
}

SamAccountName: types.#StringEnum & {
	symbol: "SamAccountName"
	schema: "ad"
	name: "SAM Account Name"
	representation: "samaccountname"
}

HomeDirectory: types.#StringEnum & {
	symbol: "HomeDirectory"
	schema: "ad"
	name: "Home Directory"
	representation: "homedirectory"
}

CertificateMappingMethodsRaw: types.#StringEnum & {
	symbol:         "CertificateMappingMethodsRaw"
	schema:         "ad"
	name:           "Certificate Mapping Methods (Raw)"
	representation: "certificatemappingmethodsraw"
}

CertificateMappingMethods: types.#StringEnum & {
	symbol:         "CertificateMappingMethods"
	schema:         "ad"
	name:           "Certificate Mapping Methods"
	representation: "certificatemappingmethods"
}

StrongCertificateBindingEnforcementRaw: types.#StringEnum & {
	symbol:         "StrongCertificateBindingEnforcementRaw"
	schema:         "ad"
	name:           "Strong Certificate Binding Enforcement (Raw)"
	representation: "strongcertificatebindingenforcementraw"
}

StrongCertificateBindingEnforcement: types.#StringEnum & {
	symbol:         "StrongCertificateBindingEnforcement"
	schema:         "ad"
	name:           "Strong Certificate Binding Enforcement"
	representation: "strongcertificatebindingenforcement"
}

CrossCertificatePair: types.#StringEnum & {
	symbol: "CrossCertificatePair"
	schema: "ad"
	name: "Cross Certificate Pair"
	representation: "crosscertificatepair"
}

EKUs: types.#StringEnum & {
	symbol: "EKUs"
	schema: "ad"
	name: "Enhanced Key Usage"
	representation: "ekus"
}

SubjectAltRequireUPN: types.#StringEnum & {
	symbol: "SubjectAltRequireUPN"
	schema: "ad"
	name: "Subject Alternative Name Require UPN"
	representation: "subjectaltrequireupn"
}

SubjectAltRequireDNS: types.#StringEnum & {
	symbol: "SubjectAltRequireDNS"
	schema: "ad"
	name: "Subject Alternative Name Require DNS"
	representation: "subjectaltrequiredns"
}

SubjectAltRequireDomainDNS: types.#StringEnum & {
	symbol: "SubjectAltRequireDomainDNS"
	schema: "ad"
	name: "Subject Alternative Name Require Domain DNS"
	representation: "subjectaltrequiredomaindns"
}

SubjectAltRequireEmail: types.#StringEnum & {
	symbol: "SubjectAltRequireEmail"
	schema: "ad"
	name: "Subject Alternative Name Require Email"
	representation: "subjectaltrequireemail"
}

SubjectAltRequireSPN: types.#StringEnum & {
	symbol: "SubjectAltRequireSPN"
	schema: "ad"
	name: "Subject Alternative Name Require SPN"
	representation: "subjectaltrequirespn"
}

SubjectRequireEmail: types.#StringEnum & {
	symbol: "SubjectRequireEmail"
	schema: "ad"
	name: "Subject Require Email"
	representation: "subjectrequireemail"
}

AuthorizedSignatures: types.#StringEnum & {
	symbol: "AuthorizedSignatures"
	schema: "ad"
	name: "Authorized Signatures Required"
	representation: "authorizedsignatures"
}

ApplicationPolicies: types.#StringEnum & {
	symbol: "ApplicationPolicies"
	schema: "ad"
	name: "Application Policies"
	representation: "applicationpolicies"
}

IssuancePolicies: types.#StringEnum & {
	symbol: "IssuancePolicies"
	schema: "ad"
	name: "Issuance Policies"
	representation: "issuancepolicies"
}

SchemaVersion: types.#StringEnum & {
	symbol: "SchemaVersion"
	schema: "ad"
	name: "Schema Version"
	representation: "schemaversion"
}

RequiresManagerApproval: types.#StringEnum & {
	symbol: "RequiresManagerApproval"
	schema: "ad"
	name: "Requires Manager Approval"
	representation: "requiresmanagerapproval"
}

AuthenticationEnabled: types.#StringEnum & {
	symbol: "AuthenticationEnabled"
	schema: "ad"
	name: "Authentication Enabled"
	representation: "authenticationenabled"
}

EnrolleeSuppliesSubject: types.#StringEnum & {
	symbol: "EnrolleeSuppliesSubject"
	schema: "ad"
	name: "Enrollee Supplies Subject"
	representation: "enrolleesuppliessubject"
}

CertificateApplicationPolicy: types.#StringEnum & {
	symbol: "CertificateApplicationPolicy"
	schema: "ad"
	name: "Certificate Application Policies"
	representation: "certificateapplicationpolicy"
}

CertificateNameFlag: types.#StringEnum & {
	symbol: "CertificateNameFlag"
	schema: "ad"
	name: "Certificate Name Flags"
	representation: "certificatenameflag"
}

EffectiveEKUs: types.#StringEnum & {
	symbol: "EffectiveEKUs"
	schema: "ad"
	name: "Effective EKUs"
	representation: "effectiveekus"
}

EnrollmentFlag: types.#StringEnum & {
	symbol: "EnrollmentFlag"
	schema: "ad"
	name: "Enrollment Flags"
	representation: "enrollmentflag"
}

Flags: types.#StringEnum & {
	symbol: "Flags"
	schema: "ad"
	name: "Flags"
	representation: "flags"
}

NoSecurityExtension: types.#StringEnum & {
	symbol: "NoSecurityExtension"
	schema: "ad"
	name: "No Security Extension"
	representation: "nosecurityextension"
}

RenewalPeriod: types.#StringEnum & {
	symbol: "RenewalPeriod"
	schema: "ad"
	name: "Renewal Period"
	representation: "renewalperiod"
}

ValidityPeriod: types.#StringEnum & {
	symbol: "ValidityPeriod"
	schema: "ad"
	name: "Validity Period"
	representation: "validityperiod"
}

OID: types.#StringEnum & {
	symbol: "OID"
	schema: "ad"
	name: "OID"
	representation: "oid"
}

Properties: [
	AdminCount,
	CASecurityCollected,
	CAName,
	CertChain,
	CertName,
	CertThumbprint,
	CertThumbprints,
	HasEnrollmentAgentRestrictions,
	EnrollmentAgentRestrictionsCollected,
	IsUserSpecifiesSanEnabled,
	IsUserSpecifiesSanEnabledCollected,
	HasBasicConstraints,
	BasicConstraintPathLength,
	DNSHostname,
	CrossCertificatePair,
	DistinguishedName,
	DomainFQDN,
	DomainSID,
	Sensitive,
	HighValue,
	BlocksInheritance,
	IsACL,
	IsACLProtected,
	IsDeleted,
	Enforced,
	Department,
	HasCrossCertificatePair,
	HasSPN,
	UnconstrainedDelegation,
	LastLogon,
	LastLogonTimestamp,
	IsPrimaryGroup,
	HasLAPS,
	DontRequirePreAuth,
	LogonType,
	HasURA,
	PasswordNeverExpires,
	PasswordNotRequired,
	FunctionalLevel,
	TrustType,
	SidFiltering,
	TrustedToAuth,
	SamAccountName,
	CertificateMappingMethodsRaw,
	CertificateMappingMethods,
	StrongCertificateBindingEnforcementRaw,
	StrongCertificateBindingEnforcement,
	EKUs,
	SubjectAltRequireUPN,
	SubjectAltRequireDNS,
	SubjectAltRequireDomainDNS,
	SubjectAltRequireEmail,
	SubjectAltRequireSPN,
	SubjectRequireEmail,
	AuthorizedSignatures,
	ApplicationPolicies,
	IssuancePolicies,
	SchemaVersion,
	RequiresManagerApproval,
	AuthenticationEnabled,
	EnrolleeSuppliesSubject,
	CertificateApplicationPolicy,
	CertificateNameFlag,
	EffectiveEKUs,
	EnrollmentFlag,
	Flags,
	NoSecurityExtension,
	RenewalPeriod,
	ValidityPeriod,
	OID,
	HomeDirectory
]

// Kinds
Entity: types.#Kind & {
	symbol:         "Entity"
	schema:         "active_directory"
	representation: "Base"
}

User: types.#Kind & {
	symbol: "User"
	schema: "active_directory"
}

Computer: types.#Kind & {
	symbol: "Computer"
	schema: "active_directory"
}

Group: types.#Kind & {
	symbol: "Group"
	schema: "active_directory"
}

LocalGroup: types.#Kind & {
	symbol:         "LocalGroup"
	schema:         "active_directory"
	representation: "ADLocalGroup"
}

LocalUser: types.#Kind & {
	symbol:         "LocalUser"
	schema:         "active_directory"
	representation: "ADLocalUser"
}

GPO: types.#Kind & {
	symbol: "GPO"
	schema: "active_directory"
}

OU: types.#Kind & {
	symbol: "OU"
	schema: "active_directory"
}

Container: types.#Kind & {
	symbol: "Container"
	schema: "active_directory"
}

Domain: types.#Kind & {
	symbol: "Domain"
	schema: "active_directory"
}

AIACA: types.#Kind & {
	symbol: "AIACA"
	schema: "active_directory"
}

RootCA: types.#Kind & {
	symbol: "RootCA"
	schema: "active_directory"
}

EnterpriseCA: types.#Kind & {
	symbol: "EnterpriseCA"
	schema: "active_directory"
}

NTAuthStore: types.#Kind & {
	symbol: "NTAuthStore"
	schema: "active_directory"
}

CertTemplate: types.#Kind & {
	symbol: "CertTemplate"
	schema: "active_directory"
}

NodeKinds: [
	Entity,
	User,
	Computer,
	Group,
	GPO,
	OU,
	Container,
	Domain,
	LocalGroup,
	LocalUser,
	AIACA,
	RootCA,
	EnterpriseCA,
	NTAuthStore,
	CertTemplate
]

Owns: types.#Kind & {
	symbol: "Owns"
	schema: "active_directory"
}

GenericAll: types.#Kind & {
	symbol: "GenericAll"
	schema: "active_directory"
}

GenericWrite: types.#Kind & {
	symbol: "GenericWrite"
	schema: "active_directory"
}

WriteOwner: types.#Kind & {
	symbol: "WriteOwner"
	schema: "active_directory"
}

WriteDACL: types.#Kind & {
	symbol:         "WriteDACL"
	schema:         "active_directory"
	representation: "WriteDacl"
}

MemberOf: types.#Kind & {
	symbol: "MemberOf"
	schema: "active_directory"
}

MemberOfLocalGroup: types.#Kind & {
	symbol: "MemberOfLocalGroup"
	schema: "active_directory"
}

LocalToComputer: types.#Kind & {
	symbol: "LocalToComputer"
	schema: "active_directory"
}

ForceChangePassword: types.#Kind & {
	symbol: "ForceChangePassword"
	schema: "active_directory"
}

AllExtendedRights: types.#Kind & {
	symbol: "AllExtendedRights"
	schema: "active_directory"
}

AddMember: types.#Kind & {
	symbol: "AddMember"
	schema: "active_directory"
}

HasSession: types.#Kind & {
	symbol: "HasSession"
	schema: "active_directory"
}

Contains: types.#Kind & {
	symbol: "Contains"
	schema: "active_directory"
}

GPLink: types.#Kind & {
	symbol: "GPLink"
	schema: "active_directory"
}

AllowedToDelegate: types.#Kind & {
	symbol: "AllowedToDelegate"
	schema: "active_directory"
}

GetChanges: types.#Kind & {
	symbol: "GetChanges"
	schema: "active_directory"
}

GetChangesAll: types.#Kind & {
	symbol: "GetChangesAll"
	schema: "active_directory"
}

TrustedBy: types.#Kind & {
	symbol: "TrustedBy"
	schema: "active_directory"
}

AllowedToAct: types.#Kind & {
	symbol: "AllowedToAct"
	schema: "active_directory"
}

AdminTo: types.#Kind & {
	symbol: "AdminTo"
	schema: "active_directory"
}

CanPSRemote: types.#Kind & {
	symbol: "CanPSRemote"
	schema: "active_directory"
}

CanRDP: types.#Kind & {
	symbol: "CanRDP"
	schema: "active_directory"
}

ExecuteDCOM: types.#Kind & {
	symbol: "ExecuteDCOM"
	schema: "active_directory"
}

HasSIDHistory: types.#Kind & {
	symbol: "HasSIDHistory"
	schema: "active_directory"
}

AddSelf: types.#Kind & {
	symbol: "AddSelf"
	schema: "active_directory"
}

DCSync: types.#Kind & {
	symbol: "DCSync"
	schema: "active_directory"
}

ReadLAPSPassword: types.#Kind & {
	symbol: "ReadLAPSPassword"
	schema: "active_directory"
}

ReadGMSAPassword: types.#Kind & {
	symbol: "ReadGMSAPassword"
	schema: "active_directory"
}

DumpSMSAPassword: types.#Kind & {
	symbol: "DumpSMSAPassword"
	schema: "active_directory"
}

SQLAdmin: types.#Kind & {
	symbol: "SQLAdmin"
	schema: "active_directory"
}

AddAllowedToAct: types.#Kind & {
	symbol: "AddAllowedToAct"
	schema: "active_directory"
}

WriteSPN: types.#Kind & {
	symbol: "WriteSPN"
	schema: "active_directory"
}

AddKeyCredentialLink: types.#Kind & {
	symbol: "AddKeyCredentialLink"
	schema: "active_directory"
}

RemoteInteractiveLogonPrivilege: types.#Kind & {
	symbol: "RemoteInteractiveLogonPrivilege"
	schema: "active_directory"
}

SyncLAPSPassword: types.#Kind & {
	symbol: "SyncLAPSPassword"
	schema: "active_directory"
}

WriteAccountRestrictions: types.#Kind & {
	symbol: "WriteAccountRestrictions"
	schema: "active_directory"
}

GetChangesInFilteredSet: types.#Kind & {
	symbol: "GetChangesInFilteredSet"
	schema: "active_directory"
}

RootCAFor: types.#Kind & {
	symbol: "RootCAFor"
	schema: "active_directory"
}

DCFor: types.#Kind & {
	symbol: "DCFor"
	schema: "active_directory"
}

PublishedTo: types.#Kind & {
	symbol: "PublishedTo"
	schema: "active_directory"
}

ManageCertificates: types.#Kind & {
	symbol: "ManageCertificates"
	schema: "active_directory"
}

ManageCA: types.#Kind & {
	symbol: "ManageCA"
	schema: "active_directory"
}

DelegatedEnrollmentAgent: types.#Kind & {
	symbol: "DelegatedEnrollmentAgent"
	schema: "active_directory"
}

Enroll: types.#Kind & {
	symbol: "Enroll"
	schema: "active_directory"
}

HostsCAService: types.#Kind & {
	symbol: "HostsCAService"
	schema: "active_directory"
}

WritePKIEnrollmentFlag: types.#Kind & {
	symbol: "WritePKIEnrollmentFlag"
	schema: "active_directory"
}

WritePKINameFlag: types.#Kind & {
	symbol: "WritePKINameFlag"
	schema: "active_directory"
}

NTAuthStoreFor: types.#Kind & {
	symbol: "NTAuthStoreFor"
	schema: "active_directory"
}

TrustedForNTAuth: types.#Kind & {
	symbol: "TrustedForNTAuth"
	schema: "active_directory"
}

EnterpriseCAFor: types.#Kind & {
	symbol: "EnterpriseCAFor"
	schema: "active_directory"
}

CanAbuseUPNCertMapping: types.#Kind & {
	symbol: "CanAbuseUPNCertMapping"
	schema: "active_directory"
}

CanAbuseWeakCertBinding: types.#Kind & {
	symbol: "CanAbuseWeakCertBinding"
	schema: "active_directory"
}

IssuedSignedBy: types.#Kind & {
	symbol: "IssuedSignedBy"
	schema: "active_directory"
}

GoldenCert: types.#Kind & {
	symbol: "GoldenCert"
	schema: "active_directory"
}

EnrollOnBehalfOf: types.#Kind & {
	symbol: "EnrollOnBehalfOf"
	schema: "active_directory"
}

ADCSESC1: types.#Kind & {
	symbol: "ADCSESC1"
	schema: "active_directory"
}

ADCSESC3: types.#Kind & {
	symbol: "ADCSESC3"
	schema: "active_directory"
}

ADCSESC4: types.#Kind & {
	symbol: "ADCSESC4"
	schema: "active_directory"
}

ADCSESC5: types.#Kind & {
	symbol: "ADCSESC5"
	schema: "active_directory"
}

ADCSESC6a: types.#Kind & {
	symbol: "ADCSESC6a"
	schema: "active_directory"
}

ADCSESC6b: types.#Kind & {
	symbol: "ADCSESC6b"
	schema: "active_directory"
}

ADCSESC7: types.#Kind & {
	symbol: "ADCSESC7"
	schema: "active_directory"
}

ADCSESC9a: types.#Kind & {
	symbol: "ADCSESC9a"
	schema: "active_directory"
}

ADCSESC9b: types.#Kind & {
	symbol: "ADCSESC9b"
	schema: "active_directory"
}

ADCSESC10a: types.#Kind & {
	symbol: "ADCSESC10a"
	schema: "active_directory"
}

ADCSESC10b: types.#Kind & {
	symbol: "ADCSESC10b"
	schema: "active_directory"
}

// Relationship Kinds
RelationshipKinds: [
	Owns,
	GenericAll,
	GenericWrite,
	WriteOwner,
	WriteDACL,
	MemberOf,
	ForceChangePassword,
	AllExtendedRights,
	AddMember,
	HasSession,
	Contains,
	GPLink,
	AllowedToDelegate,
	GetChanges,
	GetChangesAll,
	GetChangesInFilteredSet,
	TrustedBy,
	AllowedToAct,
	AdminTo,
	CanPSRemote,
	CanRDP,
	ExecuteDCOM,
	HasSIDHistory,
	AddSelf,
	DCSync,
	ReadLAPSPassword,
	ReadGMSAPassword,
	DumpSMSAPassword,
	SQLAdmin,
	AddAllowedToAct,
	WriteSPN,
	AddKeyCredentialLink,
	LocalToComputer,
	MemberOfLocalGroup,
	RemoteInteractiveLogonPrivilege,
	SyncLAPSPassword,
	WriteAccountRestrictions,
	RootCAFor,
	DCFor,
	PublishedTo,
	ManageCertificates,
	ManageCA,
	DelegatedEnrollmentAgent,
	Enroll,
	HostsCAService,
	WritePKIEnrollmentFlag,
	WritePKINameFlag,
	NTAuthStoreFor,
	TrustedForNTAuth,
	EnterpriseCAFor,
	CanAbuseUPNCertMapping,
	CanAbuseWeakCertBinding,
	IssuedSignedBy,
	GoldenCert,
	EnrollOnBehalfOf,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC5,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC7,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b
]

// ACL Relationships
ACLRelationships: [
	AllExtendedRights,
	ForceChangePassword,
	AddMember,
	AddAllowedToAct,
	GenericAll,
	WriteDACL,
	WriteOwner,
	GenericWrite,
	ReadLAPSPassword,
	ReadGMSAPassword,
	Owns,
	AddSelf,
	WriteSPN,
	AddKeyCredentialLink,
	GetChanges,
	GetChangesAll,
	GetChangesInFilteredSet,
	WriteAccountRestrictions,
	SyncLAPSPassword,
	DCSync,
	ManageCertificates,
	ManageCA,
	Enroll,
	WritePKIEnrollmentFlag,
	WritePKINameFlag
]

// Edges that are used in pathfinding
PathfindingRelationships: [
	Owns,
	GenericAll,
	GenericWrite,
	WriteOwner,
	WriteDACL,
	MemberOf,
	ForceChangePassword,
	AllExtendedRights,
	AddMember,
	HasSession,
	Contains,
	GPLink,
	AllowedToDelegate,
	TrustedBy,
	AllowedToAct,
	AdminTo,
	CanPSRemote,
	CanRDP,
	ExecuteDCOM,
	HasSIDHistory,
	AddSelf,
	DCSync,
	ReadLAPSPassword,
	ReadGMSAPassword,
	DumpSMSAPassword,
	SQLAdmin,
	AddAllowedToAct,
	WriteSPN,
	AddKeyCredentialLink,
	SyncLAPSPassword,
	WriteAccountRestrictions,
	GoldenCert,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC5,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC7,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b,
	DCFor
]

EdgeCompositionRelationships: [
	GoldenCert,
	ADCSESC1,
	ADCSESC3,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b,
]
