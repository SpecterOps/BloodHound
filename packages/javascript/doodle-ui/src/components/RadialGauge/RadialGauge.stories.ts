import type { Meta, StoryObj } from '@storybook/react';
import { RadialGauge } from './RadialGauge';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/RadialGauge',
    component: RadialGauge,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        value: { type: 'number', control: { max: 100, min: 0 } },
        color: {
            type: 'string',
            control: 'select',
            options: ['primary', 'tertiary'],
        },
    },
} satisfies Meta<typeof RadialGauge>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const Primary: Story = {
    args: {
        value: 75,
        color: 'primary',
    },
};

export const Tertiary: Story = {
    args: {
        value: 80,
        color: 'tertiary',
    },
};

export const Custom: Story = {
    args: {
        value: 80,
        color: '#FFB6C1',
    },
};
