// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import type { Meta, StoryObj } from '@storybook/react';
import { expect, fn, userEvent, within } from '@storybook/test';
import { useEffect, useState } from 'react';
import CardWithSwitch from './CardWithSwitch';

const meta: Meta<typeof CardWithSwitch> = {
    title: 'Components/CardWithSwitch',
    component: CardWithSwitch,
    tags: ['autodocs'],
    args: {
        title: 'Scheduled analysis',
        description: 'Automatically analyze newly collected data each day.',
        isEnabled: true,
        onSwitchChange: fn(),
    },
    decorators: [
        (Story) => (
            <div className='max-w-xl'>
                <Story />
            </div>
        ),
    ],
    render: function Render(args) {
        const [isEnabled, setIsEnabled] = useState(args.isEnabled);
        useEffect(() => setIsEnabled(args.isEnabled), [args.isEnabled]);
        return (
            <CardWithSwitch
                {...args}
                isEnabled={isEnabled}
                onSwitchChange={() => {
                    args.onSwitchChange();
                    setIsEnabled(!isEnabled);
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Enabled: Story = {};
export const DisabledSetting: Story = { args: { isEnabled: false } };
export const ReadOnly: Story = { args: { disableSwitch: true } };
function SettingDetails() {
    return (
        <p>
            Next analysis: <strong>Every day at 09:00 UTC</strong>
        </p>
    );
}
export const CustomContent: Story = { args: { children: <SettingDetails /> } };
export const Toggle: Story = {
    play: async ({ canvasElement, args }) => {
        const canvas = within(canvasElement);
        const toggle = canvas.getByRole('switch');
        await expect(toggle).toBeChecked();
        await userEvent.click(toggle);
        await expect(toggle).not.toBeChecked();
        await expect(args.onSwitchChange).toHaveBeenCalled();
    },
};
