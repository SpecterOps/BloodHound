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
import { Button } from '../Button';
import { LoadingSpinner } from './LoadingSpinner';

const meta = {
    title: 'Components/LoadingSpinner',
    component: LoadingSpinner,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    argTypes: {
        className: {
            control: 'text',
            description: 'Override size and color with Tailwind classes, for example `size-6 text-secondary`.',
        },
    },
} satisfies Meta<typeof LoadingSpinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Sizes: Story = {
    render: () => (
        <div className='flex items-center gap-6'>
            <LoadingSpinner className='size-5' aria-label='Loading, small' />
            <LoadingSpinner aria-label='Loading, default' />
            <LoadingSpinner className='size-16' aria-label='Loading, large' />
        </div>
    ),
};

export const Colors: Story = {
    render: () => (
        <div className='flex items-center gap-6 text-status-success-main'>
            <LoadingSpinner />
            <LoadingSpinner className='text-secondary' />
            <LoadingSpinner className='text-inherit' />
        </div>
    ),
};

export const InButton: Story = {
    parameters: {
        docs: {
            description: {
                story:
                    'Use a small spinner with `text-inherit` to match the button color. ' +
                    'Disable the button while its action is pending and set `aria-busy` to indicate loading.',
            },
        },
    },
    render: () => (
        <div className='flex items-center gap-4'>
            <Button disabled aria-busy='true'>
                Saving…
                <LoadingSpinner className='size-4 text-inherit' aria-label='Saving' />
            </Button>
            <Button variant='secondary' disabled aria-busy='true'>
                <LoadingSpinner className='size-4 text-inherit' aria-label='Loading' />
                Loading…
            </Button>
        </div>
    ),
};
