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

import {
    EntityInfoCollapsibleSection,
    EntityInfoDataTableProps,
    InfiniteScrollingTable,
    useExploreParams,
} from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { getRACFGroupMembersQuery, getRACFGroupSubgroupsQuery, getRACFUserGroupsQuery } from './groupMembers';
import { fetchRACFRelatedNodes } from './relatedNodes';

export type RACFRelatedNodesProps = EntityInfoDataTableProps & {
    queryKey: string;
    getQuery: (databaseId: string) => string;
    fallbackKind: string;
};

export const RACFRelatedNodes = ({
    id,
    label,
    parentLabels = [],
    queryKey,
    getQuery,
    fallbackKind,
}: RACFRelatedNodesProps) => {
    const { expandedPanelSections, setExploreParams } = useExploreParams();
    const isExpanded = !!expandedPanelSections?.includes(label);

    const relatedNodesQuery = useQuery([queryKey, id], () => fetchRACFRelatedNodes(id, getQuery, fallbackKind), {
        refetchOnWindowFocus: false,
        retry: false,
    });

    const relatedNodes = relatedNodesQuery.data || [];

    const handleOnChange = (isOpen: boolean) => {
        setExploreParams({
            expandedPanelSections: isOpen ? [...parentLabels, label] : parentLabels,
        });
    };

    const handleOnClick = (member: { id: string }) => {
        setExploreParams({
            primarySearch: member.id,
            searchType: 'node',
            exploreSearchTab: 'node',
        });
    };

    return (
        <EntityInfoCollapsibleSection
            label={label}
            count={relatedNodes.length}
            isExpanded={isExpanded}
            isLoading={relatedNodesQuery.isLoading}
            isError={relatedNodesQuery.isError}
            error={relatedNodesQuery.error}
            onChange={handleOnChange}>
            <InfiniteScrollingTable
                itemCount={relatedNodes.length}
                fetchDataCallback={({ skip, limit }) =>
                    Promise.resolve({
                        data: relatedNodes.slice(skip, skip + limit),
                        total: relatedNodes.length,
                        skip,
                        limit,
                    })
                }
                onClick={handleOnClick}
            />
        </EntityInfoCollapsibleSection>
    );
};

export const RACFGroupMembers = (props: EntityInfoDataTableProps) => (
    <RACFRelatedNodes
        {...props}
        queryKey='racf-group-members'
        getQuery={getRACFGroupMembersQuery}
        fallbackKind='RACFUser'
    />
);

export const RACFGroupSubgroups = (props: EntityInfoDataTableProps) => (
    <RACFRelatedNodes
        {...props}
        queryKey='racf-group-subgroups'
        getQuery={getRACFGroupSubgroupsQuery}
        fallbackKind='RACFGroup'
    />
);

export const RACFUserGroups = (props: EntityInfoDataTableProps) => (
    <RACFRelatedNodes
        {...props}
        queryKey='racf-user-groups'
        getQuery={getRACFUserGroupsQuery}
        fallbackKind='RACFGroup'
    />
);
