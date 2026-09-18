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
        const shortText = 'Heading';
        const longText =
            'Lorem ipsum dolor sit amet, consectetur adipisicing elit. Quos blanditiis tenetur unde suscipit, quam beatae rerum inventore consectetur, neque doloribus, cupiditate numquam dignissimos laborum fugiat deleniti? Eum quasi quidem quibusdam.';

        const headings = Object.keys(variantMapping).map((variant, i) => {
            return (
                <Typography variant={variant as Variant} className='mb-4' key={variant}>
                    {variant}. {i < 6 ? shortText : longText}
                </Typography>
            );
        });

        return <div>{headings}</div>;
    },
};
