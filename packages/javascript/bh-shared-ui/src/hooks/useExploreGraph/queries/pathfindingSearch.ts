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

import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import {
    createPathFilterString,
    ExploreGraphQuery,
    ExploreGraphQueryError,
    ExploreGraphQueryKey,
    ExploreGraphQueryOptions,
    INITIAL_FILTER_TYPES,
    sharedGraphQueryOptions,
} from './utils';

// Only need to create our default filters once
const DEFAULT_FILTERS = createPathFilterString(INITIAL_FILTER_TYPES);

// Deep Sniff Query Variant 1 (EnableDCSync): original DCSync enablement query
const buildEnableDCSyncCypher = (sourceNodeId: string, destinationNodeId: string) => `
MATCH p_changes = (x1:Base)-[:GetChanges]->(d:Domain)
MATCH p_changesall = (x2:Base)-[:GetChangesAll]->(d)
WHERE x1:Group OR x2:Group
MATCH p_tochanges = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x1))
WHERE n.objectid = "${sourceNodeId}"
MATCH p_tochangesall = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x2))
MATCH p_totarget = (d)-[:Contains*0..]->(target)
WHERE target.objectid = "${destinationNodeId}"
RETURN p_changes,p_tochanges,p_changesall,p_tochangesall,p_totarget`;

// Deep Sniff Query Variant 2 (EnableADCSESC3): executes if EnableDCSync returns no results
const buildEnableADCSESC3Cypher = (sourceNodeId: string, destinationNodeId: string) => `
MATCH p1 = (x1:Base)-[:GenericAll|Enroll|AllExtendedRights]->(ct1:CertTemplate)-[:PublishedTo]->(eca1:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d:Domain)
WHERE ct1.requiresmanagerapproval = false
AND (ct1.schemaversion = 1 OR ct1.authorizedsignatures = 0)
AND (
    x1:Group
    OR x1:Computer
    OR (
    x1:User
    AND ct1.subjectaltrequiredns = false
    AND ct1.subjectaltrequiredomaindns = false
    )
)
MATCH p2 = (x2:Base)-[:GenericAll|Enroll|AllExtendedRights]->(ct2:CertTemplate)-[:PublishedTo]->(eca2:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
WHERE ct2.authenticationenabled = true
AND ct2.requiresmanagerapproval = false

MATCH p3 = (ct1)-[:EnrollOnBehalfOf]->(ct2)

MATCH p4 = (x3:Base)-[:Enroll]->(eca1)
WHERE x1:Group OR x3:Group

MATCH p5 = (x4:Base)-[:Enroll]->(eca2)
WHERE x2:Group OR x4:Group

MATCH p6 = (eca1)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)
MATCH p7 = (eca2)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)

MATCH p_tox1 = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x1))
WHERE n.objectid = "${sourceNodeId}"
MATCH p_tox2 = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x2))
MATCH p_tox3 = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x3))
MATCH p_tox4 = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x4))

MATCH p_totarget = (d)-[:Contains*0..]->(target)
WHERE target.objectid = "${destinationNodeId}"

OPTIONAL MATCH p8 = (x1)-[:MemberOf*0..]->()-[:DelegatedEnrollmentAgent]->(ct2)

WITH *
WHERE (
    NOT eca2.hasenrollmentagentrestrictions = True
    OR p8 IS NOT NULL
)

RETURN p1,p2,p3,p4,p5,p6,p7,p8,p_tox1,p_tox2,p_tox3,p_tox4,p_totarget`;

// Helper to run deep sniff variants based on user selection
const runDeepSniffWithPreferences = async (
    primarySearch: string,
    secondarySearch: string,
    signal: AbortSignal | undefined,
    variantsPref: ('EnableDCSync' | 'EnableADCSESC3')[] | null
) => {
    const includeProperties = true;
    const wantDCSync = !variantsPref || variantsPref.includes('EnableDCSync');
    const wantESC3 = !variantsPref || variantsPref.includes('EnableADCSESC3');

    // Execution ordering: always try DCSync first if selected, else ESC3
    if (wantDCSync) {
        try {
            const dcsyncCypher = buildEnableDCSyncCypher(primarySearch, secondarySearch);
            const res = await apiClient.cypherSearch(dcsyncCypher, { signal }, includeProperties);
            const nodes = (res?.data as any)?.data?.nodes ?? {};
            const edges = (res?.data as any)?.data?.edges ?? [];
            if (Object.keys(nodes).length > 0 || (Array.isArray(edges) && edges.length > 0)) {
                return { ...(res.data as any), deepSniff: true, deepSniffVariant: 'EnableDCSync' } as any;
            }
        } catch (err: any) {
            if (err?.response?.status !== 404) throw err; // propagate non-404
        }
    }
    if (wantESC3) {
        const esc3Cypher = buildEnableADCSESC3Cypher(primarySearch, secondarySearch);
        const esc3Res = await apiClient.cypherSearch(esc3Cypher, { signal }, includeProperties);
        return { ...(esc3Res.data as any), deepSniff: true, deepSniffVariant: 'EnableADCSESC3' } as any;
    }
    // If no variants actually selected (shouldn't happen due to UI validation) throw not found equivalent
    const error: any = new Error('No deep sniff variants selected');
    error.response = { status: 404 };
    throw error;
};

export const pathfindingSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { searchType, primarySearch, secondarySearch, pathFilters, pathSearchMode, deepSniffVariants } = paramOptions;

    // Query should occur whether or not pathFilters exist
    if (!primarySearch || !searchType || !secondarySearch) {
        return { enabled: false };
    }

    const filter = pathFilters?.length ? createPathFilterString(pathFilters) : DEFAULT_FILTERS;

    const mode: 'hybrid' | 'path' | 'deepsniff' =
        pathSearchMode === 'path' || pathSearchMode === 'deepsniff' ? pathSearchMode : 'hybrid';

    return {
        ...sharedGraphQueryOptions,
        queryKey: [
            ExploreGraphQueryKey,
            searchType,
            primarySearch,
            secondarySearch,
            filter,
            mode,
            deepSniffVariants?.join(',') ?? 'all',
        ],
        queryFn: ({ signal }) => {
            // Deep sniff only mode: skip shortest path entirely
            if (mode === 'deepsniff') {
                return runDeepSniffWithPreferences(primarySearch, secondarySearch, signal, deepSniffVariants ?? null);
            }
            // Path only mode: do not fall back to deep sniff
            if (mode === 'path') {
                return apiClient
                    .getShortestPathV2(primarySearch, secondarySearch, filter, { signal })
                    .then((res) => res.data);
            }
            // Hybrid: attempt path then deep sniff fallback
            return apiClient
                .getShortestPathV2(primarySearch, secondarySearch, filter, { signal })
                .then((res) => res.data)
                .catch((error) => {
                    const statusCode = error?.response?.status;
                    if (statusCode === 404) {
                        return runDeepSniffWithPreferences(
                            primarySearch,
                            secondarySearch,
                            signal,
                            deepSniffVariants ?? null
                        );
                    }
                    throw error;
                });
        },
        enabled: !!(searchType && primarySearch && secondarySearch),
    };
};

const getPathfindingErrorMessage = (error: any): ExploreGraphQueryError => {
    const statusCode = error?.response?.status;
    if (statusCode === 404) {
        return { message: 'Path not found.', key: 'shortestPathNotFound' };
    } else if (statusCode === 503) {
        return {
            message:
                'Calculating the requested Attack Path exceeded memory limitations due to the complexity of paths involved.',
            key: 'ShortestPathOutOfMemory',
        };
    } else if (statusCode === 504) {
        return {
            message: 'The results took too long to compute, possibly due to the complexity of paths involved.',
            key: 'ShortestPathTimeout',
        };
    } else {
        return { message: 'An unknown error occurred. Please try again.', key: 'ShortestPathUnknown' };
    }
};

export const pathfindingSearchQuery: ExploreGraphQuery = {
    getQueryConfig: pathfindingSearchGraphQuery,
    getErrorMessage: getPathfindingErrorMessage,
};
