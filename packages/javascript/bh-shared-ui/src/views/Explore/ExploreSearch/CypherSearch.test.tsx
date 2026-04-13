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

import userEvent from '@testing-library/user-event';
import { GraphData } from 'js-client-library';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { cypherTestResponse } from '../../../mocks/factories/explore';
import { render, waitFor } from '../../../test-utils';
import { mockCodemirrorLayoutMethods } from '../../../utils';
import CypherSearch from './CypherSearch';

const mockClearSelectedItem = vi.fn(() => {
    const url = new URL(window.location.href);
    url.searchParams.delete('selectedItem');
    url.searchParams.delete('expandedPanelSections');
    window.history.replaceState({}, '', url.toString());
});

vi.mock('../../../hooks', async () => {
    const actual = await vi.importActual('../../../hooks');

    return {
        ...actual,
        useExploreSelectedItem: () => ({
            clearSelectedItem: mockClearSelectedItem,
            selectedItem: '123',
        }),
    };
});

const CYPHER_SEARCH_ROUTE = '?searchType=cypher&cypherSearch=bWF0Y2ggKG4pIHJldHVybiBuIGxpbWl0IDEw';

const multiNodeGraphResponse = cypherTestResponse;

const singleNodeGraphResponse = {
    data: {
        nodes: { '108': cypherTestResponse.data.nodes['108'] },
        edges: [],
    },
};

const zeroNodeGraphResponse = {
    data: {
        nodes: {},
        edges: [{ source: '1', target: '2', label: 'HasSession', kind: 'HasSession', lastSeen: '2023-01-01' }],
    },
};

const CYPHER = 'match (n) return n limit 5';
const INCOMPLETE_CYPHER = 'match (n:';

describe('CypherSearch', () => {
    const testPerformSearch = vi.fn();

    const testState = {
        cypherQuery: '',
        setCypherQuery: vi.fn(),
        performSearch: testPerformSearch,
    };
    const setup = async (state = testState, route = '/', disableQueryLimit = false) => {
        const autoRun = true;
        const handleAutoRun = () => {};
        const testOnRunSearchClick = vi.fn();
        const handleDisableQueryLimit = () => {};

        const screen = render(
            <CypherSearch
                cypherSearchState={state}
                autoRun={autoRun}
                setAutoRun={handleAutoRun}
                disableQueryLimit={disableQueryLimit}
                setDisableQueryLimit={handleDisableQueryLimit}
            />,
            { route }
        );
        const user = userEvent.setup();

        return { state, screen, user, testOnRunSearchClick };
    };

    const server = setupServer(
        rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
            return res(
                ctx.json({
                    data: { kinds: ['Tier Zero', 'Tier One', 'Tier Two'] },
                })
            );
        }),
        rest.get('/api/v2/features', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [{ id: 1, key: 'tier_management_engine', enabled: true }],
                })
            );
        }),
        rest.get('/api/v2/config', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [
                        {
                            key: 'analysis.tiering',
                            name: 'Multi-Tier Analysis Configuration',
                            value: {
                                tier_limit: 3,
                                label_limit: 10,
                                multi_tier_analysis_enabled: true,
                            },
                            id: 8,
                        },
                    ],
                })
            );
        }),
        rest.get('/api/v2/saved-queries', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
        }),
        rest.get('/api/v2/self', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {},
                })
            );
        }),
        rest.get('/api/v2/asset-group-tags', async (req, res, ctx) => {
            return res(
                ctx.json({
                    data: [],
                })
            );
        }),
        rest.get('/api/v2/bloodhound-users-minimal', (req, res, ctx) => {
            return res(
                ctx.json({
                    data: {
                        users: [],
                    },
                })
            );
        })
    );

    beforeAll(() => {
        server.listen();
    });
    beforeEach(mockCodemirrorLayoutMethods);
    afterEach(() => {
        vi.restoreAllMocks();
        server.resetHandlers();
        mockClearSelectedItem.mockClear();
    });
    afterAll(() => {
        server.close();
    });

    it('should render', async () => {
        const { screen } = await setup();
        expect(screen.getByText(/cypher query/i)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /Learn more about cypher/i })).toBeInTheDocument();
    });

    it('should call the setCypherQuery handler when the value in the editor changes', async () => {
        const { screen, user, state } = await setup();
        const searchbox = screen.getAllByRole('textbox');

        await user.type(searchbox[1], CYPHER);

        expect(state.setCypherQuery).toHaveBeenCalled();
    });

    it('should display a dropdown when a user types a partial query that can be autocompleted', async () => {
        const { screen, user } = await setup();
        const searchbox = screen.getAllByRole('textbox');

        await user.type(searchbox[1], INCOMPLETE_CYPHER);

        const autocomplete = await screen.findByRole('listbox');

        expect(autocomplete).toBeVisible();
    });

    it('should call performSearch on keyboard press alt+R', async () => {
        const { user } = await setup({ ...testState, cypherQuery: 'Anything' });

        expect(testPerformSearch).not.toHaveBeenCalled();

        await user.keyboard('{Alt>}r{/Alt}');

        expect(testPerformSearch).toHaveBeenCalled();
    });

    it('should open save dialog on keyboard press alt+S', async () => {
        const { user, screen } = await setup({ ...testState, cypherQuery: 'Anything' });

        expect(screen.queryByTestId('save-query-dialog')).not.toBeInTheDocument();

        await user.keyboard('{Alt>}s{/Alt}');

        expect(screen.queryByTestId('save-query-dialog')).toBeInTheDocument();
    });

    describe('Minimize explorer page elements when multiple nodes are returned', () => {
        // node '108' is T1_TONYMONTANA@PHANTOM.CORP — a real node from the mock cypher response
        const cypherSearchState = { ...testState, cypherQuery: CYPHER };
        const cypherSearchRoute = `${CYPHER_SEARCH_ROUTE}&selectedItem=108`;

        const mockCypherEndpoint = (response: { data: GraphData }) => {
            server.use(rest.post('/api/v2/graphs/cypher', (_req, res, ctx) => res(ctx.json(response))));
        };

        it('closes saved queries panel and entity info panel when a cypher query returns multiple nodes', async () => {
            const { screen, user } = await setup(cypherSearchState, cypherSearchRoute, true);

            const toggleButton = screen.getByTestId('common-queries-toggle');
            await user.click(toggleButton);

            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());

            await waitFor(() =>
                expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'true')
            );
            expect(mockClearSelectedItem).not.toHaveBeenCalled();

            mockCypherEndpoint(multiNodeGraphResponse);
            await user.click(screen.getByRole('button', { name: /run cypher query/i }));

            // Saved queries panel is closed
            await waitFor(() =>
                expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'false')
            );

            // Entity info panel is closed
            expect(mockClearSelectedItem).toHaveBeenCalled();
        });

        it('does not close saved queries panel or entity info panel when search returns zero nodes', async () => {
            mockCypherEndpoint(singleNodeGraphResponse);
            // selectedItem in URL simulates a previously selected node (entity info panel open)
            const { screen, user } = await setup(cypherSearchState, cypherSearchRoute, true);

            const toggleButton = screen.getByTestId('common-queries-toggle');
            await user.click(toggleButton);

            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());

            await waitFor(() =>
                expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'true')
            );
            expect(mockClearSelectedItem).not.toHaveBeenCalled();

            mockCypherEndpoint(zeroNodeGraphResponse);
            await user.click(screen.getByRole('button', { name: /run cypher query/i }));

            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());

            // Saved queries panel stayed open
            expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'true');

            // Entity info panel stayed open
            expect(mockClearSelectedItem).not.toHaveBeenCalled();
            await waitFor(() => expect(window.location.search).toContain('selectedItem='));
        });

        it('does not close saved queries panel or entity info panel when search returns one node', async () => {
            mockCypherEndpoint(singleNodeGraphResponse);
            const { screen, user } = await setup(cypherSearchState, cypherSearchRoute, true);

            const toggleButton = screen.getByTestId('common-queries-toggle');
            await user.click(toggleButton);

            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());

            await waitFor(() =>
                expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'true')
            );
            expect(mockClearSelectedItem).not.toHaveBeenCalled();

            mockCypherEndpoint(singleNodeGraphResponse);
            await user.click(screen.getByRole('button', { name: /run cypher query/i }));

            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());

            // Saved queries panel stayed open
            expect(screen.getByTestId('common-queries-toggle')).toHaveAttribute('aria-expanded', 'true');

            // Entity info panel stayed open
            expect(mockClearSelectedItem).not.toHaveBeenCalled();
            await waitFor(() => expect(window.location.search).toContain('selectedItem='));
        });

        it('allows user to reopen saved queries panel after auto-close', async () => {
            mockCypherEndpoint(singleNodeGraphResponse);
            const { screen, user } = await setup(cypherSearchState, cypherSearchRoute, true);

            const toggleButton = screen.getByTestId('common-queries-toggle');
            await user.click(toggleButton);
            await waitFor(() => expect(screen.getByRole('button', { name: /run cypher query/i })).not.toBeDisabled());
            await waitFor(() => expect(toggleButton).toHaveAttribute('aria-expanded', 'true'));

            mockCypherEndpoint(multiNodeGraphResponse);
            await user.click(screen.getByRole('button', { name: /run cypher query/i }));
            await waitFor(() => expect(toggleButton).toHaveAttribute('aria-expanded', 'false'));

            await user.click(toggleButton);
            expect(toggleButton).toHaveAttribute('aria-expanded', 'true');
        });
    });
});
