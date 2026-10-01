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
import { CopyToClipboardButton } from './CopyToClipboardButton';

const meta = {
    title: 'Components/CopyToClipboardButton',
    component: CopyToClipboardButton,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    args: { value: 'Example value to copy' },
    render: (args) => (
        <div className='rounded border border-neutral-3 px-4 py-2'>
            <div className='group relative inline-flex items-center'>
                <CopyToClipboardButton {...args} />
            </div>
        </div>
    ),
} satisfies Meta<typeof CopyToClipboardButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ArrayValue: Story = {
    args: { value: ['First value', 'Second value'] },
};
