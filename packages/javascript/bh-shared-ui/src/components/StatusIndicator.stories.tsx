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
import { StatusIndicator } from './StatusIndicator';

const meta: Meta<typeof StatusIndicator> = {
    title: 'Components/StatusIndicator',
    component: StatusIndicator,
    tags: ['autodocs'],
    args: { status: 'good', label: 'Completed', pulse: false },
    argTypes: {
        status: {
            control: 'select',
            options: ['good', 'bad', 'pending'],
            description: 'Determines the status dot color.',
            table: { type: { summary: "'good' | 'bad' | 'pending'" } },
        },
        label: {
            control: 'text',
            description: 'Text displayed beside the status dot.',
            table: { type: { summary: 'string' } },
        },
        pulse: {
            control: 'boolean',
            description: 'Animates the status dot while work is in progress.',
            table: { type: { summary: 'boolean' } },
        },
    },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'Displays a status dot with an optional text label. Choose `good`, `bad`, or `pending`, and enable `pulse` to indicate ongoing activity.',
            },
        },
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Good: Story = {};
export const Bad: Story = { args: { status: 'bad', label: 'Failed' } };
export const Pending: Story = { args: { status: 'pending', label: 'Processing', pulse: true } };
export const WithoutLabel: Story = { args: { label: '' } };
