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

// Property name enumerations
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
	name:           "Sensitive"
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

Properties: [
	AdminCount,
	DistinguishedName,
	DomainFQDN,
	DomainSID,
	Sensitive,
	HighValue,
	BlocksInheritance,
	IsACL,
	IsACLProtected,
	Enforced,
	Department,
	HasSPN,
	UnconstrainedDelegation,
	LastLogon,
	LastLogonTimestamp,
	IsPrimaryGroup,
	HasLAPS,
	DontRequirePreAuth,
	LogonType,
	HasURA
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
	WriteAccountRestrictions
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
	WriteAccountRestrictions
]
