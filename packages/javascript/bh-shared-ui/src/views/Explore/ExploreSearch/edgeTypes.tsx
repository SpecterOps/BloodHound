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

import { ActiveDirectoryRelationshipKind, AzureRelationshipKind } from '../../../graphSchema';

export type Category = {
    categoryName: string;
    subcategories: Subcategory[];
};

export type Subcategory = {
    name: string;
    edgeTypes: string[];
};

export const AllEdgeTypes: Category[] = [
    {
        categoryName: 'Active Directory',
        subcategories: [
            {
                name: 'Active Directory Structure',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.Contains,
                    ActiveDirectoryRelationshipKind.GPLink,
                    ActiveDirectoryRelationshipKind.HasSIDHistory,
                    ActiveDirectoryRelationshipKind.MemberOf,
                    ActiveDirectoryRelationshipKind.TrustedBy,
                ],
            },
            {
                name: 'Lateral Movement',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.AdminTo,
                    ActiveDirectoryRelationshipKind.AllowedToAct,
                    ActiveDirectoryRelationshipKind.AllowedToDelegate,
                    ActiveDirectoryRelationshipKind.CanPSRemote,
                    ActiveDirectoryRelationshipKind.CanRDP,
                    ActiveDirectoryRelationshipKind.ExecuteDCOM,
                    ActiveDirectoryRelationshipKind.SQLAdmin,
                ],
            },
            {
                name: 'Credential Access',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.DCSync,
                    ActiveDirectoryRelationshipKind.DumpSMSAPassword,
                    ActiveDirectoryRelationshipKind.HasSession,
                    ActiveDirectoryRelationshipKind.ReadGMSAPassword,
                    ActiveDirectoryRelationshipKind.ReadLAPSPassword,
                    ActiveDirectoryRelationshipKind.SyncLAPSPassword,
                ],
            },
            {
                name: 'Basic Object Manipulation',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.AddMember,
                    ActiveDirectoryRelationshipKind.AddSelf,
                    ActiveDirectoryRelationshipKind.AllExtendedRights,
                    ActiveDirectoryRelationshipKind.ForceChangePassword,
                    ActiveDirectoryRelationshipKind.GenericAll,
                    ActiveDirectoryRelationshipKind.Owns,
                    ActiveDirectoryRelationshipKind.GenericWrite,
                    ActiveDirectoryRelationshipKind.WriteDACL,
                    ActiveDirectoryRelationshipKind.WriteOwner,
                ],
            },
            {
                name: 'Advanced Object Manipulation',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.AddAllowedToAct,
                    ActiveDirectoryRelationshipKind.AddKeyCredentialLink,
                    ActiveDirectoryRelationshipKind.WriteAccountRestrictions,
                    ActiveDirectoryRelationshipKind.WriteSPN,
                ],
            },
            {
                name: 'Active Directory Certificate Services',
                edgeTypes: [
                    ActiveDirectoryRelationshipKind.GoldenCert,
                    ActiveDirectoryRelationshipKind.ADCSESC1,
                    ActiveDirectoryRelationshipKind.ADCSESC3,
                    ActiveDirectoryRelationshipKind.ADCSESC6a,
                    ActiveDirectoryRelationshipKind.ADCSESC6b,
                    ActiveDirectoryRelationshipKind.ADCSESC9a,
                    ActiveDirectoryRelationshipKind.ADCSESC9b,
                    ActiveDirectoryRelationshipKind.ADCSESC10a,
                    ActiveDirectoryRelationshipKind.ADCSESC10b,
                ],
            },
        ],
    },
    {
        categoryName: 'Azure',
        subcategories: [
            {
                name: 'Structure',
                edgeTypes: [
                    AzureRelationshipKind.AppAdmin,
                    AzureRelationshipKind.CloudAppAdmin,
                    AzureRelationshipKind.Contains,
                    AzureRelationshipKind.GlobalAdmin,
                    AzureRelationshipKind.HasRole,
                    AzureRelationshipKind.ManagedIdentity,
                    AzureRelationshipKind.MemberOf,
                    AzureRelationshipKind.NodeResourceGroup,
                    AzureRelationshipKind.PrivilegedAuthAdmin,
                    AzureRelationshipKind.PrivilegedRoleAdmin,
                    AzureRelationshipKind.RunsAs,
                ],
            },
            {
                name: 'Basic AzureAD Object Manipulation',
                edgeTypes: [
                    AzureRelationshipKind.AddMembers,
                    AzureRelationshipKind.AddOwner,
                    AzureRelationshipKind.AddSecret,
                    AzureRelationshipKind.ExecuteCommand,
                    AzureRelationshipKind.Grant,
                    AzureRelationshipKind.GrantSelf,
                    AzureRelationshipKind.Owns,
                    AzureRelationshipKind.ResetPassword,
                ],
            },
            {
                name: 'MS Graph App Role Abuses',
                edgeTypes: [
                    AzureRelationshipKind.AZMGAddMember,
                    AzureRelationshipKind.AZMGAddOwner,
                    AzureRelationshipKind.AZMGAddSecret,
                    AzureRelationshipKind.AZMGGrantAppRoles,
                    AzureRelationshipKind.AZMGGrantRole,
                ],
            },
            {
                name: 'Secret/Credential Access',
                edgeTypes: [
                    AzureRelationshipKind.GetCertificates,
                    AzureRelationshipKind.GetKeys,
                    AzureRelationshipKind.GetSecrets,
                ],
            },
            {
                name: 'Basic AzureRM Object Manipulation',
                edgeTypes: [
                    AzureRelationshipKind.AvereContributor,
                    AzureRelationshipKind.KeyVaultContributor,
                    AzureRelationshipKind.Owner,
                    AzureRelationshipKind.Contributor,
                    AzureRelationshipKind.UserAccessAdministrator,
                    AzureRelationshipKind.VMAdminLogin,
                    AzureRelationshipKind.VMContributor,
                ],
            },
            {
                name: 'Advanced AzureRM Object Manipulation',
                edgeTypes: [
                    AzureRelationshipKind.AKSContributor,
                    AzureRelationshipKind.AutomationContributor,
                    AzureRelationshipKind.LogicAppContributor,
                    AzureRelationshipKind.WebsiteContributor,
                ],
            },
        ],
    },
];
