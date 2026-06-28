// Copyright 2026 Specter Ops, Inc.
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

import { CommonSearchType } from './types';
import { RACF_NODE_KINDS } from './utils/racfNodeIcons';

const categoryRACF = 'RACF';

const relationship = {
    CanAccessKey: 'RACFCanAccessKey',
    CanRead: 'RACFCanRead',
    CanWrite: 'RACFCanWrite',
    CertificateFor: 'RACFCertificateFor',
    ClassAuth: 'RACFClassAuth',
    GroupRevoke: 'RACFGroupRevoke',
    GroupScopeOper: 'RACFGroupScopeOper',
    GroupScopeSpecial: 'RACFGroupScopeSpecial',
    HasPrivilege: 'RACFHasPrivilege',
    HasSubgroup: 'RACFHasSubgroup',
    MemberOf: 'RACFMemberOf',
    Owns: 'RACFOwns',
    PassticketFor: 'RACFPassticketFor',
    StartedTaskRunsAs: 'RACFStartedTaskRunsAs',
    SurrogateFor: 'RACFSurrogateFor',
} as const;

export const RACF_RELATIONSHIP_KINDS: string[] = Object.values(relationship);
export const RACF_NODE_KIND_VALUES: string[] = Object.values(RACF_NODE_KINDS);

const effectivePrivilegeQuery = (privilegeNames: string[]) =>
    `MATCH p=(source:${RACF_NODE_KINDS.User})-[:${relationship.MemberOf}|${relationship.SurrogateFor}|${relationship.PassticketFor}*0..6]->(principal)-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})\n` +
    `WHERE toUpper(privilege.name) IN [${privilegeNames.map((name) => `'${name}'`).join(', ')}]\n` +
    `RETURN p\nLIMIT 500`;

const effectiveDatasetWriteQuery = (property: string) =>
    `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.MemberOf}*0..1]->(principal)-[:${relationship.CanWrite}]->(dataset:${RACF_NODE_KINDS.Dataset})\n` +
    `WHERE dataset.${property} = true\n` +
    `RETURN p\nLIMIT 500`;

export const RACFCommonSearches: CommonSearchType[] = [
    {
        subheader: 'Identities and Membership',
        category: categoryRACF,
        queries: [
            {
                name: 'RACF direct user-to-group memberships',
                description:
                    'Direct RACF group connections. Superior-group permissions are not inferred for subgroup members.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.MemberOf}]->(g:${RACF_NODE_KINDS.Group})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF users with TSO access',
                description: 'Users with an active TSO segment and therefore an interactive execution path.',
                query: `MATCH (u:${RACF_NODE_KINDS.User})
WHERE u.hasTSO = true
RETURN u
LIMIT 500`,
            },
            {
                name: 'RACF protected non-interactive users',
                description:
                    'Protected identities that cannot authenticate with a password and normally back services or jobs.',
                query: `MATCH (u:${RACF_NODE_KINDS.User})
WHERE u.nopwd = 'PRO'
RETURN u
LIMIT 500`,
            },
            {
                name: 'RACF started-task identities',
                description: 'Maps started tasks to the RACF user IDs under which they execute.',
                query: `MATCH p=(task:${RACF_NODE_KINDS.StartedTask})-[:${relationship.StartedTaskRunsAs}]->(u:${RACF_NODE_KINDS.User})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF revoked group connections',
                description:
                    'User-to-group connections that are individually revoked but retain latent access if restored.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.GroupRevoke}]->(g:${RACF_NODE_KINDS.Group})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF orphaned ACL principals',
                description: 'ACL entries that reference undefined or deleted authorization IDs.',
                query: `MATCH (principal:${RACF_NODE_KINDS.Undefined})
WHERE principal.isOrphan = true
RETURN principal
LIMIT 500`,
            },
        ],
    },
    {
        subheader: 'Administrative Privileges',
        category: categoryRACF,
        queries: [
            {
                name: 'RACF effective paths to SPECIAL',
                description:
                    'Directed paths through direct membership, SURROGAT, or passticket transitions followed by a SPECIAL grant. Zero transit hops include direct SPECIAL users.',
                query: effectivePrivilegeQuery(['SPECIAL']),
            },
            {
                name: 'RACF effective paths to OPERATIONS',
                description: 'Directed identity, membership, and privilege-grant paths to OPERATIONS.',
                query: effectivePrivilegeQuery(['OPERATIONS']),
            },
            {
                name: 'RACF effective paths to BPX.SUPERUSER',
                description: 'Directed identity, membership, and privilege-grant paths to BPX.SUPERUSER.',
                query: effectivePrivilegeQuery(['BPX.SUPERUSER']),
            },
            {
                name: 'RACF effective paths to IRR.PASSWORD.RESET',
                description: 'Directed identity, membership, and privilege-grant paths to IRR.PASSWORD.RESET.',
                query: effectivePrivilegeQuery(['IRR.PASSWORD.RESET']),
            },
            {
                name: 'RACF effective paths to STGADMIN dump or restore',
                description:
                    'Directed identity, membership, and privilege-grant paths to ADRDSSU dump or restore rights.',
                query: effectivePrivilegeQuery(['STGADMIN.ADR.DUMP', 'STGADMIN.ADR.RESTORE']),
            },
            {
                name: 'RACF FACILITY and SURROGAT class authorities',
                description: 'Users with CLAUTH for security-sensitive general-resource classes.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.ClassAuth}]->(class:${RACF_NODE_KINDS.Class})
WHERE toUpper(class.name) IN ['FACILITY', 'SURROGAT']
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF Group-SPECIAL administrative scope',
                description:
                    'Shows each Group-SPECIAL grant and its subordinate-group administrative scope. This is the only built-in query that traverses RACFHasSubgroup; it does not imply permission inheritance.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.GroupScopeSpecial}]->(scopeRoot:${RACF_NODE_KINDS.Group})-[:${relationship.HasSubgroup}*0..10]->(inScope:${RACF_NODE_KINDS.Group})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF users with group-OPERATIONS',
                description:
                    'Direct group-OPERATIONS grants. Subgroup hierarchy is intentionally not traversed because the authority follows resource scope, not superior-group permission inheritance.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.GroupScopeOper}]->(g:${RACF_NODE_KINDS.Group})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF TRUSTED and PRIVILEGED started tasks',
                description: 'Started tasks carrying TRUSTED or PRIVILEGED authority.',
                query: `MATCH p=(task:${RACF_NODE_KINDS.StartedTask})-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})
WHERE toUpper(privilege.name) IN ['TRUSTED', 'PRIVILEGED']
RETURN p
LIMIT 500`,
            },
        ],
    },
    {
        subheader: 'Authentication and Delegation',
        category: categoryRACF,
        queries: [
            {
                name: 'RACF privileged users without MFA',
                description: 'Users with a direct named privilege and no registered MFA factor.',
                query: `MATCH p=(u:${RACF_NODE_KINDS.User})-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})
WHERE (u.hasMFA IS NULL OR u.hasMFA = false)
AND (u.nopwd IS NULL OR u.nopwd <> 'PRO')
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF users with legacy password algorithms',
                description: 'Users whose active password or passphrase is marked as using a legacy algorithm.',
                query: `MATCH (u:${RACF_NODE_KINDS.User})
WHERE (u.pwd_alg <> 'NOPASSWORD' AND u.pwd_alg='LEGACY')
OR (u.phr_alg <> 'NOPHRASE' AND u.phr_alg='LEGACY')
RETURN u
LIMIT 500`,
            },
            {
                name: 'RACF certificate-to-user associations',
                description: 'Certificates that authenticate as RACF user IDs.',
                query: `MATCH p=(certificate:${RACF_NODE_KINDS.Certificate})-[:${relationship.CertificateFor}]->(u:${RACF_NODE_KINDS.User})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF certificate paths to SPECIAL',
                description: 'Certificates that authenticate directly as a user holding SPECIAL.',
                query: `MATCH p=(certificate:${RACF_NODE_KINDS.Certificate})-[:${relationship.CertificateFor}]->(u:${RACF_NODE_KINDS.User})-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})
WHERE toUpper(privilege.name) = 'SPECIAL'
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF passticket grants',
                description: 'Principals that can generate a passticket for another RACF identity.',
                query: `MATCH p=(principal)-[:${relationship.PassticketFor}]->(u:${RACF_NODE_KINDS.User})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF passticket paths to SPECIAL',
                description: 'Passticket rights targeting a user who directly holds SPECIAL.',
                query: `MATCH p=(principal)-[:${relationship.PassticketFor}]->(u:${RACF_NODE_KINDS.User})-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})
WHERE toUpper(privilege.name) = 'SPECIAL'
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF direct surrogate grants',
                description: 'Direct SURROGAT authority to submit work under a target RACF user ID.',
                query: `MATCH p=(principal)-[:${relationship.SurrogateFor}]->(u:${RACF_NODE_KINDS.User})
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF surrogate chains to SPECIAL',
                description: 'Directed SURROGAT chains ending at a user who directly holds SPECIAL.',
                query: `MATCH p=(source)-[:${relationship.SurrogateFor}*1..10]->(u:${RACF_NODE_KINDS.User})-[:${relationship.HasPrivilege}]->(privilege:${RACF_NODE_KINDS.Privilege})
WHERE toUpper(privilege.name) = 'SPECIAL'
RETURN p
LIMIT 500`,
            },
        ],
    },
    {
        subheader: 'Sensitive Dataset and Resource Access',
        category: categoryRACF,
        queries: [
            {
                name: 'RACF effective APF-library write paths',
                description:
                    'Users that can write an APF-authorized library directly or through one direct group connection.',
                query: effectiveDatasetWriteQuery('isAPF'),
            },
            {
                name: 'RACF effective PARMLIB write paths',
                description: 'Users that can write PARMLIB directly or through one direct group connection.',
                query: effectiveDatasetWriteQuery('isPARMLIB'),
            },
            {
                name: 'RACF effective PROCLIB write paths',
                description: 'Users that can write PROCLIB directly or through one direct group connection.',
                query: effectiveDatasetWriteQuery('isPROCLIB'),
            },
            {
                name: 'RACF owners of APF-library profiles',
                description: 'Principals that own the RACF dataset profile protecting an APF-authorized library.',
                query: `MATCH p=(principal)-[:${relationship.Owns}]->(dataset:${RACF_NODE_KINDS.Dataset})
WHERE dataset.isAPF = true
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF WARNING-mode APF libraries',
                description: 'APF-authorized libraries whose protecting profile warns rather than denies access.',
                query: `MATCH p=(principal)-[:${relationship.CanRead}|${relationship.CanWrite}]->(dataset:${RACF_NODE_KINDS.Dataset})
WHERE dataset.isAPF = true
AND dataset.warning = true
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF world-writable datasets',
                description: 'Dataset write grants assigned to the synthetic PUBLIC principal.',
                query: `MATCH p=(public:${RACF_NODE_KINDS.User})-[:${relationship.CanWrite}]->(dataset:${RACF_NODE_KINDS.Dataset})
WHERE toUpper(public.name) = 'PUBLIC'
RETURN p
LIMIT 500`,
            },
            {
                name: 'RACF ICSF cryptographic-key access',
                description: 'Principals with access to ICSF CSFKEYS resources.',
                query: `MATCH p=(principal)-[:${relationship.CanAccessKey}]->(resource:${RACF_NODE_KINDS.Resource})
RETURN p
LIMIT 500`,
            },
        ],
    },
];
