import type { Meta, StoryObj } from '@storybook/react';
import { RiskBadge } from './RiskBadge';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/RiskBadge',
    component: RiskBadge,
    parameters: {
        layout: 'centered',
    },
    argTypes: {
        type: { type: 'string', control: 'select', options: ['labeled', 'sm-circle', 'md-circle'] },
        color: {
            type: 'string',
            control: 'color',
        },
        outlined: { type: 'boolean' },
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
} satisfies Meta<typeof RiskBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const Default: Story = {
    args: {
        type: 'md-circle',
        color: 'primary',
        outlined: false,
    },
};

export const Labeled: Story = {
    args: {
        type: 'labeled',
        color: 'secondary',
        label: 'Critical',
        outlined: false,
    },
};

export const OutlinedLabel: Story = {
    args: {
        type: 'labeled',
        color: 'secondary',
        label: 'Critical',
        outlined: true,
    },
};

export const Small: Story = {
    args: {
        type: 'sm-circle',
        color: 'tertiary',
        outlined: false,
    },
};

export const Outlined: Story = {
    args: {
        type: 'md-circle',
        color: 'primary',
        outlined: true,
    },
};
