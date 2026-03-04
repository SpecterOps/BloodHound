import type { Meta, StoryObj } from '@storybook/react';
import { FontSizes } from './FontSizes';

const storyDescription = `<h4>Tailwind classes for font-sizes</h4><p>Usage: <code>className="text-headline-1"</code></p>`;

const meta = {
    title: 'Styleguide/FontSizes',
    component: FontSizes,
    tags: ['autodocs'],
    parameters: {
        docs: {
            description: {
                story: storyDescription,
            },
        },
    },
} satisfies Meta<typeof FontSizes>;

export default meta;
type Story = StoryObj<typeof meta>;

export const FontSize: Story = {
    args: {},
};
