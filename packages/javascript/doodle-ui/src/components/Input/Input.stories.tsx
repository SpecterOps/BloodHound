import type { Meta, StoryObj } from '@storybook/react';

import { Button } from 'components/Button';
import { Label } from 'components/Label/Label';
import { Input } from './Input';

/**
 * Displays a form input field or a component that looks like an input field.
 */
const meta = {
    title: 'Components/Input',
    component: Input,
    tags: ['autodocs'],
    argTypes: { variant: { options: ['outlined', 'underlined'], control: 'select' } },
    args: {
        className: 'w-96',
        type: 'email',
        placeholder: 'email@example.com',
        disabled: false,
    },
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Input>;

export default meta;

type Story = StoryObj<typeof meta>;

/**
 * The default form of the input field.
 */
export const Default: Story = {};

export const Outlined: Story = { args: { variant: 'outlined' } };

/**
 * Use the `disabled` prop to make the input non-interactive and appears faded,
 * indicating that input is not currently accepted.
 */
export const Disabled: Story = {
    args: { disabled: true },
};

/**
 * Use the `Label` component to includes a clear, descriptive label above or
 * alongside the input area to guide users.
 */
export const WithLabel: Story = {
    render: (args) => (
        <div className='grid items-center gap-1.5'>
            <Label htmlFor='email'>Email</Label>
            <Input {...args} id='email' />
        </div>
    ),
};

/**
 * Use a text element below the input field to provide additional instructions
 * or information to users.
 */
export const WithHelperText: Story = {
    render: (args) => (
        <div className='grid items-center gap-1.5'>
            <Label htmlFor='email-2'>Email</Label>
            <Input {...args} id='email-2' />
            <p className='text-sm text-neutral-dark-1 dark:text-neutral-light-1 opacity-50'>
                Enter your email address.
            </p>
        </div>
    ),
};

/**
 * Use the `Button` component to indicate that the input field can be submitted
 * or used to trigger an action.
 */
export const WithButton: Story = {
    render: (args) => (
        <div className='flex items-center space-x-2'>
            <Input {...args} />
            <Button className='rounded-md'>Subscribe</Button>
        </div>
    ),
};

export const WithFile: Story = {
    args: {
        type: 'file',
        placeholder: 'No file selected.',
    },
    render: (args) => (
        <div className='grid w-full max-w-sm items-center gap-1.5'>
            <Label htmlFor='picture'>Upload Image</Label>
            <Input id='picture' type='file' {...args} variant={'outlined'} />
        </div>
    ),
};
