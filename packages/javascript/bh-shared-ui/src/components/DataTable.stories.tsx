// Copyright 2026 Specter Ops, Inc.
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
import type { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, within } from '@storybook/test';
import { ComponentProps, useState } from 'react';
import DataTable from './DataTable';
import { StatusIndicator } from './StatusIndicator';

const headers = [{ label: 'Environment' }, { label: 'Objects', alignment: 'right' as const }, { label: 'Status' }];
const data = [
    ['Production', 12450, 'Complete'],
    ['Test', 875, 'Pending'],
    ['Development', 240, 'Failed'],
];
const meta: Meta<typeof DataTable> = {
    title: 'Components/DataTable',
    component: DataTable,
    tags: ['autodocs'],
    args: { headers, data },
    decorators: [
        (Story) => (
            <div className='max-w-3xl'>
                <Story />
            </div>
        ),
    ],
};

export default meta;
type Story = StoryObj<typeof DataTable>;

export const Default: Story = {};
export const Loading: Story = { args: { isLoading: true } };
export const Empty: Story = { args: { data: [] } };
function RichCells(args: ComponentProps<typeof DataTable>) {
    return (
        <DataTable
            {...args}
            data={[
                ['Production', 12450, <StatusIndicator key='good' status='good' label='Complete' />],
                ['Test', 875, <StatusIndicator key='pending' status='pending' label='Pending' pulse />],
                ['Development', 240, <StatusIndicator key='bad' status='bad' label='Failed' />],
            ]}
        />
    );
}
export const ReactCells: Story = { render: (args) => <RichCells {...args} /> };
const paginatedData = Array.from({ length: 23 }, (_, index) => [
    `Environment ${index + 1}`,
    (index + 1) * 100,
    'Complete',
]);
function PaginatedTable(args: ComponentProps<typeof DataTable>) {
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(5);
    return (
        <DataTable
            {...args}
            data={paginatedData.slice(page * rowsPerPage, (page + 1) * rowsPerPage)}
            showPaginationControls
            paginationProps={{
                page,
                rowsPerPage,
                count: paginatedData.length,
                onPageChange: (_, nextPage) => setPage(nextPage),
                onRowsPerPageChange: (event) => {
                    setRowsPerPage(Number(event.target.value));
                    setPage(0);
                },
            }}
        />
    );
}
export const Pagination: Story = {
    render: (args) => <PaginatedTable {...args} />,
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        await expect(canvas.getByText('Environment 1')).toBeVisible();
        await userEvent.click(canvas.getByRole('button', { name: 'Go to next page' }));
        await expect(canvas.getByText('Environment 6')).toBeVisible();
        await expect(canvas.queryByText('Environment 1')).not.toBeInTheDocument();
    },
};
