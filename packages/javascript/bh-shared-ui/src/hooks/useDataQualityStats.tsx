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

import { useQuery } from 'react-query';
import { apiClient } from '../utils/api';

export const useActiveDirectoryDataQualityHistoryQuery = (id: string) => {
    return useQuery(['active-directory-data-quality-history', id], ({ signal }) => {
        return apiClient
            .getADQualityStats(id, undefined, undefined, undefined, 'created_at', { signal })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve AD quality history');
                return response.data;
            });
    });
};

export const useActiveDirectoryDataQualityStatsQuery = (id: string) => {
    return useQuery(['active-directory-data-quality-stats', id], ({ signal }) => {
        return apiClient.getADQualityStats(id, undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve AD quality stats');
            return response.data;
        });
    });
};

export const useAzureDataQualityHistoryQuery = (id: string) => {
    return useQuery(['azure-data-quality-history', id], ({ signal }) => {
        return apiClient
            .getAzureQualityStats(id, undefined, undefined, undefined, 'created_at', { signal })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve Azure quality history');
                return response.data;
            });
    });
};

export const useAzureDataQualityStatsQuery = (id: string) => {
    return useQuery(['azure-data-quality-stats', id], ({ signal }) => {
        return apiClient.getAzureQualityStats(id, undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve Azure quality stats');
            return response.data;
        });
    });
};

export const useActiveDirectoryPlatformsDataQualityHistoryQuery = () => {
    return useQuery('active-directory-platform-data-quality-history', ({ signal }) =>
        apiClient
            .getPlatformQualityStats('ad', undefined, undefined, undefined, 'created_at', { signal })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve AD platform quality history');
                return response.data;
            })
    );
};

export const useActiveDirectoryPlatformsDataQualityStatsQuery = () => {
    return useQuery('active-directory-platform-data-quality-stats', ({ signal }) =>
        apiClient.getPlatformQualityStats('ad', undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve AD platform quality stats');
            return response.data;
        })
    );
};

export const useAzurePlatformsDataQualityHistoryQuery = () => {
    return useQuery('azure-platform-data-quality-history', ({ signal }) =>
        apiClient
            .getPlatformQualityStats('azure', undefined, undefined, undefined, 'created_at', { signal })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve Azure platform quality history');
                return response.data;
            })
    );
};

export const useAzurePlatformsDataQualityStatsQuery = () => {
    return useQuery('azure-platform-data-quality-stats', ({ signal }) =>
        apiClient.getPlatformQualityStats('azure', undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve Azure platform quality stats');
            return response.data;
        })
    );
};

// OpenGraph DQ queries pass schemaEnvironmentKindId so extension-level views stay scoped to one environment kind.
export const useOpenGraphDataQualityStatsQuery = (
    environmentId: string,
    extensionId?: number | null,
    schemaEnvironmentKindId?: number | null
) => {
    return useQuery(
        ['opengraph-data-quality-stats', environmentId, extensionId, schemaEnvironmentKindId],
        ({ signal }) =>
            apiClient
                .getOpenGraphQualityStats(
                    environmentId,
                    extensionId ?? undefined,
                    schemaEnvironmentKindId ?? undefined,
                    undefined,
                    undefined,
                    1000,
                    '-created_at',
                    {
                        signal,
                    }
                )
                .then((response) => {
                    if (!response.data) throw new Error('Unable to retrieve OpenGraph quality stats');
                    return response.data;
                })
    );
};

export const useOpenGraphDataQualityHistoryQuery = (
    environmentId: string,
    extensionId?: number | null,
    schemaEnvironmentKindId?: number | null
) => {
    return useQuery(
        ['opengraph-data-quality-history', environmentId, extensionId, schemaEnvironmentKindId],
        ({ signal }) =>
            apiClient
                .getOpenGraphQualityStats(
                    environmentId,
                    extensionId ?? undefined,
                    schemaEnvironmentKindId ?? undefined,
                    undefined,
                    undefined,
                    undefined,
                    'created_at',
                    { signal }
                )
                .then((response) => {
                    if (!response.data) throw new Error('Unable to retrieve OpenGraph quality history');
                    return response.data;
                })
    );
};

export const useOpenGraphDataQualityAggregationsQuery = (
    extensionId?: number | null,
    schemaEnvironmentKindId?: number | null
) => {
    return useQuery(
        ['opengraph-data-quality-aggregations', extensionId, schemaEnvironmentKindId],
        ({ signal }) =>
            apiClient
                .getOpenGraphQualityAggregations(
                    extensionId ?? undefined,
                    schemaEnvironmentKindId ?? undefined,
                    undefined,
                    undefined,
                    1000,
                    '-created_at',
                    {
                        signal,
                    }
                )
                .then((response) => {
                    if (!response.data) throw new Error('Unable to retrieve OpenGraph quality aggregations');
                    return response.data;
                }),
        { enabled: extensionId !== undefined && extensionId !== null }
    );
};

export const useOpenGraphDataQualityAggregationsHistoryQuery = (
    extensionId?: number | null,
    schemaEnvironmentKindId?: number | null
) => {
    return useQuery(
        ['opengraph-data-quality-aggregations-history', extensionId, schemaEnvironmentKindId],
        ({ signal }) =>
            apiClient
                .getOpenGraphQualityAggregations(
                    extensionId ?? undefined,
                    schemaEnvironmentKindId ?? undefined,
                    undefined,
                    undefined,
                    undefined,
                    'created_at',
                    { signal }
                )
                .then((response) => {
                    if (!response.data) throw new Error('Unable to retrieve OpenGraph quality aggregations history');
                    return response.data;
                }),
        { enabled: extensionId !== undefined && extensionId !== null }
    );
};
