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

import (
	"list"
	"pkg.specterops.io/schemas/bh/types:types"
)

// Exported requirements
Properties: [...types.#StringEnum]
NodeKinds: [...types.#Kind]
RelationshipKinds: [...types.#Kind]
ACLRelationships: [...types.#Kind]
PathfindingRelationships: [...types.#Kind]
InboundRelationshipKinds: [...types.#Kind]
OutboundRelationshipKinds: [...types.#Kind]
EdgeCompositionRelationships: [...types.#Kind]
PostProcessedRelationships: [...types.#Kind]

// Property name enumerations

CertChain: types.#StringEnum & {
	symbol:         "CertChain"
	schema:         "ad"
	name:           "Certificate Chain"
	representation: "certchain"
}

CertName: types.#StringEnum & {
	symbol:         "CertName"
	schema:         "ad"
	name:           "Certificate Name"
	representation: "certname"
}

CertThumbprint: types.#StringEnum & {
	symbol:         "CertThumbprint"
	schema:         "ad"
	name:           "Certificate Thumbprint"
	representation: "certthumbprint"
}

CertThumbprints: types.#StringEnum & {
	symbol:         "CertThumbprints"
	schema:         "ad"
	name:           "Certificate Thumbprints"
	representation: "certthumbprints"
}

CAName: types.#StringEnum & {
	symbol:         "CAName"
	schema:         "ad"
	name:           "CA Name"
	representation: "caname"
}

CASecurityCollected: types.#StringEnum & {
	symbol:         "CASecurityCollected"
	schema:         "ad"
	name:           "CA Security Collected"
	representation: "casecuritycollected"
}

HasEnrollmentAgentRestrictions: types.#StringEnum & {
	symbol:         "HasEnrollmentAgentRestrictions"
	schema:         "ad"
	name:           "Has Enrollment Agent Restrictions"
	representation: "hasenrollmentagentrestrictions"
}

EnrollmentAgentRestrictionsCollected: types.#StringEnum & {
	symbol:         "EnrollmentAgentRestrictionsCollected"
	schema:         "ad"
	name:           "Enrollment Agent Restrictions Collected"
	representation: "enrollmentagentrestrictionscollected"
}

IsUserSpecifiesSanEnabled: types.#StringEnum & {
	symbol:         "IsUserSpecifiesSanEnabled"
	schema:         "ad"
	name:           "Is User Specifies San Enabled"
	representation: "isuserspecifiessanenabled"
}

IsUserSpecifiesSanEnabledCollected: types.#StringEnum & {
	symbol:         "IsUserSpecifiesSanEnabledCollected"
	schema:         "ad"
	name:           "Is User Specifies San Enabled Collected"
	representation: "isuserspecifiessanenabledcollected"
}

RoleSeparationEnabled: types.#StringEnum & {
	symbol:         "RoleSeparationEnabled"
	schema:         "ad"
	name:           "Role Separation Enabled"
	representation: "roleseparationenabled"
}

RoleSeparationEnabledCollected: types.#StringEnum & {
	symbol:         "RoleSeparationEnabledCollected"
	schema:         "ad"
	name:           "Role Separation Enabled Collected"
	representation: "roleseparationenabledcollected"
}

HasBasicConstraints: types.#StringEnum & {
	symbol:         "HasBasicConstraints"
	schema:         "ad"
	name:           "Has Basic Constraints"
	representation: "hasbasicconstraints"
}

BasicConstraintPathLength: types.#StringEnum & {
	symbol:         "BasicConstraintPathLength"
	schema:         "ad"
	name:           "Basic Constraint Path Length"
	representation: "basicconstraintpathlength"
}

UnresolvedPublishedTemplates: types.#StringEnum & {
	symbol:         "UnresolvedPublishedTemplates"
	schema:         "ad"
	name:           "Unresolved Published Certificate Templates"
	representation: "unresolvedpublishedtemplates"
}

DNSHostname: types.#StringEnum & {
	symbol:         "DNSHostname"
	schema:         "ad"
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

BlocksInheritance: types.#StringEnum & {
	symbol:         "BlocksInheritance"
	schema:         "ad"
	name:           "Blocks GPO Inheritance"
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

InheritanceHashes: types.#StringEnum & {
	symbol:         "InheritanceHashes"
	schema:         "ad"
	name:           "ACL Inheritance Hashes"
	representation: "inheritancehashes"
}

InheritanceHash: types.#StringEnum & {
	symbol:         "InheritanceHash"
	schema:         "ad"
	name:           "ACL Inheritance Hash"
	representation: "inheritancehash"
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
	symbol:         "HasLAPS"
	schema:         "ad"
	name:           "LAPS Enabled"
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

AdminSDHolderProtected: types.#StringEnum & {
	symbol:         "AdminSDHolderProtected"
	schema:         "ad"
	name:           "AdminSDHolder Protected"
	representation: "adminsdholderprotected"
}

DontRequirePreAuth: types.#StringEnum & {
	symbol:         "DontRequirePreAuth"
	schema:         "ad"
	name:           "Do Not Require Pre-Authentication"
	representation: "dontreqpreauth"
}

HasURA: types.#StringEnum & {
	symbol:         "HasURA"
	schema:         "ad"
	name:           "Has User Rights Assignment Collection"
	representation: "hasura"
}

PasswordNeverExpires: types.#StringEnum & {
	symbol:         "PasswordNeverExpires"
	schema:         "ad"
	name:           "Password Never Expires"
	representation: "pwdneverexpires"
}

PasswordNotRequired: types.#StringEnum & {
	symbol:         "PasswordNotRequired"
	schema:         "ad"
	name:           "Password Not Required"
	representation: "passwordnotreqd"
}

FunctionalLevel: types.#StringEnum & {
	symbol:         "FunctionalLevel"
	schema:         "ad"
	name:           "Functional Level"
	representation: "functionallevel"
}

TrustType: types.#StringEnum & {
	symbol:         "TrustType"
	schema:         "ad"
	name:           "Trust Type"
	representation: "trusttype"
}

SpoofSIDHistoryBlocked: types.#StringEnum & {
	symbol:         "SpoofSIDHistoryBlocked"
	schema:         "ad"
	name:           "Spoof SID History Blocked"
	representation: "spoofsidhistoryblocked"
}

TrustedToAuth: types.#StringEnum & {
	symbol:         "TrustedToAuth"
	schema:         "ad"
	name:           "Trusted For Constrained Delegation"
	representation: "trustedtoauth"
}

SamAccountName: types.#StringEnum & {
	symbol:         "SamAccountName"
	schema:         "ad"
	name:           "SAM Account Name"
	representation: "samaccountname"
}

HomeDirectory: types.#StringEnum & {
	symbol:         "HomeDirectory"
	schema:         "ad"
	name:           "Home Directory"
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

VulnerableNetlogonSecurityDescriptor: types.#StringEnum & {
	symbol:         "VulnerableNetlogonSecurityDescriptor"
	schema:         "ad"
	name:           "Vulnerable Netlogon Security Descriptor"
	representation: "vulnerablenetlogonsecuritydescriptor"
}

VulnerableNetlogonSecurityDescriptorCollected: types.#StringEnum & {
	symbol:         "VulnerableNetlogonSecurityDescriptorCollected"
	schema:         "ad"
	name:           "Vulnerable Netlogon Security Descriptor Collected"
	representation: "vulnerablenetlogonsecuritydescriptorcollected"
}

CrossCertificatePair: types.#StringEnum & {
	symbol:         "CrossCertificatePair"
	schema:         "ad"
	name:           "Cross Certificate Pair"
	representation: "crosscertificatepair"
}

EKUs: types.#StringEnum & {
	symbol:         "EKUs"
	schema:         "ad"
	name:           "Enhanced Key Usage"
	representation: "ekus"
}

SubjectAltRequireUPN: types.#StringEnum & {
	symbol:         "SubjectAltRequireUPN"
	schema:         "ad"
	name:           "Subject Alternative Name Require UPN"
	representation: "subjectaltrequireupn"
}

SubjectAltRequireDNS: types.#StringEnum & {
	symbol:         "SubjectAltRequireDNS"
	schema:         "ad"
	name:           "Subject Alternative Name Require DNS"
	representation: "subjectaltrequiredns"
}

SubjectAltRequireDomainDNS: types.#StringEnum & {
	symbol:         "SubjectAltRequireDomainDNS"
	schema:         "ad"
	name:           "Subject Alternative Name Require Domain DNS"
	representation: "subjectaltrequiredomaindns"
}

SubjectAltRequireEmail: types.#StringEnum & {
	symbol:         "SubjectAltRequireEmail"
	schema:         "ad"
	name:           "Subject Alternative Name Require Email"
	representation: "subjectaltrequireemail"
}

SubjectAltRequireSPN: types.#StringEnum & {
	symbol:         "SubjectAltRequireSPN"
	schema:         "ad"
	name:           "Subject Alternative Name Require SPN"
	representation: "subjectaltrequirespn"
}

SubjectRequireEmail: types.#StringEnum & {
	symbol:         "SubjectRequireEmail"
	schema:         "ad"
	name:           "Subject Require Email"
	representation: "subjectrequireemail"
}

AuthorizedSignatures: types.#StringEnum & {
	symbol:         "AuthorizedSignatures"
	schema:         "ad"
	name:           "Authorized Signatures Required"
	representation: "authorizedsignatures"
}

ApplicationPolicies: types.#StringEnum & {
	symbol:         "ApplicationPolicies"
	schema:         "ad"
	name:           "Application Policies Required"
	representation: "applicationpolicies"
}

IssuancePolicies: types.#StringEnum & {
	symbol:         "IssuancePolicies"
	schema:         "ad"
	name:           "Issuance Policies Required"
	representation: "issuancepolicies"
}

SchemaVersion: types.#StringEnum & {
	symbol:         "SchemaVersion"
	schema:         "ad"
	name:           "Schema Version"
	representation: "schemaversion"
}

RequiresManagerApproval: types.#StringEnum & {
	symbol:         "RequiresManagerApproval"
	schema:         "ad"
	name:           "Requires Manager Approval"
	representation: "requiresmanagerapproval"
}

AuthenticationEnabled: types.#StringEnum & {
	symbol:         "AuthenticationEnabled"
	schema:         "ad"
	name:           "Authentication Enabled"
	representation: "authenticationenabled"
}

SchannelAuthenticationEnabled: types.#StringEnum & {
	symbol:         "SchannelAuthenticationEnabled"
	schema:         "ad"
	name:           "Schannel Authentication Enabled"
	representation: "schannelauthenticationenabled"
}

EnrolleeSuppliesSubject: types.#StringEnum & {
	symbol:         "EnrolleeSuppliesSubject"
	schema:         "ad"
	name:           "Enrollee Supplies Subject"
	representation: "enrolleesuppliessubject"
}

CertificateApplicationPolicy: types.#StringEnum & {
	symbol:         "CertificateApplicationPolicy"
	schema:         "ad"
	name:           "Application Policy Extensions"
	representation: "certificateapplicationpolicy"
}

CertificateNameFlag: types.#StringEnum & {
	symbol:         "CertificateNameFlag"
	schema:         "ad"
	name:           "Certificate Name Flags"
	representation: "certificatenameflag"
}

EffectiveEKUs: types.#StringEnum & {
	symbol:         "EffectiveEKUs"
	schema:         "ad"
	name:           "Effective EKUs"
	representation: "effectiveekus"
}

EnrollmentFlag: types.#StringEnum & {
	symbol:         "EnrollmentFlag"
	schema:         "ad"
	name:           "Enrollment Flags"
	representation: "enrollmentflag"
}

Flags: types.#StringEnum & {
	symbol:         "Flags"
	schema:         "ad"
	name:           "Flags"
	representation: "flags"
}

NoSecurityExtension: types.#StringEnum & {
	symbol:         "NoSecurityExtension"
	schema:         "ad"
	name:           "No Security Extension"
	representation: "nosecurityextension"
}

RenewalPeriod: types.#StringEnum & {
	symbol:         "RenewalPeriod"
	schema:         "ad"
	name:           "Renewal Period"
	representation: "renewalperiod"
}

ValidityPeriod: types.#StringEnum & {
	symbol:         "ValidityPeriod"
	schema:         "ad"
	name:           "Validity Period"
	representation: "validityperiod"
}

OID: types.#StringEnum & {
	symbol:         "OID"
	schema:         "ad"
	name:           "OID"
	representation: "oid"
}

CertificatePolicy: types.#StringEnum & {
	symbol:         "CertificatePolicy"
	schema:         "ad"
	name:           "Issuance Policy Extensions"
	representation: "certificatepolicy"
}

CertTemplateOID: types.#StringEnum & {
	symbol:         "CertTemplateOID"
	schema:         "ad"
	name:           "Certificate Template OID"
	representation: "certtemplateoid"
}

GroupLinkID: types.#StringEnum & {
	symbol:         "GroupLinkID"
	schema:         "ad"
	name:           "Group Link ID"
	representation: "grouplinkid"
}

ObjectGUID: types.#StringEnum & {
	symbol:         "ObjectGUID"
	schema:         "ad"
	name:           "Object GUID"
	representation: "objectguid"
}

ExpirePasswordsOnSmartCardOnlyAccounts: types.#StringEnum & {
	symbol:         "ExpirePasswordsOnSmartCardOnlyAccounts"
	schema:         "ad"
	name:           "Expire Passwords on Smart Card only Accounts"
	representation: "expirepasswordsonsmartcardonlyaccounts"
}

MachineAccountQuota: types.#StringEnum & {
	symbol:         "MachineAccountQuota"
	schema:         "ad"
	name:           "Machine Account Quota"
	representation: "machineaccountquota"
}

SupportedKerberosEncryptionTypes: types.#StringEnum & {
	symbol:         "SupportedKerberosEncryptionTypes"
	schema:         "ad"
	name:           "Supported Kerberos Encryption Types"
	representation: "supportedencryptiontypes"
}

TGTDelegation: types.#StringEnum & {
	symbol:         "TGTDelegation"
	schema:         "ad"
	name:           "TGT Delegation"
	representation: "tgtdelegation"
}

PasswordStoredUsingReversibleEncryption: types.#StringEnum & {
	symbol:         "PasswordStoredUsingReversibleEncryption"
	schema:         "ad"
	name:           "Password Stored Using Reversible Encryption"
	representation: "encryptedtextpwdallowed"
}

SmartcardRequired: types.#StringEnum & {
	symbol:         "SmartcardRequired"
	schema:         "ad"
	name:           "Smartcard Required"
	representation: "smartcardrequired"
}

UseDESKeyOnly: types.#StringEnum & {
	symbol:         "UseDESKeyOnly"
	schema:         "ad"
	name:           "Use DES Key Only"
	representation: "usedeskeyonly"
}

LogonScriptEnabled: types.#StringEnum & {
	symbol:         "LogonScriptEnabled"
	schema:         "ad"
	name:           "Logon Script Enabled"
	representation: "logonscriptenabled"
}

LockedOut: types.#StringEnum & {
	symbol:         "LockedOut"
	schema:         "ad"
	name:           "Locked Out"
	representation: "lockedout"
}

UserCannotChangePassword: types.#StringEnum & {
	symbol:         "UserCannotChangePassword"
	schema:         "ad"
	name:           "User Cannot Change Password"
	representation: "passwordcantchange"
}

PasswordExpired: types.#StringEnum & {
	symbol:         "PasswordExpired"
	schema:         "ad"
	name:           "Password Expired"
	representation: "passwordexpired"
}

DSHeuristics: types.#StringEnum & {
	symbol:         "DSHeuristics"
	schema:         "ad"
	name:           "DSHeuristics"
	representation: "dsheuristics"
}

UserAccountControl: types.#StringEnum & {
	symbol:         "UserAccountControl"
	schema:         "ad"
	name:           "User Account Control"
	representation: "useraccountcontrol"
}

TrustAttributesInbound: types.#StringEnum & {
	symbol:         "TrustAttributesInbound"
	schema:         "ad"
	name:           "Trust Attributes (Inbound)"
	representation: "trustattributesinbound"
}

TrustAttributesOutbound: types.#StringEnum & {
	symbol:         "TrustAttributesOutbound"
	schema:         "ad"
	name:           "Trust Attributes (Outbound)"
	representation: "trustattributesoutbound"
}

LockoutDuration: types.#StringEnum & {
	symbol:         "LockoutDuration"
	schema:         "ad"
	name:           "Lockout Duration"
	representation: "lockoutduration"
}

LockoutObservationWindow: types.#StringEnum & {
	symbol:         "LockoutObservationWindow"
	schema:         "ad"
	name:           "Lockout Observation Window"
	representation: "lockoutobservationwindow"
}

MaxPwdAge: types.#StringEnum & {
	symbol:         "MaxPwdAge"
	schema:         "ad"
	name:           "Maximum Password Age"
	representation: "maxpwdage"
}

MinPwdAge: types.#StringEnum & {
	symbol:         "MinPwdAge"
	schema:         "ad"
	name:           "Minimum Password Age"
	representation: "minpwdage"
}

LockoutThreshold: types.#StringEnum & {
	symbol:         "LockoutThreshold"
	schema:         "ad"
	name:           "Lockout Threshold"
	representation: "lockoutthreshold"
}

PwdHistoryLength: types.#StringEnum & {
	symbol:         "PwdHistoryLength"
	schema:         "ad"
	name:           "Password History Length"
	representation: "pwdhistorylength"
}

PwdProperties: types.#StringEnum & {
	symbol:         "PwdProperties"
	schema:         "ad"
	name:           "Password Properties"
	representation: "pwdproperties"
}

MinPwdLength: types.#StringEnum & {
	symbol:         "MinPwdLength"
	schema:         "ad"
	name:           "Minimum password length"
	representation: "minpwdlength"
}

GMSA: types.#StringEnum & {
 	symbol: "GMSA"
 	schema: "ad"
 	name: "GMSA"
 	representation: "gmsa"
}

MSA: types.#StringEnum & {
 	symbol: "MSA"
 	schema: "ad"
 	name: "MSA"
 	representation: "msa"
}

SMBSigning: types.#StringEnum & {
	symbol:         "SMBSigning"
	schema:         "ad"
	name:           "SMB Signing"
	representation: "smbsigning"
}

WebClientRunning: types.#StringEnum & {
	symbol: "WebClientRunning"
	schema: "ad"
	name: "WebClient Running"
	representation: "webclientrunning"
}

RestrictOutboundNTLM: types.#StringEnum & {
	symbol:         "RestrictOutboundNTLM"
	schema:         "ad"
	name:           "Restrict Outbound NTLM"
	representation: "restrictoutboundntlm"
}

ADCSWebEnrollmentHTTP: types.#StringEnum & {
	symbol: "ADCSWebEnrollmentHTTP"
	schema: "ad"
	name: "ADCS Web Enrollment HTTP"
	representation: "adcswebenrollmenthttp"
}

ADCSWebEnrollmentHTTPS: types.#StringEnum & {
	symbol: "ADCSWebEnrollmentHTTPS"
	schema: "ad"
	name: "ADCS Web Enrollment HTTPS"
	representation: "adcswebenrollmenthttps"
}

ADCSWebEnrollmentHTTPSEPA: types.#StringEnum & {
	symbol: "ADCSWebEnrollmentHTTPSEPA"
	schema: "ad"
	name: "ADCS Web Enrollment HTTPS EPA"
	representation: "adcswebenrollmenthttpsepa"
}

DoesAnyAceGrantOwnerRights: types.#StringEnum & {
 	symbol: "DoesAnyAceGrantOwnerRights"
 	schema: "ad"
 	name: "Does Any ACE Grant Owner Rights"
 	representation: "doesanyacegrantownerrights"
}

DoesAnyInheritedAceGrantOwnerRights: types.#StringEnum & {
 	symbol: "DoesAnyInheritedAceGrantOwnerRights"
 	schema: "ad"
 	name: "Does Any Inherited ACE Grant Owner Rights"
 	representation: "doesanyinheritedacegrantownerrights"
}

OwnerSid: types.#StringEnum & {
	symbol: "OwnerSid"
 	schema: "ad"
 	name: "Owner SID"
 	representation: "ownersid"
}

LDAPSigning: types.#StringEnum & {
	symbol: "LDAPSigning"
	schema: "ad"
	name: "LDAP Signing"
	representation: "ldapsigning"
}

LDAPAvailable: types.#StringEnum & {
	symbol: "LDAPAvailable"
	schema: "ad"
	name: "LDAP Available"
	representation: "ldapavailable"
}

LDAPSAvailable: types.#StringEnum & {
	symbol: "LDAPSAvailable"
	schema: "ad"
	name: "LDAPS Available"
	representation: "ldapsavailable"
}

LDAPSEPA: types.#StringEnum & {
	symbol: "LDAPSEPA"
	schema: "ad"
	name: "LDAPS EPA"
	representation: "ldapsepa"
}

RelayableToDCLDAP: types.#StringEnum & {
	symbol: "RelayableToDCLDAP"
	schema: "ad"
	name: "Relayable To DC LDAP"
	representation: "replayabletodcldap"
}

RelayableToDCLDAPS: types.#StringEnum & {
	symbol: "RelayableToDCLDAPS"
	schema: "ad"
	name: "Relayable To DC LDAPS"
	representation: "replayabletodcldaps"
}

WebClientRunning: types.#StringEnum & {
	symbol: "WebClientRunning"
	schema: "ad"
	name: "WebClient Running"
	representation: "webclientrunning"
}

IsDC: types.#StringEnum & {
	symbol: "IsDC"
	schema: "ad"
	name: "Is Domain Controller"
	representation: "isdc"
}

IsReadOnlyDC: types.#StringEnum & {
	symbol: "IsReadOnlyDC"
	schema: "ad"
	name: "Read-Only DC"
	representation: "isreadonlydc"
}

HTTPEnrollmentEndpoints: types.#StringEnum & {
	symbol: "HTTPEnrollmentEndpoints"
	schema: "ad"
	name:"HTTP Enrollment Endpoints"
	representation: "httpenrollmentendpoints"
}

HTTPSEnrollmentEndpoints: types.#StringEnum & {
	symbol: "HTTPSEnrollmentEndpoints"
	schema: "ad"
	name:"HTTPS Enrollment Endpoints"
	representation: "httpsenrollmentendpoints"
}

HasVulnerableEndpoint: types.#StringEnum & {
	symbol: "HasVulnerableEndpoint"
	schema: "ad"
	name:"Has Vulnerable Endpoint"
	representation: "hasvulnerableendpoint"
}

RequireSecuritySignature: types.#StringEnum & {
	symbol: "RequireSecuritySignature"
	schema: "ad"
	name: "Require Security Signature"
	representation: "requiresecuritysignature"
}

EnableSecuritySignature: types.#StringEnum & {
	symbol: "EnableSecuritySignature"
	schema: "ad"
	name: "Enable Security Signature"
	representation: "enablesecuritysignature"
}

RestrictReceivingNTLMTraffic: types.#StringEnum & {
	symbol: "RestrictReceivingNTLMTraffic"
	schema: "ad"
	name: "Restrict Receiving NTLM Traffic"
	representation: "restrictreceivingntmltraffic"
}

NTLMMinServerSec: types.#StringEnum & {
	symbol: "NTLMMinServerSec"
	schema: "ad"
	name: "NTLM Min Server Sec"
	representation: "ntlmminserversec"
}

NTLMMinClientSec: types.#StringEnum & {
	symbol: "NTLMMinClientSec"
	schema: "ad"
	name: "NTLM Min Client Sec"
	representation: "ntlmminclientsec"
}
LMCompatibilityLevel: types.#StringEnum & {
	symbol: "LMCompatibilityLevel"
	schema: "ad"
	name: "LM Compatibility Level"
	representation: "lmcompatibilitylevel"
}

UseMachineID: types.#StringEnum & {
	symbol: "UseMachineID"
	schema: "ad"
	name: "Use Machine ID"
	representation: "usemachineid"
}

ClientAllowedNTLMServers: types.#StringEnum & {
	symbol: "ClientAllowedNTLMServers"
	schema: "ad"
	name: "Client Allowed NTLM Servers"
	representation: "clientallowedntlmservers"
}

Transitive: types.#StringEnum & {
	symbol: "Transitive"
	schema: "ad"
	name:"Transitive"
	representation: "transitive"
}

GroupScope: types.#StringEnum & {
	symbol:         "GroupScope"
	schema:         "ad"
	name:           "Group Scope"
	representation: "groupscope"
}

NetBIOS: types.#StringEnum & {
	symbol:         "NetBIOS"
	schema:         "ad"
	name:           "NetBIOS"
	representation: "netbios"
}

ServicePrincipalNames: types.#StringEnum & {
	symbol:         "ServicePrincipalNames"
	schema:         "ad"
	name:           "Service Principal Names"
	representation: "serviceprincipalnames"
}

GPOStatusRaw: types.#StringEnum & {
	symbol:         "GPOStatusRaw"
	schema:         "ad"
	name:           "GPO Status (Raw)"
	representation: "gpostatusraw"
}

GPOStatus: types.#StringEnum & {
	symbol:         "GPOStatus"
	schema:         "ad"
	name:           "GPO Status"
	representation: "gpostatus"
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
	RoleSeparationEnabled,
	RoleSeparationEnabledCollected,
	HasBasicConstraints,
	BasicConstraintPathLength,
	UnresolvedPublishedTemplates,
	DNSHostname,
	CrossCertificatePair,
	DistinguishedName,
	DomainFQDN,
	DomainSID,
	Sensitive,
	BlocksInheritance,
	IsACL,
	IsACLProtected,
	InheritanceHash,
	InheritanceHashes,
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
	SpoofSIDHistoryBlocked,
	TrustedToAuth,
	SamAccountName,
	CertificateMappingMethodsRaw,
	CertificateMappingMethods,
	StrongCertificateBindingEnforcementRaw,
	StrongCertificateBindingEnforcement,
	VulnerableNetlogonSecurityDescriptor,
	VulnerableNetlogonSecurityDescriptorCollected,
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
	SchannelAuthenticationEnabled,
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
	HomeDirectory,
	CertificatePolicy,
	CertTemplateOID,
	GroupLinkID,
	ObjectGUID,
	ExpirePasswordsOnSmartCardOnlyAccounts,
	MachineAccountQuota,
	SupportedKerberosEncryptionTypes,
	TGTDelegation,
	PasswordStoredUsingReversibleEncryption,
	SmartcardRequired,
	UseDESKeyOnly,
	LogonScriptEnabled,
	LockedOut,
	UserCannotChangePassword,
	PasswordExpired,
	DSHeuristics,
	UserAccountControl,
	TrustAttributesInbound,
	TrustAttributesOutbound,
	MinPwdLength,
	PwdProperties,
	PwdHistoryLength,
	LockoutThreshold,
	MinPwdAge,
	MaxPwdAge,
	LockoutDuration,
	LockoutObservationWindow,
	OwnerSid,
	SMBSigning,
	WebClientRunning,
	RestrictOutboundNTLM,
	GMSA,
	MSA,
	DoesAnyAceGrantOwnerRights,
	DoesAnyInheritedAceGrantOwnerRights,
	ADCSWebEnrollmentHTTP,
	ADCSWebEnrollmentHTTPS,
	ADCSWebEnrollmentHTTPSEPA,
	LDAPSigning,
	LDAPAvailable,
	LDAPSAvailable,
	LDAPSEPA,
	IsDC,
	IsReadOnlyDC,
	HTTPEnrollmentEndpoints,
	HTTPSEnrollmentEndpoints,
	HasVulnerableEndpoint,
	RequireSecuritySignature,
	EnableSecuritySignature,
	RestrictReceivingNTLMTraffic,
	NTLMMinServerSec,
	NTLMMinClientSec,
	LMCompatibilityLevel,
	UseMachineID,
	ClientAllowedNTLMServers,
	Transitive,
	GroupScope,
	NetBIOS,
	AdminSDHolderProtected,
	ServicePrincipalNames,
	GPOStatusRaw,
	GPOStatus,
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

IssuancePolicy: types.#Kind & {
	symbol: "IssuancePolicy"
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
	CertTemplate,
	IssuancePolicy,
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

CoerceToTGT: types.#Kind & {
	symbol: "CoerceToTGT"
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

CrossForestTrust: types.#Kind & {
	symbol: "CrossForestTrust"
	schema: "active_directory"
}

SameForestTrust: types.#Kind & {
	symbol: "SameForestTrust"
	schema: "active_directory"
}

SpoofSIDHistory: types.#Kind & {
	symbol: "SpoofSIDHistory"
	schema: "active_directory"
}

AbuseTGTDelegation: types.#Kind & {
	symbol: "AbuseTGTDelegation"
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

RemoteInteractiveLogonRight: types.#Kind & {
	symbol: "RemoteInteractiveLogonRight"
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

WriteGPLink: types.#Kind & {
	symbol: "WriteGPLink"
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

OIDGroupLink: types.#Kind & {
	symbol: "OIDGroupLink"
	schema: "active_directory"
}

ExtendedByPolicy: types.#Kind & {
	symbol: "ExtendedByPolicy"
	schema: "active_directory"
}

ExtendedByPolicy: types.#Kind & {
	symbol: "ExtendedByPolicy"
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

ADCSESC6a: types.#Kind & {
	symbol: "ADCSESC6a"
	schema: "active_directory"
}

ADCSESC6b: types.#Kind & {
	symbol: "ADCSESC6b"
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

ADCSESC13: types.#Kind & {
	symbol: "ADCSESC13"
	schema: "active_directory"
}

SyncedToEntraUser: types.#Kind & {
	symbol: "SyncedToEntraUser"
	schema: "active_directory"
}

CoerceAndRelayNTLMToSMB: types.#Kind & {
	symbol: "CoerceAndRelayNTLMToSMB"
	schema: "active_directory"
}

CoerceAndRelayNTLMToADCS: types.#Kind & {
	symbol: "CoerceAndRelayNTLMToADCS"
	schema: "active_directory"
}

WriteOwnerLimitedRights: types.#Kind & {
	symbol: "WriteOwnerLimitedRights"
	schema: "active_directory"
}

WriteOwnerRaw: types.#Kind & {
	symbol: "WriteOwnerRaw"
	schema: "active_directory"
}

OwnsLimitedRights: types.#Kind & {
	symbol: "OwnsLimitedRights"
	schema: "active_directory"
}

OwnsRaw: types.#Kind & {
	symbol: "OwnsRaw"
	schema: "active_directory"
}

CoerceAndRelayNTLMToLDAP: types.#Kind & {
	symbol: "CoerceAndRelayNTLMToLDAP"
	schema: "active_directory"
}

CoerceAndRelayNTLMToLDAPS: types.#Kind & {
	symbol: "CoerceAndRelayNTLMToLDAPS"
	schema: "active_directory"
}

ProtectAdminGroups: types.#Kind & {
	symbol:         "ProtectAdminGroups"
	schema:         "active_directory"
}

HasTrustKeys: types.#Kind & {
	symbol: "HasTrustKeys"
	schema: "active_directory"
}

ClaimSpecialIdentity: types.#Kind & {
	symbol: "ClaimSpecialIdentity"
	schema: "active_directory"
}

ContainsIdentity: types.#Kind & {
	symbol: "ContainsIdentity"
	schema: "active_directory"
}

PropagatesACEsTo: types.#Kind & {
	symbol: "PropagatesACEsTo"
	schema: "active_directory"
}

GPOAppliesTo: types.#Kind & {
	symbol: "GPOAppliesTo"
	schema: "active_directory"
}

CanApplyGPO: types.#Kind & {
	symbol: "CanApplyGPO"
	schema: "active_directory"
}

CanBackup: types.#Kind & {
	symbol: "CanBackup"
	schema: "active_directory"
}

BackupPrivilege: types.#Kind & {
	symbol: "BackupPrivilege"
	schema: "active_directory"
}

RestorePrivilege: types.#Kind & {
	symbol: "RestorePrivilege"
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
	CoerceToTGT,
	GetChanges,
	GetChangesAll,
	GetChangesInFilteredSet,
	CrossForestTrust,
	SameForestTrust,
	SpoofSIDHistory,
	AbuseTGTDelegation,
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
	RemoteInteractiveLogonRight,
	SyncLAPSPassword,
	WriteAccountRestrictions,
	WriteGPLink,
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
	IssuedSignedBy,
	GoldenCert,
	EnrollOnBehalfOf,
	OIDGroupLink,
	ExtendedByPolicy,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b,
	ADCSESC13,
	SyncedToEntraUser,
	CoerceAndRelayNTLMToSMB,
	CoerceAndRelayNTLMToADCS,
	WriteOwnerLimitedRights,
	WriteOwnerRaw,
	OwnsLimitedRights,
	OwnsRaw,
	ClaimSpecialIdentity,
	CoerceAndRelayNTLMToLDAP,
	CoerceAndRelayNTLMToLDAPS,
	ContainsIdentity,
	PropagatesACEsTo,
	GPOAppliesTo,
	CanApplyGPO,
	HasTrustKeys,
	ProtectAdminGroups,
	CanBackup,
	BackupPrivilege,
	RestorePrivilege,
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
	WriteGPLink,
	SyncLAPSPassword,
	DCSync,
	ManageCertificates,
	ManageCA,
	Enroll,
	WritePKIEnrollmentFlag,
	WritePKINameFlag,
	WriteOwnerLimitedRights,
	OwnsLimitedRights,
]

// these edges are common to inbound/outbound/pathfinding
SharedRelationshipKinds: [
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
	GPLink,
	AllowedToDelegate,
	CoerceToTGT,
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
	WriteGPLink,
	GoldenCert,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b,
	ADCSESC13,
	SyncedToEntraUser,
	CoerceAndRelayNTLMToSMB,
	CoerceAndRelayNTLMToADCS,
	WriteOwnerLimitedRights,
	OwnsLimitedRights,
	ClaimSpecialIdentity,
	CoerceAndRelayNTLMToLDAP,
	CoerceAndRelayNTLMToLDAPS,
	ContainsIdentity,
	PropagatesACEsTo,
	GPOAppliesTo,
	CanApplyGPO,
	HasTrustKeys,
	ManageCA,
	ManageCertificates,
	CanBackup,
]

// Edges that are used during inbound traversal
InboundRelationshipKinds: list.Concat([SharedRelationshipKinds, [Contains]])

// Edges that are used during outbound traversal
OutboundRelationshipKinds: list.Concat([SharedRelationshipKinds,[Contains, DCFor]])

// Edges that are used in pathfinding
PathfindingRelationships: list.Concat([SharedRelationshipKinds,[Contains, DCFor, SameForestTrust, SpoofSIDHistory, AbuseTGTDelegation]])

EdgeCompositionRelationships: [
	GoldenCert,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC10a,
	ADCSESC10b,
	ADCSESC13,
	CoerceAndRelayNTLMToSMB,
	CoerceAndRelayNTLMToADCS,
	CoerceAndRelayNTLMToLDAP,
	CoerceAndRelayNTLMToLDAPS,
	GPOAppliesTo,
	CanApplyGPO,
]

PostProcessedRelationships: [
	DCSync,
	ProtectAdminGroups,
	SyncLAPSPassword,
	CanRDP,
	AdminTo,
	CanPSRemote,
	ExecuteDCOM,
	TrustedForNTAuth,
	IssuedSignedBy,
	EnterpriseCAFor,
	GoldenCert,
	ADCSESC1,
	ADCSESC3,
	ADCSESC4,
	ADCSESC6a,
	ADCSESC6b,
	ADCSESC10a,
	ADCSESC10b,
	ADCSESC9a,
	ADCSESC9b,
	ADCSESC13,
	EnrollOnBehalfOf,
	SyncedToEntraUser,
	Owns,
	WriteOwner,
	ExtendedByPolicy,
	CoerceAndRelayNTLMToADCS,
	CoerceAndRelayNTLMToSMB,
	CoerceAndRelayNTLMToLDAP,
	CoerceAndRelayNTLMToLDAPS,
	GPOAppliesTo,
	CanApplyGPO,
	HasTrustKeys,
	CanBackup,
]
