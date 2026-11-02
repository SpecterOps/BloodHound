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
    args: { variant: 'inline' },
    argTypes: {
        variant: {
            control: 'select',
            options: ['inline', 'overlay'],
            description:
                'Use inline for a standalone spinner or overlay to cover the nearest positioned ancestor with a dimmed backdrop.',
        },
        spinnerClassName: {
            control: 'text',
            if: { arg: 'variant', eq: 'overlay' },
            description:
                'Only available with `variant="overlay"`. ' +
                'Styles the centered icon without changing the backdrop, for example `size-6 text-secondary`. ' +
                'For inline spinners, use `className` instead.',
        },
        className: {
            control: 'text',
            description:
                'Styles the root span. ' +
                'For inline spinners, use this for size, color, and layout (for example `size-6 text-secondary`). ' +
                'For overlays, use it for the backdrop and positioning; use `spinnerClassName` for the icon.',
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

export const Overlay: Story = {
    args: { variant: 'overlay' },
    parameters: {
        docs: {
            description: {
                story: 'Place the spinner inside a positioned container, such as `relative`. The overlay dims and blocks pointer interaction with its contents. Use `spinnerClassName` for the icon size and color, and `className` for the backdrop. The backdrop continues to fill the container regardless of icon size.',
            },
        },
    },
    render: (args) => (
        <div className='relative flex h-64 w-96 flex-col items-start gap-4 rounded border border-neutral-300 p-6'>
            <p>Content remains visible while loading.</p>
            <Button>Covered action</Button>
            <LoadingSpinner {...args} />
        </div>
    ),
};

export const CustomizedOverlay: Story = {
    ...Overlay,
    args: {
        variant: 'overlay',
        className: 'bg-black/10',
        spinnerClassName: 'size-6 text-green-700',
    },
    parameters: {
        docs: {
            description: {
                story: '`className="bg-black/10"` lightens the full backdrop. `spinnerClassName="size-6 text-green-700"` makes only the centered icon smaller and changes its color. For a regular inline spinner, keep using `className="size-6 text-green-700"`.',
            },
        },
    },
};
