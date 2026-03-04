import type { Meta, StoryObj } from '@storybook/react';
import { ChevronDown, ChevronUp, Minus, Plus } from 'lucide-react';
import { ReactNode } from 'react';
import { Badge } from './Badge';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Badge',
    component: Badge,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        icon: { control: 'select', options: ['chevron-down', 'chevron-up', 'minus', 'plus'] },
        color: { control: { type: 'color' } },
        backgroundColor: { control: { type: 'color' } },
        label: { type: 'string' },
    },
    args: {},
} satisfies Meta<typeof Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

const ICONS: Record<string, ReactNode> = {
    'chevron-down': <ChevronDown />,
    'chevron-up': <ChevronUp />,
    minus: <Minus />,
    plus: <Plus />,
};

export const PositiveBadge: Story = {
    args: {
        label: '10',
        icon: 'chevron-up',
        color: '#02c577',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};

export const NegativeBadge: Story = {
    args: {
        label: '10',
        icon: 'chevron-down',
        color: '#e15851',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};

export const ZeroBadge: Story = {
    args: {
        label: '0',
        icon: 'minus',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};

export const LongBadge: Story = {
    args: {
        label: '1,000,000',
        icon: 'plus',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};

export const CustomColorBadge: Story = {
    args: {
        label: '10',
        color: '#0AF',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};

export const CustomBackgroundColorBadge: Story = {
    args: {
        label: '95%',
        backgroundColor: '#E15851',
    },
    render: (args) => {
        return <Badge {...args} icon={ICONS[args.icon as string]} />;
    },
};
