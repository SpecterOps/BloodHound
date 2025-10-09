import type { Meta, StoryObj } from '@storybook/react';
import { $0 } from './$0';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/$0',
    component: $0,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof $0>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
