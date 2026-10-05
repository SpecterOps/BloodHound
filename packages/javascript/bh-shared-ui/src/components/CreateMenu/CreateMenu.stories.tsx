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
import CreateMenu from './CreateMenu';

const createCollector = fn();
const createEnvironment = fn();
const meta: Meta<typeof CreateMenu> = {
    title: 'Components/CreateMenu',
    component: CreateMenu,
    tags: ['autodocs'],
    parameters: { layout: 'centered' },
    args: {
        createMenuTitle: 'Create',
        menuItems: [
            { title: 'Create collector', onClick: createCollector },
            { title: 'Create environment', onClick: createEnvironment },
        ],
    },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const MultipleActions: Story = {};
export const SingleAction: Story = { args: { menuItems: [{ title: 'Create collector', onClick: createCollector }] } };
export const Disabled: Story = { args: { disabled: true } };
export const SelectAction: Story = {
    play: async ({ canvasElement }) => {
        createCollector.mockClear();
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Create' }));
        await userEvent.click(
            await within(canvasElement.ownerDocument.body).findByRole('menuitem', { name: 'Create collector' })
        );
        await expect(createCollector).toHaveBeenCalled();
    },
};
