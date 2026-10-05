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
import { expect, within } from '@storybook/test';
import { useInfiniteQuery } from 'react-query';
import { PaginatedResult } from '../../utils/paginatedFetcher';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from './InfiniteQueryFixedList';

type Item = { id: number; name: string };
type Scenario = 'populated' | 'loading' | 'empty';
const items: Item[] = Array.from({ length: 80 }, (_, index) => ({ id: index + 1, name: `Environment ${index + 1}` }));
type ExampleProps = Partial<
    Pick<InfiniteQueryFixedListProps<Item>, 'itemSize' | 'placeholderCount' | 'overscanCount' | 'thresholdCount'>
> & { scenario?: Scenario };
function InfiniteListExample({
    scenario = 'populated',
    itemSize = 40,
    placeholderCount = 4,
    overscanCount,
    thresholdCount,
}: ExampleProps) {
    const queryResult = useInfiniteQuery<PaginatedResult<Item>>({
        queryKey: ['storybook-environments', scenario],
        queryFn: async ({ pageParam = { skip: 0, limit: 20 } }) => {
            if (scenario === 'loading') return new Promise(() => {});
            if (scenario === 'empty') return { items: [] };
            const { skip, limit } = pageParam;
            return {
                items: items.slice(skip, skip + limit),
                nextPageParam: skip + limit < items.length ? { skip: skip + limit, limit } : undefined,
            };
        },
        getNextPageParam: (page) => page.nextPageParam,
    });
    return (
        <div className='max-w-xl'>
            {queryResult.isSuccess && scenario === 'empty' && <p>No environments found.</p>}
            <InfiniteQueryFixedList<Item>
                itemSize={itemSize}
                queryResult={queryResult}
                placeholderCount={placeholderCount}
                overscanCount={overscanCount}
                thresholdCount={thresholdCount}
                renderRow={(item, _, style) => (
                    <div style={style} className='border-b border-neutral-3 px-3 py-2'>
                        {item.name}
                    </div>
                )}
                renderLoadingRow={(_, style) => (
                    <div style={style} className='px-3 py-2 animate-pulse'>
                        Loading environments…
                    </div>
                )}
            />
        </div>
    );
}
const meta: Meta<typeof InfiniteQueryFixedList<Item>> = {
    title: 'Components/InfiniteQueryFixedList',
    component: InfiniteQueryFixedList<Item>,
    tags: ['autodocs'],
    args: { itemSize: 40, placeholderCount: 4, overscanCount: 5, thresholdCount: 5 },
    parameters: {
        docs: {
            description: {
                component:
                    'A virtualized list backed by a real React Query infinite query over local fixtures. Scroll to fetch later pages. Empty-state presentation belongs to the caller.',
            },
        },
    },
    argTypes: {
        queryResult: { control: false },
        renderRow: { control: false },
        renderLoadingRow: { control: false },
        listRef: { control: false },
    },
    render: (args) => <InfiniteListExample {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Paginated: Story = {
    play: async ({ canvasElement }) => {
        await expect(await within(canvasElement).findByText('Environment 1')).toBeVisible();
    },
};
export const Loading: Story = { render: (args) => <InfiniteListExample {...args} scenario='loading' /> };
export const Empty: Story = { render: (args) => <InfiniteListExample {...args} scenario='empty' /> };
