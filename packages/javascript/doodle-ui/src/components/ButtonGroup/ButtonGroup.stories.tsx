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
import { faCaretDown, faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { expect, fn, userEvent, within } from '@storybook/test';
import { Button } from '../Button';
import { Input } from '../Input';
import { Menu, MenuContent, MenuItem, MenuTrigger } from '../Menu';
import { ButtonGroup, ButtonGroupSeparator, ButtonGroupText } from './ButtonGroup';

const meta = {
    title: 'Components/ButtonGroup',
    component: ButtonGroup,
    tags: ['autodocs'],
    parameters: {
        layout: 'centered',
        docs: {
            description: {
                component:
                    'Group related actions with joined borders. Label the group with aria-label or aria-labelledby. Tab moves between buttons; use ToggleGroup for state selection. Set orientation on the group and the perpendicular orientation on separators. Use secondary buttons for outlined groups, and add a separator between filled buttons. ButtonGroupText supports the Base UI render prop for labels or other elements.',
            },
        },
    },
    argTypes: {
        orientation: { control: 'inline-radio', options: ['horizontal', 'vertical'] },
    },
    args: { orientation: 'horizontal', 'aria-label': 'Query actions' },
} satisfies Meta<typeof ButtonGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
    render: (args) => (
        <ButtonGroup {...args}>
            <Button variant='secondary'>Copy</Button>
            <Button variant='secondary'>Paste</Button>
            <Button variant='secondary' disabled>
                Delete
            </Button>
        </ButtonGroup>
    ),
};

export const Vertical: Story = {
    args: { orientation: 'vertical', 'aria-label': 'Zoom controls' },
    render: (args) => (
        <ButtonGroup {...args}>
            <Button variant='secondary' aria-label='Zoom in'>
                <FontAwesomeIcon icon={faPlus} />
            </Button>
            <Button variant='secondary' aria-label='Zoom out'>
                <FontAwesomeIcon icon={faMinus} />
            </Button>
        </ButtonGroup>
    ),
};

export const WithSeparator: Story = {
    render: ({ orientation, ...args }) => (
        <ButtonGroup {...args} orientation={orientation}>
            <Button>Copy</Button>
            <ButtonGroupSeparator orientation={orientation === 'vertical' ? 'horizontal' : 'vertical'} />
            <Button>Paste</Button>
        </ButtonGroup>
    ),
};

export const WithInput: Story = {
    args: { 'aria-label': 'Search nodes' },
    render: (args) => (
        <ButtonGroup {...args}>
            <ButtonGroupText render={<label htmlFor='button-group-search'>Nodes</label>}>Nodes</ButtonGroupText>
            <Input id='button-group-search' variant='outlined' placeholder='Search nodes...' />
            <Button variant='secondary'>Search</Button>
        </ButtonGroup>
    ),
};

export const Nested: Story = {
    args: { 'aria-label': 'Query toolbar' },
    render: (args) => (
        <ButtonGroup {...args}>
            <ButtonGroup aria-label='Clipboard actions'>
                <Button variant='secondary'>Copy</Button>
                <Button variant='secondary'>Paste</Button>
            </ButtonGroup>
            <ButtonGroup aria-label='History actions'>
                <Button variant='secondary'>Undo</Button>
                <Button variant='secondary'>Redo</Button>
            </ButtonGroup>
        </ButtonGroup>
    ),
};

const save = fn();
const saveAs = fn();

export const SplitButton: Story = {
    args: { 'aria-label': 'Save query' },
    render: (args) => (
        <ButtonGroup {...args}>
            <Button variant='secondary' onClick={save}>
                Save
            </Button>
            <Menu>
                <MenuTrigger asChild>
                    <Button variant='secondary' aria-label='Show save query options'>
                        <FontAwesomeIcon icon={faCaretDown} />
                    </Button>
                </MenuTrigger>
                <MenuContent>
                    <MenuItem onSelect={saveAs}>Save As</MenuItem>
                </MenuContent>
            </Menu>
        </ButtonGroup>
    ),
    play: async ({ canvasElement }) => {
        const canvas = within(canvasElement);
        await userEvent.click(canvas.getByRole('button', { name: 'Save' }));
        await expect(save).toHaveBeenCalled();
        await userEvent.click(canvas.getByRole('button', { name: 'Show save query options' }));
        const page = within(canvasElement.ownerDocument.body);
        await userEvent.click(await page.findByRole('menuitem', { name: 'Save As' }));
        await expect(saveAs).toHaveBeenCalled();
    },
};
