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
import DropdownTriggerContents from './DropdownTriggerContents';

const meta: Meta<typeof DropdownTriggerContents> = {
    title: 'Components/DropdownTriggerContents',
    component: DropdownTriggerContents,
    tags: ['autodocs'],
    args: { selectedText: 'Production environment', open: false },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'The standalone trigger presentation used by dropdowns and static reports. Use DropdownSelector for an interactive menu.',
            },
        },
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};
export const Open: Story = { args: { open: true } };
export const ReadOnly: Story = { args: { readOnly: true } };
export const Disabled: Story = { args: { disabled: true } };
export const Truncated: Story = {
    args: { className: 'max-w-48', selectedText: 'A very long production environment name' },
};
