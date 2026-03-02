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
import { Dispatch, SetStateAction } from 'react';
import { DisableQueryLimitContext } from '../../views/Explore/providers/DisableQueryLimitProvider/DisableQueryLimitContext';
import { ExploreQueryParams } from '../useExploreParams';
import { exploreGraphQueryFactory, useUserSettings } from './useExploreGraph';

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

        it('returns a prefer wait in the header when state of is disable query limit is true ', () => {
            const mockSetState = vi.fn() as Dispatch<SetStateAction<boolean>>;
            const mockValue = { setIsDisableQueryLimit: mockSetState, isDisableQueryLimit: true };

            const wrapper = ({ children }: { children: React.ReactNode }) => (
                <DisableQueryLimitContext.Provider value={mockValue}>{children}</DisableQueryLimitContext.Provider>
            );

            const { result } = renderHook(() => useUserSettings(), { wrapper });

            expect(result.current).toEqual({ headers: { Prefer: 'wait=-1' } });
        });

        it('returns undefined for headers when state of is disable query limit is false ', () => {
            const mockSetState = vi.fn() as Dispatch<SetStateAction<boolean>>;
            const mockValue = { setIsDisableQueryLimit: mockSetState, isDisableQueryLimit: false };

            const wrapper = ({ children }: { children: React.ReactNode }) => (
                <DisableQueryLimitContext.Provider value={mockValue}>{children}</DisableQueryLimitContext.Provider>
            );

            const { result } = renderHook(() => useUserSettings(), { wrapper });
            expect(result.current.headers).toEqual(undefined);
        });
    });
});
