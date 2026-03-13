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

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectPortal,
    SelectSeparator,
    SelectTrigger,
    SelectValue,
} from './Select';

/**
 * Displays a list of options for the user to pick from—triggered by a button.
 */

const meta: Meta<typeof Select> = {
    title: 'Components/Select',
    component: Select,
    tags: ['autodocs'],
    argTypes: {},
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Select>;

export default meta;

type Story = StoryObj<typeof meta>;

/**
 * The default form of the select.
 */
export const Default: Story = {
    render: (args) => (
        <Select {...args}>
            <SelectTrigger className='w-60'>
                <SelectValue placeholder='Pick your favorite color' />
            </SelectTrigger>
            <SelectPortal>
                <SelectContent>
                    <SelectItem value='red'>Red</SelectItem>
                    <SelectItem value='green'>Green</SelectItem>
                    <SelectItem value='blue'>Blue</SelectItem>
                    <SelectItem value='orange'>Orange</SelectItem>
                    <SelectItem value='yellow'>Yellow</SelectItem>
                    <SelectItem value='purple'>Purple</SelectItem>
                    <SelectItem value='pink'>Pink</SelectItem>
                </SelectContent>
            </SelectPortal>
        </Select>
    ),
};

/**
 * SelectTrigger variant underlined.
 */
export const Underlined: Story = {
    render: (args) => (
        <Select {...args}>
            <SelectTrigger variant='underlined' className='w-60'>
                <SelectValue placeholder='Pick your favorite color' />
            </SelectTrigger>
            <SelectPortal>
                <SelectContent>
                    <SelectItem value='red'>Red</SelectItem>
                    <SelectItem value='green'>Green</SelectItem>
                    <SelectItem value='blue'>Blue</SelectItem>
                    <SelectItem value='orange'>Orange</SelectItem>
                    <SelectItem value='yellow'>Yellow</SelectItem>
                    <SelectItem value='purple'>Purple</SelectItem>
                    <SelectItem value='pink'>Pink</SelectItem>
                </SelectContent>
            </SelectPortal>
        </Select>
    ),
};

/**
 * Align currently selected item in popout with the input field.
 */
export const ItemAligned: Story = {
    render: (args) => (
        <Select {...args}>
            <SelectTrigger className='w-60'>
                <SelectValue placeholder='Select a month' />
            </SelectTrigger>
            <SelectPortal>
                <SelectContent position='item-aligned'>
                    <SelectItem value='January'>January</SelectItem>
                    <SelectItem value='February'>February</SelectItem>
                    <SelectItem value='March'>March</SelectItem>
                    <SelectItem value='April'>April</SelectItem>
                    <SelectItem value='May'>May</SelectItem>
                    <SelectItem value='June'>June</SelectItem>
                    <SelectItem value='July'>July</SelectItem>
                    <SelectItem value='August'>August</SelectItem>
                    <SelectItem value='September'>September</SelectItem>
                    <SelectItem value='October'>October</SelectItem>
                    <SelectItem value='November'>November</SelectItem>
                    <SelectItem value='December'>December</SelectItem>
                </SelectContent>
            </SelectPortal>
        </Select>
    ),
};

export const WithHierarchies: Story = {
    render: (args) => (
        <Select {...args}>
            <SelectTrigger className='w-60'>
                <SelectValue placeholder='Select a fruit' />
            </SelectTrigger>
            <SelectPortal>
                <SelectContent>
                    <SelectGroup>
                        <SelectLabel>Fruits</SelectLabel>
                        <SelectItem value='apple'>Apple</SelectItem>
                        <SelectItem value='banana'>Banana</SelectItem>
                        <SelectItem value='blueberry'>Blueberry</SelectItem>
                        <SelectItem value='grapes'>Grapes</SelectItem>
                        <SelectItem value='pineapple'>Pineapple</SelectItem>
                    </SelectGroup>
                    <SelectSeparator />
                    <SelectGroup>
                        <SelectLabel>Vegetables</SelectLabel>
                        <SelectItem value='aubergine'>Aubergine</SelectItem>
                        <SelectItem value='broccoli'>Broccoli</SelectItem>
                        <SelectItem value='carrot' disabled>
                            Carrot
                        </SelectItem>
                        <SelectItem value='courgette'>Courgette</SelectItem>
                        <SelectItem value='leek'>Leek</SelectItem>
                    </SelectGroup>
                    <SelectSeparator />
                    <SelectGroup>
                        <SelectLabel>Meat</SelectLabel>
                        <SelectItem value='beef'>Beef</SelectItem>
                        <SelectItem value='chicken'>Chicken</SelectItem>
                        <SelectItem value='lamb'>Lamb</SelectItem>
                        <SelectItem value='pork'>Pork</SelectItem>
                    </SelectGroup>
                </SelectContent>
            </SelectPortal>
        </Select>
    ),
};
