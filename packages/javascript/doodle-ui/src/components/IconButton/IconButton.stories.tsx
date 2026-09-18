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

import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { IconButton } from './IconButton';

const meta = {
    title: 'Components/IconButton',
    component: IconButton,
    tags: ['autodocs'],
    args: {
        variant: 'primary',
        disabled: false,
        size: 16,
        'aria-label': 'Show information',
        children: <AppIcon.Info />,
    },
    argTypes: {
        variant: {
            options: ['primary', 'secondary'],
            control: 'select',
        },
        children: {
            control: false,
            table: {
                disable: true,
            },
        },
        'aria-label': {
            description: 'Required accessible name describing the action performed by the icon button.',
            control: 'text',
            table: {
                category: 'Accessibility',
                type: {
                    summary: 'string',
                },
            },
        },
        size: {
            description:
                'Sets the icon width and height in pixels. Defaults to 16. The square button resizes with the icon.',
            control: {
                type: 'number',
                min: 8,
                step: 1,
            },
            table: {
                category: 'Appearance',
                defaultValue: {
                    summary: '16',
                },
                type: {
                    summary: 'number',
                },
            },
        },
    },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `Use IconButton for an action represented by an icon without a visible text label.

### Sizing

The \`size\` prop sets the icon's width and height in pixels. The button automatically resizes around the icon while preserving its square shape and consistent padding.

\`\`\`tsx
<IconButton aria-label='Open filters' size={16}>
    <AppIcon.FilterOutline />
</IconButton>
\`\`\`

The default icon size is \`16px\`. By default, the button adds \`8px\` of padding on every side, so its total width and height are the icon size plus \`16px\`. For example, \`size={16}\` produces a \`32px × 32px\` button.

Use the \`size\` prop to resize the icon and button together. Use \`className\` only when you need to override spacing or other presentation.

### Variants

The \`primary\` and \`secondary\` variants use the same visual styles as their Button counterparts while preserving IconButton's icon-only shape and sizing. IconButton defaults to \`primary\`, just like Button. Select the variant directly instead of adding \`ButtonVariants\` through \`className\`.

\`\`\`tsx
<IconButton variant='primary' aria-label='Show filter options'>
    <AppIcon.FilterOutline />
</IconButton>
\`\`\`

### Color

Icons inherit the button's computed text color through \`currentColor\`. Prefer a variant when the icon should follow a standard button color. Use \`className\` for a local color override.

\`\`\`tsx
<IconButton aria-label='Show information' className='text-primary'>
    <AppIcon.Info />
</IconButton>
\`\`\`

### Accessible label

- Every IconButton requires an \`aria-label\` describing the action performed by the button.
- IconButton forwards that label to its internal \`Icon\`, so the rendered icon retains the same accessible label and always displays it in a tooltip.

Do not use the icon's name as the label when it does not describe the action. For example, prefer \`"Show filter options"\` over \`"Filter icon"\`.

Consumers only need to provide the label on IconButton; IconButton supplies it to the internal Icon for its accessible name and tooltip.`,
            },
        },
    },
    render: ({ ...iconButtonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <IconButton {...iconButtonProps}>
                    <AppIcon.Info />
                </IconButton>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <div className='flex flex-col items-center gap-4'>
                    <IconButton aria-label='Delete item' size={18} variant='primary'>
                        <FontAwesomeIcon icon={faTrash} />
                    </IconButton>
                    Primary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButton aria-label='Show filter options' size={24} variant='secondary'>
                        <AppIcon.FilterOutline />
                    </IconButton>
                    Secondary
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButton aria-label='Show filter options' disabled size={24} variant='primary'>
                        <AppIcon.FilterOutline />
                    </IconButton>
                    Disabled
                </div>
            </div>
        </>
    ),
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
