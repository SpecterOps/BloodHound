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

import { CommonSearches } from './commonSearchesAGI';
import { RACF_NODE_KIND_VALUES, RACF_RELATIONSHIP_KINDS } from './commonSearchesRACF';
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
        CommonSearches.forEach((commonSearchType: CommonSearchType) => {
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
                                const isRACFNode = RACF_NODE_KIND_VALUES.includes(kind);
                                const isRACFEdge = RACF_RELATIONSHIP_KINDS.includes(kind);
                                const inSchema =
                                    isADNode || isADEdge || isAZNode || isAZEdge || isRACFNode || isRACFEdge;

                                expect(inSchema).toBeTruthy();
                            });
                    });
                }
            });
        });
    });

    test('RACF queries do not use unrestricted relationship traversal', () => {
        const racfQueries = CommonSearches.filter(({ category }) => category === 'RACF').flatMap(
            ({ queries }) => queries
        );

        racfQueries.forEach(({ query }) => {
            expect(query).not.toMatch(/\[\s*\*/);
        });
    });

    test('RACFHasSubgroup is only used to describe Group-SPECIAL administrative scope', () => {
        const subgroupQueries = CommonSearches.filter(({ category }) => category === 'RACF')
            .flatMap(({ queries }) => queries)
            .filter(({ query }) => query.includes('RACFHasSubgroup'));

        expect(subgroupQueries).toHaveLength(1);
        expect(subgroupQueries[0].name).toBe('RACF Group-SPECIAL administrative scope');
    });

    test('effective SPECIAL paths use only explicit identity, membership, and privilege transitions', () => {
        const specialQuery = CommonSearches.flatMap(({ queries }) => queries).find(
            ({ name }) => name === 'RACF effective paths to SPECIAL'
        )?.query;

        expect(specialQuery).toContain('RACFMemberOf|RACFSurrogateFor|RACFPassticketFor*0..6');
        expect(specialQuery).toContain('RACFHasPrivilege');
        expect(specialQuery).not.toContain('RACFCanRead');
        expect(specialQuery).not.toContain('RACFCanWrite');
        expect(specialQuery).not.toContain('RACFOwns');
        expect(specialQuery).not.toContain('RACFHasSubgroup');
    });

    test('legacy credential query checks password and passphrase algorithms', () => {
        const legacyCredentialQuery = CommonSearches.flatMap(({ queries }) => queries).find(
            ({ name }) => name === 'RACF users with legacy password algorithms'
        )?.query;

        expect(legacyCredentialQuery).toContain('MATCH (u:RACFUser)');
        expect(legacyCredentialQuery).toContain("u.pwd_alg <> 'NOPASSWORD' AND u.pwd_alg='LEGACY'");
        expect(legacyCredentialQuery).toContain("u.phr_alg <> 'NOPHRASE' AND u.phr_alg='LEGACY'");
    });

    test('RACF node kinds are declared in MATCH patterns rather than WHERE clauses', () => {
        const racfQueries = CommonSearches.filter(({ category }) => category === 'RACF').flatMap(
            ({ queries }) => queries
        );

        racfQueries.forEach(({ query }) => {
            expect(query).not.toMatch(/(?:WHERE|AND)\s+\w+:RACF\w+/);
        });
    });

    test('RACF built-in queries limit graph results to 500', () => {
        const racfQueries = CommonSearches.filter(({ category }) => category === 'RACF').flatMap(
            ({ queries }) => queries
        );

        racfQueries.forEach(({ query }) => {
            expect(query).toContain('LIMIT 500');
            expect(query).not.toContain('LIMIT 1000');
        });
    });

    test('RACF queries use the unprefixed RACF graph contract', () => {
        const racfQueries = CommonSearches.filter(({ category }) => category === 'RACF').flatMap(
            ({ queries }) => queries
        );

        racfQueries.forEach(({ query }) => expect(query).not.toContain('racf_'));
    });

    test('RACF built-in query names are unique and documented', () => {
        const racfQueries = CommonSearches.filter(({ category }) => category === 'RACF').flatMap(
            ({ queries }) => queries
        );
        const queryNames = racfQueries.map(({ name }) => name);

        expect(new Set(queryNames).size).toBe(queryNames.length);
        racfQueries.forEach(({ description }) => expect(description.length).toBeGreaterThan(0));
    });
});
