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
import { BaseColumnHeader } from './ColumnHeaders';

const meta: Meta<typeof BaseColumnHeader> = {
    title: 'Components/ColumnHeaders/BaseColumnHeader',
    component: BaseColumnHeader,
    tags: ['autodocs'],
    args: { title: 'Environment', textAlign: 'left' },
    decorators: [
        (Story) => (
            <div className='max-w-md border border-neutral-3 p-4'>
                <Story />
            </div>
        ),
    ],
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Left: Story = {};
export const Center: Story = { args: { textAlign: 'center' } };
export const Right: Story = { args: { textAlign: 'right' } };
