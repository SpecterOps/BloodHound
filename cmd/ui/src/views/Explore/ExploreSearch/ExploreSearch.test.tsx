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

import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { act, render, screen } from 'src/test-utils';
import ExploreSearch from './ExploreSearch';

import userEvent from '@testing-library/user-event';
import { mockCodemirrorLayoutMethods } from 'bh-shared-ui';

const comboboxLookaheadOptions = {
    data: [
        {
            name: 'admin',
            objectid: '1',
            type: 'User',
        },
        {
            name: 'computer',
            objectid: '2',
            type: 'Computer',
        },
    ],
};

const server = setupServer(
    rest.get('/api/v2/search', (req, res, ctx) => {
        return res(ctx.json(comboboxLookaheadOptions));
    }),
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [{ id: 1, key: 'tier_management_engine', enabled: true }],
            })
        );
    }),
    rest.get('/api/v2/graphs/kinds', async (_req, res, ctx) => {
        return res(
            ctx.json({
                data: { kinds: ['Tier Zero', 'Tier One', 'Tier Two'] },
            })
        );
    }),
    rest.get(`/api/v2/customnode`, async (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
beforeEach(() => mockCodemirrorLayoutMethods());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const setup = async (exploreSearchTab?: string) => {
    const url = exploreSearchTab ? `/?exploreSearchTab=${exploreSearchTab}` : '/';

    const screen = await act(async () => {
        return render(<ExploreSearch />, { route: url });
    });

    const user = userEvent.setup();

    return { screen, user };
};

// Example

describe('ExploreSearch rendering per tab', async () => {
    it('should render', async () => {
        await setup();
        expect(screen.getByLabelText('Search Nodes')).toBeInTheDocument();

        expect(screen.getByRole('tab', { name: /search/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /pathfinding/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /cypher/i })).toBeInTheDocument();
    });

    it('should render the pathfinding search controls when searchType is pathfinding', async () => {
        await setup('pathfinding');

        expect(screen.getByLabelText(/start node/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/destination node/i)).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /right-left/i })).toBeDisabled();
        expect(screen.getByRole('button', { name: /filter/i })).toBeInTheDocument();
    });

    it('should render the cypher search controls when user clicks on cypher tab ', async () => {
        await setup('cypher');

        expect(screen.getByText(/cypher query/i)).toBeInTheDocument();

        expect(screen.getByRole('link', { name: /help/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /run/ })).toBeInTheDocument();
    });
    // To do: Work on this when TW css classes are applied in test environment
    it.todo('should hide/expand search widget when user clicks minus/plus button', async () => {
        const { user } = await setup();
        const widgetBody = screen.getByLabelText('Search Nodes');
        expect(widgetBody).toBeVisible();

        const toggleWidgetButton = screen.getByRole('button', { name: /minus/i });

        await user.click(toggleWidgetButton);

        expect(widgetBody).not.toBeVisible();
        // button changes from minus to plus
        expect(toggleWidgetButton).toHaveAccessibleName('plus');
    });
});

describe('ExploreSearch sets searchType on tab changing', async () => {
    it('sets exploreSearchTab param to node when the user clicks the `Search` tab', async () => {
        const { screen, user } = await setup('pathfinding');

        const exploreSearchTab = screen.getByRole('tab', { name: /search/i });
        await user.click(exploreSearchTab);

        expect(window.location.search).toContain('exploreSearchTab=node');
    });

    it('sets exploreSearchTab param to pathfinding when the user clicks the `pathfinding` tab', async () => {
        const { screen, user } = await setup();

        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });
        await user.click(pathfindingTab);

        expect(window.location.search).toContain('exploreSearchTab=pathfinding');
    });

    it('sets exploreSearchTab param to cypher when the user clicks the `cypher` tab', async () => {
        const { screen, user } = await setup();

        const cypherTab = screen.getByRole('tab', { name: /cypher/i });
        await user.click(cypherTab);

        expect(window.location.search).toContain('exploreSearchTab=cypher');
    });

    it('initializes search tab to node search if the exploreSearchTab is not a supported tab name on first render', async () => {
        const { screen } = await setup('unsupported_tab');
        const primarySearchInput = screen.getByPlaceholderText('Search Nodes');
        expect(primarySearchInput).toBeInTheDocument();
    });

    it('initializes search tab to the exploreSearchTab on initial render', async () => {
        const { screen } = await setup('pathfinding');
        const startNodeInput = screen.getByPlaceholderText('Start Node');
        const endNodeInput = screen.getByPlaceholderText('Destination Node');
        expect(startNodeInput).toBeInTheDocument();
        expect(endNodeInput).toBeInTheDocument();
    });
});

// Clicking a new tab in these tests cause a query param update but not an actual tab change -- maybe a bad interaction
// between createMemoryHistory and useExploreParams
describe('ExploreSearch interaction', () => {
    it('when user performs a single node search, the selected node carries over to the `start node` input field on the pathfinding tab', async () => {
        const { screen, user } = await setup();
        const searchInput = screen.getByPlaceholderText('Search Nodes');
        const userSuppliedSearchTerm = 'admin';
        await user.type(searchInput, userSuppliedSearchTerm);

        // select first option from list and check that text field input is updated
        const firstOption = await screen.findByRole('option', { name: /admin/i });
        await user.click(firstOption);
        expect(searchInput).toHaveValue(userSuppliedSearchTerm);

        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });
        await user.click(pathfindingTab);

        const startNodeInputField = screen.getByPlaceholderText(/start node/i);
        expect(startNodeInputField).toHaveValue(userSuppliedSearchTerm);
    });

    it('when user performs a pathfinding search, the selection for the start node is carried over to the `search` tab', async () => {
        const { screen, user } = await setup();
        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });
        await user.click(pathfindingTab);

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        const exploreSearchTab = screen.getByRole('tab', { name: /search/i });
        await user.click(exploreSearchTab);

        const searchInput = screen.getByPlaceholderText('Search Nodes');
        expect(searchInput).toHaveValue('admin');
    });
});
