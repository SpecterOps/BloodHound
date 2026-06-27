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
    apiClient,
    EntityInfoCollapsibleSection,
    EntityInfoDataTableProps,
    InfiniteScrollingTable,
    useExploreParams,
} from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { getRACFGroupMembersQuery, getRACFGroupSubgroupsQuery, getRACFUserGroupsQuery } from './groupMembers';

type RACFRelatedNode = {
    objectID: string;
    name: string;
    label: string;
};

const normalizeRelatedNode = (node: Record<string, any>, fallbackKind: string): RACFRelatedNode => ({
    objectID: node.objectId || node.properties?.objectid || '',
    name: node.label || node.properties?.name || node.objectId || node.properties?.objectid || 'Unknown',
    label: node.kind || node.kinds?.[0] || fallbackKind,
});

const fetchRelatedNodes = async (
    databaseId: string,
    getQuery: (databaseId: string) => string,
    fallbackKind: string
) => {
    try {
        const response = await apiClient.cypherSearch(getQuery(databaseId), undefined, true);
        const nodes = response.data?.data?.nodes || {};

        return Object.values(nodes).map((node) => normalizeRelatedNode(node as Record<string, any>, fallbackKind));
    } catch (error) {
        const status = (error as { response?: { status?: number } }).response?.status;

        if (status === 404) {
            return [];
        }

        throw error;
    }
};

type RACFRelatedNodesProps = EntityInfoDataTableProps & {
    queryKey: string;
    getQuery: (databaseId: string) => string;
    fallbackKind: string;
};

const RACFRelatedNodes = ({
    id,
    label,
    parentLabels = [],
    queryKey,
    getQuery,
    fallbackKind,
}: RACFRelatedNodesProps) => {
    const { expandedPanelSections, setExploreParams } = useExploreParams();
    const isExpanded = !!expandedPanelSections?.includes(label);

    const relatedNodesQuery = useQuery([queryKey, id], () => fetchRelatedNodes(id, getQuery, fallbackKind), {
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
