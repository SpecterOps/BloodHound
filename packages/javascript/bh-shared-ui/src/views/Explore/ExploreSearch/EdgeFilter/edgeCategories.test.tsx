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
import { ActiveDirectoryPathfindingEdgesMatchFrontend, AzurePathfindingEdges } from '../../../../graphSchema';
import { BUILTIN_EDGE_CATEGORIES } from './edgeCategories';

describe('Make sure pathfinding filterable edges match schemagen', () => {
    it('matches all AD edges', () => {
        const adEdges = new Set(getEdgeListFromCategory('Active Directory'));
        const adSchemaEdges = new Set(ActiveDirectoryPathfindingEdgesMatchFrontend());

        // @ts-ignore
        expect(adEdges.symmetricDifference(adSchemaEdges).size).toEqual(0);
    });

    it('matches all AZ edges', () => {
        const azEdges = new Set(getEdgeListFromCategory('Azure'));
        const azSchemaEdges = new Set(AzurePathfindingEdges());

        // @ts-ignore
        expect(azEdges.symmetricDifference(azSchemaEdges).size).toEqual(0);
    });
});

function getEdgeListFromCategory(categoryName: string) {
    const category = BUILTIN_EDGE_CATEGORIES.find((category) => category.categoryName === categoryName);
    return category?.subcategories.flatMap((subcategory) => subcategory.edgeTypes);
}
