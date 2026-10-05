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
import { fn } from '@storybook/test';
import InfiniteScrollingTable from './InfiniteScrollingTable';

const items = Array.from({ length: 300 }, (_, index) => ({
    name: `USER-${index + 1}@EXAMPLE.COM`,
    objectID: `S-1-5-21-1000-${index + 1}`,
    label: 'User',
}));
const fetchData: React.ComponentProps<typeof InfiniteScrollingTable>['fetchDataCallback'] = async ({
    skip,
    limit,
}) => ({
    data: items.slice(skip, skip + limit),
    total: items.length,
    skip,
    limit,
});
const meta: Meta<typeof InfiniteScrollingTable> = {
    title: 'Components/InfiniteScrollingTable',
    component: InfiniteScrollingTable,
    tags: ['autodocs'],
    args: { fetchDataCallback: fetchData, itemCount: items.length, onClick: fn() },
    argTypes: { fetchDataCallback: { control: false } },
    decorators: [
        (Story) => (
            <div className='max-w-xl'>
                <Story />
            </div>
        ),
    ],
    parameters: {
        docs: {
            description: {
                component:
                    'Pages deterministic local data through the real fetch callback. Scroll to request more rows; no API calls are made.',
            },
        },
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Paginated: Story = {};
export const ShortList: Story = { args: { itemCount: 3 } };
export const Loading: Story = { args: { itemCount: 10, fetchDataCallback: () => new Promise(() => {}) } };
export const ReadOnly: Story = { args: { onClick: undefined } };
