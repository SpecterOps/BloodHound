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
import { AccessibleIcon } from './AccessibleIcon';

const meta = {
    title: 'Components/AccessibleIcon',
    component: AccessibleIcon,
    tags: ['autodocs'],
    args: {
        children: <AppIcon.Info size={24} />,
        hideTooltip: false,
        label: 'Information',
    },
    argTypes: {
        children: {
            control: false,
            description:
                'Exactly one SVG icon element, such as an AppIcon or Font Awesome icon. AccessibleIcon clones this element to apply its accessibility attributes.',
            table: { category: 'Content' },
        },
        'aria-label': {
            control: 'text',
            description:
                'Accessible name applied to the SVG. Use it with label when assistive technology needs a more descriptive name than the visible tooltip. When label is omitted, this value is also used by the tooltip.',
            table: { category: 'Accessibility' },
        },
        label: {
            control: 'text',
            description:
                'Visible tooltip content and, unless aria-label overrides it, the SVG accessible name. Describe the meaning conveyed by the icon rather than its appearance.',
            table: { category: 'Content' },
        },
        'aria-hidden': {
            control: false,
            description:
                'Marks the SVG as decorative and suppresses its tooltip. IconButton sets this automatically because the button supplies the accessible name.',
            table: {
                category: 'Accessibility',
                defaultValue: { summary: 'false' },
            },
        },
        className: {
            control: 'text',
            description:
                'Classes applied to the inline-flex wrapper. Use the child icon props for its size and the wrapper or child className for inherited color.',
            table: { category: 'Appearance' },
        },
        hideTooltip: {
            control: 'boolean',
            description:
                'Suppresses the visual tooltip while preserving the SVG accessible name. Use this when nearby visible text already communicates the same information.',
            table: {
                category: 'Accessibility',
                defaultValue: { summary: 'false' },
            },
        },
    },
    parameters: {
        docs: {
            description: {
                component: `AccessibleIcon gives a meaningful, non-interactive SVG an accessible name and an optional explanatory tooltip. It accepts exactly one icon child, clones it with the appropriate ARIA attribute, and wraps it in an inline layout container.

### When to use it

Use AccessibleIcon when an icon communicates status, context, or other information but does not perform an action. Examples include a lock that means an item cannot be moved, a warning status, or an informational symbol beside a field label.

Do not add click handlers to AccessibleIcon. Use <code>IconButton</code> when activating the icon performs an action. Use the icon element directly, or set <code>aria-hidden</code>, when the icon is decorative and nearby text already conveys its meaning.

### Label and accessible name

Provide either <code>label</code> or <code>aria-label</code>.

- <code>label</code> supplies both the visible tooltip and the SVG accessible name. This is preferred when both audiences need the same message.
- <code>aria-label</code> supplies the SVG accessible name. When it is the only label provided, the tooltip uses it too.
- Provide both when the concise visible tooltip and the assistive-technology description should differ.

Describe the meaning conveyed by the icon, not its visual appearance. Prefer <code>"Query saved"</code> over <code>"Checkmark icon"</code> and <code>"Tier Zero cannot be moved"</code> over <code>"Lock icon"</code>.

~~~tsx
<AccessibleIcon label='Tier Zero cannot be moved'>
    <AppIcon.Lock />
</AccessibleIcon>
~~~

Use <code>aria-label</code> as an override when the accessible name should provide additional context:

~~~tsx
<AccessibleIcon label='Learn more' aria-label='Learn more about saved queries'>
    <AppIcon.Info />
</AccessibleIcon>
~~~

### Tooltip behavior

The tooltip is displayed on pointer hover by default. AccessibleIcon remains non-interactive and does not add a tab stop. Its accessible name is always applied to the SVG, so the icon's meaning does not depend on the tooltip being open.

Use <code>hideTooltip</code> when equivalent visible text is already adjacent to the icon. This removes only the visual tooltip; it does not remove the SVG accessible name.

~~~tsx
<span className='inline-flex items-center gap-2'>
    <AccessibleIcon label='Query saved' hideTooltip>
        <AppIcon.Info />
    </AccessibleIcon>
    Query saved
</span>
~~~

### Decorative icons

Set <code>aria-hidden</code> when the icon adds no information. This removes the SVG from the accessibility tree and suppresses its tooltip.

~~~tsx
<AccessibleIcon aria-hidden label='Decorative information icon'>
    <AppIcon.Info />
</AccessibleIcon>
~~~

When AccessibleIcon is nested in IconButton, IconButton marks the child as decorative and suppresses its tooltip automatically. The button owns the accessible name and displays the only tooltip, preventing duplicate announcements.

~~~tsx
<IconButton aria-label='Show query information'>
    <AccessibleIcon label='Query information'>
        <AppIcon.Info />
    </AccessibleIcon>
</IconButton>
~~~

### Sizing and styling

Set size and icon-specific styling on the child. Use AccessibleIcon's <code>className</code> for wrapper layout, spacing, and inherited text color.

~~~tsx
<AccessibleIcon label='Important information' className='mr-2 text-status-warning-main'>
    <AppIcon.Info size={20} />
</AccessibleIcon>
~~~`,
            },
        },
    },
} satisfies Meta<typeof AccessibleIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Hover the icon to see its label. The same text is applied to the SVG as its accessible name.',
            },
        },
    },
};

export const DistinctAccessibleLabel: Story = {
    args: {
        label: 'Learn more',
        'aria-label': 'Learn more about saved queries',
    },
    parameters: {
        docs: {
            description: {
                story: 'Use separate values when the visible tooltip should be concise but the accessible name needs additional context.',
            },
        },
    },
};

export const TooltipHidden: Story = {
    args: { hideTooltip: true },
    parameters: {
        docs: {
            description: {
                story: 'Hide the tooltip when equivalent information is already visible nearby. The SVG retains its accessible name.',
            },
        },
    },
};

export const Decorative: Story = {
    args: {
        'aria-hidden': true,
        label: 'Decorative information icon',
    },
    parameters: {
        docs: {
            description: {
                story: 'Decorative icons are hidden from assistive technology and do not display a tooltip.',
            },
        },
    },
};
