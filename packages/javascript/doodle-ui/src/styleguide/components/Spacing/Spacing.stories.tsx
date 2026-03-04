import type { Meta, StoryObj } from '@storybook/react';
import { Spacing } from './Spacing';

const storyDescription = `<div><h4>Visual reference for Tailwind spacing values.</h4><p>Usage: <code>className="w-4"</code></p></div>`;
const meta = {
    title: 'Styleguide/Spacing',
    component: Spacing,
    tags: ['autodocs'],
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                story: storyDescription,
            },
        },
    },
} satisfies Meta<typeof Spacing>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
