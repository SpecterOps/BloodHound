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
import { PRIMARY_SEARCH, SECONDARY_SEARCH } from 'src/ducks/searchbar/types';

describe('Pathfinding: edge filter dialog', () => {
    beforeEach(async () => {
        await act(async () => {
            render(<PathfindingSearch />);
        });
    });

    it('should open path filtering dialog', async () => {
        const user = userEvent.setup();

        const dialog = screen.queryByRole('dialog', { name: /path edge filtering/i });
        expect(dialog).toBeNull();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        expect(screen.queryByRole('dialog', { name: /path edge filtering/i })).toBeInTheDocument();
    });

    it('should close the path filtering dialog when user clicks cancel button', async () => {
        const user = userEvent.setup();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        const dialog = screen.queryByRole('dialog', { name: /path edge filtering/i });
        expect(dialog).toBeInTheDocument();

        const cancelButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancelButton);

        expect(dialog).not.toBeVisible();
    });

    it('should close the path filtering dialog when user clicks apply button', async () => {
        const user = userEvent.setup();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        const cancelButton = screen.getByRole('button', { name: /apply/i });
        await user.click(cancelButton);

        const dialog = screen.getByRole('dialog', { name: /path edge filtering/i });
        expect(dialog).not.toBeVisible();
    });

    it('filter selections are rolled back if user closes modal with the cancel button', async () => {
        const user = userEvent.setup();

        // 1: open dialog
        const toggleDialogButton = screen.getByRole('button', { name: /filter/i });
        await user.click(toggleDialogButton);

        const activeDirectoryCategoryCheckbox = screen.getByRole('checkbox', { name: /active directory/i });
        expect(activeDirectoryCategoryCheckbox).toBeChecked();

        // 2: click active directory category, deselecting those edges
        await user.click(activeDirectoryCategoryCheckbox);
        expect(activeDirectoryCategoryCheckbox).not.toBeChecked();

        // 3: click apply to persist changes
        const applyButton = screen.getByRole('button', { name: /apply/i });
        await user.click(applyButton);

        // 4. open dialog again
        await user.click(toggleDialogButton);

        // 5. click active directory category, re-selecting those edges
        await user.click(activeDirectoryCategoryCheckbox);

        // 6. close the dialog with the cancel button, undoing the changes made above
        const cancelButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancelButton);

        // 7. open dialog a third time, active directory category should be unselected
        await user.click(toggleDialogButton);
        expect(activeDirectoryCategoryCheckbox).not.toBeChecked();
    });

    it('filter selections are persisted if user closes modal with the apply button', async () => {
        const user = userEvent.setup();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        const categoryADCheckbox = screen.getByRole('checkbox', { name: /active directory/i });
        const categoryAzureCheckbox = screen.getByRole('checkbox', { name: /azure/i });
        expect(categoryADCheckbox).toBeChecked();
        expect(categoryAzureCheckbox).toBeChecked();

        await user.click(categoryADCheckbox);
        await user.click(categoryAzureCheckbox);

        expect(categoryADCheckbox).not.toBeChecked();
        expect(categoryAzureCheckbox).not.toBeChecked();

        const applyButton = screen.getByRole('button', { name: /apply/i });
        await user.click(applyButton);

        await user.click(pathfindingButton);
        expect(categoryADCheckbox).not.toBeChecked();
        expect(categoryAzureCheckbox).not.toBeChecked();
    });
});

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
        const spy = vi.spyOn(actions, 'startSearchSelected');

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(spy).toHaveBeenLastCalledWith(PRIMARY_SEARCH);
        expect(spy).not.toHaveBeenCalledWith(SECONDARY_SEARCH);
    });

    it('executes a pathfinding search when both a source and destination node are provided', async () => {
        const user = userEvent.setup();
        const spy = vi.spyOn(actions, 'startSearchSelected');

        const startInput = screen.getByPlaceholderText(/start node/i);
        await user.type(startInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(spy).toHaveBeenLastCalledWith(PRIMARY_SEARCH);

        const destinationInput = screen.getByPlaceholderText(/destination node/i);
        await user.type(destinationInput, 'admin');
        await user.click(await screen.findByRole('option', { name: /admin/i }));

        expect(spy).toHaveBeenLastCalledWith(SECONDARY_SEARCH);
    });
});
