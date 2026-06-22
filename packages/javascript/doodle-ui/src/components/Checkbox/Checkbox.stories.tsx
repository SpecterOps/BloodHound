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
import { Label } from '../Label';
import { Checkbox } from './Checkbox';

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
        },
    },
    args: { size: 'md' },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    render: (args) => <Checkbox {...args} />,
};

export const IconOnly: Story = {
    render: (args) => (
        <div className='flex items-center gap-4'>
            <Checkbox aria-label='Unchecked checkbox' checked={false} {...args} />
            <Checkbox aria-label='Checked checkbox' checked={true} {...args} />
            <Checkbox aria-label='Indeterminate checkbox' checked='indeterminate' {...args} />
        </div>
    ),
};

export const Unchecked: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox id='checkbox-unchecked' checked={false} {...args} />
            <Label htmlFor='checkbox-unchecked'>Label</Label>
        </div>
    ),
};

export const Checked: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox id='checkbox-checked' checked={true} {...args} />
            <Label htmlFor='checkbox-checked'>Label</Label>
        </div>
    ),
};

export const Indeterminate: Story = {
    render: (args) => (
        <div className='flex items-center gap-2'>
            <Checkbox id='checkbox-indeterminate' checked='indeterminate' {...args} />
            <Label htmlFor='checkbox-indeterminate'>Label</Label>
        </div>
    ),
};

export const Labeled: Story = {
    render: () => {
        return (
            <div className='flex justify-center flex-row items-center'>
                <Checkbox id='test-id' />
                <Label htmlFor='test-id' className='pl-2'>
                    Testing Label
                </Label>
            </div>
        );
    },
};

export const Disabled: Story = {
    render: (args) => (
        <div className='flex flex-col gap-4'>
            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-disabled-unchecked' disabled checked={false} {...args} />
                <Label htmlFor='checkbox-disabled-unchecked'>Unchecked disabled</Label>
            </div>

            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-disabled-checked' disabled checked {...args} />
                <Label htmlFor='checkbox-disabled-checked'>Checked disabled</Label>
            </div>

            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-disabled-indeterminate' disabled checked='indeterminate' {...args} />
                <Label htmlFor='checkbox-disabled-indeterminate'>Indeterminate disabled</Label>
            </div>
        </div>
    ),
};

export const Error: Story = {
    render: (args) => (
        <div className='flex flex-col gap-4'>
            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-error-unchecked' aria-invalid checked={false} {...args} />
                <Label htmlFor='checkbox-error-unchecked'>Unchecked error</Label>
            </div>

            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-error-checked' aria-invalid checked={true} {...args} />
                <Label htmlFor='checkbox-error-checked'>Checked error</Label>
            </div>

            <div className='flex items-center gap-2'>
                <Checkbox id='checkbox-error-indeterminate' aria-invalid checked='indeterminate' {...args} />
                <Label htmlFor='checkbox-error-indeterminate'>Indeterminate error</Label>
            </div>
        </div>
    ),
};
