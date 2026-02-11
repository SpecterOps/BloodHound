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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { renderHook, waitFor } from '../../../../test-utils';
import { BUILTIN_EDGE_CATEGORIES } from './edgeCategories';
import { useEdgeCategories } from './useEdgeCategories';
import { getEdgeListFromCategory } from './utils';

const server = setupServer(
    rest.get('/api/v2/graph-schema/edges', async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    }),
    rest.get('/api/v2/features', async (req, res, ctx) => res(ctx.json(createOpenGraphFeatureFlag(true))))
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('useEdgeCategories', async () => {
    it('initializes the categories with just the built-in defaults when edge endpoint returns nothing', async () => {
        const hook = renderHook(() => useEdgeCategories());
        expect(hook.result.current.edgeCategories).toEqual(BUILTIN_EDGE_CATEGORIES);
    });

    it('includes all traversable OpenGraph edges in their own category when they are available via the API', async () => {
        const testEdges = [...edgeSchemaA, ...edgeSchemaB];

        server.use(
            rest.get('/api/v2/graph-schema/edges', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: testEdges,
                    })
                );
            })
        );

        const hook = renderHook(() => useEdgeCategories());

        // wait until the OpenGraph category is available in the hook's return data
        await waitFor(() => expect(hook.result.current.edgeCategories.length).toEqual(3));

        const edgeList = getEdgeListFromCategory('OpenGraph', hook.result.current.edgeCategories);

        // check that non-traversable edges are excluded
        expect(edgeList?.length).toEqual(8);
        expect(edgeList).not.toEqual(expect.arrayContaining(['SchemaB_EdgeD', 'SchemaB_EdgeE']));
    });

    it('excludes API edges that are associated with one of our built-in schema types', async () => {
        const testEdges = [...edgeSchemaA, ...edgeSchemaB, ...edgeSchemaBuiltin];

        server.use(
            rest.get('/api/v2/graph-schema/edges', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: testEdges,
                    })
                );
            })
        );

        const hook = renderHook(() => useEdgeCategories());

        await waitFor(() => expect(hook.result.current.edgeCategories.length).toEqual(3));

        const adEdges = getEdgeListFromCategory('Active Directory', hook.result.current.edgeCategories);
        const azEdges = getEdgeListFromCategory('Azure', hook.result.current.edgeCategories);
        const ogEdges = getEdgeListFromCategory('OpenGraph', hook.result.current.edgeCategories);

        const edgeList = [...(adEdges ?? []), ...(azEdges ?? []), ...(ogEdges ?? [])];

        expect(edgeList).not.toEqual(expect.arrayContaining(['AdEdge_ShouldntAppear', 'AzEdge_ShouldntAppear']));
    });

    it('does not include OpenGraph edges when the associated feature flag is disabled', async () => {
        const testEdges = [...edgeSchemaA, ...edgeSchemaB, ...edgeSchemaBuiltin];

        server.use(
            rest.get('/api/v2/graph-schema/edges', (req, res, ctx) => {
                return res(
                    ctx.json({
                        data: testEdges,
                    })
                );
            }),
            rest.get('/api/v2/features', async (req, res, ctx) => res(ctx.json(createOpenGraphFeatureFlag(false))))
        );

        const hook = renderHook(() => useEdgeCategories());

        await waitFor(() => expect(hook.result.current.isLoading).toEqual(false));

        // check that only built-in categories are included
        expect(hook.result.current.edgeCategories).toEqual(BUILTIN_EDGE_CATEGORIES);
    });
});

const createOpenGraphFeatureFlag = (enabled: boolean = false) => {
    return {
        data: [
            {
                key: 'opengraph_extension_management',
                enabled,
            },
        ],
    };
};

// all these edges are traversable and should be present in the OpenGraph edge category
const edgeSchemaA = [
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
];

// this group of edges includes two non-traversable edges that should not appear in the resulting OpenGraph category
const edgeSchemaB = [
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
];

// these edges are part of the built-in schema and should be overridden by the hardcoded edge categories in edgeCategories.ts
const edgeSchemaBuiltin = [
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
];
