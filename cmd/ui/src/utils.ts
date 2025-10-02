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
import type { SigmaNodeEventPayload } from 'sigma/sigma';
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

export const endpoints = [
    'post ​/api​/v2​/login',
    'post ​/api​/v2​/logout',
    'get ​/api​/v2​/self',
    'get ​/api​/v2​/saml',
    'get ​/api​/v2​/saml​/sso',
    'post ​/api​/v2​/saml​/providers',
    'get ​/api​/v2​/saml​/providers​/{saml_provider_id}',
    'delete ​/api​/v2​/saml​/providers​/{saml_provider_id}',
    'get ​/api​/v2​/sso-providers',
    'post ​/api​/v2​/sso-providers​/oidc',
    'post ​/api​/v2​/sso-providers​/saml',
    'patch ​/api​/v2​/sso-providers​/{sso_provider_id}',
    'delete ​/api​/v2​/sso-providers​/{sso_provider_id}',
    'get ​/api​/v2​/sso-providers​/{sso_provider_id}​/signing-certificate',
    'get ​/api​/v2​/permissions',
    'get ​/api​/v2​/permissions​/{permission_id}',
    'get ​/api​/v2​/roles',
    'get ​/api​/v2​/roles​/{role_id}',
    'get ​/api​/v2​/tokens',
    'post ​/api​/v2​/tokens',
    'delete ​/api​/v2​/tokens​/{token_id}',
    'get ​/api​/v2​/bloodhound-users',
    'post ​/api​/v2​/bloodhound-users',
    'get ​/api​/v2​/bloodhound-users-minimal',
    'get ​/api​/v2​/bloodhound-users​/{user_id}',
    'patch ​/api​/v2​/bloodhound-users​/{user_id}',
    'delete ​/api​/v2​/bloodhound-users​/{user_id}',
    'put ​/api​/v2​/bloodhound-users​/{user_id}​/secret',
    'delete ​/api​/v2​/bloodhound-users​/{user_id}​/secret',
    'post ​/api​/v2​/bloodhound-users​/{user_id}​/mfa',
    'delete ​/api​/v2​/bloodhound-users​/{user_id}​/mfa',
    'get ​/api​/v2​/bloodhound-users​/{user_id}​/mfa-activation',
    'post ​/api​/v2​/bloodhound-users​/{user_id}​/mfa-activation',
    'get ​/api​/v2​/collectors​/{collector_type}',
    'get ​/api​/v2​/collectors​/{collector_type}​/{release_tag}',
    'get ​/api​/v2​/collectors​/{collector_type}​/{release_tag}​/checksum',
    'get ​/api​/v2​/kennel​/manifest',
    'get ​/api​/v2​/kennel​/enterprise-manifest',
    'get ​/api​/v2​/kennel​/download​/{asset_name}',
    'get ​/api​/v2​/file-upload',
    'post ​/api​/v2​/file-upload​/start',
    'post ​/api​/v2​/file-upload​/{file_upload_job_id}',
    'get ​/api​/v2​/file-upload​/{file_upload_job_id}​/completed-tasks',
    'post ​/api​/v2​/file-upload​/{file_upload_job_id}​/end',
    'get ​/api​/v2​/file-upload​/accepted-types',
    'get ​/api​/v2​/custom-nodes',
    'post ​/api​/v2​/custom-nodes',
    'get ​/api​/v2​/custom-nodes​/{kind_name}',
    'put ​/api​/v2​/custom-nodes​/{kind_name}',
    'delete ​/api​/v2​/custom-nodes​/{kind_name}',
    'get ​/api​/version',
    'get ​/api​/v2​/spec​/openapi.yaml',
    'get ​/api​/v2​/search',
    'get ​/api​/v2​/available-domains',
    'get ​/api​/v2​/audit',
    'get ​/api​/v2​/config',
    'put ​/api​/v2​/config',
    'get ​/api​/v2​/features',
    'put ​/api​/v2​/features​/{feature_id}​/toggle',
    'get ​/api​/v2​/asset-groups',
    'post ​/api​/v2​/asset-groups',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}',
    'put ​/api​/v2​/asset-groups​/{asset_group_id}',
    'delete ​/api​/v2​/asset-groups​/{asset_group_id}',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}​/collections',
    'post ​/api​/v2​/asset-groups​/{asset_group_id}​/selectors',
    'put ​/api​/v2​/asset-groups​/{asset_group_id}​/selectors',
    'delete ​/api​/v2​/asset-groups​/{asset_group_id}​/selectors​/{asset_group_selector_id}',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}​/custom-selectors',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}​/members',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}​/members​/counts',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/members​/{asset_group_member_id}',
    'get ​/api​/v2​/asset-group-tags',
    'post ​/api​/v2​/asset-group-tags',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}',
    'patch ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}',
    'delete ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/members',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/members​/counts',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors​/{asset_group_tag_selector_id}​/members',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors',
    'post ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors',
    'get ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors​/{asset_group_tag_selector_id}',
    'patch ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors​/{asset_group_tag_selector_id}',
    'delete ​/api​/v2​/asset-group-tags​/{asset_group_tag_id}​/selectors​/{asset_group_tag_selector_id}',
    'post ​/api​/v2​/asset-group-tags​/preview-selectors',
    'post ​/api​/v2​/asset-group-tags​/search',
    'get ​/api​/v2​/asset-group-tags-history',
    'post ​/api​/v2​/asset-group-tags-history',
    'post ​/api​/v2​/asset-group-tags​/certifications',
    'get ​/api​/v2​/asset-group-tags​/certifications',
    'get ​/api​/v2​/graphs​/kinds',
    'get ​/api​/v2​/pathfinding',
    'get ​/api​/v2​/graph-search',
    'get ​/api​/v2​/graphs​/shortest-path',
    'get ​/api​/v2​/graphs​/edge-composition',
    'get ​/api​/v2​/graphs​/relay-targets',
    'get ​/api​/v2​/graphs​/acl-inheritance',
    'get ​/api​/v2​/saved-queries',
    'post ​/api​/v2​/saved-queries',
    'get ​/api​/v2​/saved-queries​/{saved_query_id}',
    'delete ​/api​/v2​/saved-queries​/{saved_query_id}',
    'put ​/api​/v2​/saved-queries​/{saved_query_id}',
    'get ​/api​/v2​/saved-queries​/{saved_query_id}​/permissions',
    'delete ​/api​/v2​/saved-queries​/{saved_query_id}​/permissions',
    'put ​/api​/v2​/saved-queries​/{saved_query_id}​/permissions',
    'get ​/api​/v2​/saved-queries​/{saved_query_id}​/export',
    'post ​/api​/v2​/saved-queries​/import',
    'get ​/api​/v2​/saved-queries​/export',
    'post ​/api​/v2​/graphs​/cypher',
    'get ​/api​/v2​/azure​/{entity_type}',
    'get ​/api​/v2​/base​/{object_id}',
    'get ​/api​/v2​/base​/{object_id}​/controllables',
    'get ​/api​/v2​/base​/{object_id}​/controllers',
    'get ​/api​/v2​/computers​/{object_id}',
    'get ​/api​/v2​/computers​/{object_id}​/admin-rights',
    'get ​/api​/v2​/computers​/{object_id}​/admin-users',
    'get ​/api​/v2​/computers​/{object_id}​/constrained-delegation-rights',
    'get ​/api​/v2​/computers​/{object_id}​/constrained-users',
    'get ​/api​/v2​/computers​/{object_id}​/controllables',
    'get ​/api​/v2​/computers​/{object_id}​/controllers',
    'get ​/api​/v2​/computers​/{object_id}​/dcom-rights',
    'get ​/api​/v2​/computers​/{object_id}​/dcom-users',
    'get ​/api​/v2​/computers​/{object_id}​/group-membership',
    'get ​/api​/v2​/computers​/{object_id}​/ps-remote-rights',
    'get ​/api​/v2​/computers​/{object_id}​/ps-remote-users',
    'get ​/api​/v2​/computers​/{object_id}​/rdp-rights',
    'get ​/api​/v2​/computers​/{object_id}​/rdp-users',
    'get ​/api​/v2​/computers​/{object_id}​/sessions',
    'get ​/api​/v2​/computers​/{object_id}​/sql-admins',
    'get ​/api​/v2​/containers​/{object_id}',
    'get ​/api​/v2​/containers​/{object_id}​/controllers',
    'get ​/api​/v2​/domains​/{object_id}',
    'patch ​/api​/v2​/domains​/{object_id}',
    'get ​/api​/v2​/domains​/{object_id}​/computers',
    'get ​/api​/v2​/domains​/{object_id}​/controllers',
    'get ​/api​/v2​/domains​/{object_id}​/dc-syncers',
    'get ​/api​/v2​/domains​/{object_id}​/foreign-admins',
    'get ​/api​/v2​/domains​/{object_id}​/foreign-gpo-controllers',
    'get ​/api​/v2​/domains​/{object_id}​/foreign-groups',
    'get ​/api​/v2​/domains​/{object_id}​/foreign-users',
    'get ​/api​/v2​/domains​/{object_id}​/gpos',
    'get ​/api​/v2​/domains​/{object_id}​/groups',
    'get ​/api​/v2​/domains​/{object_id}​/inbound-trusts',
    'get ​/api​/v2​/domains​/{object_id}​/linked-gpos',
    'get ​/api​/v2​/domains​/{object_id}​/ous',
    'get ​/api​/v2​/domains​/{object_id}​/outbound-trusts',
    'get ​/api​/v2​/domains​/{object_id}​/users',
    'get ​/api​/v2​/domains​/{object_id}​/adcs-escalations',
    'get ​/api​/v2​/gpos​/{object_id}',
    'get ​/api​/v2​/gpos​/{object_id}​/computers',
    'get ​/api​/v2​/gpos​/{object_id}​/controllers',
    'get ​/api​/v2​/gpos​/{object_id}​/ous',
    'get ​/api​/v2​/gpos​/{object_id}​/tier-zero',
    'get ​/api​/v2​/gpos​/{object_id}​/users',
    'get ​/api​/v2​/aiacas​/{object_id}',
    'get ​/api​/v2​/aiacas​/{object_id}​/controllers',
    'get ​/api​/v2​/aiacas​/{object_id}​/pki-hierarchy',
    'get ​/api​/v2​/rootcas​/{object_id}',
    'get ​/api​/v2​/rootcas​/{object_id}​/controllers',
    'get ​/api​/v2​/rootcas​/{object_id}​/pki-hierarchy',
    'get ​/api​/v2​/enterprisecas​/{object_id}',
    'get ​/api​/v2​/enterprisecas​/{object_id}​/controllers',
    'get ​/api​/v2​/enterprisecas​/{object_id}​/pki-hierarchy',
    'get ​/api​/v2​/enterprisecas​/{object_id}​/published-templates',
    'get ​/api​/v2​/ntauthstores​/{object_id}',
    'get ​/api​/v2​/ntauthstores​/{object_id}​/controllers',
    'get ​/api​/v2​/ntauthstores​/{object_id}​/trusted-cas',
    'get ​/api​/v2​/certtemplates​/{object_id}',
    'get ​/api​/v2​/certtemplates​/{object_id}​/controllers',
    'get ​/api​/v2​/certtemplates​/{object_id}​/published-to-cas',
    'get ​/api​/v2​/ous​/{object_id}',
    'get ​/api​/v2​/ous​/{object_id}​/computers',
    'get ​/api​/v2​/ous​/{object_id}​/gpos',
    'get ​/api​/v2​/ous​/{object_id}​/groups',
    'get ​/api​/v2​/ous​/{object_id}​/users',
    'get ​/api​/v2​/users​/{object_id}',
    'get ​/api​/v2​/users​/{object_id}​/admin-rights',
    'get ​/api​/v2​/users​/{object_id}​/constrained-delegation-rights',
    'get ​/api​/v2​/users​/{object_id}​/controllables',
    'get ​/api​/v2​/users​/{object_id}​/controllers',
    'get ​/api​/v2​/users​/{object_id}​/dcom-rights',
    'get ​/api​/v2​/users​/{object_id}​/memberships',
    'get ​/api​/v2​/users​/{object_id}​/ps-remote-rights',
    'get ​/api​/v2​/users​/{object_id}​/rdp-rights',
    'get ​/api​/v2​/users​/{object_id}​/sessions',
    'get ​/api​/v2​/users​/{object_id}​/sql-admin-rights',
    'get ​/api​/v2​/groups​/{object_id}',
    'get ​/api​/v2​/groups​/{object_id}​/admin-rights',
    'get ​/api​/v2​/groups​/{object_id}​/controllables',
    'get ​/api​/v2​/groups​/{object_id}​/controllers',
    'get ​/api​/v2​/groups​/{object_id}​/dcom-rights',
    'get ​/api​/v2​/groups​/{object_id}​/members',
    'get ​/api​/v2​/groups​/{object_id}​/memberships',
    'get ​/api​/v2​/groups​/{object_id}​/ps-remote-rights',
    'get ​/api​/v2​/groups​/{object_id}​/rdp-rights',
    'get ​/api​/v2​/groups​/{object_id}​/sessions',
    'get ​/api​/v2​/completeness',
    'get ​/api​/v2​/ad-domains​/{domain_id}​/data-quality-stats',
    'get ​/api​/v2​/azure-tenants​/{tenant_id}​/data-quality-stats',
    'get ​/api​/v2​/platform​/{platform_id}​/data-quality-stats',
    'post ​/api​/v2​/clear-database',
    'get ​/api​/v2​/datapipe​/status',
    'put ​/api​/v2​/analysis',
    'put ​/api​/v2​/accept-eula',
    'get ​/api​/v2​/meta-nodes​/{domain_id}',
    'get ​/api​/v2​/meta-trees​/{domain_id}',
    'get ​/api​/v2​/asset-groups​/{asset_group_id}​/combo-node',
    'post ​/api​/v2​/ingest',
    'get ​/api​/v2​/clients',
    'post ​/api​/v2​/clients',
    'post ​/api​/v2​/clients​/error',
    'put ​/api​/v2​/clients​/update',
    'get ​/api​/v2​/clients​/{client_id}',
    'put ​/api​/v2​/clients​/{client_id}',
    'delete ​/api​/v2​/clients​/{client_id}',
    'put ​/api​/v2​/clients​/{client_id}​/token',
    'get ​/api​/v2​/clients​/{client_id}​/completed-tasks',
    'get ​/api​/v2​/clients​/{client_id}​/completed-jobs',
    'post ​/api​/v2​/clients​/{client_id}​/tasks',
    'post ​/api​/v2​/clients​/{client_id}​/jobs',
    'get ​/api​/v2​/jobs​/available',
    'get ​/api​/v2​/jobs​/finished',
    'get ​/api​/v2​/jobs',
    'get ​/api​/v2​/jobs​/current',
    'post ​/api​/v2​/jobs​/start',
    'post ​/api​/v2​/jobs​/end',
    'get ​/api​/v2​/jobs​/{job_id}',
    'put ​/api​/v2​/jobs​/{job_id}​/cancel',
    'get ​/api​/v2​/jobs​/{job_id}​/log',
    'get ​/api​/v2​/events',
    'post ​/api​/v2​/events',
    'get ​/api​/v2​/events​/{event_id}',
    'put ​/api​/v2​/events​/{event_id}',
    'delete ​/api​/v2​/events​/{event_id}',
    'get ​/api​/v2​/domains​/{domain_id}​/attack-path-findings',
    'get ​/api​/v2​/attack-paths​/finding-trends',
    'get ​/api​/v2​/attack-path-types',
    'put ​/api​/v2​/attack-paths',
    'get ​/api​/v2​/attack-paths​/details',
    'get ​/api​/v2​/domains​/{domain_id}​/available-types',
    'get ​/api​/v2​/domains​/{domain_id}​/details',
    'get ​/api​/v2​/domains​/{domain_id}​/sparkline',
    'put ​/api​/v2​/attack-paths​/{attack_path_id}​/acceptance',
    'get ​/api​/v2​/posture-stats',
    'get ​/api​/v2​/posture-history​/{data_type}',
    'get ​/api​/v2​/meta​/{object_id}',
];
