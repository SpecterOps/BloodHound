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

export const RACF_GROUP_KINDS = ['RACFGroup', 'racf_RACFGroup'] as const;
export const RACF_USER_KINDS = ['RACFUser', 'racf_RACFUser'] as const;
export const RACF_GROUP_MEMBERS_SECTION = 'All Members';
export const RACF_GROUP_SUBGROUPS_SECTION = 'Subgroups';
export const RACF_USER_GROUPS_SECTION = 'Groups';

export const isRACFGroupKind = (kind: string): boolean =>
    RACF_GROUP_KINDS.some(
        (racfGroupKind) => racfGroupKind.localeCompare(kind, undefined, { sensitivity: 'base' }) === 0
    );

export const isRACFUserKind = (kind: string): boolean =>
    RACF_USER_KINDS.some((racfUserKind) => racfUserKind.localeCompare(kind, undefined, { sensitivity: 'base' }) === 0);

export const getRACFGroupMembersQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF group database ID must be an integer');
    }

    return `
MATCH (group)
WHERE ID(group) = ${databaseId}
MATCH (member)-[:RACFMemberOf|racf_RACFMemberOf]->(group)
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
MATCH (group)-[:RACFHasSubgroup|racf_RACFHasSubgroup]->(subgroup)
RETURN DISTINCT subgroup
ORDER BY subgroup.name`;
};

export const getRACFUserGroupsQuery = (databaseId: string): string => {
    if (!/^\d+$/.test(databaseId)) {
        throw new Error('RACF user database ID must be an integer');
    }

    return `
MATCH (user)
WHERE ID(user) = ${databaseId}
MATCH (user)-[:RACFMemberOf|racf_RACFMemberOf]->(group)
RETURN DISTINCT group
ORDER BY group.name`;
};
