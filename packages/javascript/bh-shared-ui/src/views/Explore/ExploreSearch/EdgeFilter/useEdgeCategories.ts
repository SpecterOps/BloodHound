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
import { EdgeType } from 'js-client-library';
import { useMemo } from 'react';
import { useQuery } from 'react-query';
import { useFeatureFlag } from '../../../../hooks/useFeatureFlags';
import { apiClient } from '../../../../utils';
import { BUILTIN_EDGE_CATEGORIES, Category, Subcategory } from './edgeCategories';

const BUILTIN_TYPES = ['ad', 'az'];

export const useEdgeCategories = () => {
    const { data: openGraphFeatureFlag } = useFeatureFlag('opengraph_search');

    const edgeTypesQuery = useQuery({
        queryKey: ['getEdgeTypes'],
        queryFn: ({ signal }) => apiClient.getEdgeTypes({ signal }).then((res) => res.data.data),
        enabled: !!openGraphFeatureFlag?.enabled,
    });

    const edgeCategories = useMemo(() => {
        const customEdgeTypes = filterUneededTypes(edgeTypesQuery.data);

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

const filterUneededTypes = (data: EdgeType[] | undefined): EdgeType[] | undefined => {
    return data?.filter((edge) => !BUILTIN_TYPES.includes(edge.schema_name) && edge.is_traversable);
};

const mapEdgeTypesToCategory = (edgeTypes: EdgeType[], categoryName: string): Category => {
    const subcategories = new Map<string, Subcategory>();

    for (const edge of edgeTypes) {
        const existing = subcategories.get(edge.schema_name);

        if (existing) {
            existing.edgeTypes.push(edge.name);
        } else {
            subcategories.set(edge.schema_name, {
                name: edge.schema_name,
                edgeTypes: [edge.name],
            });
        }
    }

    return {
        categoryName,
        subcategories: Array.from(subcategories.values()),
    };
};
