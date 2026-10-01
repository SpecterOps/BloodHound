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
import { fn } from '@storybook/test';
import { JOB_STATUS_MAP } from '../../utils';
import { StatusSelect } from './StatusSelect';

const meta: Meta<typeof StatusSelect> = {
    title: 'Components/SelectMenus/StatusSelect',
    component: StatusSelect,
    tags: ['autodocs'],
    args: { status: undefined, onSelect: fn() },
    argTypes: {
        status: { control: 'number' },
        statusOptions: { control: 'multi-select', options: Object.values(JOB_STATUS_MAP) },
    },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'Uses local job status labels without an API request. Selection callbacks return strings; this example converts status IDs to numbers and clears the controlled value when None is selected. statusOptions filters by label.',
            },
        },
    },
    render: function Render(args) {
        const [, updateArgs] = useArgs();
        return (
            <StatusSelect
                {...args}
                onSelect={(value) => {
                    args.onSelect(value);
                    updateArgs({ status: value === '-none-' ? undefined : Number(value) });
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Selected: Story = { args: { status: 0 } };

export const Filtered: Story = {
    args: { statusOptions: ['Complete', 'Failed', 'Partially Completed'] },
};
