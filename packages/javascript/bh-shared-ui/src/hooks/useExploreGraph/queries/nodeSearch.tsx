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

import { FlatGraphResponse } from 'js-client-library';
import { apiClient } from '../../../utils';
import { ExploreQueryParams } from '../../useExploreParams';
import { ExploreGraphQueryKey, ExploreGraphQueryOptions } from './utils';

// Temporary example code. To be fully implemented in BED-5445
export const nodeSearchGraphQuery = (paramOptions: Partial<ExploreQueryParams>): ExploreGraphQueryOptions => {
    const { searchType, primarySearch } = paramOptions;
    if (!primarySearch || !searchType) {
        return {
            enabled: false,
        };
    }

    return {
        queryKey: [ExploreGraphQueryKey, searchType, primarySearch],
        queryFn: ({ signal }) =>
            apiClient
                .getSearchResult(primarySearch, 'exact', { signal })
                .then((res) => res.data.data as FlatGraphResponse),
        enabled: !!(searchType && primarySearch),
    };
};
