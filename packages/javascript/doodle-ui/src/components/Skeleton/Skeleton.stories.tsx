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
import { Skeleton } from './Skeleton';

/**
 * Use to show a placeholder while content is loading.
 */
const meta = {
    title: 'Components/Skeleton',
    component: Skeleton,
    tags: ['autodocs'],
    argTypes: {},
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Skeleton>;

export default meta;

type Story = StoryObj<typeof Skeleton>;

/**
 * The default form of the skeleton.
 */
export const Default: Story = {
    render: (args) => (
        <div className='flex items-center space-x-4'>
            <Skeleton {...args} className='h-12 w-12 rounded-full' />
            <div className='space-y-2'>
                <Skeleton {...args} className='h-4 w-[250px]' />
                <Skeleton {...args} className='h-4 w-[200px]' />
            </div>
        </div>
    ),
};
