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
import { getEdgeListFromCategory } from './utils';

describe('Make sure pathfinding filterable edges match schemagen', () => {
    it('matches all AD edges', () => {
        const adEdges = getEdgeListFromCategory('Active Directory');
        const adSchemaEdges = ActiveDirectoryPathfindingEdgesMatchFrontend();

        const difference = getDifferenceCount(adEdges, adSchemaEdges);
        expect(difference).toEqual(0);
    });

    it('matches all AZ edges', () => {
        const azEdges = getEdgeListFromCategory('Azure');
        const azSchemaEdges = AzurePathfindingEdges();

        const difference = getDifferenceCount(azEdges, azSchemaEdges);
        expect(difference).toEqual(0);
    });
});

function getDifferenceCount(a: string[] | undefined | null, b: string[] | undefined | null) {
    const setA = new Set(a);
    const setB = new Set(b);

    const result = new Set([...[...setA].filter((x) => !setB.has(x)), ...[...setB].filter((x) => !setA.has(x))]);
    return result.size;
}
