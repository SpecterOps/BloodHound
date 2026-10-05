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
import { ManagedDatePicker } from './ManagedDatePicker';

const meta: Meta<typeof ManagedDatePicker> = {
    title: 'Components/ManagedDatePicker',
    component: ManagedDatePicker,
    tags: ['autodocs'],
    args: {
        hint: 'Start date',
        value: '2026-06-15',
        fromDate: new Date(2026, 0, 1),
        toDate: new Date(2026, 11, 31),
        onDateChange: fn(),
    },
    argTypes: { fromDate: { control: false }, toDate: { control: false } },
    decorators: [
        (Story) => (
            <div className='flex flex-col max-w-xs gap-1'>
                <Story />
            </div>
        ),
    ],
    render: function Render(args) {
        const [value, setValue] = useState(args.value);
        useEffect(() => setValue(args.value), [args.value]);
        return (
            <ManagedDatePicker
                {...args}
                value={value}
                onDateChange={(value, isValid) => {
                    args.onDateChange(value, isValid);
                    if (isValid) setValue(value);
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const SelectedDate: Story = {};
export const Empty: Story = { args: { value: undefined } };
export const ValidationError: Story = { args: { validationError: 'Start date must be before the end date.' } };
export const InvalidInput: Story = {
    play: async ({ canvasElement, args }) => {
        const canvas = within(canvasElement);
        const input = canvas.getByRole('textbox', { name: 'Start date' });
        await userEvent.clear(input);
        await userEvent.type(input, '2026-02-30');
        await userEvent.tab();
        await expect(await canvas.findByText('Input is not a valid date.')).toBeVisible();
        await expect(args.onDateChange).toHaveBeenLastCalledWith('2026-02-30', false);
    },
};
