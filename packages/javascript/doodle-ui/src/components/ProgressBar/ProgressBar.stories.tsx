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
import { ProgressBar } from './ProgressBar';

/** Displays the completion percentage of a named task. Use the theme toolbar to preview light and dark colors. */
const meta = {
    title: 'Components/ProgressBar',
    component: ProgressBar,
    tags: ['autodocs'],
    parameters: {
        layout: 'padded',
    },
    argTypes: {
        value: {
            control: { type: 'range', min: 0, max: 100, step: 1 },
            description: 'The completion percentage, from 0 to 100. Use null for indeterminate progress.',
        },
        'aria-label': {
            control: 'text',
            description: 'Required accessible name for the process being tracked.',
        },
        className: {
            control: 'text',
            description: 'Additional CSS classes. Overrides the default w-full width.',
        },
    },
    args: {
        value: 50,
        'aria-label': 'File upload',
    },
} satisfies Meta<typeof ProgressBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
    args: { value: 0 },
};

export const Complete: Story = {
    args: { value: 100 },
};

export const CustomWidth: Story = {
    args: { className: 'w-64', value: 75 },
};

/** Use when the task has no measurable completion percentage. */
export const Indeterminate: Story = {
    argTypes: { value: { control: false } },
    args: { value: null, 'aria-label': 'Loading data' },
};
