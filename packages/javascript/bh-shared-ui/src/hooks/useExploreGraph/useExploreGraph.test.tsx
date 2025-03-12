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

import { ExploreQueryParams } from '../useExploreParams';
import { exploreGraphQueryFactory } from './useExploreGraph';

describe('useExploreGraph', () => {
    describe('exploreGraphQueryFactory', () => {
        it('returns {enabled: false} if there is not a match on the switch statement', () => {
            const paramOptions = {
                searchType: 'noMatch',
            } as any;
            const queryContext = exploreGraphQueryFactory(paramOptions);
            const config = queryContext.getQueryConfig(paramOptions);
            expect(config).toStrictEqual({ enabled: false });
        });

        it('runs a node search when the query param is set to "node"', () => {
            const paramOptions: Partial<ExploreQueryParams> = { searchType: 'node', primarySearch: 'test1' };
            const context = exploreGraphQueryFactory(paramOptions);

            const query = context.getQueryConfig(paramOptions);
            expect(query?.queryKey).toContain('node');
        });

        it('runs a pathfinding search when the query param is set to "pathfinding"', () => {
            const paramOptions: Partial<ExploreQueryParams> = {
                searchType: 'pathfinding',
                primarySearch: 'test1',
                secondarySearch: 'test2',
            };
            const context = exploreGraphQueryFactory(paramOptions);

            const query = context.getQueryConfig(paramOptions);
            expect(query?.queryKey).toContain('pathfinding');
        });
    });
});
