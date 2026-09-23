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
import { X } from 'lucide-react';
import { Button } from '../Button';
import {
    Drawer,
    DrawerBody,
    DrawerClose,
    DrawerContent,
    DrawerDescription,
    DrawerFooter,
    DrawerHeader,
    DrawerTitle,
    DrawerTrigger,
} from './Drawer';

const meta = {
    title: 'Components/Drawer',
    component: DrawerContent,
    parameters: {
        docs: {
            description: {
                component:
                    'A right-side drawer for longer tasks and contextual details. The backdrop, Escape key, and close button dismiss it. Put long content in DrawerBody to keep the header and footer visible.',
            },
        },
    },
    tags: ['autodocs'],
} satisfies Meta<typeof DrawerContent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Form: Story = {
    render: () => (
        <Drawer>
            <DrawerTrigger render={<Button variant='secondary' />}>Open Drawer</DrawerTrigger>
            <DrawerContent>
                <DrawerHeader>
                    <div>
                        <DrawerTitle>Form Inside Drawer</DrawerTitle>
                        <DrawerDescription className='mt-3 text-muted'>
                            This form contains a name, but does not actually save.
                        </DrawerDescription>
                    </div>
                    <DrawerClose
                        aria-label='Close drawer'
                        className='rounded-full p-2 focus:outline-none focus-visible:focus-ring'>
                        <X aria-hidden='true' size={20} />
                    </DrawerClose>
                </DrawerHeader>
                <DrawerBody className='space-y-4'>
                    <div>
                        <p>Name</p>
                        <input
                            aria-label='Name'
                            className='mt-2 block w-full rounded border p-2'
                            placeholder='Enter collection name'
                        />
                    </div>
                    <p>Content can scroll in DrawerBody.</p>
                </DrawerBody>
                <DrawerFooter className='justify-end gap-2'>
                    <DrawerClose render={<Button variant='secondary' />}>Cancel</DrawerClose>
                    <Button>Create Plan</Button>
                </DrawerFooter>
            </DrawerContent>
        </Drawer>
    ),
};

export const ScrollingList: Story = {
    render: () => (
        <Drawer>
            <DrawerTrigger render={<Button variant='secondary' />}>Open Drawer</DrawerTrigger>
            <DrawerContent>
                <DrawerHeader>
                    <DrawerTitle>Long List of Items</DrawerTitle>
                    <DrawerClose
                        aria-label='Close drawer'
                        className='rounded-full p-2 focus:outline-none focus-visible:focus-ring'>
                        <X aria-hidden='true' size={20} />
                    </DrawerClose>
                </DrawerHeader>
                <DrawerBody>
                    <DrawerDescription className='mb-6 text-muted'>
                        See how a list scrolls between header and footer.
                    </DrawerDescription>
                    <ul className='divide-y'>
                        {Array.from({ length: 20 }, (_, index) => (
                            <li className='py-3 flex justify-between' key={index}>
                                Item {index + 1}
                                <span>Success</span>
                            </li>
                        ))}
                    </ul>
                </DrawerBody>
                <DrawerFooter className='justify-end'>
                    <DrawerClose render={<Button variant='secondary' />}>Close</DrawerClose>
                </DrawerFooter>
            </DrawerContent>
        </Drawer>
    ),
};
