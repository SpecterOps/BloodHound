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
import * as hooks from '../../hooks';
import { render, screen, within } from '../../test-utils';
import { FinishedJobsFilterDialog } from './FinishedJobsFilterDialog';

const useObjectStateMock = vi.spyOn(hooks, 'useObjectState');

const mockObjectHook = (initialState = {}) => {
    const applyState = vi.fn();
    const deleteKeys = vi.fn();
    const state = initialState;

    useObjectStateMock.mockReturnValue({ applyState, deleteKeys, state } as any);

    return { applyState, deleteKeys };
};

const renderFilterDialog = async (open = true) => {
    render(<FinishedJobsFilterDialog onConfirm={() => {}} />);

    const user = userEvent.setup();
    const filterButton = await screen.findByTestId('finished_jobs_log-open_filter_dialog');

    if (open) {
        await user.click(filterButton);
    }

    return { filterButton, user };
};

// Radix Select relies on pointer events + scroll positioning under the hood
// (Popper + focus management). In JSDOM, those methods (scrollIntoView,
// hasPointerCapture, releasePointerCapture) don’t exist by default, so Radix
// crashes silently when trying to open the select dropdown.
beforeAll(() => {
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
    window.HTMLElement.prototype.hasPointerCapture = vi.fn();
    window.HTMLElement.prototype.releasePointerCapture = vi.fn();
});

describe('FinishedJobsFilterDialog', () => {
    it('renders a filter button', async () => {
        const { filterButton } = await renderFilterDialog(false);
        expect(filterButton).toBeInTheDocument();
    });

    it('opens and closes the filter', async () => {
        const { user } = await renderFilterDialog();

        const dialogTitle = screen.getByText('Filter');
        const status = screen.getByText('Status');
        const dataCollected = screen.getByText('Data Collected');

        expect(dialogTitle).toBeInTheDocument();
        expect(status).toBeInTheDocument();
        expect(dataCollected).toBeInTheDocument();

        const cancelButton = screen.getByRole('button', { name: 'Cancel' });
        expect(cancelButton).toBeInTheDocument();

        await user.click(cancelButton);
        expect(cancelButton).not.toBeInTheDocument();
    });
});

describe('FinishedJobsFilterDialog - Status Select', () => {
    const openSelect = async (user: ReturnType<typeof userEvent.setup>) => {
        // Grab the select trigger and click it
        const statusSelect = screen.getByRole('combobox', { name: 'Status Select' });
        await user.click(statusSelect);
    };

    it('has status filters', async () => {
        const { user } = await renderFilterDialog();
        await openSelect(user);

        // Grab the listbox that just opened (menu items)
        const listbox = await screen.findByRole('listbox');

        // Get all the menu items
        const options = within(listbox).getAllByRole('option');
        expect(options.map((o) => o.textContent)).toEqual(['None', 'Complete', 'Failed']);
    });

    it('filters by the selected status', async () => {
        const { applyState } = mockObjectHook({ status: '' });
        const { user } = await renderFilterDialog(true);
        await openSelect(user);

        // Select a status from the listbox
        const completeStatus = await screen.findByRole('option', { name: 'Complete' });
        await user.click(completeStatus);

        expect(applyState).toBeCalledWith({ status: '2' });
    });

    it('clears the applied filter', async () => {
        const { deleteKeys } = mockObjectHook({ status: '2' });
        const { user } = await renderFilterDialog(true);
        await openSelect(user);

        // Select a status from the listbox
        const completeStatus = await screen.findByRole('option', { name: 'None' });
        await user.click(completeStatus);

        expect(deleteKeys).toBeCalledWith('status');
    });
});

describe('FinishedJobsFilterDialog - Data Collected Select', () => {
    const openSelect = async (user: ReturnType<typeof userEvent.setup>) => {
        // Grab the select trigger and click it
        const dataCollectedSelect = screen.getByRole('combobox', { name: 'Data Collected Select' });
        await user.click(dataCollectedSelect);
    };

    it('has data collected filters', async () => {
        const { user } = await renderFilterDialog();
        openSelect(user);

        // Grab the listbox that just opened (menu items)
        const listbox = await screen.findByRole('listbox');

        // Get all the menu items
        const options = within(listbox).getAllByRole('option');
        expect(options.map((o) => o.textContent)).toEqual([
            'Sessions',
            'Local Groups',
            'AD Structure',
            'Certificate Services',
            'CA Registry',
            'DC Registry',
        ]);
    });

    it('filters by selected data collected', async () => {
        const { applyState } = mockObjectHook({});
        const { user } = await renderFilterDialog(true);
        await openSelect(user);

        // Select a couple data collecters from the listbox
        const adCollect = await screen.findByRole('option', { name: 'AD Structure' });
        const dcCollect = await screen.findByRole('option', { name: 'DC Registry' });
        await user.click(adCollect);
        await user.click(dcCollect);

        expect(applyState).toHaveBeenNthCalledWith(1, { ad_structure_collection: true });
        expect(applyState).toHaveBeenNthCalledWith(2, { dc_registry_collection: true });
    });

    it('clears the applied filter', async () => {
        const { deleteKeys } = mockObjectHook({ ad_structure_collection: true, dc_registry_collection: true });
        const { user } = await renderFilterDialog(true);
        await openSelect(user);

        // Unselect the selected data collecters
        const adCollect = await screen.findByRole('option', { name: 'AD Structure' });
        const dcCollect = await screen.findByRole('option', { name: 'DC Registry' });
        await user.click(adCollect);
        await user.click(dcCollect);

        expect(deleteKeys).toHaveBeenNthCalledWith(1, 'ad_structure_collection');
        expect(deleteKeys).toHaveBeenNthCalledWith(2, 'dc_registry_collection');
    });
});
