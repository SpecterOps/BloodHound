import type { Meta, StoryObj } from '@storybook/react';
import { AppIcons } from './AppIcons';

/**
 * Usage:
 *
 * ```javascript
 *
 * <AppIcon.CaretDown size={12} />
 * ```
 */

const meta = {
    title: 'Styleguide/AppIcons',
    component: AppIcons,
    tags: ['autodocs'],
} satisfies Meta<typeof AppIcons>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Story: Story = {
    args: {},
};
