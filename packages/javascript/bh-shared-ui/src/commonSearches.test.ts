// // Copyright 2025 Specter Ops, Inc.
// //
// // Licensed under the Apache License, Version 2.0
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.
// //
// // SPDX-License-Identifier: Apache-2.0

import { CommonSearches as CommonSearchesAGI } from './commonSearchesAGI';
import { CommonSearches as CommonSearchesAGT } from './commonSearchesAGT';
import {
    ActiveDirectoryNodeKind,
    ActiveDirectoryRelationshipKind,
    AzureNodeKind,
    AzureRelationshipKind,
} from './graphSchema';
import { CommonSearchType } from './types';

describe('common search list', () => {
    const kindPattern = /:([^ )\n\]*]+)/gm;

    test('the queries in the list only include nodes and edges that are defined in our schema', () => {
        CommonSearchesAGI.forEach((commonSearchType: CommonSearchType) => {
            commonSearchType.queries.forEach((query) => {
                const kinds = query.query.match(kindPattern);

                if (kinds) {
                    kinds.forEach((result) => {
                        result
                            .slice(1)
                            .split('|')
                            .forEach((kind) => {
                                const isADNode = Object.values(ActiveDirectoryNodeKind).includes(
                                    kind as ActiveDirectoryNodeKind
                                );
                                const isADEdge = Object.values(ActiveDirectoryRelationshipKind).includes(
                                    kind as ActiveDirectoryRelationshipKind
                                );
                                const isAZNode = Object.values(AzureNodeKind).includes(kind as AzureNodeKind);
                                const isAZEdge = Object.values(AzureRelationshipKind).includes(
                                    kind as AzureRelationshipKind
                                );
                                const inSchema = isADNode || isADEdge || isAZNode || isAZEdge;

                                expect(inSchema).toBeTruthy();
                            });
                    });
                }
            });
        });
    });

    test.each([
        {
            commonSearches: CommonSearchesAGI,
            decoyExclusion: `none(n IN nodes(p) WHERE COALESCE(n.system_tags, '') CONTAINS 'decoy')`,
            mode: 'AGI',
        },
        {
            commonSearches: CommonSearchesAGT,
            decoyExclusion: 'none(n IN nodes(p) WHERE n:Tag_Decoy)',
            mode: 'AGT',
        },
    ])('Azure shortest path queries exclude Decoy nodes in $mode mode', ({ commonSearches, decoyExclusion }) => {
        const azureShortestPathQueries = commonSearches.find(
            (commonSearchType) =>
                commonSearchType.category === 'Azure' && commonSearchType.subheader === 'Shortest Paths'
        )?.queries;

        expect(azureShortestPathQueries).toHaveLength(4);
        azureShortestPathQueries?.forEach((query) => {
            expect(query.query).toContain(decoyExclusion);
        });
    });
});
