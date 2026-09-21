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
    },
    argTypes: {
        children: {
            control: false,
            description: 'An AppIcon supplied by the consuming application.',
        },
        'aria-label': {
            control: 'text',
            description: 'Required accessible label describing the icon.',
        },
    },
    parameters: {
        docs: {
            description: {
                component: `Icon renders an \`AppIcon\` supplied by the consuming application without adding button semantics. BloodHound consumers should use \`AppIcon\` from \`bh-shared-ui\`.

Every Icon requires an \`aria-label\`. Icon applies the label to its child so the rendered SVG has an accessible name, and always displays the same label in a tooltip. The label should describe what the icon communicates, rather than its visual shape.

\`\`\`tsx
<Icon aria-label='Information about saved queries'>
    <AppIcon.Info />
</Icon>
\`\`\`

Use Icon for visual or informational icons. For an icon that performs an action, render the \`AppIcon\` inside \`IconButton\` instead. \`IconButton\` applies its required \`aria-label\` to the button and uses it for the tooltip.`,
            },
        },
    },
} satisfies Meta<typeof Icon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Filter: Story = {
    args: {
        'aria-label': 'Filter options',
        children: <AppIcon.FilterOutline size={24} />,
    },
};
