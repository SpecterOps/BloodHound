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
import { useArgs } from '@storybook/preview-api';
import type { Meta, StoryObj } from '@storybook/react';
import { useRef, useState } from 'react';
import { Slider } from './Slider';

/**
 * An input where the user selects a value from within a given range.
 */
const meta = {
    title: 'Components/Slider',
    component: Slider,
    tags: ['autodocs'],
    argTypes: {
        value: {
            control: 'number',
            description: 'The current value of the slider (position of the thumb).',
            table: { type: { summary: 'number' } },
        },
        defaultValue: { table: { disable: true } },
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
        orientation: { table: { disable: true } },
        thumbAriaLabel: { table: { disable: true } },
        trackClassName: { table: { disable: true } },
        indicatorClassName: { table: { disable: true } },
        thumbClassName: { table: { disable: true } },
        thumbDotClassName: { table: { disable: true } },
    },
    args: {
        value: 50,
        min: 0,
        max: 100,
    },
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Slider>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    args: {},
    render: (args) => {
        const [, updateArgs] = useArgs();
        // Local state is the source of truth so dragging stays responsive; the value is
        // mirrored back to the `value` arg so it shows in the Controls panel.
        const [value, setValue] = useState(args.value ?? 50);

        // Sync the thumb when the `value` control is edited from the Controls panel
        // (adjust state during render so drags don't wait on the async arg round-trip).
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
