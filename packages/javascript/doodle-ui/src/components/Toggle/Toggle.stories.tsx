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
import { Toggle } from './Toggle';

const meta = {
    title: 'Components/Toggle',
    component: Toggle,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        size: {
            options: ['sm', 'lg'],
            control: 'select',
            description: 'Size of the toggle button.',
        },
        disabled: {
            control: 'boolean',
            description: 'Disables interaction with the toggle.',
        },
    },
    args: {
        children: 'Toggle Button',
        size: 'sm',
        disabled: false,
    },
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Default/Small size toggle button with outline.
 */
export const Default: Story = {
    args: {
        children: 'Toggle Button',
    },
};

/**
 * Large size toggle button.
 */
export const Large: Story = {
    args: {
        size: 'lg',
        children: 'Toggle Button',
    },
};

/** Toggle button with an Icon on the Left. */
export const IconLeft: Story = {
    render: (args) => (
        <Toggle {...args}>
            <FontAwesomeIcon icon={faCircleInfo} />
            Toggle Button
        </Toggle>
    ),
};

/** Toggle button with an Icon on the Right. */
export const IconRight: Story = {
    render: (args) => (
        <Toggle {...args}>
            Toggle Button
            <FontAwesomeIcon icon={faStar} />
        </Toggle>
    ),
};

/**
 * The toggle button rendered in a pressed (active) state by default.
 */
export const DefaultPressed: Story = {
    args: {
        defaultPressed: true,
        children: 'Toggle Button',
    },
};

/**
 * Disabled toggle button cannot be interacted with.
 */
export const Disabled: Story = {
    args: {
        disabled: true,
        children: 'Toggle Button',
    },
};
