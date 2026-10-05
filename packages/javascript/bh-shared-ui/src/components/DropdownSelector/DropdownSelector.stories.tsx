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
import { faCheck, faClock, faXmark } from '@fortawesome/free-solid-svg-icons';
import type { Meta, StoryObj } from '@storybook/react';
import { expect, fn, userEvent, within } from '@storybook/test';
import { useEffect, useState } from 'react';
import DropdownSelector from './DropdownSelector';

const meta: Meta<typeof DropdownSelector> = {
    title: 'Components/DropdownSelector',
    component: DropdownSelector,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    args: {
        options: [
            { key: 1, value: 'All statuses' },
            { key: 2, value: 'Completed' },
            { key: 3, value: 'Pending' },
        ],
        selectedText: 'All statuses',
        caption: 'Status',
        onChange: fn(),
    },
    render: function Render(args) {
        const [selectedText, setSelectedText] = useState(args.selectedText);
        useEffect(() => setSelectedText(args.selectedText), [args.selectedText]);
        return (
            <DropdownSelector
                {...args}
                selectedText={selectedText}
                onChange={(selection) => {
                    args.onChange(selection);
                    setSelectedText(selection.value);
                }}
            />
        );
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Primary: Story = { args: { variant: 'primary', caption: undefined } };
export const WithIcons: Story = {
    args: {
        options: [
            { key: 1, value: 'Completed', icon: faCheck },
            { key: 2, value: 'Pending', icon: faClock },
            { key: 3, value: 'Failed', icon: faXmark },
        ],
    },
};
export const SelectOption: Story = {
    play: async ({ canvasElement, args }) => {
        const canvas = within(canvasElement);
        const body = within(canvasElement.ownerDocument.body);
        await userEvent.click(canvas.getByRole('button', { name: /All statuses/i }));
        await userEvent.click(await body.findByRole('button', { name: 'Completed' }));
        await expect(canvas.getByRole('button', { name: /Completed/i })).toBeVisible();
        await expect(args.onChange).toHaveBeenCalledWith({ key: 2, value: 'Completed' });
    },
};
