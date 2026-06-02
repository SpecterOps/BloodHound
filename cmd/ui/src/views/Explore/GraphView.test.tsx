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

import userEvent from '@testing-library/user-event';
import { cypherTestResponse, mockKindsHandler, singleNodeResponse } from 'bh-shared-ui';
import { GraphEdge } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen, waitFor, within } from 'src/test-utils';
import GraphView from './GraphView';

const baseGlobalView = {
    notifications: [],
    darkMode: false,
    autoRunQueries: true,
    exploreLayout: undefined,
    isExploreTableSelected: false,
    isExploreLayoutSelected: false,
    selectedExploreTableColumns: undefined,
    pinnedExploreTableColumns: undefined,
    timeoutSetting: false,
};

const searchedNode = {
    label: 'Searched Node',
    kind: 'User',
    kinds: ['User'],
    objectId: 'testing-node-123',
    isTierZero: false,
    isOwnedObject: false,
    lastSeen: '',
};
const graphSearchResponse = {
    data: { data: { nodes: { '42': searchedNode }, edges: [] } },
};

const server = setupServer(
    rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
        return res(ctx.json(cypherTestResponse));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/self', (req, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get(`/api/v2/custom-nodes`, async (_req, res, ctx) => {
        return res(ctx.json({ data: [] }));
    }),
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(ctx.json({ data: { tags: [] } }));
    }),
    rest.get('/api/v2/file-upload/accepted-types', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/saved-queries', async (_, res, ctx) => {
        return res(ctx.status(200));
    }),
    rest.get('/api/v2/graph-search', (_req, res, ctx) => {
        return res(ctx.json(graphSearchResponse));
    }),
    mockKindsHandler(),
    rest.get(`/api/v2/roles`, (req, res, ctx) => {
        return res(
            ctx.json({
                data: {
                    roles: [],
                },
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('GraphView', () => {
    it('renders a graph view', () => {
        render(<GraphView />, { route: `/explore?searchType=cypher&cypherSearch=encodedquery` });
        const container = screen.getByTestId('explore');
        expect(container).toBeInTheDocument();
    });

    it('displays an error message', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.status(500));
            })
        );
        console.error = vi.fn();
        render(<GraphView />);
        const errorAlert = await waitFor(() =>
            screen.findByText('An unexpected error has occurred. Please refresh the page and try again.')
        );

        expect(errorAlert).toBeInTheDocument();
    });

    it('a cypher search 404 does not render the GraphViewErrorAlert', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', async (req, res, ctx) => {
                const body = await req.json();

                if (body.query.includes('MigrationData')) return res(ctx.json({ data: { nodes: { 1: {} } } }));

                return res(ctx.status(404));
            })
        );

        console.error = vi.fn();
        render(<GraphView />, {
            route: '/explore?exploreSearchTab=cypher&cypherSearch=TUFUQ0gobikgd2hlcmUgbi5uYW1lID0gJ2Fyb29vb28nIHJldHVybiBu&searchType=cypher',
        });

        const throwErrorExpectation = async () => {
            await waitFor(
                async () =>
                    await screen.findByText('An unexpected error has occurred. Please refresh the page and try again.'),
                { timeout: 1000 }
            );
        };

        await expect(throwErrorExpectation).rejects.toThrowError();

        expect(
            screen.queryByText('An unexpected error has occurred. Please refresh the page and try again.')
        ).not.toBeInTheDocument();
    });

    it('renders a table if the query has NO node edges', async () => {
        render(<GraphView />, { route: `/explore?searchType=cypher&cypherSearch=encodedquery` });

        const table = await screen.findByRole('table');

        expect(table).toBeInTheDocument();

        const rows = await screen.findAllByRole('row');

        const expectedNumberOfRows = Object.keys(cypherTestResponse.data.nodes).length + 1; // plus one for the header row
        expect(rows.length).toBe(expectedNumberOfRows);
    });

    it('renders a graph if the query has any node edges', async () => {
        const tempEdges: Array<GraphEdge> = [];
        const clonedCypherResponse = {
            ...cypherTestResponse,
            data: {
                ...cypherTestResponse.data,
                edges: tempEdges,
            },
        };

        clonedCypherResponse.data.edges.push({
            source: '108',
            target: '108',
            label: 'some label',
            kind: 'some kind',
            lastSeen: 'some lastSeen',
            impactPercent: 10,
            exploreGraphId: 'some exploreGraphId',
            data: {},
        });

        server.use(
            rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                return res(ctx.json(clonedCypherResponse));
            })
        );

        render(<GraphView />, { route: `/explore?searchType=cypher&cypherSearch=encodedquery` });

        const sigma = await screen.findByTestId('sigma-container-wrapper');
        const table = screen.queryByRole('table');

        expect(sigma).toBeInTheDocument();
        expect(table).not.toBeInTheDocument();
    });

    const autoSelectRoute = `/explore?searchType=node&primarySearch=${searchedNode.objectId}`;

    it('auto-selects when a search result has a matching objectId', async () => {
        render(<GraphView />, { route: autoSelectRoute });

        await waitFor(() => expect(window.location.search).toContain('selectedItem=42'));
    });

    it('opens the entity information panel for the auto-selected node', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (_req, res, ctx) =>
                res(
                    ctx.json({
                        data: {
                            nodes: { '42': searchedNode },
                            edges: [],
                        },
                    })
                )
            )
        );

        render(<GraphView />, { route: autoSelectRoute });

        await waitFor(() => expect(window.location.search).toContain('selectedItem=42'));

        const panel = await screen.findByTestId('explore_entity-information-panel');
        expect(within(panel).getByText('Searched Node')).toBeInTheDocument();
    });

    it('clears selected item when pathfinding search returns results', async () => {
        server.use(
            rest.post('/api/v2/graphs/cypher', (_req, res, ctx) =>
                res(
                    ctx.json({
                        data: {
                            nodes: { '42': searchedNode },
                            edges: [],
                        },
                    })
                )
            )
        );

        const pathfindingRoute = `/explore?searchType=pathfinding&primarySearch=${searchedNode.objectId}&secondarySearch=some-destination`;

        render(<GraphView />, { route: pathfindingRoute });

        await waitFor(() => expect(window.location.search).not.toContain('selectedItem='));
    });

    it('renders correct search elements after keypresses', async () => {
        render(<GraphView />, { route: `/explore` });
        const user = userEvent.setup();

        await user.keyboard('{Alt>}[Tab]{/Alt}');
        await user.keyboard('{Alt>}c{/Alt}');

        const cypherSearchEl = screen.queryByTestId('cypher-search-section');
        const searchNodesEl = screen.queryByPlaceholderText('Search Nodes');
        const pathfindingEl = screen.queryByTestId('pathfinding-search');

        expect(cypherSearchEl).toBeInTheDocument();
        expect(searchNodesEl).not.toBeInTheDocument();
        expect(pathfindingEl).not.toBeInTheDocument();

        await user.keyboard('{Alt>}[Tab]{/Alt}');
        await user.keyboard('{Alt>}p{/Alt}');

        const pathfindingElAfter = screen.queryByTestId('pathfinding-search');

        expect(cypherSearchEl).not.toBeInTheDocument();
        expect(pathfindingElAfter).toBeInTheDocument();

        await user.keyboard('{Alt>}{Shift>}[Tab]{/Shift}{/Alt}');
        await user.keyboard('{Alt>}[Slash]{/Alt}');

        expect(cypherSearchEl).not.toBeInTheDocument();
        expect(pathfindingElAfter).not.toBeInTheDocument();
        const searchNodesElAfter = screen.queryByPlaceholderText('Search Nodes');

        expect(searchNodesElAfter).toBeInTheDocument();
    });

    describe('Layout selection and de-selection', () => {
        const cypherRoute = '/explore?searchType=cypher&cypherSearch=encodedquery';

        it('does not auto-display the table for a node-only cypher response when the user has already selected a graph layout', async () => {
            render(<GraphView />, {
                route: cypherRoute,
                initialState: {
                    global: {
                        view: {
                            ...baseGlobalView,
                            exploreLayout: 'standard',
                            isExploreLayoutSelected: true,
                        },
                    },
                },
            });

            expect(await screen.findByTestId('sigma-container-wrapper')).toBeInTheDocument();
            expect(screen.queryByRole('table')).not.toBeInTheDocument();
        });

        it('does not auto display the table for single node responses with no edges', async () => {
            server.use(
                rest.post('/api/v2/graphs/cypher', (req, res, ctx) => {
                    return res(ctx.json(singleNodeResponse));
                })
            );
            render(<GraphView />, {
                initialState: {
                    global: {
                        view: {
                            ...baseGlobalView,
                            exploreLayout: 'table',
                            isExploreLayoutSelected: true,
                        },
                    },
                },
            });

            expect(screen.queryByRole('table')).not.toBeInTheDocument();
        });

        it('reflects the persisted selected layout in the controls menu', async () => {
            render(<GraphView />, {
                route: cypherRoute,
                initialState: {
                    global: {
                        view: {
                            ...baseGlobalView,
                            exploreLayout: 'standard',
                            isExploreLayoutSelected: true,
                        },
                    },
                },
            });

            const user = userEvent.setup();
            const layoutMenu = await screen.findByText('Layout');
            await user.click(layoutMenu);

            const standardOption = await screen.findByTestId('explore_graph-controls_standard-buttonLabel');
            expect(standardOption).toHaveClass('Mui-selected');

            const sequentialOption = await screen.findByTestId('explore_graph-controls_sequential-buttonLabel');
            expect(sequentialOption).not.toHaveClass('Mui-selected');

            const tableOption = await screen.findByTestId('explore_graph-controls_table-buttonLabel');
            expect(tableOption).not.toHaveClass('Mui-selected');
        });

        it('reverts to the default unselected state when the user de-selects the currently selected graph layout', async () => {
            render(<GraphView />, {
                route: cypherRoute,
                initialState: {
                    global: {
                        view: {
                            ...baseGlobalView,
                            exploreLayout: 'standard',
                            isExploreLayoutSelected: true,
                        },
                    },
                },
            });

            const user = userEvent.setup();
            const layoutMenu = await screen.findByText('Layout');
            await user.click(layoutMenu);

            const standardOption = await screen.findByTestId('explore_graph-controls_standard-buttonLabel');
            await user.click(standardOption);

            await user.click(layoutMenu);

            const standardOptionAfter = await screen.findByTestId('explore_graph-controls_standard-buttonLabel');
            expect(standardOptionAfter).not.toHaveClass('Mui-selected');

            const sequentialOptionAfter = await screen.findByTestId('explore_graph-controls_sequential-buttonLabel');
            expect(sequentialOptionAfter).not.toHaveClass('Mui-selected');

            const tableOptionAfter = await screen.findByTestId('explore_graph-controls_table-buttonLabel');
            expect(tableOptionAfter).not.toHaveClass('Mui-selected');
        });

        it('reverts to the default graph view when the user de-selects the currently selected table layout', async () => {
            render(<GraphView />, {
                route: cypherRoute,
                initialState: {
                    global: {
                        view: {
                            ...baseGlobalView,
                            isExploreTableSelected: true,
                            isExploreLayoutSelected: true,
                        },
                    },
                },
            });

            expect(await screen.findByRole('table')).toBeInTheDocument();

            const user = userEvent.setup();
            const layoutMenu = screen.getByText('Layout');
            await user.click(layoutMenu);

            const closeTableBtn = await screen.findByTestId('close-button');
            await user.click(closeTableBtn);

            await waitFor(() => {
                expect(screen.queryByRole('table')).not.toBeInTheDocument();
            });
        });
    });
});
