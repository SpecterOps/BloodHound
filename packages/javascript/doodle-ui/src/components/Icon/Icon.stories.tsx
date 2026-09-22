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
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { Icon } from './Icon';

const meta = {
    title: 'Components/Icon',
    component: Icon,
    tags: ['autodocs'],
    args: {
        'aria-label': 'Information',
        children: <AppIcon.Info size={24} />,
        hideTooltip: false,
    },
    argTypes: {
        children: {
            control: false,
            description: 'A single SVG icon element, such as an AppIcon or Font Awesome icon.',
            table: {
                category: 'Content',
            },
        },
        'aria-label': {
            control: 'text',
            description: 'Required accessible label describing the icon.',
            table: {
                category: 'Accessibility',
            },
        },
        'aria-hidden': {
            control: false,
            description:
                'Removes a decorative icon from the accessibility tree and suppresses its tooltip. IconButton sets this automatically for its child icon.',
            table: {
                category: 'Accessibility',
                defaultValue: {
                    summary: 'false',
                },
            },
        },
        className: {
            control: 'text',
            description: 'Classes applied to the inline wrapper around the icon.',
            table: {
                category: 'Appearance',
            },
        },
        hideTooltip: {
            control: 'boolean',
            description: 'Suppresses the tooltip while preserving the icon accessible label.',
            table: {
                category: 'Accessibility',
                defaultValue: {
                    summary: 'false',
                },
            },
        },
    },
    parameters: {
        docs: {
            description: {
                component: `Use Icon for a non-interactive icon that communicates information. Icon adds accessibility and tooltip behavior to a single SVG child without adding button semantics. BloodHound consumers should normally supply an \`AppIcon\` from \`bh-shared-ui\`.

### Accessible label

Every Icon requires an \`aria-label\`. Icon forwards that label to its SVG child and uses it as the default tooltip. Describe the information conveyed by the icon, not its visual appearance. For example, prefer \`"Query saved"\` over \`"Checkmark icon"\`.

\`\`\`tsx
<Icon aria-label='Information about saved queries'>
    <AppIcon.Info />
</Icon>
\`\`\`

### Tooltip behavior

The tooltip is displayed by default. Use \`hideTooltip\` only when equivalent text is already visible nearby. Hiding the tooltip does not remove the icon's accessible label.

\`\`\`tsx
<Icon aria-label='Information already shown in nearby text' hideTooltip>
    <AppIcon.Info />
</Icon>
\`\`\`

### Decorative and interactive icons

Set \`aria-hidden\` when an icon is purely decorative. This removes the SVG from the accessibility tree and suppresses its tooltip. When Icon is nested in IconButton, IconButton does this automatically because the button owns the accessible label and tooltip.

Do not attach click behavior to Icon. Use \`IconButton\` when the icon performs an action.`,
            },
        },
    },
} satisfies Meta<typeof Icon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    parameters: {
        docs: {
            description: {
                story: 'The accessible label is applied to the SVG and displayed in a tooltip.',
            },
        },
    },
};

export const TooltipHidden: Story = {
    args: {
        hideTooltip: true,
    },
    parameters: {
        docs: {
            description: {
                story: 'Hide the tooltip when equivalent information is already visible. The SVG retains its accessible label.',
            },
        },
    },
};
