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

import type { EntityTables } from 'bh-shared-ui';
import { RACFClassUsersWithCLAUTH } from 'src/racfhound/RACFClassRelationships';
import { RACFGroupMembers, RACFGroupSubgroups, RACFUserGroups } from 'src/racfhound/RACFGroupMembers';
import { RACFGroupOutboundRelationships } from 'src/racfhound/RACFGroupRelationships';
import { RACFUserInboundRelationships, RACFUserOutboundRelationships } from 'src/racfhound/RACFUserRelationships';
import {
    isRACFClassKind,
    isRACFGroupKind,
    isRACFUserKind,
    RACF_CLASS_USERS_WITH_CLAUTH_SECTION,
    RACF_GROUP_MEMBERS_SECTION,
    RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
    RACF_GROUP_SUBGROUPS_SECTION,
    RACF_USER_GROUPS_SECTION,
    RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
    RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
} from 'src/racfhound/groupMembers';

// getRACFTables returns the RACF-specific relationship tables for a given node kind, or
// undefined when the kind is not a RACF kind. `databaseId` is the numeric graph node id
// (the `selectedItem` explore param) that the RACF cypher queries filter on via ID(n).
const getRACFTables = (nodeType: string, databaseId: string): EntityTables | undefined => {
    if (isRACFGroupKind(nodeType)) {
        return [
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_GROUP_MEMBERS_SECTION,
                },
                TableComponent: RACFGroupMembers,
            },
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_GROUP_SUBGROUPS_SECTION,
                },
                TableComponent: RACFGroupSubgroups,
            },
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_GROUP_OUTBOUND_RELATIONSHIPS_SECTION,
                },
                TableComponent: RACFGroupOutboundRelationships,
            },
        ];
    }

    if (isRACFUserKind(nodeType)) {
        return [
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_USER_GROUPS_SECTION,
                },
                TableComponent: RACFUserGroups,
            },
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_USER_OUTBOUND_RELATIONSHIPS_SECTION,
                },
                TableComponent: RACFUserOutboundRelationships,
            },
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_USER_INBOUND_RELATIONSHIPS_SECTION,
                },
                TableComponent: RACFUserInboundRelationships,
            },
        ];
    }

    if (isRACFClassKind(nodeType)) {
        return [
            {
                sectionProps: {
                    id: databaseId,
                    label: RACF_CLASS_USERS_WITH_CLAUTH_SECTION,
                },
                TableComponent: RACFClassUsersWithCLAUTH,
            },
        ];
    }

    return undefined;
};

// getRACFAdditionalTables adapts getRACFTables to the upstream node response shape, which
// carries a `kinds` array of { name } refs rather than a single `kind` string. It selects the
// first RACF kind present and returns its relationship tables, or undefined when the node is
// not a RACF node.
export const getRACFAdditionalTables = (
    node: { kinds?: { name: string }[] },
    databaseId: string
): EntityTables | undefined => {
    const racfKind = (node.kinds ?? [])
        .map((kind) => kind.name)
        .find((name) => isRACFGroupKind(name) || isRACFUserKind(name) || isRACFClassKind(name));

    if (!racfKind) {
        return undefined;
    }

    return getRACFTables(racfKind, databaseId);
};
