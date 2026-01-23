// Copyright 2025 Specter Ops, Inc.
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

import { FlatGraphResponse, GraphResponse, StyledGraphEdge, StyledGraphNode, type GraphData } from 'js-client-library';
import { UseQueryOptions } from 'react-query';
import { BUILTIN_EDGE_CATEGORIES } from '../../../views/Explore/ExploreSearch/EdgeFilter/edgeCategories';
import { ExploreQueryParams } from '../../useExploreParams';
import { getInitialPathFilters } from '../utils';

type QueryKeys = ('explore-graph-query' | string | undefined)[];

export type ExploreGraphQueryOptions = UseQueryOptions<
    GraphResponse | FlatGraphResponse,
    unknown,
    GraphResponse | FlatGraphResponse,
    QueryKeys
>;

export type GraphItemMutationFn = (items: any) => unknown;

export type ExploreGraphQueryError = { message: string; key: string };

export type ExploreGraphQuery = {
    getQueryConfig: (paramOptions: Partial<ExploreQueryParams>) => ExploreGraphQueryOptions;
    getErrorMessage: (error: any) => ExploreGraphQueryError;
};

export const ExploreGraphQueryKey = 'explore-graph-query';

export const INITIAL_FILTERS = getInitialPathFilters(BUILTIN_EDGE_CATEGORIES);
export const EMPTY_FILTER_VALUE = 'empty';

export const sharedGraphQueryOptions: ExploreGraphQueryOptions = {
    retry: false,
    refetchOnWindowFocus: false,
};

// Checks if a list of path filters consists of the string 'empty', which indicates all filters are unchecked
export const areFiltersEmpty = (types: string[] | null | undefined) => {
    return !!(types && types[0] === EMPTY_FILTER_VALUE);
};

// creates an inclusive filter string formatted for the API from a list of edge types
export const createPathFilterString = (types: string[]) => {
    return `in:${types.join(',')}`;
};

// Converts between two different respresentations of graph data returned by our API for endpoints that feed the explore page
export const transformFlatGraphResponse = (graph: FlatGraphResponse): GraphData => {
    const result: GraphData = {
        nodes: {},
        edges: [],
    };

    for (const [key, item] of Object.entries(graph)) {
        if (isNode(item)) {
            const node = item as StyledGraphNode;
            const lastSeen = getLastSeenValue(node);
            result.nodes[key] = {
                label: node.label.text || '',
                kind: node.data.nodetype || '',
                kinds: node.data.kinds || [],
                objectId: node.data.objectid || '',
                isTierZero: !!(node.data.system_tags && node.data.system_tags.indexOf('admin_tier_0') !== -1),
                isOwnedObject: !!(node.data.system_tags && node.data.system_tags.indexOf('owned') !== -1),
                lastSeen: lastSeen,
            };
        } else if (isLink(item)) {
            const edge = item as StyledGraphEdge;
            const lastSeen = getLastSeenValue(edge);
            result.edges.push({
                impactPercent: edge.data ? edge.data.composite_risk_impact_percent : undefined,
                source: edge.id1,
                target: edge.id2,
                label: edge.label.text || '',
                kind: edge.label.text || '',
                lastSeen: lastSeen,
                exploreGraphId: key || `${edge.id1}_${edge.label.text}_${edge.id2}`,
                data: { ...(edge.data || {}), lastseen: lastSeen },
            });
        }
    }

    return result;
};

const getLastSeenValue = (object: any): string => {
    if (object.lastSeen) return object.lastSeen;
    if (object.data) {
        if (object.data.lastSeen) return object.data.lastSeen;
        if (object.data.lastseen) return object.data.lastseen;
    }

    return '';
};

const isLink = (item: any): boolean => {
    return item?.id1 !== undefined;
};

const isNode = (item: any): boolean => {
    return !isLink(item);
};

export const isGraphResponse = (graphData: GraphResponse | FlatGraphResponse): graphData is GraphResponse => {
    return !!(graphData as GraphResponse)?.data?.nodes && !!(graphData as GraphResponse)?.data?.edges;
};
