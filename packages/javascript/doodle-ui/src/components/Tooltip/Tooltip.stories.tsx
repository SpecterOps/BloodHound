import type { Meta, StoryObj } from '@storybook/react';
import { Button } from 'components/Button';
import { Tooltip, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger } from './Tooltip';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Tooltip',
    component: Tooltip,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        open: { type: 'boolean', control: 'boolean' },
        defaultOpen: { type: 'boolean', control: 'boolean' },
        onOpenChange: { type: 'function' },
        delayDuration: { type: 'number', control: 'number' },
        contentWidth: { type: 'string', control: 'select', options: ['default', 'match_trigger', 'lg', 'md', 'sm'] },
    },
    args: {},
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const BasicTooltip: Story = {
    args: {
        tooltip: 'This is a basic tooltip description',
    },
};

export const InlineTooltip: Story = {
    args: {
        tooltip: 'Example of a tooltip in context',
        triggerProps: {
            className: 'ml-2',
        },
    },
    render: (args) => {
        return (
            <div className='flex font-bold items-center justify-center'>
                Severity
                <Tooltip {...args} />
            </div>
        );
    },
};

export const LongTooltip: Story = {
    args: {
        tooltip: `Lorem ipsum dolor sit amet consectetur adipisicing elit. Repudiandae voluptate assumenda eius ipsum ab, eum earum corrupti provident sit quaerat quasi? Repudiandae minima esse nisi nihil nam quam vel ex?`,
        contentWidth: 'lg',
    },
    render: (args) => {
        return <Tooltip {...args} />;
    },
};

export const CustomTrigger: Story = {
    args: {
        tooltip: 'This tooltip has a custom trigger to demonstrate that you can add a tooltip to most things',
        contentWidth: 'lg',
    },
    render: (args) => {
        return (
            <Tooltip {...args}>
                <Button>Example</Button>
            </Tooltip>
        );
    },
};

export const DeconstructedTooltip: Story = {
    args: {
        tooltip: '',
    },
    render: () => {
        return (
            <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <button>Example</button>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent>
                            This is an example of a deconstructed tooltip. You can control much more than the shipped
                            Tooltip component. However, the shipped tooltip should cover most cases.
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>
        );
    },
};
