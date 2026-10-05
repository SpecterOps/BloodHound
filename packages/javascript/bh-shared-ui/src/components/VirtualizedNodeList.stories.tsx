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
import { ComponentProps, useState } from 'react';
import VirtualizedNodeList, { NormalizedNodeItem } from './VirtualizedNodeList';

const onSelect = fn();
const nodes: NormalizedNodeItem[] = Array.from({ length: 60 }, (_, index) => ({
    name: index % 2 === 0 ? `USER-${index + 1}@EXAMPLE.COM` : `COMPUTER-${index + 1}.EXAMPLE.COM`,
    objectId: `S-1-5-21-1000-${index + 1}`,
    kind: index % 2 === 0 ? 'User' : 'Computer',
    onClick: onSelect,
}));

function SelectableNodeList(args: ComponentProps<typeof VirtualizedNodeList<NormalizedNodeItem>>) {
    const [selectedIndex, setSelectedIndex] = useState<number>();
    const selectedNode = selectedIndex === undefined ? undefined : args.nodes[selectedIndex];
    const hasSelectableNodes = args.nodes.some((node) => typeof node.onClick === 'function');
    const selectableNodes = args.nodes.map((node) => ({
        ...node,
        onClick:
            typeof node.onClick === 'function'
                ? (index: number) => {
                      node.onClick?.(index);
                      setSelectedIndex(index);
                  }
                : undefined,
    }));

    return (
        <>
            <VirtualizedNodeList {...args} nodes={selectableNodes} />
            {hasSelectableNodes && (
                <div
                    role='status'
                    className='mt-3 min-h-24 rounded border border-neutral-3 bg-neutral-2 p-3 break-words'>
                    <p className='mb-1 text-sm font-semibold'>Selected node</p>
                    {selectedNode ? (
                        <>
                            <p>{selectedNode.name || selectedNode.objectId || 'NO NAME'}</p>
                            <p className='text-sm'>
                                {selectedNode.kind} · {selectedNode.objectId}
                            </p>
                        </>
                    ) : (
                        <p className='text-sm'>Click a row or focus it and press Enter to view its details.</p>
                    )}
                </div>
            )}
        </>
    );
}

const meta: Meta<typeof VirtualizedNodeList<NormalizedNodeItem>> = {
    title: 'Components/VirtualizedNodeList',
    component: VirtualizedNodeList<NormalizedNodeItem>,
    tags: ['autodocs'],
    parameters: {
        docs: {
            description: {
                component:
                    'A virtualized list of nodes. Select a row with the mouse or keyboard to view its details below the list. Selection callbacks are also logged in Actions.',
            },
        },
    },
    args: { nodes, heightScalar: 8 },
    decorators: [
        (Story) => (
            <div className='max-w-xl'>
                <Story />
            </div>
        ),
    ],
    render: (args) => <SelectableNodeList {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Scrollable: Story = {};
export const ShortList: Story = { args: { nodes: nodes.slice(0, 3) } };
export const ReadOnly: Story = { args: { nodes: nodes.slice(0, 3).map((node) => ({ ...node, onClick: undefined })) } };
export const NameFallback: Story = { args: { nodes: [{ name: '', objectId: 'S-1-5-21-1000-500', kind: 'User' }] } };
export const SelectWithMouse: Story = {
    play: async ({ canvasElement }) => {
        onSelect.mockClear();
        await userEvent.click(within(canvasElement).getAllByRole('button')[1]);
        await expect(onSelect).toHaveBeenCalledWith(1);
        const selection = within(canvasElement).getByRole('status');
        await expect(selection).toHaveTextContent('COMPUTER-2.EXAMPLE.COM');
        await expect(selection).toHaveTextContent('Computer · S-1-5-21-1000-2');
    },
};
export const SelectWithKeyboard: Story = {
    play: async ({ canvasElement }) => {
        onSelect.mockClear();
        const first = within(canvasElement).getAllByRole('button')[0];
        first.focus();
        await userEvent.keyboard('{Enter}');
        await expect(onSelect).toHaveBeenCalledWith(0);
        const selection = within(canvasElement).getByRole('status');
        await expect(selection).toHaveTextContent('USER-1@EXAMPLE.COM');
        await expect(selection).toHaveTextContent('User · S-1-5-21-1000-1');
    },
};
