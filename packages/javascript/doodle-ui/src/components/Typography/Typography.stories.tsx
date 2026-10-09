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
import { Card, CardContent, CardFooter, CardHeader } from '../Card';
import { Input } from '../Input';
import { Link } from '../Link';
import { Typography } from './Typography';
import { tagOptions, Variant, variantMapping } from './utils';

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
 * The default story. Use the controls to change the `variant` and, optionally, the rendered
 * `component`. Each `variant` applies a preset type style and a matching default HTML element; the
 * `component` prop overrides that element without changing the visual style.
 *
 * ```tsx
 * <Typography variant='h1'>Attack path exposure</Typography>
 * <Typography variant='h1' component='span'>Same style, rendered as a span</Typography>
 * ```
 */
export const TypographyComponent: Story = {
    name: 'Typography',
    args: {
        variant: 'h1',
        children: 'The quick brown fox jumps over the lazy dog',
    },
    render: (args) => <Typography {...args} />,
};

/**
 * The complete type scale. Each row pairs the `variant` token and the HTML element it renders by
 * default with a specimen of the style, making it easy to compare size, weight, and spacing.
 */
export const Variants: Story = {
    render: () => (
        <div className='w-[40rem] max-w-full space-y-6'>
            {(Object.keys(variantMapping) as Variant[]).map((variant) => (
                <div className='grid grid-cols-[10rem_1fr] items-baseline gap-6' key={variant}>
                    <Typography variant='caption' className='text-text-muted'>
                        {variant} · &lt;{variantMapping[variant]}&gt;
                    </Typography>
                    <Typography variant={variant}>The quick brown fox jumps over the lazy dog</Typography>
                </div>
            ))}
        </div>
    ),
};

/**
 * How the variants read together in flowing, real-world content. Text wraps naturally and preserves
 * its size and line height across multiple lines and constrained widths.
 */
export const MultilineAndLongText: Story = {
    render: () => (
        <article className='w-96 max-w-full space-y-3'>
            <Typography variant='h2'>Review attack path exposure</Typography>
            <Typography variant='subtitle1'>A summary of the highest-risk paths in your environment</Typography>
            <Typography variant='body1'>
                Attack paths chain together the misconfigurations and privileges that let a principal move toward your
                most critical assets. Prioritize the paths that expose Tier Zero, then work outward.
            </Typography>
            <Typography variant='body2'>
                Longer explanatory copy stays comfortable to read across several lines, preserving spacing and line
                height even when the available width is reduced by browser zoom or a narrow container.
            </Typography>
            <Typography variant='caption'>Last calculated 2 minutes ago</Typography>
        </article>
    ),
};

/**
 * A compact, information-dense layout that combines `subtitle2`, `body2`, and `caption` to present
 * many rows of related data, such as a list of groups or assets.
 */
export const DenseList: Story = {
    render: () => (
        <div className='w-[36rem] max-w-full divide-y divide-neutral-400 rounded border border-neutral-400'>
            {['Domain Admins', 'Enterprise Admins', 'Remote Desktop Users', 'Backup Operators'].map((name, index) => (
                <div className='grid grid-cols-[1fr_auto] gap-4 p-3' key={name}>
                    <div>
                        <Typography variant='subtitle2'>{name}</Typography>
                        <Typography variant='body2'>Active Directory group · Tier {index % 2}</Typography>
                    </div>
                    <Typography variant='caption'>{12 + index * 7} members</Typography>
                </div>
            ))}
        </div>
    ),
};

/**
 * Typography composed with other components inside a Card: a heading and timestamp, supporting body
 * copy with an inline Link, and a caption used as an accessible label for an Input.
 */
export const CardComposition: Story = {
    render: () => (
        <Card className='w-[28rem] max-w-full'>
            <CardHeader>
                <Typography variant='h3'>Review attack-path exposure</Typography>
                <Typography variant='subtitle2'>Updated a few seconds ago</Typography>
            </CardHeader>
            <CardContent className='space-y-3'>
                <Typography variant='body1'>
                    Prioritize the paths that give principals control of your most critical assets.
                </Typography>
                <Typography variant='body2'>
                    Learn how exposure is calculated in the{' '}
                    <Link href='https://bloodhound.specterops.io/' className='inline-flex'>
                        BloodHound documentation
                    </Link>
                    .
                </Typography>
                <div>
                    <Typography id='typography-card-filter-label' variant='caption' component='span'>
                        Filter by asset name
                    </Typography>
                    <Input
                        id='typography-card-filter'
                        aria-labelledby='typography-card-filter-label'
                        className='mt-1'
                        placeholder='Search assets'
                    />
                </div>
            </CardContent>
            <CardFooter className='gap-2'>
                <Button size='small'>Review paths</Button>
                <Button size='small' variant='secondary'>
                    Dismiss
                </Button>
            </CardFooter>
        </Card>
    ),
};
