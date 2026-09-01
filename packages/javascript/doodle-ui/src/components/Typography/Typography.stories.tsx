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
import { Card, CardContent, CardHeader } from '../Card';
import { Typography } from './Typography';
import { tagOptions, variantMapping } from './utils';

const meta = {
    title: 'Components/Typography',
    component: Typography,
    parameters: {
        layout: 'centered',
    },
    tags: ['autodocs'],
    argTypes: {
        variant: {
            type: 'string',
            options: Object.keys(variantMapping),
            control: 'select',
            description: 'Applies default styling based on heading/tag level:',
        },
        component: {
            options: tagOptions,
            control: 'select',
            description: 'Applies selected html tag. Overrides default tag from variant mapping.',
        },
    },
    args: {},
} satisfies Meta<typeof Typography>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Use the default body style for primary explanatory content in a product view.
 * Leave `component` unset when a semantic paragraph is appropriate.
 */
export const TypographyComponent: Story = {
    name: 'Typography',
    args: {
        variant: 'body1',
        children:
            'Review the attack paths that give principals control of your most critical assets before prioritizing remediation.',
    },
};

/**
 * Use a heading variant to establish hierarchy within a view, so people can
 * scan related content before reading its details.
 */
export const Variants: Story = {
    args: {
        variant: 'h2',
        children: 'Attack paths requiring review',
    },
};

/**
 * Use body text for long-form explanations that must remain readable when the
 * available content width is constrained.
 */
export const MultilineAndLongText: Story = {
    args: {
        variant: 'body1',
        children:
            'This explanation remains readable when a narrow side panel or browser zoom reduces the available width. It wraps naturally without changing the intended visual hierarchy.',
    },
    render: (args) => (
        <div className='w-80 max-w-full'>
            <Typography {...args} />
        </div>
    ),
};

/**
 * Pair a group name with supporting context and a compact count when users
 * need to compare many related records without losing the important detail.
 */
export const DenseList: Story = {
    args: {
        variant: 'subtitle2',
    },
    render: (args) => (
        <div className='w-[36rem] max-w-full divide-y divide-neutral-400 rounded border border-neutral-400'>
            {['Domain Admins', 'Enterprise Admins', 'Remote Desktop Users', 'Backup Operators'].map((name, index) => (
                <div className='grid grid-cols-[1fr_auto] gap-4 p-3' key={name}>
                    <div>
                        <Typography variant={args.variant}>{name}</Typography>
                        <Typography variant='body2'>Active Directory group · Tier {index % 2}</Typography>
                    </div>
                    <Typography variant='caption'>{12 + index * 7} members</Typography>
                </div>
            ))}
        </div>
    ),
};

/**
 * Combine a card title, status, and explanation to communicate a single
 * focused decision without making the card compete with surrounding content.
 */
export const CardComposition: Story = {
    args: {
        variant: 'h3',
    },
    render: (args) => (
        <Card className='w-[28rem] max-w-full'>
            <CardHeader className='space-y-1'>
                <Typography variant={args.variant}>Review attack-path exposure</Typography>
                <Typography variant='subtitle2'>Updated a few seconds ago</Typography>
            </CardHeader>
            <CardContent>
                <Typography variant='body1'>
                    Prioritize the paths that give principals control of your most critical assets.
                </Typography>
            </CardContent>
        </Card>
    ),
};
