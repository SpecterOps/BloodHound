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

import { apiClient } from 'bh-shared-ui';
import { FlatGraphResponse, GraphData, GraphResponse, StyledGraphEdge, StyledGraphNode } from 'js-client-library';
import identity from 'lodash/identity';
import throttle from 'lodash/throttle';
import { Coordinates } from 'sigma/types';
import { logout } from 'src/ducks/auth/authSlice';
import { addSnackbar } from 'src/ducks/global/actions';
import { isLink, isNode } from 'src/ducks/graph/utils';
import { Glyph } from 'src/rendering/programs/node.glyphs';
import { store } from 'src/store';

export const getDatesInRange = (startDate: Date, endDate: Date) => {
    const date = new Date(startDate.getTime());

    date.setDate(date.getDate());

    const dates = [];

    while (date < endDate) {
        dates.push(new Date(date));
        date.setDate(date.getDate() + 1);
    }

    return dates;
};

export const getUsername = (user: any): string | undefined => {
    if (user?.first_name && user?.last_name) {
        return `${user.first_name} ${user.last_name}`;
    } else if (user?.first_name) {
        return user.first_name;
    } else if (user?.principal_name) {
        return user.principal_name;
    }
    return undefined;
};

const throttledLogout = throttle(() => {
    store.dispatch(logout());
}, 2000);

export const initializeBHEClient = () => {
    // attach session token from store to each request
    apiClient.baseClient.interceptors.request.use(
        (request) => {
            const sessionToken = store.getState().auth.sessionToken;
            if (sessionToken) {
                request.headers['Authorization'] = `Bearer ${sessionToken}`;
            }
            return request;
        },
        (error) => Promise.reject(error)
    );

    // logout on 401, show notification on 403
    apiClient.baseClient.interceptors.response.use(
        identity,

        (error) => {
            if (error?.response) {
                if (
                    error?.response?.status === 401 &&
                    error?.response?.config.url !== '/api/v2/login' &&
                    error?.response?.config.url !== '/api/v2/logout'
                ) {
                    throttledLogout();
                } else if (error?.response?.status === 403) {
                    store.dispatch(addSnackbar('Permission denied!', 'permissionDenied'));
                }
            }
            return Promise.reject(error);
        }
    );
};

export type NodeParams = {
    x: number;
    y: number;
    size?: number;
    color?: string;
    borderColor?: string;
    type?: string;
    highlighted?: boolean;
    image?: string;
    label?: string;
    glyphs?: Glyph[];
    forceLabel?: boolean;
};

export interface Index<T> {
    [id: string]: T;
}

export type Items = Record<string, any>;

export enum EdgeDirection {
    FORWARDS = 1,
    BACKWARDS = -1,
}

export type EdgeParams = {
    size: number;
    type: string;
    label: string;
    color: string;
    exploreGraphId: string;
    groupPosition?: number;
    groupSize?: number;
    direction?: EdgeDirection;
    control?: Coordinates;
    controlInViewport?: Coordinates;
    forceLabel?: boolean;
};

const getLastSeenValue = (object: any): string => {
    if (object.lastSeen) return object.lastSeen;
    if (object.data) {
        if (object.data.lastSeen) return object.data.lastSeen;
        if (object.data.lastseen) return object.data.lastseen;
    }

    return '';
};

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
                kind: node.data.nodetype,
                objectId: node.data.objectid,
                isTierZero: !!(node.data.system_tags && node.data.system_tags.indexOf('admin_tier_0') !== -1),
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

export const transformToFlatGraphResponse = (graph: GraphResponse) => {
    const result: any = {};
    for (const [key, value] of Object.entries(graph.data.nodes)) {
        const lastSeen = getLastSeenValue(value);
        result[key] = {
            label: {
                text: value.label,
            },
            data: {
                nodetype: value.kind,
                name: value.label,
                objectid: value.objectId,
                system_tags: value.isTierZero ? 'admin_tier_0' : undefined,
                lastseen: lastSeen,
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
            data: { ...(edge.data || {}), lastseen: lastSeen },
        };
    }
    return result;
};
