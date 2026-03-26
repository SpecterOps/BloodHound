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

import { renderHook } from '@testing-library/react';
import { ExploreQueryParams } from '../useExploreParams';
import { exploreGraphQueryFactory, useUserSettings } from './useExploreGraph';

const mockUseTimeoutLimitConfiguration = vi.fn();

vi.mock('../useConfiguration', () => ({
    useTimeoutLimitConfiguration: () => mockUseTimeoutLimitConfiguration(),
}));

describe('useExploreGraph', () => {
    describe('exploreGraphQueryFactory', () => {
        it('returns {enabled: false} if there is not a match on the switch statement', () => {
            const paramOptions = {
                searchType: 'noMatch',
            } as any;

            const userSettings = {};

            const queryContext = exploreGraphQueryFactory(paramOptions, userSettings);

            const config = queryContext.getQueryConfig();
            expect(config).toStrictEqual({ enabled: false });
        });

        it('runs a node search when the query param is set to "node"', () => {
            const paramOptions: Partial<ExploreQueryParams> = { searchType: 'node', primarySearch: 'test1' };
            const userSettings = {};

            const context = exploreGraphQueryFactory(paramOptions, userSettings);

            const query = context.getQueryConfig();
            expect(query?.queryKey).toContain('node');
        });

        it('runs a pathfinding search when the query param is set to "pathfinding"', () => {
            const paramOptions: Partial<ExploreQueryParams> = {
                searchType: 'pathfinding',
                primarySearch: 'test1',
                secondarySearch: 'test2',
            };

            const userSettings = {};

            const context = exploreGraphQueryFactory(paramOptions, userSettings);

            const query = context.getQueryConfig();
            expect(query?.queryKey).toContain('pathfinding');
        });

        describe('relationship search queries', () => {
            it('returns query config when searchType is relationship and all required params are passed', () => {
                const paramOptions: Partial<ExploreQueryParams> = {
                    searchType: 'relationship',
                    relationshipQueryItemId: 'testId',
                    relationshipQueryType: 'user-member_of',
                };

                const userSettings = {};

                const context = exploreGraphQueryFactory(paramOptions, userSettings);
                const query = context.getQueryConfig();

                expect(query.enabled).toBeUndefined();
                expect(query.queryKey).toContain('relationship');
            });

            it.each([
                {
                    relationshipQueryItemId: 'testId',
                    relationshipQueryType: 'user-member_of',
                },
                {
                    searchType: 'relationship',
                    relationshipQueryType: 'user-member_of',
                },
                {
                    searchType: 'relationship',
                    relationshipQueryItemId: 'testId',
                },
            ])(
                'returns disabled config when any required param is falsey',
                ({ searchType, relationshipQueryItemId, relationshipQueryType }) => {
                    {
                        const paramOptions: Partial<ExploreQueryParams> = {
                            searchType,
                            relationshipQueryItemId,
                            relationshipQueryType,
                        } as any;

                        const userSettings = {};

                        const context = exploreGraphQueryFactory(paramOptions, userSettings);
                        const query = context.getQueryConfig();

                        expect(query.enabled).toBeFalsy();
                    }
                }
            );
        });

        describe('composition search queries', () => {
            it('returns query config when searchType is composition and all required params are passed', () => {
                const paramOptions: Partial<ExploreQueryParams> = {
                    searchType: 'composition',
                    relationshipQueryItemId: 'rel_1234_member_5678',
                };

                const userSettings = {};

                const context = exploreGraphQueryFactory(paramOptions, userSettings);

                const query = context.getQueryConfig();

                expect(query.enabled).toBeUndefined();
                expect(query.queryKey).toContain('composition');
            });

            it.each([{ relationshipQueryItemId: 'testId' }, { searchType: 'relationship' }])(
                'returns disabled config when any required param is falsey',
                ({ searchType, relationshipQueryItemId }) => {
                    {
                        const paramOptions: Partial<ExploreQueryParams> = {
                            searchType,
                            relationshipQueryItemId,
                        } as any;

                        const userSettings = {};

                        const context = exploreGraphQueryFactory(paramOptions, userSettings);

                        const query = context.getQueryConfig();
                        expect(query.enabled).toBeFalsy();
                    }
                }
            );

            it('returns disabled if relationshipQueryItemId does not have a matching sourceId, edgeType, targetId', () => {
                const paramOptions: Partial<ExploreQueryParams> = {
                    searchType: 'composition',
                    relationshipQueryItemId: 'rel_broken-member_5678',
                };

                const userSettings = {};

                const context = exploreGraphQueryFactory(paramOptions, userSettings);

                const query = context.getQueryConfig();
                expect(query.enabled).toBeFalsy();
            });
        });
        it('runs a cypher search when the query param is set to "cypher"', () => {
            const paramOptions: Partial<ExploreQueryParams> = {
                searchType: 'cypher',
                cypherSearch: 'test1',
            };

            const userSettings = {};

            const context = exploreGraphQueryFactory(paramOptions, userSettings);

            const query = context.getQueryConfig();
            expect(query?.queryKey).toContain('cypher');
        });
        it('maps cypher 504 errors to timeout messaging', () => {
            const params: Partial<ExploreQueryParams> = {
                searchType: 'cypher',
                cypherSearch: 'dGVzdA==',
            };

            const userSettings = {};

            const query = exploreGraphQueryFactory(params, userSettings);
            const result = query.getErrorMessage({ response: { status: 504 } });
            expect(result).toStrictEqual({
                message: 'The results took too long to compute, possibly due to the complexity of the query.',
                key: 'CypherSearchQueryTimeout',
            });
        });

        describe('userSettings', () => {
            const setLocalStorageTimeoutSetting = (timeoutSetting: boolean) => {
                localStorage.setItem('persistedState', JSON.stringify({ global: { view: { timeoutSetting } } }));
            };

            const renderUserSettings = () => renderHook(() => useUserSettings()).result.current;

            beforeEach(() => {
                localStorage.clear();
            });

            it('returns a prefer wait in the header when db config timeout limit setting is disabled and state of is disable query limit is true', () => {
                // Sets the DB value that determines if the checkbox is shown in the UI ( false shows the checkbox )
                mockUseTimeoutLimitConfiguration.mockReturnValue(false);
                // Sets the value of the checkbox to disable query timeout
                setLocalStorageTimeoutSetting(true);

                const { headers } = renderUserSettings();

                expect(headers).toEqual({ Prefer: 'wait=-1' });
            });

            it('returns undefined for headers when db config timeout limit setting is disabled and state of is disable query limit is false', () => {
                // Sets the DB value that determines if the checkbox is shown in the UI ( false shows the checkbox )
                mockUseTimeoutLimitConfiguration.mockReturnValue(false);
                // Sets the value of the checkbox to disable query timeout
                setLocalStorageTimeoutSetting(false);

                const { headers } = renderUserSettings();

                expect(headers).toEqual(undefined);
            });
            // This test is to cover the possibility that the configuration is set to hide the checkbox but the user had it set to true previously when it was showing
            it('returns undefined for headers when db config timeout limit setting is enabled and state of is disable query limit is true', () => {
                // Sets the DB value that determines if the checkbox is shown in the UI ( true hides the checkbox )
                mockUseTimeoutLimitConfiguration.mockReturnValue(true);
                // Sets the value of the checkbox to disable query timeout
                setLocalStorageTimeoutSetting(true);

                const { headers } = renderUserSettings();

                expect(headers).toEqual(undefined);
            });
        });
    });
});
