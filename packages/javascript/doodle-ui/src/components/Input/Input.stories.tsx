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

import { Button } from '../Button';
import { Label } from '../Label/Label';
import { Input } from './Input';

/**
 * Displays a form input field or a component that looks like an input field.
 */
const meta = {
    title: 'Components/Input',
    component: Input,
    tags: ['autodocs'],
    argTypes: { variant: { options: ['outlined', 'underlined'], control: 'select' } },
    args: {
        className: 'w-96',
        type: 'email',
        placeholder: 'email@example.com',
        disabled: false,
        variant: 'outlined',
    },
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Input>;

export default meta;

type Story = StoryObj<typeof meta>;

/**
 * The default form of the input field.
 */
export const Default: Story = {};

export const Required: Story = {
    args: {
        label: 'Email address',
        required: true,
        variant: 'outlined',
    },
};

export const RequiredWithoutLabel: Story = {
    args: {
        required: true,
    },
};

export const Error: Story = {
    args: {
        defaultValue: 'not-an-email',
        error: true,
        errorMessage: 'Enter a valid email address.',
        label: 'Email address',
        required: true,
        variant: 'outlined',
    },
};

/**
 * Use the `disabled` prop to make the input non-interactive and appears faded,
 * indicating that input is not currently accepted.
 */
export const Disabled: Story = {
    args: { disabled: true, variant: 'outlined' },
};

export const WithLabel: Story = {
    args: {
        id: 'email',
        label: 'Email',
    },
};

/**
 * Use the helperText prop to provide additional instructions
 * or information to users.
 */
export const WithHelperText: Story = {
    args: {
        id: 'email-2',
        label: 'Email',
        helperText: 'Enter your email address.',
        required: true,
        variant: 'outlined',
    },
};

/**
 * Use the `Button` component to indicate that the input field can be submitted
 * or used to trigger an action.
 */
export const WithButton: Story = {
    render: (args) => (
        <div className='flex items-center space-x-2'>
            <Input {...args} />
            <Button className='rounded-md'>Subscribe</Button>
        </div>
    ),
};

export const WithFile: Story = {
    args: {
        type: 'file',
        placeholder: 'No file selected.',
    },
    render: (args) => (
        <div className='grid w-full max-w-sm items-center gap-1.5'>
            <Label htmlFor='picture'>Upload Image</Label>
            <Input id='picture' type='file' {...args} variant={'outlined'} />
        </div>
    ),
};

export const TimeInput: Story = {
    args: {
        intent: 'time',
        type: 'time',
    },
    render: (args) => (
        <div className='grid items-center gap-1.5'>
            <Label htmlFor='time'>Time</Label>
            <Input {...args} id='time' />
        </div>
    ),
};
