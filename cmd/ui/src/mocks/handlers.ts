// Copyright 2023 Specter Ops, Inc.
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

import { rest } from 'msw';
import { createMockSearchResult } from 'src/mocks/factories';

export const handlers = [
    rest.get('/api/v2/search', (req, res, ctx) => {
        const nodeType = req.url.searchParams.get('type');
        return res(
            ctx.json({
                data: Array(10)
                    .fill(null)
                    .map(() => {
                        if (nodeType) return createMockSearchResult(nodeType);
                        return createMockSearchResult();
                    }),
            })
        );
    }),
    rest.get('/api/v2/queries', (req, res, ctx) => {
        return res(
            ctx.json({
                data: Array(1).fill({
                    user_id: 'abcdefgh',
                    query: 'match (n) return n limit 5',
                    name: 'myCustomQuery1',
                }),
            })
        );
    }),
    rest.post('/api/v2/queries', (req, res, ctx) => {
        return res(
            ctx.json({
                user_id: 'abcdefgh',
                query: 'Match(n) return n;',
                name: 'myCustomQuery 2',
            })
        );
    }),
];
