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

package azure

import "pkg.specterops.io/schemas/bh/types:types"

// Exported requirements
Properties: [...types.#StringEnum]
NodeKinds: [...types.#Kind]
RelationshipKinds: [...types.#Kind]
AppRoleTransitRelationshipKinds: [...types.#Kind]
AbusableAppRoleRelationshipKinds: [... types.#Kind]
ControlRelationshipKinds: [...types.#Kind]
ExecutionPrivilegeKinds: [...types.#Kind]
PathfindingRelationships: [...types.#Kind]

// Property name enumerations
AppOwnerOrganizationID: types.#StringEnum & {
	symbol:         "AppOwnerOrganizationID"
	schema:         "azure"
	name:           "App Owner Organization ID"
	representation: "appownerorganizationid"
}

AppDescription: types.#StringEnum & {
	symbol:         "AppDescription"
	schema:         "azure"
	name:           "App Description"
	representation: "appdescription"
}

AppDisplayName: types.#StringEnum & {
	symbol:         "AppDisplayName"
	schema:         "azure"
	name:           "App Display Name"
	representation: "appdisplayname"
}

ServicePrincipalType: types.#StringEnum & {
	symbol:         "ServicePrincipalType"
	schema:         "azure"
	name:           "Service Principal Type"
	representation: "serviceprincipaltype"
}

UserType: types.#StringEnum & {
	symbol:         "UserType"
	schema:         "azure"
	name:           "User Type"
	representation: "usertype"
}

Scope: types.#StringEnum & {
	symbol:         "Scope"
	schema:         "azure"
	name:           "Scope"
	representation: "scope"
}

MFAEnabled: types.#StringEnum & {
	symbol:         "MFAEnabled"
	schema:         "azure"
	name:           "MFA Enabled"
	representation: "mfaenabled"
}

OnPremSyncEnabled: types.#StringEnum & {
	symbol:         "OnPremSyncEnabled"
	schema:         "azure"
	name:           "On Prem Sync Enabled"
	representation: "onpremsyncenabled"
}

SecurityEnabled: types.#StringEnum & {
	symbol:         "SecurityEnabled"
	schema:         "azure"
	name:           "Security Enabled"
	representation: "securityenabled"
}

SecurityIdentifier: types.#StringEnum & {
	symbol:         "SecurityIdentifier"
	schema:         "azure"
	name:           "Security Identifier"
	representation: "securityidentifier"
}

EnableRBACAuthorization: types.#StringEnum & {
	symbol:         "EnableRBACAuthorization"
	schema:         "azure"
	name:           "RBAC Authorization Enabled"
	representation: "enablerbacauthorization"
}

License: types.#StringEnum & {
	symbol:         "License"
	schema:         "azure"
	representation: "license"
}

Licenses: types.#StringEnum & {
	symbol:         "Licenses"
	schema:         "azure"
	representation: "licenses"
}

NodeResourceGroupID: types.#StringEnum & {
	symbol:         "NodeResourceGroupID"
	schema:         "azure"
	name:           "Node Resource Group ID"
	representation: "noderesourcegroupid"
}

Offer: types.#StringEnum & {
	symbol:         "Offer"
	schema:         "azure"
	representation: "offer"
}

MFAEnforced: types.#StringEnum & {
	symbol:         "MFAEnforced"
	schema:         "azure"
	name:           "MFA Enforced"
	representation: "mfaenforced"
}

UserPrincipalName: types.#StringEnum & {
	symbol:         "UserPrincipalName"
	schema:         "azure"
	name:           "User Principal Name"
	representation: "userprincipalname"
}

IsAssignableToRole: types.#StringEnum & {
	symbol:         "IsAssignableToRole"
	schema:         "azure"
	name:           "Is Role Assignable"
	representation: "isassignabletorole"
}

PublisherDomain: types.#StringEnum & {
	symbol:         "PublisherDomain"
	schema:         "azure"
	name:           "Publisher Domain"
	representation: "publisherdomain"
}

SignInAudience: types.#StringEnum & {
	symbol:         "SignInAudience"
	schema:         "azure"
	name:           "Sign In Audience"
	representation: "signinaudience"
}

OperatingSystemVersion: types.#StringEnum & {
	symbol:         "OperatingSystemVersion"
	schema:         "azure"
	name:           "Operating System Version"
	representation: "operatingsystemversion"
}

TrustType: types.#StringEnum & {
	symbol:         "TrustType"
	schema:         "azure"
	name:           "Trust Type"
	representation: "trustype"
}

IsBuiltIn: types.#StringEnum & {
	symbol:         "IsBuiltIn"
	schema:         "azure"
	name:           "Is Built In"
	representation: "isbuiltin"
}

AppID: types.#StringEnum & {
	symbol:         "AppID"
	schema:         "azure"
	name:           "App ID"
	representation: "appid"
}

AppRoleID: types.#StringEnum & {
	symbol:         "AppRoleID"
	schema:         "azure"
	name:           "App Role ID"
	representation: "approleid"
}

DeviceID: types.#StringEnum & {
	symbol:         "DeviceID"
	schema:         "azure"
	name:           "Device ID"
	representation: "deviceid"
}

OnPremID: types.#StringEnum & {
	symbol:         "OnPremID"
	schema:         "azure"
	name:           "On Prem ID"
	representation: "onpremid"
}

RoleTemplateID: types.#StringEnum & {
	symbol:         "RoleTemplateID"
	schema:         "azure"
	name:           "Role Template ID"
	representation: "templateid"
}

ServicePrincipalID: types.#StringEnum & {
	symbol:         "ServicePrincipalID"
	schema:         "azure"
	name:           "Service Principal ID"
	representation: "service_principal_id"
}

ServicePrincipalNames: types.#StringEnum & {
	symbol:         "ServicePrincipalNames"
	schema:         "azure"
	name:           "Service Principal Names"
	representation: "service_principal_names"
}

TenantID: types.#StringEnum & {
	symbol:         "TenantID"
	schema:         "azure"
	name:           "Tenant ID"
	representation: "tenantid"
}

Properties: [
	AppOwnerOrganizationID,
	AppDescription,
	AppDisplayName,
	ServicePrincipalType,
	UserType,
	TenantID,
	ServicePrincipalID,
	ServicePrincipalNames,
	OperatingSystemVersion,
	TrustType,
	IsBuiltIn,
	AppID,
	AppRoleID,
	DeviceID,
	NodeResourceGroupID,
	OnPremID,
	OnPremSyncEnabled,
	SecurityEnabled,
	SecurityIdentifier,
	EnableRBACAuthorization,
	Scope,
	Offer,
	MFAEnabled,
	License,
	Licenses,
	MFAEnforced,
	UserPrincipalName,
	IsAssignableToRole,
	PublisherDomain,
	SignInAudience,
	RoleTemplateID,
]

// Kinds
Entity: types.#Kind & {
	symbol:         "Entity"
	schema:         "azure"
	representation: "AZBase"
}

App: types.#Kind & {
	symbol:         "App"
	schema:         "azure"
	representation: "AZApp"
}

VMScaleSet: types.#Kind & {
	symbol:         "VMScaleSet"
	schema:         "azure"
	representation: "AZVMScaleSet"
}

Role: types.#Kind & {
	symbol:         "Role"
	schema:         "azure"
	representation: "AZRole"
}

Device: types.#Kind & {
	symbol:         "Device"
	schema:         "azure"
	representation: "AZDevice"
}

FunctionApp: types.#Kind & {
	symbol:         "FunctionApp"
	schema:         "azure"
	representation: "AZFunctionApp"
}

Group: types.#Kind & {
	symbol:         "Group"
	schema:         "azure"
	representation: "AZGroup"
}

KeyVault: types.#Kind & {
	symbol:         "KeyVault"
	schema:         "azure"
	representation: "AZKeyVault"
}

ManagementGroup: types.#Kind & {
	symbol:         "ManagementGroup"
	schema:         "azure"
	representation: "AZManagementGroup"
}

ResourceGroup: types.#Kind & {
	symbol:         "ResourceGroup"
	schema:         "azure"
	representation: "AZResourceGroup"
}

ServicePrincipal: types.#Kind & {
	symbol:         "ServicePrincipal"
	schema:         "azure"
	representation: "AZServicePrincipal"
}

Subscription: types.#Kind & {
	symbol:         "Subscription"
	schema:         "azure"
	representation: "AZSubscription"
}

Tenant: types.#Kind & {
	symbol:         "Tenant"
	schema:         "azure"
	representation: "AZTenant"
}

User: types.#Kind & {
	symbol:         "User"
	schema:         "azure"
	representation: "AZUser"
}

VM: types.#Kind & {
	symbol:         "VM"
	schema:         "azure"
	representation: "AZVM"
}

ManagedCluster: types.#Kind & {
	symbol:         "ManagedCluster"
	schema:         "azure"
	representation: "AZManagedCluster"
}

ContainerRegistry: types.#Kind & {
	symbol:         "ContainerRegistry"
	schema:         "azure"
	representation: "AZContainerRegistry"
}

WebApp: types.#Kind & {
	symbol:         "WebApp"
	schema:         "azure"
	representation: "AZWebApp"
}

LogicApp: types.#Kind & {
	symbol:         "LogicApp"
	schema:         "azure"
	representation: "AZLogicApp"
}

AutomationAccount: types.#Kind & {
  symbol:         "AutomationAccount"
	schema:         "azure"
	representation: "AZAutomationAccount"
}

NodeKinds: [
	Entity,
	VMScaleSet,
	App,
	Role,
	Device,
	FunctionApp,
	Group,
	KeyVault,
	ManagementGroup,
	ResourceGroup,
	ServicePrincipal,
	Subscription,
	Tenant,
	User,
	VM,
	ManagedCluster,
	ContainerRegistry,
	WebApp,
	LogicApp,
	AutomationAccount,
]

AvereContributor: types.#Kind & {
	symbol:         "AvereContributor"
	schema:         "azure"
	representation: "AZAvereContributor"
}

WebsiteContributor: types.#Kind & {
	symbol:         "WebsiteContributor"
	schema:         "azure"
	representation: "AZWebsiteContributor"
}

VMAdminLogin: types.#Kind & {
	symbol:         "VMAdminLogin"
	schema:         "azure"
	representation: "AZVMAdminLogin"
}

ExecuteCommand: types.#Kind & {
	symbol:         "ExecuteCommand"
	schema:         "azure"
	representation: "AZExecuteCommand"
}

CloudAppAdmin: types.#Kind & {
	symbol:         "CloudAppAdmin"
	schema:         "azure"
	representation: "AZCloudAppAdmin"
}

AppAdmin: types.#Kind & {
	symbol:         "AppAdmin"
	schema:         "azure"
	representation: "AZAppAdmin"
}

ApplicationReadWriteAll: types.#Kind & {
	symbol:         "ApplicationReadWriteAll"
	schema:         "azure"
	representation: "AZMGApplication_ReadWrite_All"
}

AppRoleAssignmentReadWriteAll: types.#Kind & {
	symbol:         "AppRoleAssignmentReadWriteAll"
	schema:         "azure"
	representation: "AZMGAppRoleAssignment_ReadWrite_All"
}

DirectoryReadWriteAll: types.#Kind & {
	symbol:         "DirectoryReadWriteAll"
	schema:         "azure"
	representation: "AZMGDirectory_ReadWrite_All"
}

GroupReadWriteAll: types.#Kind & {
	symbol:         "GroupReadWriteAll"
	schema:         "azure"
	representation: "AZMGGroup_ReadWrite_All"
}

GroupMemberReadWriteAll: types.#Kind & {
	symbol:         "GroupMemberReadWriteAll"
	schema:         "azure"
	representation: "AZMGGroupMember_ReadWrite_All"
}

RoleManagementReadWriteDirectory: types.#Kind & {
	symbol:         "RoleManagementReadWriteDirectory"
	schema:         "azure"
	representation: "AZMGRoleManagement_ReadWrite_Directory"
}

ServicePrincipalEndpointReadWriteAll: types.#Kind & {
	symbol:         "ServicePrincipalEndpointReadWriteAll"
	schema:         "azure"
	representation: "AZMGServicePrincipalEndpoint_ReadWrite_All"
}

AZMGAddSecret: types.#Kind & {
	symbol:         "AZMGAddSecret"
	schema:         "azure"
	representation: "AZMGAddSecret"
}

AZMGAddOwner: types.#Kind & {
	symbol:         "AZMGAddOwner"
	schema:         "azure"
	representation: "AZMGAddOwner"
}

AZMGGrantAppRoles: types.#Kind & {
	symbol:         "AZMGGrantAppRoles"
	schema:         "azure"
	representation: "AZMGGrantAppRoles"
}

AZMGAddMember: types.#Kind & {
	symbol:         "AZMGAddMember"
	schema:         "azure"
	representation: "AZMGAddMember"
}

AZMGGrantRole: types.#Kind & {
	symbol:         "AZMGGrantRole"
	schema:         "azure"
	representation: "AZMGGrantRole"
}

Contains: types.#Kind & {
	symbol:         "Contains"
	schema:         "azure"
	representation: "AZContains"
}

Contributor: types.#Kind & {
	symbol:         "Contributor"
	schema:         "azure"
	representation: "AZContributor"
}

GetCertificates: types.#Kind & {
	symbol:         "GetCertificates"
	schema:         "azure"
	representation: "AZGetCertificates"
}

GetKeys: types.#Kind & {
	symbol:         "GetKeys"
	schema:         "azure"
	representation: "AZGetKeys"
}

GetSecrets: types.#Kind & {
	symbol:         "GetSecrets"
	schema:         "azure"
	representation: "AZGetSecrets"
}

HasRole: types.#Kind & {
	symbol:         "HasRole"
	schema:         "azure"
	representation: "AZHasRole"
}

MemberOf: types.#Kind & {
	symbol:         "MemberOf"
	schema:         "azure"
	representation: "AZMemberOf"
}

Owner: types.#Kind & {
	symbol:         "Owner"
	schema:         "azure"
	representation: "AZOwner"
}

Owns: types.#Kind & {
	symbol:         "Owns"
	schema:         "azure"
	representation: "AZOwns"
}

ScopedTo: types.#Kind & {
	symbol:         "ScopedTo"
	schema:         "azure"
	representation: "AZScopedTo"
}

RunsAs: types.#Kind & {
	symbol:         "RunsAs"
	schema:         "azure"
	representation: "AZRunsAs"
}

VMContributor: types.#Kind & {
	symbol:         "VMContributor"
	schema:         "azure"
	representation: "AZVMContributor"
}

AKSContributor: types.#Kind & {
	symbol:         "AKSContributor"
	schema:         "azure"
	representation: "AZAKSContributor"
}

NodeResourceGroup: types.#Kind & {
	symbol:         "NodeResourceGroup"
	schema:         "azure"
	representation: "AZNodeResourceGroup"
}

AutomationContributor: types.#Kind & {
  symbol:         "AutomationContributor"
	schema:         "azure"
	representation: "AZAutomationContributor"
}

KeyVaultContributor: types.#Kind & {
	symbol:         "KeyVaultContributor"
	schema:         "azure"
	representation: "AZKeyVaultContributor"
}

VMAdminLogin: types.#Kind & {
	symbol:         "VMAdminLogin"
	schema:         "azure"
	representation: "AZVMAdminLogin"
}

AddMembers: types.#Kind & {
	symbol:         "AddMembers"
	schema:         "azure"
	representation: "AZAddMembers"
}

AddSecret: types.#Kind & {
	symbol:         "AddSecret"
	schema:         "azure"
	representation: "AZAddSecret"
}

ExecuteCommand: types.#Kind & {
	symbol:         "ExecuteCommand"
	schema:         "azure"
	representation: "AZExecuteCommand"
}

GlobalAdmin: types.#Kind & {
	symbol:         "GlobalAdmin"
	schema:         "azure"
	representation: "AZGlobalAdmin"
}

PrivilegedAuthAdmin: types.#Kind & {
	symbol:         "PrivilegedAuthAdmin"
	schema:         "azure"
	representation: "AZPrivilegedAuthAdmin"
}

Grant: types.#Kind & {
	symbol:         "Grant"
	schema:         "azure"
	representation: "AZGrant"
}

GrantSelf: types.#Kind & {
	symbol:         "GrantSelf"
	schema:         "azure"
	representation: "AZGrantSelf"
}

PrivilegedRoleAdmin: types.#Kind & {
	symbol:         "PrivilegedRoleAdmin"
	schema:         "azure"
	representation: "AZPrivilegedRoleAdmin"
}

ResetPassword: types.#Kind & {
	symbol:         "ResetPassword"
	schema:         "azure"
	representation: "AZResetPassword"
}

UserAccessAdministrator: types.#Kind & {
	symbol:         "UserAccessAdministrator"
	schema:         "azure"
	representation: "AZUserAccessAdministrator"
}

AddOwner: types.#Kind & {
	symbol:         "AddOwner"
	schema:         "azure"
	representation: "AZAddOwner"
}

ManagedIdentity: types.#Kind & {
	symbol:         "ManagedIdentity"
	schema:         "azure"
	representation: "AZManagedIdentity"
}

LogicAppContributor: types.#Kind & {
	symbol:			"LogicAppContributor"
	schema:			"azure"
	representation:	"AZLogicAppContributor"
}

RelationshipKinds: [
	AvereContributor,
	Contains,
	Contributor,
	GetCertificates,
	GetKeys,
	GetSecrets,
	HasRole,
	MemberOf,
	Owner,
	RunsAs,
	VMContributor,
	AutomationContributor,
	KeyVaultContributor,
	VMAdminLogin,
	AddMembers,
	AddSecret,
	ExecuteCommand,
	GlobalAdmin,
	PrivilegedAuthAdmin,
	Grant,
	GrantSelf,
	PrivilegedRoleAdmin,
	ResetPassword,
	UserAccessAdministrator,
	Owns,
	ScopedTo,
	CloudAppAdmin,
	AppAdmin,
	AddOwner,
	ManagedIdentity,
	ApplicationReadWriteAll,
	AppRoleAssignmentReadWriteAll,
	DirectoryReadWriteAll,
	GroupReadWriteAll,
	GroupMemberReadWriteAll,
	RoleManagementReadWriteDirectory,
	ServicePrincipalEndpointReadWriteAll,
	AKSContributor,
	NodeResourceGroup,
	WebsiteContributor,
	LogicAppContributor,
	AZMGAddMember,
	AZMGAddOwner,
	AZMGAddSecret,
	AZMGGrantAppRoles,
	AZMGGrantRole,
]

AppRoleTransitRelationshipKinds: [
	AZMGAddMember,
	AZMGAddOwner,
	AZMGAddSecret,
	AZMGGrantAppRoles,
	AZMGGrantRole,
]

AbusableAppRoleRelationshipKinds: [
	ApplicationReadWriteAll,
	AppRoleAssignmentReadWriteAll,
	DirectoryReadWriteAll,
	GroupReadWriteAll,
	GroupMemberReadWriteAll,
	RoleManagementReadWriteDirectory,
	ServicePrincipalEndpointReadWriteAll,
]

ControlRelationshipKinds: [
	AvereContributor,
	Contributor,
	Owner,
	VMContributor,
	AutomationContributor,
	KeyVaultContributor,
	AddMembers,
	AddSecret,
	ExecuteCommand,
	GlobalAdmin,
	Grant,
	GrantSelf,
	PrivilegedRoleAdmin,
	ResetPassword,
	UserAccessAdministrator,
	Owns,
	CloudAppAdmin,
	AppAdmin,
	AddOwner,
	ManagedIdentity,
	AKSContributor,
	WebsiteContributor,
	LogicAppContributor,
	AZMGAddMember,
	AZMGAddOwner,
	AZMGAddSecret,
	AZMGGrantAppRoles,
	AZMGGrantRole,
]

ExecutionPrivilegeKinds: [
	VMAdminLogin,
	VMContributor,
	AvereContributor,
	WebsiteContributor,
	Contributor,
	ExecuteCommand,
]

PathfindingRelationships: [
	AvereContributor,
	Contains,
	Contributor,
	GetCertificates,
	GetKeys,
	GetSecrets,
	HasRole,
	MemberOf,
	Owner,
	RunsAs,
	VMContributor,
	AutomationContributor,
	KeyVaultContributor,
	VMAdminLogin,
	AddMembers,
	AddSecret,
	ExecuteCommand,
	GlobalAdmin,
	PrivilegedAuthAdmin,
	Grant,
	GrantSelf,
	PrivilegedRoleAdmin,
	ResetPassword,
	UserAccessAdministrator,
	Owns,
	CloudAppAdmin,
	AppAdmin,
	AddOwner,
	ManagedIdentity,
	AKSContributor,
	NodeResourceGroup,
	WebsiteContributor,
	LogicAppContributor,
	AZMGAddMember,
	AZMGAddOwner,
	AZMGAddSecret,
	AZMGGrantAppRoles,
	AZMGGrantRole,
]
