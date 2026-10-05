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
import LoadingOverlay from './LoadingOverlay';

const meta: Meta<typeof LoadingOverlay> = {
    title: 'Components/LoadingOverlay',
    component: LoadingOverlay,
    tags: ['autodocs'],
    args: { loading: true },
    render: (args) => (
        <div className='relative h-48 max-w-xl rounded border border-neutral-3 p-6'>
            <h2 className='font-bold mb-4'>Collection history</h2>
            <p>The overlay covers its positioned parent while data is loading.</p>
            <LoadingOverlay {...args} />
        </div>
    ),
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Loading: Story = {};
export const Idle: Story = { args: { loading: false } };
