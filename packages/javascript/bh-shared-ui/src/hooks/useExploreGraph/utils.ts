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

import { AllEdgeTypes, Category, EdgeCheckboxType, Subcategory } from '../../edgeTypes';

export const extractEdgeTypes = (edges: EdgeCheckboxType[]): string[] => {
    return edges.filter((edge) => edge.checked).map((edge) => edge.edgeType);
};

export const mapParamsToFilters = (params: string[], initial: EdgeCheckboxType[]): EdgeCheckboxType[] => {
    return initial.map((edge) => ({
        ...edge,
        checked: !!params.includes(edge.edgeType),
    }));
};

export const compareEdgeTypes = (initial: string[], comparison: string[]): boolean => {
    const a = initial.slice(0).sort();
    const b = comparison.slice(0).sort();

    return a.length === b.length && a.every((item, index) => item === b[index]);
};

// Create a list of all edge types to initialize pathfinding filter state
export const getInitialPathFilters = (): EdgeCheckboxType[] => {
    const initialPathFilters: EdgeCheckboxType[] = [];

    AllEdgeTypes.forEach((category: Category) => {
        category.subcategories.forEach((subcategory: Subcategory) => {
            subcategory.edgeTypes.forEach((edgeType: string) => {
                initialPathFilters.push({
                    category: category.categoryName,
                    subcategory: subcategory.name,
                    edgeType,
                    checked: true,
                });
            });
        });
    });

    return initialPathFilters;
};
