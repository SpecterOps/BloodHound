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
import { act, render, screen, waitFor } from 'src/test-utils';
import ExploreSearch from '.';
import { setupServer } from 'msw/node';
import { rest } from 'msw';
import * as actions from 'src/ducks/searchbar/actions';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, SECONDARY_SEARCH } from 'src/ducks/searchbar/types';
import { initialSearchState } from 'src/ducks/searchbar/reducer';

describe('ExploreSearch rendering per tab', async () => {
    let container: HTMLElement;
    beforeEach(async () => {
        await act(async () => {
            const { container: c } = render(<ExploreSearch />);
            container = c;
        });
    });
    const user = userEvent.setup();

    it('should render', () => {
        screen.debug();
        expect(screen.getByLabelText('Search Nodes')).toBeInTheDocument();

        expect(screen.getByRole('tab', { name: /search/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /pathfinding/i })).toBeInTheDocument();
        expect(screen.getByRole('tab', { name: /cypher/i })).toBeInTheDocument();
    });

    it('should render the pathfinding search controls when user clicks on pathfinding tab ', async () => {
        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });

        await user.click(pathfindingTab);

        expect(screen.getByLabelText(/start node/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/destination node/i)).toBeInTheDocument();

        expect(screen.getByRole('button', { name: /right-left/i })).toBeDisabled();
        expect(screen.getByRole('button', { name: /filter/i })).toBeInTheDocument();
    });

    it('should render the cypher search controls when user clicks on cypher tab ', async () => {
        const cypherTab = screen.getByRole('tab', { name: /cypher/i });

        await user.click(cypherTab);

        expect(screen.getByText(/cypher search/i)).toBeInTheDocument();

        expect(screen.getByRole('link', { name: /help/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument();
    });

    it('should hide/expand search widget when user clicks minus/plus button', async () => {
        const widgetBody = screen.getByLabelText('Search Nodes');
        expect(widgetBody).toBeVisible();

        const toggleWidgetButton = screen.getByRole('button', { name: /minus/i });

        await user.click(toggleWidgetButton);

        // mui applies 300ms of animation to the collapse element, so we need to wait for the hidden class to be in the document
        await waitFor(() => expect(container.querySelector('.MuiCollapse-hidden')).toBeInTheDocument());

        expect(widgetBody).not.toBeVisible();
        // button changes from minus to plus
        expect(toggleWidgetButton).toHaveAccessibleName('plus');
    });
});

describe('ExploreSearch handles search on tab changing', async () => {
    const sourceNode = {
        objectid: '1',
        label: 'beep',
        type: 'Computer',
        name: 'computer a',
    };

    const server = setupServer(
        rest.get(`/api/v2/search`, (req, res, ctx) => {
            return res(ctx.json([]));
        })
    );

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('should perform a primary search when the user clicks the `Search` tab', async () => {
        await act(async () => {
            render(<ExploreSearch />, {
                initialState: {
                    search: {
                        ...initialSearchState,
                        primary: {
                            value: sourceNode,
                            searchTerm: sourceNode.name,
                        },
                        activeTab: SECONDARY_SEARCH,
                    },
                },
            });
        });

        const user = userEvent.setup();
        const startSearchSelectedSpy = vi.spyOn(actions, 'startSearchSelected');
        const setSearchValueSpy = vi.spyOn(actions, 'setSearchValue');
        const setActiveTabSpy = vi.spyOn(actions, 'setActiveTab');

        const searchTab = screen.getByRole('tab', { name: /search/i });
        await user.click(searchTab);

        expect(startSearchSelectedSpy).toHaveBeenLastCalledWith(PRIMARY_SEARCH);
        expect(setSearchValueSpy).toHaveBeenLastCalledWith(null, SECONDARY_SEARCH, SEARCH_TYPE_EXACT);

        expect(setActiveTabSpy).toHaveBeenCalledTimes(1);
        expect(setActiveTabSpy).toHaveBeenCalledWith(PRIMARY_SEARCH);
    });

    it('should perform a pathfinding search when the user clicks the `pathfinding` tab', async () => {
        await act(async () => {
            render(<ExploreSearch />, {
                initialState: {
                    search: {
                        ...initialSearchState,
                        primary: {
                            value: sourceNode,
                            searchTerm: sourceNode.name,
                        },
                        secondary: {
                            value: sourceNode,
                            searchTerm: sourceNode.name,
                        },
                    },
                },
            });
        });

        const user = userEvent.setup();
        const startSearchSelectedSpy = vi.spyOn(actions, 'startSearchSelected');
        const setActiveTabSpy = vi.spyOn(actions, 'setActiveTab');

        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });
        await user.click(pathfindingTab);

        expect(startSearchSelectedSpy).toHaveBeenLastCalledWith(SECONDARY_SEARCH);

        expect(setActiveTabSpy).toHaveBeenCalledTimes(1);
        expect(setActiveTabSpy).toHaveBeenCalledWith(SECONDARY_SEARCH);
    });
});

describe('ExploreSearch interaction', () => {
    const user = userEvent.setup();

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
        })
    );

    beforeEach(async () => {
        await act(async () => {
            render(<ExploreSearch />);
        });
    });

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('when user performs a single node search, the selected node carries over to the `start node` input field on the pathfinding tab', async () => {
        const searchInput = screen.getByPlaceholderText(/search nodes/i);
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
        const pathfindingTab = screen.getByRole('tab', { name: /pathfinding/i });
        await user.click(pathfindingTab);

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        const searchTab = screen.getByRole('tab', { name: /search/i });
        await user.click(searchTab);

        const searchInput = screen.getByPlaceholderText(/search nodes/i);
        expect(searchInput).toHaveValue('admin');
    });
});
