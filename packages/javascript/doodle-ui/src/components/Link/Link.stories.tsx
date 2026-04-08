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
- Applies secure \`rel="noopener noreferrer"\`
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
        <p className='text-sm text-neutral-dark-1'>
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
                code: `<p className='text-sm text-neutral-dark-1'>
    Read the{' '}
    <Link href='https://example.com'>documentation</Link>{' '}
    for more details.
  </p>`,
            },
        },
    },
};
