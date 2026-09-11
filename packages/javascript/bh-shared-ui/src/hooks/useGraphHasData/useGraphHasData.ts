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

import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { apiClient, GenericQueryOptions } from '../../utils';

export const getGraphHasData = async (options: RequestOptions): Promise<boolean> => {
    try {
        const res = await apiClient.cypherSearch('MATCH (A) WHERE NOT A:MigrationData RETURN A LIMIT 1', options);
        return Object.keys(res?.data?.data?.nodes).length > 0;
    } catch (err: any) {
        if (err?.response?.status === 404) {
            return false; // API returns 404 when response contains 0 entities – treat as no data available
        }
        throw err;
    }
};

export const useGraphHasData = (queryOptions?: GenericQueryOptions<boolean>): UseQueryResult<boolean> => {
    return useQuery({
        queryKey: ['getGraphHasData'],
        queryFn: ({ signal }) => getGraphHasData({ signal }),
        ...queryOptions,
    });
};
