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
import { Button } from 'doodle-ui';
import { ConditionalTooltip } from './ConditionalTooltip';

const meta = {
    title: 'Components/ConditionalTooltip',
    component: ConditionalTooltip,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    args: {
        children: <Button>Hover or focus me</Button>,
        condition: true,
        tooltip: 'Tooltip shown when the condition is true',
    },
    argTypes: {
        children: { control: false },
    },
} satisfies Meta<typeof ConditionalTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithTooltip: Story = {};

export const WithoutTooltip: Story = {
    args: { condition: false },
};

export const RightSide: Story = {
    args: { side: 'right' },
};
