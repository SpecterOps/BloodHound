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
import { faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { Accordion, AccordionContent, AccordionHeader, AccordionItem } from './Accordion';

const meta = {
    title: 'Components/Accordion',
    component: Accordion,
    parameters: {},
    argTypes: {
        type: { type: 'string', control: 'select', options: ['single', 'multiple'] },
    },
    tags: ['autodocs'],
    render: (args) => (
        <Accordion {...args}>
            <AccordionItem value='item-1'>
                <AccordionHeader>
                    <FontAwesomeIcon icon={faChevronUp} className='p-1' />
                    <span>Accordion Item 1</span>
                </AccordionHeader>
                <AccordionContent>
                    <p className='font-bold mb-2'>Description</p>
                    <p className='mb-2'>Lorem ipsum dolor sit amet.</p>
                </AccordionContent>
            </AccordionItem>
            <AccordionItem value='item-2'>
                <AccordionHeader>
                    <FontAwesomeIcon icon={faChevronUp} className='p-1' />
                    <span>Accordion Item 2</span>
                </AccordionHeader>
                <AccordionContent>
                    <p className='font-bold mb-2'>Description</p>
                    <p className='mb-2'>Lorem ipsum dolor sit amet.</p>
                </AccordionContent>
            </AccordionItem>
            <AccordionItem value='item-3'>
                <AccordionHeader>
                    <FontAwesomeIcon icon={faChevronUp} className='p-1' />
                    <span>Accordion Item 3</span>
                </AccordionHeader>
                <AccordionContent>
                    <p className='font-bold mb-2'>Description</p>
                    <p className='mb-2'>Lorem ipsum dolor sit amet.</p>
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    ),
} satisfies Meta<typeof Accordion>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SingleUncontrolled: Story = {
    args: {
        type: 'single',
        collapsible: true,
        defaultValue: 'item-1',
    },
};

export const SingleControlled: Story = {
    args: {
        type: 'single',
        collapsible: true,
        value: 'item-1',
        onValueChange: fn(),
    },
};

export const MultipleUncontrolled: Story = {
    args: {
        type: 'multiple',
        defaultValue: ['item-1', 'item-3'],
    },
};

export const MultipleControlled: Story = {
    args: {
        type: 'multiple',
        value: ['item-1', 'item-3'],
        onValueChange: fn(),
    },
};
