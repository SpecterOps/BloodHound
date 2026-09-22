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

import { faEllipsisVertical } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { Icon } from '../Icon';
import { IconButton } from './IconButton';

const meta = {
    title: 'Components/IconButton',
    component: IconButton,
    tags: ['autodocs'],
    args: {
        variant: 'default',
        disabled: false,
        hideTooltip: false,
        size: 16,
        'aria-label': 'More options',
        children: <FontAwesomeIcon icon={faEllipsisVertical} />,
    },
    argTypes: {
        variant: {
            options: ['default', 'primary', 'secondary'],
            control: 'select',
            description: 'Sets the visual emphasis of the button.',
            table: {
                category: 'Appearance',
                defaultValue: {
                    summary: 'default',
                },
            },
        },
        children: {
            control: false,
            description: 'A single AppIcon, Font Awesome icon, or Icon element.',
            table: {
                category: 'Content',
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
        tooltip: {
            description:
                'Visible tooltip content. Defaults to aria-label; override it when the tooltip needs additional context.',
            control: 'text',
            table: {
                category: 'Accessibility',
            },
        },
        hideTooltip: {
            description: 'Suppresses the tooltip without removing the button accessible label.',
            control: 'boolean',
            table: {
                category: 'Accessibility',
                defaultValue: {
                    summary: 'false',
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
        disabled: {
            control: 'boolean',
            description: 'Disables button interaction. The tooltip remains available unless hideTooltip is true.',
            table: {
                category: 'State',
                defaultValue: {
                    summary: 'false',
                },
            },
        },
        className: {
            control: 'text',
            description: 'Classes applied to the button element.',
            table: {
                category: 'Appearance',
            },
        },
    },
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `Use IconButton for an action represented only by an icon. If the action also has visible text, use Button or TextButton instead.

### Basic usage

Provide one icon child and an \`aria-label\` that describes the action. The button uses the label as its accessible name and default tooltip. Describe what happens when the button is activated—not the icon's appearance.

\`\`\`tsx
<IconButton aria-label='Show filter options'>
    <AppIcon.FilterOutline />
</IconButton>
\`\`\`

Prefer \`"Show filter options"\` over \`"Filter icon"\`. IconButton marks its child as decorative because the button itself owns the accessible name.

### Supported icons

IconButton accepts one \`AppIcon\`, Font Awesome icon, or \`Icon\` element. Passing an icon directly is the simplest option. If an \`Icon\` is nested, IconButton suppresses its label and tooltip to prevent duplicate announcements and tooltips.

\`\`\`tsx
<IconButton aria-label='More options'>
    <FontAwesomeIcon icon={faEllipsisVertical} />
</IconButton>

<IconButton aria-label='Show information'>
    <Icon label='Information'>
        <AppIcon.Info />
    </Icon>
</IconButton>
\`\`\`

### Tooltip behavior

The tooltip defaults to \`aria-label\`. Use \`tooltip\` when the visible message needs extra context while the concise accessible name should remain unchanged. Use \`hideTooltip\` only when the action is already identified by nearby visible text.

\`\`\`tsx
<IconButton
    aria-label='Show filter options'
    tooltip='Filters are unavailable while data loads'
    disabled
>
    <AppIcon.FilterOutline />
</IconButton>
\`\`\`

Disabled buttons retain their tooltip so users can still learn what the action does or why it is unavailable. Set \`hideTooltip\` to suppress it explicitly.

### Variants

- \`default\` is transparent and works well for standalone controls and toolbars.
- \`primary\` gives the action the strongest emphasis.
- \`secondary\` provides emphasis without competing with a primary action.

Use the \`variant\` prop rather than applying Button variant classes through \`className\`.

### Sizing and color

The \`size\` prop controls the icon's width and height in pixels. It defaults to \`16\`. IconButton adds \`8px\` of padding on each side, so \`size={16}\` produces a \`32px × 32px\` button.

Icons inherit the button's text color through \`currentColor\`. Prefer a variant for standard colors and use \`className\` only for a deliberate local override.

\`\`\`tsx
<IconButton aria-label='More options' size={20} className='text-primary'>
    <FontAwesomeIcon icon={faEllipsisVertical} />
</IconButton>
\`\`\``,
            },
        },
    },
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

const renderVariantStory =
    (variant: 'default' | 'primary' | 'secondary'): Story['render'] =>
    ({ children, ...iconButtonProps }) => (
        <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <IconButton {...iconButtonProps}>{children}</IconButton>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center justify-center gap-8'>
                <div className='flex flex-col items-center gap-4'>
                    <IconButton aria-label='More options' size={18} variant={variant}>
                        <Icon label='More options'>
                            <FontAwesomeIcon icon={faEllipsisVertical} />
                        </Icon>
                    </IconButton>
                    Enabled
                </div>
                <div className='flex flex-col items-center gap-4'>
                    <IconButton aria-label='Show filter options' disabled size={18} variant={variant}>
                        <AppIcon.FilterOutline />
                    </IconButton>
                    Disabled
                </div>
            </div>
        </>
    );

export const Default: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Use the transparent default variant for icon actions in toolbars or other compact control groups.',
            },
        },
    },
    render: renderVariantStory('default'),
};

export const Primary: Story = {
    args: {
        variant: 'primary',
    },
    parameters: {
        docs: {
            description: {
                story: 'Use the primary variant when the icon action is the main action in its immediate context.',
            },
        },
    },
    render: renderVariantStory('primary'),
};

export const Secondary: Story = {
    args: {
        variant: 'secondary',
    },
    parameters: {
        docs: {
            description: {
                story: 'Use the secondary variant for a supporting icon action alongside a primary action.',
            },
        },
    },
    render: renderVariantStory('secondary'),
};

export const TooltipHidden: Story = {
    args: {
        hideTooltip: true,
    },
    parameters: {
        docs: {
            description: {
                story: 'Hide the tooltip only when nearby visible content already identifies the action. The button keeps its accessible name.',
            },
        },
    },
};
