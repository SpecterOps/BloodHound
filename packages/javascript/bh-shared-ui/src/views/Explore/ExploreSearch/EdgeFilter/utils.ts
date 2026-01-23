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
import { BUILTIN_EDGE_CATEGORIES, Category, Subcategory } from './edgeCategories';

// these are the schema types that should be ignored from the API in favor of our built-in categories
export const BUILTIN_TYPES = ['ad', 'az'];

// map from our API EdgeType format to a single Category type that can be consumed by our edge filter dialog
export const mapEdgeTypesToCategory = (edgeTypes: EdgeType[], categoryName: string): Category => {
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

// removes all built-in and non-traversable edges from a list of edge types
export const filterUneededTypes = (data: EdgeType[] | undefined): EdgeType[] | undefined => {
    return data?.filter((edge) => !BUILTIN_TYPES.includes(edge.schema_name) && edge.is_traversable);
};

// maps one category from an array of Category types back to a flat list of edge names. useful for verification
export const getEdgeListFromCategory = (categoryName: string, categories: Category[] = BUILTIN_EDGE_CATEGORIES) => {
    const category = categories.find((category) => category.categoryName === categoryName);
    return category?.subcategories.flatMap((subcategory) => subcategory.edgeTypes);
};
