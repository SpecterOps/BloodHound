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
import { Toggle } from './toggle';

const meta = {
    title: 'Components/Toggle',
    component: Toggle,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        variant: {
            options: ['default', 'outline'],
            control: 'select',
            description: 'Visual style of the toggle button.',
        },
        size: {
            options: ['default', 'sm', 'lg'],
            control: 'select',
            description: 'Size of the toggle button.',
        },
        disabled: {
            control: 'boolean',
            description: 'Disables interaction with the toggle.',
        },
        pressed: {
            control: 'boolean',
            description: 'Controlled pressed state.',
        },
        defaultPressed: {
            control: 'boolean',
            description: 'Initial pressed state (uncontrolled).',
        },
    },
    args: {
        children: 'Toggle',
        variant: 'default',
        size: 'default',
        disabled: false,
    },
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * The default form of the toggle with transparent background.
 */
export const Default: Story = {
    args: {
        children: 'Toggle',
    },
};

/**
 * The outline variant adds a border around the toggle button.
 */
export const Outline: Story = {
    args: {
        variant: 'outline',
        children: 'Toggle',
    },
};

/**
 * A smaller toggle suitable for compact layouts.
 */
export const Small: Story = {
    args: {
        size: 'sm',
        children: 'Toggle',
    },
};

/**
 * A larger toggle for prominent actions.
 */
export const Large: Story = {
    args: {
        size: 'lg',
        children: 'Toggle',
    },
};

/**
 * The toggle rendered in a pressed (active) state by default.
 */
export const DefaultPressed: Story = {
    args: {
        defaultPressed: true,
        children: 'Toggle',
    },
};

/**
 * A disabled toggle cannot be interacted with.
 */
export const Disabled: Story = {
    args: {
        disabled: true,
        children: 'Toggle',
    },
};

/**
 * Interaction test: clicking the toggle switches between pressed and unpressed states.
 */
export const Interaction: Story = {
    args: {
        children: 'Toggle',
    },
    play: async ({ canvasElement, step }) => {
        const canvas = within(canvasElement);
        const toggle = await canvas.findByRole('button');

        await step('Toggle starts unpressed', async () => {
            expect(toggle).toHaveAttribute('data-state', 'off');
        });

        await step('Click to press the toggle', async () => {
            await userEvent.click(toggle);
            expect(toggle).toHaveAttribute('data-state', 'on');
        });

        await step('Click again to unpress the toggle', async () => {
            await userEvent.click(toggle);
            expect(toggle).toHaveAttribute('data-state', 'off');
        });
    },
};
