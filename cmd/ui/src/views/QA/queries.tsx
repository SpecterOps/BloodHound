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
import { DateTime } from 'luxon';
import apiClient from 'src/api';
import { ActiveDirectoryDataQualityResponse, AzureDataQualityResponse } from './types';

const now = DateTime.now();

export const useActiveDirectoryDataQualityHistoryQuery = (id: string) => {
    return useQuery(['active-directory-data-quality-history', id], ({ signal }) => {
        return apiClient
            .getADQualityStats(id, now.minus({ days: 30 }).toJSDate(), now.toJSDate(), undefined, undefined, { signal })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve AD quality history');
                return response.data as ActiveDirectoryDataQualityResponse;
            });
    });
};

export const useActiveDirectoryDataQualityStatsQuery = (id: string) => {
    return useQuery(['active-directory-data-quality-stats', id], ({ signal }) => {
        return apiClient.getADQualityStats(id, undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve AD quality stats');
            return response.data as ActiveDirectoryDataQualityResponse;
        });
    });
};

export const useAzureDataQualityHistoryQuery = (id: string) => {
    return useQuery(['azure-data-quality-history', id], ({ signal }) => {
        return apiClient
            .getAzureQualityStats(id, now.minus({ days: 30 }).toJSDate(), now.toJSDate(), undefined, undefined, {
                signal,
            })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve Azure quality history');
                return response.data as AzureDataQualityResponse;
            });
    });
};

export const useAzureDataQualityStatsQuery = (id: string) => {
    return useQuery(['azure-data-quality-stats', id], ({ signal }) => {
        return apiClient.getAzureQualityStats(id, undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve Azure quality stats');
            return response.data as AzureDataQualityResponse;
        });
    });
};

export const useActiveDirectoryPlatformsDataQualityHistoryQuery = () => {
    return useQuery('active-directory-platform-data-quality-history', ({ signal }) =>
        apiClient
            .getPlatformQualityStats('ad', now.minus({ days: 30 }).toJSDate(), now.toJSDate(), undefined, undefined, {
                signal,
            })
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve AD platform quality history');
                return response.data as ActiveDirectoryDataQualityResponse;
            })
    );
};

export const useActiveDirectoryPlatformsDataQualityStatsQuery = () => {
    return useQuery('active-directory-platform-data-quality-stats', ({ signal }) =>
        apiClient.getPlatformQualityStats('ad', undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve AD platform quality stats');
            return response.data as ActiveDirectoryDataQualityResponse;
        })
    );
};

export const useAzurePlatformsDataQualityHistoryQuery = () => {
    return useQuery('azure-platform-data-quality-history', ({ signal }) =>
        apiClient
            .getPlatformQualityStats(
                'azure',
                now.minus({ days: 30 }).toJSDate(),
                now.toJSDate(),
                undefined,
                undefined,
                { signal }
            )
            .then((response) => {
                if (!response.data) throw new Error('Unable to retrieve Azure platform quality history');
                return response.data as AzureDataQualityResponse;
            })
    );
};

export const useAzurePlatformsDataQualityStatsQuery = () => {
    return useQuery('azure-platform-data-quality-stats', ({ signal }) =>
        apiClient.getPlatformQualityStats('azure', undefined, undefined, 1, undefined, { signal }).then((response) => {
            if (!response.data) throw new Error('Unable to retrieve Azure platform quality stats');
            return response.data as AzureDataQualityResponse;
        })
    );
};
