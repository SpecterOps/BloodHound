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
import { useFeatureFlag } from '../../../../hooks';
import { apiClient } from '../../../../utils';
import { AllEdgeTypes, Category, Subcategory } from './edgeTypes';

const AD_SCHEMA_TYPE = 'ad';
const AZ_SCHEMA_TYPE = 'az';
const BUILT_IN_TYPES = [AD_SCHEMA_TYPE, AZ_SCHEMA_TYPE];

export const useEdgeTypes = () => {
    const { data: openGraphFeatureFlag } = useFeatureFlag('opengraph_search');

    const edgeTypesQuery = useQuery({
        queryKey: ['getEdgeTypes'],
        queryFn: ({ signal }) => apiClient.getEdgeTypes({ signal }).then((res) => res.data.data),
        enabled: !!openGraphFeatureFlag?.enabled,
    });

    const combinedEdgeTypes = useMemo(() => {
        const customEdgeTypes = filterUneededTypes(edgeTypesQuery.data);

        if (customEdgeTypes && customEdgeTypes.length > 0) {
            return [...AllEdgeTypes, mapEdgeTypesToCategory(customEdgeTypes, 'OpenGraph')];
        } else {
            return AllEdgeTypes;
        }
    }, [edgeTypesQuery.data]);

    return {
        isLoading: edgeTypesQuery.isLoading,
        isError: edgeTypesQuery.isError,
        edgeTypes: combinedEdgeTypes,
    };
};

const filterUneededTypes = (data: EdgeType[] | undefined): EdgeType[] | undefined => {
    return data?.filter((edge) => !BUILT_IN_TYPES.includes(edge.schema_name) && edge.is_traversable);
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
