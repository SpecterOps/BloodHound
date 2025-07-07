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

import { GraphData, GraphResponse, SigmaFlatGraphData, StyledGraphEdge, StyledGraphNode } from 'js-client-library';
import { UseQueryOptions } from 'react-query';
import { ExploreQueryParams } from '../../useExploreParams';
import { extractEdgeTypes, getInitialPathFilters } from '../utils';

type QueryKeys = ('explore-graph-query' | string | undefined)[];

export type ExploreGraphQueryOptions = UseQueryOptions<
    GraphResponse | SigmaFlatGraphData,
    unknown,
    GraphResponse | SigmaFlatGraphData,
    QueryKeys
>;

export type GraphItemMutationFn = (items: any) => unknown;

export type ExploreGraphQueryError = { message: string; key: string };

export type ExploreGraphQuery = {
    getQueryConfig: (paramOptions: Partial<ExploreQueryParams>) => ExploreGraphQueryOptions;
    getErrorMessage: (error: any) => ExploreGraphQueryError;
};

export const ExploreGraphQueryKey = 'explore-graph-query';

export const INITIAL_FILTERS = getInitialPathFilters();
export const INITIAL_FILTER_TYPES = extractEdgeTypes(INITIAL_FILTERS);
export const EMPTY_FILTER_VALUE = 'empty';

export const sharedGraphQueryOptions: ExploreGraphQueryOptions = {
    retry: false,
    refetchOnWindowFocus: false,
};

// creates a filter string in our API format, handling the case that our 'empty' value is in the url param
export const createPathFilterString = (types: string[]) => {
    if (types[0] === EMPTY_FILTER_VALUE) {
        return `nin:${INITIAL_FILTER_TYPES.join(',')}`;
    }
    return `in:${types.join(',')}`;
};

// Converts between two different respresentations of graph data returned by our API for endpoints that feed the explore page
export const transformSigmaFlatGraphData = (graph: SigmaFlatGraphData): GraphData => {
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

// Converts the same data types in the opposite direction. We have some typing issues here due to the "lastSeen" property we are adding that should be addressed
export const transformToSigmaFlatGraphData = (graph: GraphResponse) => {
    const result: any = {};
    for (const [key, value] of Object.entries(graph.data.nodes)) {
        const lastSeen = getLastSeenValue(value);
        // Check and add needed system_tags to node
        const tags = [];
        {
            value.isTierZero ? tags.push('admin_tier_0') : null;
        }
        {
            value.isOwnedObject ? tags.push('owned') : null;
        }
        result[key] = {
            label: {
                text: value.label,
            },
            data: {
                nodetype: value.kind,
                name: value.label,
                objectid: value.objectId,
                system_tags: tags.join(' '),
                lastseen: lastSeen,
                isTierZero: value.isTierZero,
                ...(value?.properties || {}),
            },
        };
    }
    for (const edge of graph.data.edges) {
        const lastSeen = getLastSeenValue(edge);
        result[`${edge.source}_${edge.kind}_${edge.target}`] = {
            id1: edge.source,
            id2: edge.target,
            label: {
                text: edge.label,
            },
            lastSeen: lastSeen,
            data: {
                ...(edge.data || {}),
                lastseen: lastSeen,
                ...(edge.data?.properties || {}),
            },
        };
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

export const isGraphResponse = (graphData: GraphResponse | SigmaFlatGraphData): graphData is GraphResponse => {
    return !!(graphData as GraphResponse)?.data?.nodes && !!(graphData as GraphResponse)?.data?.edges;
};
