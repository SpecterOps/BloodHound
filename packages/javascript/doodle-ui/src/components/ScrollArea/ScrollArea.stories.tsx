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
import { ScrollArea } from './ScrollArea';

const TAGS = Array.from({ length: 50 }).map(
    (_, i, a) =>
        `v1.2.0-beta.${a.length - i} Lorem ipsum, dolor sit amet consectetur adipisicing elit. Libero quo voluptas quia optio perspiciatis voluptatibus, blanditiis necessitatibus`
);

const meta = {
    title: 'Components/ScrollArea',
    component: ScrollArea,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        type: {
            control: { type: 'select' },
            options: ['auto', 'hover', 'always', 'scroll'],
            description: 'Determines when and how the scrollbars display.',
            table: { defaultValue: { summary: 'auto' } },
        },
        scrollbarWidth: {
            description: 'Scrollbar width in px',
        },
        thumbHeight: {
            description: 'Thumbnail height in px',
        },
        scrollbarColor: { control: 'color' },
        thumbColor: { control: 'color' },
    },

    render: (args) => (
        <ScrollArea {...args} className='h-72 w-48 rounded-md border'>
            {TAGS.map((tag) => (
                <div
                    className='relative mt-2.5 border-t pt-2.5 text-[13px] leading-[18px] w-full text-nowrap'
                    key={tag}>
                    {tag}
                </div>
            ))}
        </ScrollArea>
    ),
} satisfies Meta<typeof ScrollArea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Auto: Story = {
    args: {
        type: 'auto',
    },
    parameters: {
        docs: {
            description: {
                story: 'Scrollbars are visible when content is overflowing on the corresponding orientation.',
            },
        },
    },
};

export const Hover: Story = {
    args: {
        type: 'hover',
    },
    parameters: {
        docs: {
            description: {
                story: 'Scrollbars are visible when the user is scrolling along its corresponding orientation and when the user is hovering over the scroll area.',
            },
        },
    },
};

export const Always: Story = {
    args: {
        type: 'always',
    },
    parameters: {
        docs: {
            description: {
                story: 'Scrollbars are always visible regardless of whether the content is overflowing.',
            },
        },
    },
};

export const Scroll: Story = {
    args: {
        type: 'scroll',
    },
    parameters: {
        docs: {
            description: {
                story: 'Scrollbars are visible when the user is scrolling along its corresponding orientation.',
            },
        },
    },
};

export const OverrideStyles: Story = {
    args: {
        type: 'auto',
    },
    parameters: {
        docs: {
            description: {
                story: `<div>
                <p>usage: </p>
                <div><code><ScrollArea</code></div>
                <div><code>scrollbarWidth={14}</code></div>
                <div><code>thumbColor='lime'</code></div>
                <div><code>scrollbarColor='hsla(209, 82%, 64%, 0.35)'></code></div>
                </div>`,
            },
        },
    },
    render: (args) => (
        <div className='w-full flex items-center justify-center'>
            <ScrollArea
                {...args}
                className='h-72 w-48 rounded-none  border'
                scrollbarWidth={14}
                thumbColor='lime'
                scrollbarColor='hsla(209, 82%, 64%, 0.35)'>
                {TAGS.map((tag) => (
                    <div
                        className='relative mt-2.5 border-t pt-2.5 text-[13px] leading-[18px] w-full text-nowrap'
                        key={tag}>
                        {tag}
                    </div>
                ))}
            </ScrollArea>
        </div>
    ),
};
