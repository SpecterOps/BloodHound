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
import { Controls, Description, Primary, Stories, Subtitle, Title } from '@storybook/blocks';
import { useArgs } from '@storybook/preview-api';
import type { Meta, StoryObj } from '@storybook/react';
import { useRef, useState } from 'react';
import { Slider } from './Slider';

// Custom autodocs page: identical to Storybook's default, except the Stories list excludes the
// primary story so the interactive Playground is not repeated below its top preview.
const DocsPage = () => (
    <>
        <Title />
        <Subtitle />
        <Description />
        <Primary />
        <Controls />
        <Stories includePrimary={false} />
    </>
);

/**
 * An input where the user selects a value from within a given range.
 */
const meta: Meta<typeof Slider> = {
    title: 'Components/Slider',
    component: Slider,
    tags: ['autodocs'],
    argTypes: {
        value: {
            control: 'number',
            description: 'The controlled value of the slider.',
            table: { type: { summary: 'number' } },
        },
        defaultValue: {
            control: 'number',
            description: 'The initial value of an uncontrolled slider.',
            table: { type: { summary: 'number' } },
        },
        onValueChange: {
            action: 'value changed',
            description: 'Called when the slider value changes.',
            table: { type: { summary: '(value: number, eventDetails) => void' } },
        },
        min: {
            control: 'number',
            description: 'The minimum allowed value of the slider.',
            table: { type: { summary: 'number' }, defaultValue: { summary: '0' } },
        },
        max: {
            control: 'number',
            description: 'The maximum allowed value of the slider.',
            table: { type: { summary: 'number' }, defaultValue: { summary: '100' } },
        },
        disabled: {
            control: 'boolean',
            description: 'Whether the slider should ignore user interaction.',
            table: { type: { summary: 'boolean' }, defaultValue: { summary: 'false' } },
        },
        step: { table: { disable: true } },
        thumbAriaLabel: { table: { disable: true } },
    },
    args: {
        min: 0,
        max: 100,
        thumbAriaLabel: 'Value',
    },
    parameters: {
        layout: 'centered',
        docs: {
            page: DocsPage,
        },
    },
} satisfies Meta<typeof Slider>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * The interactive example. Dragging the thumb updates the `value` prop shown in the Controls
 * panel, and editing that control moves the thumb. This is the Primary story at the top of the
 * Docs page and the only story bound to the `value` prop.
 */
export const Playground: Story = {
    args: {
        value: 50,
    },
    render: (args) => {
        const [, updateArgs] = useArgs();
        // Local state keeps dragging responsive; we mirror it back to the `value` arg so the
        // Controls panel reflects the current value.
        const [value, setValue] = useState(args.value ?? 50);

        // Move the thumb when the `value` control is edited from the Controls panel.
        const previousArgValue = useRef(args.value);
        if (args.value !== previousArgValue.current) {
            previousArgValue.current = args.value;
            if (typeof args.value === 'number') setValue(args.value);
        }

        return (
            <div className='w-64'>
                <Slider
                    {...args}
                    value={value}
                    onValueChange={(nextValue, eventDetails) => {
                        setValue(nextValue);
                        updateArgs({ value: nextValue });
                        args.onValueChange?.(nextValue, eventDetails);
                    }}
                />
            </div>
        );
    },
};

/**
 * A plain, uncontrolled example. It starts at 50 and can be dragged, but it is not bound to the
 * `value` prop, so interacting with it never changes the Controls panel value.
 */
export const Default: Story = {
    args: {
        defaultValue: 50,
    },
    render: (args) => (
        <div className='w-64'>
            <Slider {...args} value={undefined} />
        </div>
    ),
};

export const Disabled: Story = {
    args: {
        disabled: true,
        defaultValue: 50,
    },
    parameters: {
        docs: {
            description: {
                story: 'A non-interactive slider. The thumb cannot be moved and interaction styles are suppressed.',
            },
        },
    },
    render: (args) => (
        <div className='w-64'>
            <Slider {...args} />
        </div>
    ),
};
