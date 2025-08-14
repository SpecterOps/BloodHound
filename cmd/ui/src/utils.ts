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

import { GlyphIconInfo, apiClient } from 'bh-shared-ui';
import identity from 'lodash/identity';
import throttle from 'lodash/throttle';
import type { SigmaEdgeEventPayload, SigmaNodeEventPayload } from 'sigma/sigma';
import type { Coordinates, MouseCoords } from 'sigma/types';
import { logout } from 'src/ducks/auth/authSlice';
import { addSnackbar } from 'src/ducks/global/actions';
import { Glyph } from 'src/rendering/programs/node.glyphs';
import { store } from 'src/store';

const IGNORE_401_LOGOUT = ['/api/v2/login', '/api/v2/logout', '/api/v2/features'];

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

/** Type guard used to narrow Sigma event payloads to Node type */
export const isNodeEvent = (event: SigmaEdgeEventPayload | SigmaNodeEventPayload): event is SigmaNodeEventPayload =>
    'node' in event;

/**
 * Reusable method to prevent defaults for mouse move, right click, and double click
 *
 * @param event Sigma or mouse event object used to cancel defaults
 */
export const preventAllDefaults = (event: SigmaNodeEventPayload | MouseCoords) => {
    if ('preventSigmaDefault' in event && typeof event.preventSigmaDefault === 'function') {
        event.preventSigmaDefault();
    }

    // Prevent events for MouseCoords type
    if ('original' in event && event.original instanceof MouseEvent) {
        event.original.preventDefault();
        event.original.stopPropagation();
    }
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
                if (error?.response?.status === 401) {
                    if (IGNORE_401_LOGOUT.includes(error?.response?.config.url) === false) {
                        throttledLogout();
                    }
                } else if (
                    error?.response?.status === 403 &&
                    !error?.response?.config.url.match('/api/v2/bloodhound-users/[a-z0-9-]+/secret')
                ) {
                    store.dispatch(addSnackbar('Permission denied!', 'permissionDenied'));
                }
            }
            return Promise.reject(error);
        }
    );
};

type ThemedLabels = {
    labelColor: string;
    backgroundColor: string;
    highlightedBackground: string;
    highlightedText: string;
};

type ThemedGlyph = {
    colors: {
        backgroundColor: string;
        color: string;
    };
    tierZeroGlyph: GlyphIconInfo;
    ownedObjectGlyph: GlyphIconInfo;
};

export type ThemedOptions = {
    labels: ThemedLabels;
    nodeBorderColor: string;
    glyph: ThemedGlyph;
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
    hidden?: boolean;
} & ThemedLabels;

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
    exploreGraphId: string;
    groupPosition?: number;
    groupSize?: number;
    direction?: EdgeDirection;
    control?: Coordinates;
    controlInViewport?: Coordinates;
    forceLabel?: boolean;
} & ThemedLabels;
