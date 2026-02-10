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
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { usePathfindingFilters } from '../../../../hooks/useExploreGraph/usePathfindingFilters';
import { act, render, screen, waitFor } from '../../../../test-utils';
import { EdgeFilter } from './EdgeFilter';

const server = setupServer(
    rest.get('/api/v2/features', (req, res, ctx) => {
        return res(
            ctx.json({
                data: [],
            })
        );
    })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const WrappedEdgeFilter = () => {
    const pathfindingFilterState = usePathfindingFilters();
    return <EdgeFilter pathfindingFilterState={pathfindingFilterState} />;
};

describe('EdgeFilter', () => {
    beforeEach(async () => {
        await act(async () => render(<WrappedEdgeFilter />));
    });

    it('should open edge filtering dialog', async () => {
        const user = userEvent.setup();

        const dialog = screen.queryByRole('dialog', { name: /path edge filtering/i });
        expect(dialog).toBeNull();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        expect(screen.queryByRole('dialog', { name: /path edge filtering/i })).toBeInTheDocument();
    });

    it('should close the edge filtering dialog when user clicks cancel button', async () => {
        const user = userEvent.setup();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        const dialog = screen.queryByRole('dialog', { name: /path edge filtering/i });
        expect(dialog).toBeInTheDocument();

        const cancelButton = screen.getByRole('button', { name: /cancel/i });
        await user.click(cancelButton);

        expect(dialog).not.toBeVisible();
    });

    it('should close the edge filtering dialog when user clicks apply button', async () => {
        const user = userEvent.setup();

        const pathfindingButton = screen.getByRole('button', { name: /filter/i });
        await user.click(pathfindingButton);

        const applyButton = screen.getByRole('button', { name: /apply/i });
        await user.click(applyButton);

        await waitFor(() => {
            expect(screen.queryByRole('dialog', { name: /path edge filtering/i })).not.toBeInTheDocument();
        });
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
        expect(await screen.findByRole('checkbox', { name: /active directory/i })).not.toBeChecked();
    });

    // Skipping this since our url param state is not syncing correctly in tests
    it.skip('filter selections are persisted if user closes modal with the apply button', async () => {
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
