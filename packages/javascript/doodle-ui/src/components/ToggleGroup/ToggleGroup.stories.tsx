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
import { faCircleInfo, faStar } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { userEvent, within } from '@storybook/test';
import { ToggleGroup, ToggleGroupItem } from './ToggleGroup';

const meta = {
    title: 'Components/ToggleGroup',
    component: ToggleGroup,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        size: {
            options: ['sm', 'lg'],
            control: 'select',
            description: 'Size applied to all items in the group.',
        },
        disabled: {
            control: 'boolean',
            description: 'Disables all items in the group.',
        },
    },
    args: {
        type: 'single',
        size: 'lg',
        disabled: false,
    },
} satisfies Meta<typeof ToggleGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Small size toggle group.
 */
export const Small: Story = {
    render: (args) => (
        <ToggleGroup {...args} size='sm'>
            <ToggleGroupItem value='a'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='b'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Small size toggle group.
 */
export const Large: Story = {
    render: (args) => (
        <ToggleGroup {...args} size='lg'>
            <ToggleGroupItem value='a'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='b'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Toggle group items with icons on the left.
 */
export const WithIconsLeft: Story = {
    render: (args) => (
        <ToggleGroup {...args}>
            <ToggleGroupItem value='info'>
                <FontAwesomeIcon icon={faCircleInfo} />
                Toggle Button
            </ToggleGroupItem>
            <ToggleGroupItem value='star'>
                <FontAwesomeIcon icon={faStar} />
                Toggle Button
            </ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Toggle group items with icons on the right.
 */
export const WithIconsRight: Story = {
    render: (args) => (
        <ToggleGroup {...args}>
            <ToggleGroupItem value='info'>
                Toggle Button
                <FontAwesomeIcon icon={faCircleInfo} />
            </ToggleGroupItem>
            <ToggleGroupItem value='star'>
                Toggle Button
                <FontAwesomeIcon icon={faStar} />
            </ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * All items in the group are disabled.
 */
export const Disabled: Story = {
    render: (args) => (
        <ToggleGroup {...args} disabled>
            <ToggleGroupItem value='a'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='b'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Interaction test: clicking items in a single-select group selects one at a time.
 */
export const Interaction: Story = {
    render: (args) => (
        <ToggleGroup {...args}>
            <ToggleGroupItem value='a'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='b'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
    play: async ({ canvasElement, step }) => {
        const canvas = within(canvasElement);
        const [itemA, itemB] = await canvas.findAllByRole('radio');

        await step('Click Toggle Button A to select it', async () => {
            await userEvent.click(itemA);
        });

        await step('Click Toggle Button B — Toggle Button A deselects', async () => {
            await userEvent.click(itemB);
        });
    },
};
