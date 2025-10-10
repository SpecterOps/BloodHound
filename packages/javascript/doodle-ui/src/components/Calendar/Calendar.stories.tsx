import { action } from '@storybook/addon-actions';
import type { Meta, StoryObj } from '@storybook/react';
import { Calendar } from './Calendar';
import { addDays } from './utils';

/**
 * A date field component that allows users to enter and edit date.
 */

const meta = {
    title: 'Components/Calendar',
    component: Calendar,
    tags: ['autodocs'],
    argTypes: {},
    args: {
        mode: 'single',
        selected: new Date(),
        onSelect: action('onDayClick'),
        className: 'rounded-md border w-fit',
    },
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Calendar>;

export default meta;

type Story = StoryObj<typeof meta>;

/**
 * The default form of the calendar.
 */
export const Default: Story = {};

/**
 * Use the `multiple` mode to select multiple dates.
 */
export const Multiple: Story = {
    args: {
        min: 1,
        selected: [new Date(), addDays(new Date(), 2), addDays(new Date(), 8)],
        mode: 'multiple',
    },
};

/**
 * Use the `range` mode to select a range of dates.
 */
export const Range: Story = {
    args: {
        selected: {
            from: new Date(),
            to: addDays(new Date(), 9),
        },
        mode: 'range',
    },
};

/**
 * Set the `captionLayout` prop to `'dropdown-buttons'` and provide start and end dates to display dropdown selectors for the month and year.
 */
export const DropdownMenus: Story = {
    args: {
        captionLayout: 'dropdown-buttons',
        fromYear: 2007,
        toYear: 2023,
    },
};

/**
 * Use the `disabled` prop to disable specific dates.
 */
export const Disabled: Story = {
    args: {
        disabled: [addDays(new Date(), 1), addDays(new Date(), 2), addDays(new Date(), 3), addDays(new Date(), 5)],
    },
};

/**
 * Use the `numberOfMonths` prop to display multiple months.
 */
export const MultipleMonths: Story = {
    args: {
        numberOfMonths: 2,
        showOutsideDays: false,
    },
};
