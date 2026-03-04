import type { Meta, StoryObj } from '@storybook/react';
import { ColorPalette } from './ColorPalette';

const storyDescription = `<div><p>Usage: <code>className="text-primary"</code></p></div>`;

const meta = {
    title: 'Styleguide/ColorPalette',
    component: ColorPalette,
    tags: ['autodocs'],
    parameters: {
        docs: {
            description: {
                story: storyDescription,
            },
        },
    },
} satisfies Meta<typeof ColorPalette>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
