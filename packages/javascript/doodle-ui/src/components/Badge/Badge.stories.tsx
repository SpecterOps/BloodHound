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
import { faChevronDown, faChevronUp, faEyeSlash, faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { Badge } from './Badge';

const iconMap = {
    // noIcon: undefined,
    chevronUp: <FontAwesomeIcon icon={faChevronUp} />,
    chevronDown: <FontAwesomeIcon icon={faChevronDown} />,
    plus: <FontAwesomeIcon icon={faPlus} />,
    minus: <FontAwesomeIcon icon={faMinus} />,
    eyeSlash: <FontAwesomeIcon icon={faEyeSlash} />,
} as const;

type IconName = keyof typeof iconMap;

type StoryArgs = React.ComponentProps<typeof Badge> & {
    iconName?: IconName;
};

const iconSourceMap: Record<IconName, string | undefined> = {
    // noIcon: undefined,
    chevronUp: '<FontAwesomeIcon icon={faChevronUp} />',
    chevronDown: '<FontAwesomeIcon icon={faChevronDown} />',
    plus: '<FontAwesomeIcon icon={faPlus} />',
    minus: '<FontAwesomeIcon icon={faMinus} />',
    eyeSlash: '<FontAwesomeIcon icon={faEyeSlash} />',
};

const meta = {
    title: 'Components/Badge',
    component: Badge,
    tags: ['autodocs'],
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component: `
Badges are used to communicate **status, counts, or small pieces of metadata** in a compact, visually distinct way.

### When to use
- To display **status** (e.g., active, inactive, warning)
- To show **counts or indicators** (e.g., notifications, totals)
- To highlight **categorization or labels** (e.g., tags, types)

Badges are best used sparingly and should remain **short, scannable, and non-interactive**.
                `,
            },
            canvas: {
                sourceState: 'shown',
            },
        },
    },
    argTypes: {
        label: {
            control: 'text',
            description: 'Text displayed inside the badge.',
            table: {
                type: { summary: 'string' },
            },
        },
        color: {
            control: 'select',
            options: ['indeterminate', 'primary', 'secondary', 'grey', 'red', 'orange', 'green', 'blue', 'purple'],
            description: 'Color of the badge.',
            table: {
                type: {
                    summary:
                        "'indeterminate' | 'primary' | 'secondary' | 'grey' | 'red' | 'orange' | 'green' | 'blue' | 'purple'",
                },
                defaultValue: { summary: 'indeterminate' },
            },
        },
        iconPosition: {
            control: 'select',
            options: ['left', 'right'],
            description: 'Position of the icon relative to the label. Only applies when an icon is provided.',
            table: {
                type: { summary: "'left' | 'right'" },
                defaultValue: { summary: 'left' },
            },
        },

        // Storybook-only control
        iconName: {
            name: 'icon',
            control: 'select',
            options: ['chevronUp', 'chevronDown', 'plus', 'minus', 'eyeSlash'],
            labels: {
                // noIcon: 'None',
                chevronUp: 'Chevron Up',
                chevronDown: 'Chevron Down',
                plus: 'Plus',
                minus: 'Minus',
                eyeSlash: 'Eye Slash',
            },
            description: 'Select an icon to display.',
            table: {
                type: { summary: 'ReactNode' },
                defaultValue: { summary: 'undefined' },
            },
        },

        // hide real prop from controls table
        icon: {
            control: false,
            table: {
                disable: true,
            },
        },

        labelClassName: {
            control: 'text',
            description: 'Additional class names applied to the label.',
            table: {
                type: { summary: 'string' },
                defaultValue: { summary: 'undefined' },
            },
        },
        iconClassName: {
            control: 'text',
            description: 'Additional class names applied to the icon wrapper.',
            table: {
                type: { summary: 'string' },
                defaultValue: { summary: 'undefined' },
            },
        },
    },
    args: {
        label: 'Badge',
        color: 'indeterminate',
        iconPosition: 'left',
    },
} satisfies Meta<StoryArgs>;

export default meta;

type Story = StoryObj<StoryArgs>;

export const Default: Story = {
    tags: ['!dev'],
    render: ({ iconName, ...args }) => <Badge {...args} icon={iconName ? iconMap[iconName] : undefined} />,
    parameters: {
        docs: {
            source: {
                transform: (_src: string, context: { args: StoryArgs }) => {
                    const { iconName, ...args } = context.args;
                    const iconCode = iconName ? iconSourceMap[iconName] : undefined;

                    const props = [
                        `label="${args.label}"`,
                        args.color && `color="${args.color}"`,
                        iconCode && args.iconPosition && `iconPosition="${args.iconPosition}"`,
                        iconCode && `icon={${iconCode}}`,
                    ]
                        .filter(Boolean)
                        .join(' ');

                    return `<Badge ${props} />`;
                },
            },
        },
    },
};

export const WithoutIcon: Story = {
    args: {
        label: 'No Icon',
        iconName: undefined,
    },
    argTypes: {
        iconName: {
            control: false,
            table: { disable: true },
        },
        iconPosition: {
            control: false,
            table: { disable: true },
        },
    },
    render: ({ iconName, iconPosition, ...args }) => <Badge {...args} />,
};

export const WithIcon: Story = {
    args: {
        label: 'With Icon',
        iconName: 'chevronUp',
        iconPosition: 'left',
    },
    render: ({ iconName, ...args }) => <Badge {...args} icon={iconName ? iconMap[iconName] : undefined} />,
    parameters: {
        docs: {
            source: {
                transform: (_src: string, context: { args: StoryArgs }) => {
                    const { iconName, ...args } = context.args;
                    const iconCode = iconName ? iconSourceMap[iconName] : undefined;

                    const props = [
                        `label="${args.label}"`,
                        args.color && `color="${args.color}"`,
                        iconCode && args.iconPosition && `iconPosition="${args.iconPosition}"`,
                        iconCode && `icon={${iconCode}}`,
                    ]
                        .filter(Boolean)
                        .join(' ');

                    return `<Badge ${props} />`;
                },
            },
        },
    },
};
