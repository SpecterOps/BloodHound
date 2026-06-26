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
import { Checkbox, CheckboxWithLabel } from './Checkbox';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Checkbox',
    component: Checkbox,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        size: {
            control: 'select',
            options: ['lg', 'md', 'sm'],
            description: 'Size of the Checkbox.',
        },
        disabled: {
            control: 'boolean',
            description: 'Disables interaction with the Checkbox.',
        },
    },
    args: { size: 'md', disabled: false },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    render: (args) => <Checkbox {...args} />,
};

export const WithLabel: Story = {
    render: (args) => <CheckboxWithLabel {...args} label='Label' defaultChecked={false} />,
};

/**
 *Checkboxes with a label.
 */
export const LabelExamples: Story = {
    render: ({ size }) => (
        <div className='flex flex-col items-start gap-4'>
            <CheckboxWithLabel size={size} label='Unchecked' checked={false} />
            <CheckboxWithLabel size={size} label='Checked' checked />
            <CheckboxWithLabel size={size} label='Indeterminate' checked='indeterminate' />
            <CheckboxWithLabel size={size} label='Error' error checked />
            <CheckboxWithLabel size={size} label='Disabled' disabled checked />
        </div>
    ),
};

/**
 *Icons only Checkboxes.
 */
export const CheckboxWithoutLabel: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox aria-label='Unchecked checkbox' checked={false} {...args} />
            <Checkbox aria-label='Checked checkbox' checked={true} {...args} />
            <Checkbox aria-label='Indeterminate checkbox' checked='indeterminate' {...args} />
        </div>
    ),
};

/**
 * Disabled Checkbox cannot be interacted with.
 */
export const Disabled: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox id='checkbox-disabled-unchecked' checked={false} {...args} disabled />
            <Checkbox id='checkbox-disabled-checked' checked {...args} disabled />
            <Checkbox id='checkbox-disabled-indeterminate' checked='indeterminate' {...args} disabled />
        </div>
    ),
};

/**
 * Error Checkboxes.
 */
export const Error: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox id='checkbox-error-unchecked' checked={false} {...args} aria-invalid />
            <Checkbox id='checkbox-error-checked' checked={true} {...args} aria-invalid />
            <Checkbox id='checkbox-error-indeterminate' checked='indeterminate' {...args} aria-invalid />
        </div>
    ),
};
