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
import { render, screen } from '../../test-utils';
import { FinishedJobsFilterDialog } from './FinishedJobsFilterDialog';

const renderFilterDialog = async (open = true) => {
    render(<FinishedJobsFilterDialog onConfirm={() => {}} />);

    const user = userEvent.setup();
    const filterButton = await screen.findByTestId('finished_jobs_log-open_filter_dialog');

    if (open) {
        await user.click(filterButton);
    }

    return { filterButton, user };
};

describe('FinishedJobsFilterDialog', () => {
    it('renders a filter button', async () => {
        const { filterButton } = await renderFilterDialog(false);
        expect(filterButton).toBeInTheDocument();
    });

    it('displays the filter', async () => {
        await renderFilterDialog();

        const dialogTitle = screen.getByText('Filter');
        const status = screen.getByText('Status');
        const dataCollected = screen.getByText('Data Collected');

        expect(dialogTitle).toBeInTheDocument();
        expect(status).toBeInTheDocument();
        expect(dataCollected).toBeInTheDocument();
    });

    it('closes the filter', async () => {
        const { user } = await renderFilterDialog();

        const cancelButton = screen.getByRole('button', { name: 'Cancel' });
        expect(cancelButton).toBeInTheDocument();

        await user.click(cancelButton);
        expect(cancelButton).not.toBeInTheDocument();
    });
});

// describe('FinishedJobsFilterDialog - Status Select', () => {
//     it('filters by status', async () => {
//         await renderFilterDialog();

//         const statusSelect = screen.getByRole('combobox', { name: 'Status Select' });
//         expect(statusSelect).toBeInTheDocument();
//     });
// });

// describe('FinishedJobsFilterDialog - Status Select', () => {
//     it('has select by data collected', async () => {
//         const { user } = await renderFilterDialog();

//         const dataCollectedSelect = screen.getByRole('combobox', { name: 'Data Collected Select' });
//         await expect(dataCollectedSelect).toBeInTheDocument();
//     });
// });
