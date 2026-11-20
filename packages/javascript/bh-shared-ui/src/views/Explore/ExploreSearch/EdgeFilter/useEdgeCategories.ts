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
import { useEdgeTypesQuery } from '../../../../hooks/useEdgeTypes';
import { useIsEnterprise } from '../../../../providers/AppNameProvider';
import { BUILTIN_EDGE_CATEGORIES, COMMON_EDGE_CATEGORY } from './edgeCategories';
import { filterUnneededTypes, mapEdgeTypesToCategory } from './utils';

// this hook combines our hardcoded edge categories with an OpenGraph category pulled from the API
export const useEdgeCategories = () => {
    const edgeTypesQuery = useEdgeTypesQuery();
    const isEnterprise = useIsEnterprise();

    // append the common category (Enterprise only) and traversable opengraph edges (if the query is enabled and they
    // exist) to our built-in categories from edgeCategories.ts
    const edgeCategories = useMemo(() => {
        const builtinCategories = isEnterprise
            ? [...BUILTIN_EDGE_CATEGORIES, COMMON_EDGE_CATEGORY]
            : BUILTIN_EDGE_CATEGORIES;
        const customEdgeTypes = filterUnneededTypes(edgeTypesQuery.data);

        if (customEdgeTypes && customEdgeTypes.length > 0) {
            return [...builtinCategories, mapEdgeTypesToCategory(customEdgeTypes, 'OpenGraph')];
        } else {
            return builtinCategories;
        }
    }, [edgeTypesQuery.data, isEnterprise]);

    return {
        isLoading: edgeTypesQuery.isLoading,
        isError: edgeTypesQuery.isError,
        edgeCategories,
    };
};
