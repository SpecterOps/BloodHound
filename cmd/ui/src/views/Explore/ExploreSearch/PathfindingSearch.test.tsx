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

import { act } from 'react-dom/test-utils';
import { render, screen } from 'src/test-utils';
import PathfindingSearch from './PathfindingSearch';
import userEvent from '@testing-library/user-event';
import { setupServer } from 'msw/node';
import { rest } from 'msw';
import * as actions from 'src/ducks/searchbar/actions';

describe('Pathfinding: interaction', () => {
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
            render(<PathfindingSearch />);
        });
    });

    beforeAll(() => server.listen());
    afterEach(() => server.resetHandlers());
    afterAll(() => server.close());

    it('when user performs a pathfinding search, the swap button is disabled until both the start and destination nodes are provided', async () => {
        const user = userEvent.setup();

        const swapButton = screen.getByRole('button', { name: /right-left/i });
        expect(swapButton).toBeDisabled();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(swapButton).toBeDisabled();

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(swapButton).toBeEnabled();
    });

    it('when user performs a pathfinding search, and then clicks the swap button, the start and destination inputs are swapped', async () => {
        const user = userEvent.setup();

        const swapButton = screen.getByRole('button', { name: /right-left/i });
        expect(swapButton).toBeDisabled();

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'computer');
        await user.click(await screen.findByRole('option', { name: /computer/i }));

        expect(startInput).toHaveValue('admin');
        expect(destinationInput).toHaveValue('computer');

        await user.click(swapButton);

        expect(startInput).toHaveValue('computer');
        expect(destinationInput).toHaveValue('admin');
    });

    it('executes a primary search when only a source node is provided', async () => {
        const user = userEvent.setup();
        const spy = vi.spyOn(actions, 'sourceNodeSelected');

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(spy).toHaveBeenCalledTimes(1);
        expect(spy).toHaveBeenCalledWith(comboboxLookaheadOptions.data[0]);
    });

    it('executes a pathfinding search when both a source and destination node are provided', async () => {
        const user = userEvent.setup();
        const sourceNodeSelectedSpy = vi.spyOn(actions, 'sourceNodeSelected');
        const destinationNodeSelectedSpy = vi.spyOn(actions, 'destinationNodeSelected');

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(sourceNodeSelectedSpy).toHaveBeenCalledTimes(1);
        expect(sourceNodeSelectedSpy).toHaveBeenCalledWith(comboboxLookaheadOptions.data[0]);

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'computer');
        await user.click(await screen.findByRole('option', { name: /computer/i }));

        expect(destinationNodeSelectedSpy).toHaveBeenCalledTimes(1);
        expect(destinationNodeSelectedSpy).toHaveBeenCalledWith(comboboxLookaheadOptions.data[1]);
    });
});
