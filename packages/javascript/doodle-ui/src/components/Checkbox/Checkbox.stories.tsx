import type { Meta, StoryObj } from '@storybook/react';
import { Label } from 'components/Label';
import { Checkbox } from './Checkbox';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Checkbox',
    component: Checkbox,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        size: {
            control: 'select',
            options: ['lg', 'md', 'sm'],
        },
    },
    args: { size: 'md' },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    render: (args) => <Checkbox {...args} />,
};

export const Labeled: Story = {
    render: () => {
        return (
            <div className='flex justify-center flex-row items-center'>
                <Checkbox id='test-id' />
                <Label htmlFor='test-id' className='pl-2'>
                    Testing Label
                </Label>
            </div>
        );
    },
};
