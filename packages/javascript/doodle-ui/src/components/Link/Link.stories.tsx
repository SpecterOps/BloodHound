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
import { Link } from './Link';

const meta = {
    title: 'Components/Link',
    component: Link,
    tags: ['autodocs'],
    args: {
        href: 'https://example.com',
        children: 'External link',
        variant: 'styled',
    },
    argTypes: {
        variant: {
            control: 'select',
            options: ['styled', 'unstyled'],
        },
    },
    parameters: {
        docs: {
            description: {
                component: `
### When to use

- **Use this for external links**

This component:
- Always opens links in a new tab
- Always applies \`noopener\`
- Applies \`noreferrer\` by default
- Supports \`allowReferrer\` to omit \`noreferrer\`
- Displays an external link icon
- Announces "(opens in a new tab)" to screen readers
        `,
            },
        },
    },
} satisfies Meta<typeof Link>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Styled: Story = {};

export const Unstyled: Story = {
    args: {
        variant: 'unstyled',
        children: 'Unstyled external link',
    },
};

export const InText: Story = {
    name: 'In text',
    render: (args) => (
        <p className='text-sm text-neutral-dark-1 dark:text-white'>
            Read the <Link {...args} /> for more details.
        </p>
    ),
    args: {
        href: 'https://example.com',
        children: 'documentation',
        variant: 'styled',
    },
    parameters: {
        docs: {
            description: {
                story: `
  ### Inline usage
  
  **Do** use explicit spacing:
  
  \`\`\`tsx
  Read the{' '}
  <Link href="https://example.com">documentation</Link>{' '}
  for more details
  \`\`\`
  
  **Don’t** rely on implicit spacing.
          `,
            },
            source: {
                code: `<p className='text-sm text-neutral-dark-1 dark:text-white'>
    Read the{' '}
    <Link href='https://example.com'>documentation</Link>{' '}
    for more details.
  </p>`,
            },
        },
    },
};
