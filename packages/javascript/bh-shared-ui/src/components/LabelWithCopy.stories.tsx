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
import LabelWithCopy from './LabelWithCopy';

const meta: Meta<typeof LabelWithCopy> = {
    title: 'Components/LabelWithCopy',
    component: LabelWithCopy,
    tags: ['autodocs'],
    args: { label: 'service.account@example.com', valueToCopy: 'service.account@example.com' },
    parameters: { layout: 'centered' },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const AlwaysVisible: Story = {};
export const HoverOnly: Story = { args: { hoverOnly: true } };
export const DifferentCopyValue: Story = {
    args: { label: 'Directory object ID', valueToCopy: 'S-1-5-21-1000-1000-1000-500' },
};
export const NumericValue: Story = { args: { label: 'Job 1234', valueToCopy: 1234 } };
