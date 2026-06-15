// Copyright 2026 Specter Ops, Inc.
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
import { useMemo } from 'react';
import { useQuery } from 'react-query';
import { useFeatureFlag } from '../../../../hooks/useFeatureFlags';
import { apiClient } from '../../../../utils';
import { BUILTIN_EDGE_CATEGORIES } from './edgeCategories';
import { filterUnneededTypes, mapEdgeTypesToCategory } from './utils';

// this hook combines our hardcoded edge categories with an OpenGraph category pulled from the API
export const useEdgeCategories = () => {
    const { data: openGraphFeatureFlag } = useFeatureFlag('opengraph_extension_management');

    const edgeTypesQuery = useQuery({
        queryKey: ['getEdgeTypes'],
        queryFn: ({ signal }) => apiClient.getEdgeTypes({ signal }).then((res) => res.data.data),
        enabled: !!openGraphFeatureFlag?.enabled,
    });

    // append traversable opengraph edges (if the query is enabled and they exist) to our built-in categories from edgeCategories.ts
    const edgeCategories = useMemo(() => {
        const customEdgeTypes = filterUnneededTypes(edgeTypesQuery.data);

        if (customEdgeTypes && customEdgeTypes.length > 0) {
            return [...BUILTIN_EDGE_CATEGORIES, mapEdgeTypesToCategory(customEdgeTypes, 'OpenGraph')];
        } else {
            return BUILTIN_EDGE_CATEGORIES;
        }
    }, [edgeTypesQuery.data]);

    return {
        isLoading: edgeTypesQuery.isLoading,
        isError: edgeTypesQuery.isError,
        edgeCategories,
    };
};
