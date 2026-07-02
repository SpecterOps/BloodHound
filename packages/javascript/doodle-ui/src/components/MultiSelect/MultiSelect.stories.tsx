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
import { useState } from 'react';
import type { MultiSelectProps } from './MultiSelect';
import { MultiSelect, MultiSelectOptionRow, MultiSelectTrigger } from './MultiSelect';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/MultiSelect',
    component: MultiSelect,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof MultiSelect>;

export default meta;
type Story = StoryObj<typeof meta>;

const itemOptions = [
    { value: 'option-a', label: 'Option A' },
    { value: 'option-b', label: 'Option B' },
    { value: 'option-c', label: 'Option C' },
];

const renderMultiSelect = (args: MultiSelectProps) => (
    <div className='w-60'>
        <MultiSelect {...args} />
    </div>
);

export const Default: Story = {
    args: {
        options: itemOptions,
        value: [],
        onValueChange: () => {},
        placeholder: 'All Zones',
    },
    render: (args) => {
        const [value, setValue] = useState<string[]>(args.value ?? []);

        return (
            <div className='w-60'>
                <MultiSelect {...args} value={value} onValueChange={setValue} />
            </div>
        );
    },
};

export const Trigger: Story = {
    args: { options: [], value: [], onValueChange: () => {} },
    render: () => (
        <div className='w-60'>
            <MultiSelectTrigger>Placeholder</MultiSelectTrigger>
        </div>
    ),
};

export const TriggerOpen: Story = {
    args: { options: [], value: [], onValueChange: () => {} },
    render: () => (
        <div className='w-60'>
            <MultiSelectTrigger open>Placeholder</MultiSelectTrigger>
        </div>
    ),
};

export const TriggerError: Story = {
    args: { options: [], value: [], onValueChange: () => {} },
    render: () => (
        <div className='w-60'>
            <MultiSelectTrigger aria-invalid='true'>Placeholder</MultiSelectTrigger>
        </div>
    ),
};

export const TriggerDisabled: Story = {
    args: { options: [], value: [], onValueChange: () => {} },
    render: () => (
        <div className='w-60'>
            <MultiSelectTrigger disabled>Placeholder</MultiSelectTrigger>
        </div>
    ),
};

export const OptionRows: Story = {
    args: { options: [], value: [], onValueChange: () => {} },
    render: () => (
        <div className='w-60 border rounded-md'>
            <MultiSelectOptionRow
                option={{ value: 'option-a', label: 'Option A' }}
                checked={false}
                onSelect={() => {}}
            />
            <MultiSelectOptionRow
                option={{ value: 'option-b', label: 'Option B' }}
                checked={true}
                onSelect={() => {}}
            />
            <MultiSelectOptionRow
                option={{
                    value: 'option-c',
                    label: 'A very long label that should be truncated when it overflows the container',
                }}
                checked={false}
                onSelect={() => {}}
            />
            <MultiSelectOptionRow
                option={{ value: 'option-d', label: 'Option D (disabled)', disabled: true }}
                checked={false}
                onSelect={() => {}}
            />
        </div>
    ),
};

export const Placeholder: Story = {
    args: {
        options: itemOptions,
        value: [],
        onValueChange: () => {},
        placeholder: 'Select',
    },
    render: renderMultiSelect,
};

export const OneSelected: Story = {
    args: {
        options: itemOptions,
        value: ['option-a'],
        onValueChange: () => {},
        placeholder: 'Select',
    },
    render: renderMultiSelect,
};

export const MultipleSelected: Story = {
    args: {
        options: itemOptions,
        value: ['option-a', 'option-b'],
        onValueChange: () => {},
        placeholder: 'Select',
    },
    render: renderMultiSelect,
};

export const Disabled: Story = {
    args: {
        options: itemOptions,
        value: [],
        onValueChange: () => {},
        placeholder: 'All Zones',
        disabled: true,
    },
    render: renderMultiSelect,
};

export const WithDisabledOption: Story = {
    args: {
        options: [...itemOptions, { value: 'option-d', label: 'Option D', disabled: true }],
        value: [],
        onValueChange: () => {},
        placeholder: 'All Zones',
    },
    render: (args) => {
        const [value, setValue] = useState<string[]>(args.value ?? []);

        return (
            <div className='w-60'>
                <MultiSelect {...args} value={value} onValueChange={setValue} />
            </div>
        );
    },
};
