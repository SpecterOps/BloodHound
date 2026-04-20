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
            table: { category: 'props' },
        },
        disabled: {
            control: 'boolean',
            description: 'Disables all items in the group.',
            table: { category: 'props' },
        },
    },
    args: {
        type: 'single',
        size: 'sm',
        disabled: false,
    },
} satisfies Meta<typeof ToggleGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Default/Small toggle group, use the controls panel above to change size and disabled.
 */
export const Default: Story = {
    render: (args) => (
        <ToggleGroup {...args}>
            <ToggleGroupItem value='graph'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='table'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Large size toggle group.
 */
export const Large: Story = {
    render: (args) => (
        <ToggleGroup {...args} size='lg'>
            <ToggleGroupItem value='a'>Toggle Button</ToggleGroupItem>
            <ToggleGroupItem value='b'>Toggle Button</ToggleGroupItem>
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
            <ToggleGroupItem value='graph'>Graph</ToggleGroupItem>
            <ToggleGroupItem value='table'>Table</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Multiple selection toggle group.
 */
export const Multiple: Story = {
    render: ({ size, disabled }) => (
        <ToggleGroup type='multiple' size={size} disabled={disabled}>
            <ToggleGroupItem value='a'>Toggle A</ToggleGroupItem>
            <ToggleGroupItem value='b'>Toggle B</ToggleGroupItem>
        </ToggleGroup>
    ),
};

/**
 * Three toggle button group.
 */
export const Three: Story = {
    render: (args) => (
        <ToggleGroup {...args}>
            <ToggleGroupItem value='a'>Toggle A</ToggleGroupItem>
            <ToggleGroupItem value='b'>Toggle B</ToggleGroupItem>
            <ToggleGroupItem value='c'>Toggle C</ToggleGroupItem>
        </ToggleGroup>
    ),
};
