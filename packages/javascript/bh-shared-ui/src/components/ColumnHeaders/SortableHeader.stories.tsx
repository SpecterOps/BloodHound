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
import { expect, fn, userEvent, within } from '@storybook/test';
import { useEffect, useState } from 'react';
import { SortableHeader } from './ColumnHeaders';

const meta: Meta<typeof SortableHeader> = {
    title: 'Components/ColumnHeaders/SortableHeader',
    component: SortableHeader,
    tags: ['autodocs'],
    args: { title: 'Name', onSort: fn() },
    decorators: [
        (Story) => (
            <div className='p-6'>
                <Story />
            </div>
        ),
    ],
    render: function Render(args) {
        const [sortOrder, setSortOrder] = useState(args.sortOrder);
        useEffect(() => setSortOrder(args.sortOrder), [args.sortOrder]);
        return (
            <div className='w-fit max-w-fit'>
                <SortableHeader
                    {...args}
                    sortOrder={sortOrder}
                    onSort={() => {
                        args.onSort();
                        setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
                    }}
                />
            </div>
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Unsorted: Story = {};
export const Ascending: Story = { args: { sortOrder: 'asc' } };
export const Descending: Story = { args: { sortOrder: 'desc' } };
export const WithTooltip: Story = { args: { tooltipText: 'Sort environments alphabetically by name.' } };
export const Disabled: Story = { args: { disable: true } };
export const Sort: Story = {
    play: async ({ canvasElement, args }) => {
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Sort by Name' }));
        await expect(args.onSort).toHaveBeenCalledTimes(1);
    },
};
