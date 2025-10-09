import type { Meta, StoryObj } from '@storybook/react';

import { Label } from './Label';

/**
 * Renders an accessible label associated with controls.
 */
const meta = {
    title: 'Components/Label',
    component: Label,
    tags: ['autodocs'],
    argTypes: {
        children: {
            control: { type: 'text' },
        },
        size: { control: 'select', options: ['small', 'medium', 'large'] },
    },
    args: {
        children: 'Email',
        htmlFor: 'email',
    },
} satisfies Meta<typeof Label>;

export default meta;

type Story = StoryObj<typeof Label>;

/**
 * The default form of the label.
 */
export const Default: Story = {};
