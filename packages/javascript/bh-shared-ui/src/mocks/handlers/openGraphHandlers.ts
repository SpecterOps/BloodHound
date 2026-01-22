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
import { DefaultBodyType, MockedRequest, rest, RestHandler } from 'msw';

export const openGraphHandlers: RestHandler<MockedRequest<DefaultBodyType>>[] = [
    rest.get('/api/v2/graph-schema/edges', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: [
                    {
                        id: 1,
                        name: 'SchemaA_EdgeA',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaA',
                    },
                    {
                        id: 2,
                        name: 'SchemaA_EdgeB',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaA',
                    },
                    {
                        id: 3,
                        name: 'SchemaA_EdgeC',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaA',
                    },
                    {
                        id: 4,
                        name: 'SchemaA_EdgeD',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaA',
                    },
                    {
                        id: 5,
                        name: 'SchemaA_EdgeE',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaA',
                    },
                    {
                        id: 6,
                        name: 'SchemaB_EdgeA',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaB',
                    },
                    {
                        id: 7,
                        name: 'SchemaB_EdgeB',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaB',
                    },
                    {
                        id: 8,
                        name: 'SchemaB_EdgeC',
                        description: '',
                        is_traversable: true,
                        schema_name: 'SchemaB',
                    },
                    {
                        id: 9,
                        name: 'SchemaB_EdgeD',
                        description: '',
                        is_traversable: false,
                        schema_name: 'SchemaB',
                    },
                    {
                        id: 10,
                        name: 'SchemaB_EdgeE',
                        description: '',
                        is_traversable: false,
                        schema_name: 'SchemaE',
                    },
                    {
                        id: 11,
                        name: 'AdEdge_ShouldntAppear',
                        description: '',
                        is_traversable: true,
                        schema_name: 'ad',
                    },
                    {
                        id: 12,
                        name: 'AzEdge_ShouldntAppear',
                        description: '',
                        is_traversable: true,
                        schema_name: 'az',
                    },
                ],
            })
        );
    }),
];
