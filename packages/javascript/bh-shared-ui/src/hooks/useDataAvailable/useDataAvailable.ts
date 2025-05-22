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

export const getDataAvailable = async (options: RequestOptions): Promise<boolean> =>
    apiClient.cypherSearch('MATCH (A) WHERE NOT A:MigrationData RETURN A LIMIT 1', options).then((res) => {
        return Object.keys(res?.data?.data?.nodes).length > 0;
    });

export const useDataAvailable = (queryOptions?: GenericQueryOptions<boolean>): UseQueryResult<boolean> => {
    return useQuery({
        queryKey: ['getDataAvailable'],
        queryFn: ({ signal }) => getDataAvailable({ signal }),
        ...queryOptions,
    });
};
