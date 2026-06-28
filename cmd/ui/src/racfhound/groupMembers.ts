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

export const RACF_GROUP_KINDS = ['RACFGroup'] as const;
export const RACF_USER_KINDS = ['RACFUser'] as const;
export const RACF_CLASS_KINDS = ['RACFClass'] as const;
export const RACF_GROUP_MEMBERS_SECTION = 'All Members';
export const RACF_GROUP_SUBGROUPS_SECTION = 'Subgroups';
export const RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION = 'Outbound Relationships';
export const RACF_GROUP_CAN_SUBMIT_AS_SECTION = 'Can Submit As';
export const RACF_USER_GROUPS_SECTION = 'Groups';
export const RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION = 'Outbound Relationships';
export const RACF_USER_INBOUND_RELATIONSHIPS_SECTION = 'Inbound Relationships';
export const RACF_USER_CAN_SUBMIT_AS_SECTION = 'Can Submit As';
export const RACF_USER_SUBMITTED_AS_BY_SECTION = 'Can Be Submitted As By';
export const RACF_USER_CLASS_AUTHORITIES_SECTION = 'Class Authorities';
export const RACF_CLASS_USERS_WITH_CLAUTH_SECTION = 'Users With CLAUTH';

export const isRACFGroupKind = (kind: string): boolean =>
    RACF_GROUP_KINDS.some(
        (racfGroupKind) => racfGroupKind.localeCompare(kind, undefined, { sensitivity: 'base' }) === 0
    );

export const isRACFUserKind = (kind: string): boolean =>
    RACF_USER_KINDS.some((racfUserKind) => racfUserKind.localeCompare(kind, undefined, { sensitivity: 'base' }) === 0);

export const isRACFClassKind = (kind: string): boolean =>
    RACF_CLASS_KINDS.some(
        (racfClassKind) => racfClassKind.localeCompare(kind, undefined, { sensitivity: 'base' }) === 0
    );

export const getRACFGroupMembersQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF group database ID must be an integer');
    }

    return `
MATCH (group)
WHERE ID(group) = ${databaseId}
MATCH (member)-[:RACFMemberOf]->(group)
RETURN DISTINCT member
ORDER BY member.name`;
};

export const getRACFGroupSubgroupsQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF group database ID must be an integer');
    }

    return `
MATCH (group)
WHERE ID(group) = ${databaseId}
MATCH (group)-[:RACFHasSubgroup]->(subgroup)
RETURN DISTINCT subgroup
ORDER BY subgroup.name`;
};

export const getRACFGroupCanSubmitAsQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF group database ID must be an integer');
    }

    return `
MATCH (group)
WHERE ID(group) = ${databaseId}
MATCH (group)-[:RACFSurrogateFor]->(target)
RETURN DISTINCT target
ORDER BY target.name`;
};

export const getRACFUserGroupsQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF user database ID must be an integer');
    }

    return `
MATCH (user)
WHERE ID(user) = ${databaseId}
MATCH (user)-[:RACFMemberOf]->(group)
RETURN DISTINCT group
ORDER BY group.name`;
};

export const getRACFUserCanSubmitAsQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF user database ID must be an integer');
    }

    return `
MATCH (user)
WHERE ID(user) = ${databaseId}
MATCH (user)-[:RACFSurrogateFor]->(target)
RETURN DISTINCT target
ORDER BY target.name`;
};

export const getRACFUserSubmittedAsByQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF user database ID must be an integer');
    }

    return `
MATCH (user)
WHERE ID(user) = ${databaseId}
MATCH (principal)-[:RACFSurrogateFor]->(user)
RETURN DISTINCT principal
ORDER BY principal.name`;
};

export const getRACFUserClassAuthoritiesQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF user database ID must be an integer');
    }

    return `
MATCH (user)
WHERE ID(user) = ${databaseId}
MATCH (user)-[:RACFClassAuth]->(class)
RETURN DISTINCT class
ORDER BY class.name`;
};

export const getRACFClassUsersWithCLAUTHQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF class database ID must be an integer');
    }

    return `
MATCH (class)
WHERE ID(class) = ${databaseId}
MATCH (user)-[:RACFClassAuth]->(class)
RETURN DISTINCT user
ORDER BY user.name`;
};
