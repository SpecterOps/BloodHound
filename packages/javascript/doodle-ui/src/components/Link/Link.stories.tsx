import type { Meta, StoryObj } from '@storybook/react';
import { Link } from './Link';

const meta = {
    title: 'Components/Link',
    component: Link,
    tags: ['autodocs'],
    args: {
        href: '#',
        children: 'Visit details',
        variant: 'styled',
    },
    argTypes: {
        variant: {
            control: 'select',
            options: ['styled', 'unstyled'],
            description: 'Controls whether the link uses default link styling or appears visually unstyled.',
        },
        href: {
            control: 'text',
            description: 'Destination URL for the anchor element.',
        },
        children: {
            control: 'text',
            description: 'Link content.',
        },
    },
} satisfies Meta<typeof Link>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Styled: Story = {
    args: {
        variant: 'styled',
        children: 'styled link',
    },
};

export const Unstyled: Story = {
    args: {
        variant: 'unstyled',
        children: 'unstyled link',
    },
};
