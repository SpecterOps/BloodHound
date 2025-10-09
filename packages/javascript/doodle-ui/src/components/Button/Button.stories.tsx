import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { Button } from './Button';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGear, faListUl } from '@fortawesome/free-solid-svg-icons';
import { within, expect } from '@storybook/test';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Button',
    component: Button,
    parameters: {
        // Optional parameter to center the component in the Canvas. More info: https://storybook.js.org/docs/configure/story-layout
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        variant: {
            options: ['primary', 'secondary', 'tertiary', 'transparent', 'icon', 'text'],
            control: 'select',
        },
        fontColor: { options: ['primary'], control: 'select' },
        size: { options: ['small', 'medium', 'large'], control: 'select' },
        asChild: { options: [false, true], control: 'select' },
    },
    // Use `fn` to spy on the onClick arg, which will appear in the actions panel once invoked: https://storybook.js.org/docs/essentials/actions#action-args
    args: { onClick: fn(), asChild: false },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

// More on writing stories with args: https://storybook.js.org/docs/writing-stories/args
export const Primary: Story = {
    args: {
        variant: 'primary',
        children: 'Next',
    },
};

export const Secondary: Story = {
    args: {
        variant: 'secondary',
        children: 'Next',
    },
};

export const Tertiary: Story = {
    args: {
        variant: 'tertiary',
        children: 'Next',
    },
};

export const Transparent: Story = {
    args: {
        variant: 'transparent',
        children: 'Next',
    },
};

export const Large: Story = {
    args: {
        size: 'large',
        children: 'Button',
    },
};

export const Small: Story = {
    args: {
        size: 'small',
        children: 'Button',
    },
};

export const IconButton: Story = {
    render: () => {
        return (
            <Button variant='icon' aria-label='Gear Icon'>
                <FontAwesomeIcon icon={faGear} style={{ fontSize: '24px' }} />
            </Button>
        );
    },
};

export const TextButton: Story = {
    render: () => {
        return (
            <Button variant='text'>
                <FontAwesomeIcon icon={faListUl} className='mb-[3px]' />
                <span className={'ml-2'}>Text Button</span>
            </Button>
        );
    },
};

export const TextButtonAlt: Story = {
    render: () => {
        return (
            <Button variant='text' fontColor={'primary'}>
                <span className={'ml-2'}>Text Button Alt</span>
            </Button>
        );
    },
};

export const DefaultType: Story = {
    render: () => {
        return <Button>Type is Button</Button>;
    },
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        const button = await canvas.findByRole('button');
        // Assert that the default type is button
        // We default to this type so that the element does not submit forms if a type is not passed
        expect(button).toHaveAttribute('type', 'button');
    },
};
