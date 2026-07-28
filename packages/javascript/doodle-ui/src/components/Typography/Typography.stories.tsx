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
 * Usage:
 *
 * ```javascript
 * <Typography variant='h1' component='optional tag name'>Lorem ipsum dolor sit amet.</Typography>
 * ```
 */

export const TypographyComponent: Story = {
    name: 'Typography',
    args: {
        variant: 'h1',
    },
    render: (args) => {
        const componentString = args.component || variantMapping[args.variant || 'body1'];
        const codeString = `<${componentString} variant='${args.variant}'${args.component ? " component='" + args.component + "'" : ''}>Lorem ipsum dolor sit amet.</${componentString}>`;

        return (
            <>
                <div className='mb-8'>
                    <Typography variant={args.variant} {...args}>
                        Lorem ipsum dolor sit amet.
                    </Typography>
                </div>
                <p>Code Output:</p>
                <code className='bg-sky-400/10 p-4 block rounded-lg'>{codeString}</code>
            </>
        );
    },
};

/**
 * #### Mapping
 *  > h1: 'h1'<br>
 *  > h2: 'h2'<br>
 *  > h3: 'h3'<br>
 *  > h4: 'h4'<br>
 *  > h5: 'h5'<br>
 *  > h6: 'h6'<br>
 *  > subtitle1: 'h6'<br>
 *  > subtitle2: 'h6'<br>
 *  > body1: 'p'<br>
 *  > body2: 'p'<br>
 *  > caption: 'span'<br>
 */

export const Variants: Story = {
    args: {},
    render: () => {
        return (
            <div className='w-[32rem] max-w-full space-y-4'>
                {Object.keys(variantMapping).map((variant) => (
                    <Typography variant={variant as Variant} key={variant}>
                        {variant}. The quick brown fox jumps over the lazy dog.
                    </Typography>
                ))}
            </div>
        );
    },
};

export const MultilineAndLongText: Story = {
    args: {},
    render: () => (
        <div className='w-80 max-w-full space-y-4'>
            <Typography variant='h2'>
                A compact heading that wraps cleanly across multiple lines without clipping
            </Typography>
            <Typography variant='body1'>
                Body1 remains comfortable for longer explanatory content. This deliberately long string verifies
                wrapping, line boxes, descenders, and spacing when the viewport or browser zoom reduces available width.
            </Typography>
            <Typography variant='body2'>
                Body2 supports dense product interfaces while preserving its existing size and line height across
                several lines of supporting information.
            </Typography>
            <Typography variant='caption'>
                Caption: extraordinarily-long-unbroken-identifier-for-overflow-validation.example
            </Typography>
        </div>
    ),
};

export const DenseList: Story = {
    args: {},
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

export const CardComposition: Story = {
    args: {},
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
