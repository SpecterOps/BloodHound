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
import { faDownload } from '@fortawesome/free-solid-svg-icons';
import type { Meta, StoryObj } from '@storybook/react';
import { expect, fn, userEvent, within } from '@storybook/test';
import { MenuItem } from 'doodle-ui';
import GraphMenu from './GraphMenu';

const onSelect = fn();
function ExportOptions() {
    return (
        <>
            <MenuItem onSelect={onSelect}>JSON</MenuItem>
            <MenuItem disabled>Image unavailable</MenuItem>
        </>
    );
}
const meta: Meta<typeof GraphMenu> = {
    title: 'Components/GraphMenu',
    component: GraphMenu,
    tags: ['autodocs'],
    args: { label: 'Export', icon: faDownload, children: <ExportOptions /> },
    argTypes: { icon: { control: false }, children: { control: false } },
    parameters: { layout: 'centered' },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const CustomTooltip: Story = { args: { tooltip: 'Export current graph results' } };
export const SelectAction: Story = {
    play: async ({ canvasElement }) => {
        onSelect.mockClear();
        await userEvent.click(within(canvasElement).getByRole('button', { name: 'Export' }));
        const body = within(canvasElement.ownerDocument.body);
        await expect(await body.findByRole('menuitem', { name: 'Image unavailable' })).toHaveAttribute(
            'aria-disabled',
            'true'
        );
        await userEvent.click(body.getByRole('menuitem', { name: 'JSON' }));
        await expect(onSelect).toHaveBeenCalled();
    },
};
