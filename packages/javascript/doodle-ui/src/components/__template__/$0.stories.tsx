import type { Meta, StoryObj } from '@storybook/react';
import { $0 } from './$0';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/$0',
    component: $0,
    parameters: {
        layout: 'centered',
    },
    // This story will not appear in Storybook's sidebar or docs page: https://storybook.js.org/docs/writing-stories/tags
    tags: ['!dev'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {},
    args: {},
} satisfies Meta<typeof $0>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
