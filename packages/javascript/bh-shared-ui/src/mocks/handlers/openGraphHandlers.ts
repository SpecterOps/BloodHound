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
