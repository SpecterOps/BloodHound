import type { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, within } from '@storybook/test';
import { Switch } from './Switch';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/Switch',
    component: Switch,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        size: { options: ['small', 'medium', 'large'], control: 'select' },
        label: { control: 'text' },
        labelPosition: { options: ['left', 'right'], control: 'select' },
    },
    args: {},
} satisfies Meta<typeof Switch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Base: Story = {
    play: async ({ canvasElement, step }) => {
        const canvas = within(canvasElement);
        const switchButton = await canvas.findByRole('switch');

        await step('Click the switch to toggle it', async () => {
            expect(switchButton).toHaveAttribute('aria-checked', 'false');

            await userEvent.click(switchButton);

            await expect(switchButton).toHaveAttribute('aria-checked', 'true');
        });

        await step('Pressing spacebar or enter on the focused element toggles the switch', async () => {
            expect(switchButton).toHaveAttribute('aria-checked', 'true');

            switchButton.focus();

            await userEvent.keyboard(' ');

            await expect(switchButton).toHaveAttribute('aria-checked', 'false');

            switchButton.focus();

            await userEvent.keyboard('{enter}');

            await expect(switchButton).toHaveAttribute('aria-checked', 'true');

            // Other keyboard inputs like tab should not toggle the switch
            await userEvent.keyboard('{tab}');

            await expect(switchButton).not.toHaveAttribute('aria-checked', 'false');
        });
    },
};

export const Disabled: Story = {
    args: {
        disabled: true,
    },
};

export const DefaultChecked: Story = {
    args: {
        defaultChecked: true,
    },
};

export const Small: Story = {
    args: {
        size: 'small',
    },
};

export const Large: Story = {
    args: {
        size: 'large',
    },
};

export const Labeled: Story = {
    args: {
        label: 'Muted',
    },
};

export const LeftLabeled: Story = {
    args: {
        label: 'Muted',
        labelPosition: 'left',
    },
};
