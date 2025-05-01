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
import { createMemoryHistory } from 'history';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { usePathfindingFilters, usePathfindingSearch } from '../../../hooks';
import { act, render } from '../../../test-utils';
import PathfindingSearch from './PathfindingSearch';

describe('Pathfinding: interaction', () => {
    const comboboxLookaheadOptions = {
        data: [
            {
                name: 'admin1',
                objectid: '1',
                type: 'User',
            },
            {
                name: 'admin2',
                objectid: '2',
                type: 'User',
            },
            {
                name: 'computer',
                objectid: '3',
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
                    data: [],
                })
            );
        })
    );

    const WrappedPathfindingSearch = () => {
        const pathfindingSearchState = usePathfindingSearch();
        const pathfindingFilterState = usePathfindingFilters();
        return (
            <PathfindingSearch
                pathfindingSearchState={pathfindingSearchState}
                pathfindingFilterState={pathfindingFilterState}
            />
        );
    };

    const setup = async () => {
        const history = createMemoryHistory({ initialEntries: ['/'] });
        const screen = await act(async () => render(<WrappedPathfindingSearch />, { history }));
        const user = userEvent.setup();

        return { screen, history, user };
    };

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('when user performs a pathfinding search, the swap button is disabled until both the start and destination nodes are provided', async () => {
        const { screen, user } = await setup();

        const swapButton = screen.getByRole('button', { name: /right-left/i });
        expect(swapButton).toBeDisabled();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin1');
        await user.click(await screen.findByRole('option', { name: /admin1/i }));

        expect(swapButton).toBeDisabled();

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'admin1');
        await user.click(await screen.findByRole('option', { name: /admin1/i }));

        expect(swapButton).toBeEnabled();
    });

    it('when user performs a pathfinding search, and then clicks the swap button, the start and destination inputs are swapped', async () => {
        const { screen, user, history } = await setup();

        const swapButton = screen.getByRole('button', { name: /right-left/i });
        expect(swapButton).toBeDisabled();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin1');
        await user.click(await screen.findByRole('option', { name: /admin1/i }));

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'computer');
        await user.click(await screen.findByRole('option', { name: /computer/i }));

        expect(history.location.search).toContain(`primarySearch=${comboboxLookaheadOptions.data[0].objectid}`);
        expect(history.location.search).toContain(`secondarySearch=${comboboxLookaheadOptions.data[2].objectid}`);

        await user.click(swapButton);

        expect(history.location.search).toContain(`primarySearch=${comboboxLookaheadOptions.data[2].objectid}`);
        expect(history.location.search).toContain(`secondarySearch=${comboboxLookaheadOptions.data[0].objectid}`);
    });

    it('executes a primary search when only a source node is provided', async () => {
        const { screen, user, history } = await setup();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin1');
        await user.click(await screen.findByRole('option', { name: /admin1/i }));

        expect(history.location.search).toContain('searchType=node');
        expect(history.location.search).toContain(`primarySearch=${comboboxLookaheadOptions.data[0].objectid}`);
    });

    it('executes a pathfinding search when both a source and destination node are provided', async () => {
        const { screen, user, history } = await setup();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin1');

        const startOption = await screen.findByRole('option', { name: /admin1/i });
        await user.click(startOption);

        // searchType is 'node' because a destination is not yet selected
        expect(history.location.search).toContain('searchType=node');
        expect(history.location.search).toContain(`primarySearch=${comboboxLookaheadOptions.data[0].objectid}`);

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'computer');

        const destinationOption = await screen.findByRole('option', { name: /computer/i });
        await user.click(destinationOption);

        expect(history.location.search).toContain('searchType=pathfinding');
        expect(history.location.search).toContain(`primarySearch=${comboboxLookaheadOptions.data[0].objectid}`);
        expect(history.location.search).toContain(`secondarySearch=${comboboxLookaheadOptions.data[2].objectid}`);
    });
});
