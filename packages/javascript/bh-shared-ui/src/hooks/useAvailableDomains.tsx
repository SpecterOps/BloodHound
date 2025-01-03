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

import { Domain } from 'js-client-library';
import { apiClient } from '../utils/api';
import { UseQueryOptions, useQuery } from 'react-query';

export const availableDomainKeys = {
    all: ['available-domains'],
} as const;

type QueryOptions = Omit<
    UseQueryOptions<unknown, unknown, Domain[], readonly ['available-domains']>,
    'queryKey' | 'queryFn'
>;
const useAvailableDomains = (options?: QueryOptions) =>
    useQuery(
        availableDomainKeys.all,
        ({ signal }) => apiClient.getAvailableDomains({ signal }).then((response) => response.data.data),
        options
    );

export default useAvailableDomains;
