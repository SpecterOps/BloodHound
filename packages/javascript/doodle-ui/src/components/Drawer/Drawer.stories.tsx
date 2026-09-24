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
import type { ComponentProps } from 'react';
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

type DrawerStoryArgs = ComponentProps<typeof Drawer> & {
    className?: ComponentProps<typeof DrawerContent>['className'];
};

const meta = {
    title: 'Components/Drawer',
    component: Drawer,
    parameters: {
        docs: {
            description: {
                component:
                    'A drawer that opens from the right by default. Put long content in DrawerBody to keep the header and footer visible. Side drawers are 75% wide on small screens and 24rem wide above the sm breakpoint; add a class such as w-[860px] to DrawerContent to change the width.',
            },
        },
    },
    tags: ['autodocs'],
    argTypes: {
        swipeDirection: {
            control: 'select',
            options: ['right', 'left', 'up', 'down'],
            description: 'The edge from which the drawer opens.',
            table: {
                category: 'Drawer',
                type: { summary: "'right' | 'left' | 'up' | 'down'" },
                defaultValue: { summary: 'right' },
            },
        },
        modal: {
            control: 'select',
            options: [true, false, 'trap-focus'],
            description:
                "Controls interaction outside the drawer. true traps focus, locks page scrolling, and disables outside pointer interaction; false allows interaction with the rest of the page; 'trap-focus' traps focus without locking scrolling or blocking outside pointer interaction.",
            table: {
                category: 'Drawer',
                type: { summary: "boolean | 'trap-focus'" },
                defaultValue: { summary: 'true' },
            },
        },
        showSwipeHandle: {
            control: 'boolean',
            description: 'Whether to show a swipe handle inside DrawerContent.',
            table: {
                category: 'Drawer',
                type: { summary: 'boolean' },
                defaultValue: { summary: 'false' },
            },
        },
        snapPoints: {
            control: 'object',
            description:
                'Snap positions for a top or bottom drawer. Numbers from 0 to 1 are fractions of the viewport height; larger numbers are pixels. Strings may use px or rem units.',
            table: {
                category: 'Drawer',
                type: { summary: '(number | string)[]' },
            },
        },
        defaultOpen: {
            control: false,
            description: 'Whether the drawer starts open when using uncontrolled state.',
            table: {
                category: 'Drawer',
                type: { summary: 'boolean' },
                defaultValue: { summary: 'false' },
            },
        },
        open: {
            control: false,
            description: 'Controls whether the drawer is open. Use with onOpenChange.',
            table: {
                category: 'Drawer',
                type: { summary: 'boolean' },
                defaultValue: { summary: 'uncontrolled' },
            },
        },
        onOpenChange: {
            control: false,
            description: 'Called when the drawer opens or closes.',
            table: {
                category: 'Drawer',
                type: { summary: '(open: boolean, eventDetails) => void' },
            },
        },
        className: {
            control: 'text',
            description:
                'Classes applied to DrawerContent. Use a width class such as w-[860px] to override the default width of side drawers.',
            table: {
                category: 'DrawerContent',
                type: { summary: 'string' },
                defaultValue: { summary: '75% / 24rem at sm+' },
            },
        },
    },
    args: { swipeDirection: 'right', modal: true, showSwipeHandle: false, className: '' },
} satisfies Meta<DrawerStoryArgs>;

export default meta;
type Story = StoryObj<DrawerStoryArgs>;

export const Default: Story = {
    parameters: {
        docs: {
            description: {
                story: 'A right-side drawer with a simple form at the default compact width. The header and footer remain visible while the body scrolls if its content grows.',
            },
        },
    },
    render: ({ className, ...drawerProps }) => (
        <Drawer {...drawerProps}>
            <DrawerTrigger render={<Button variant='secondary' />}>Open drawer</DrawerTrigger>
            <DrawerContent className={className}>
                <DrawerHeader>
                    <div>
                        <DrawerTitle>Create Item</DrawerTitle>
                        <DrawerDescription className='mt-3'>Enter a name for the new item.</DrawerDescription>
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
                            placeholder='Enter item name'
                        />
                    </div>
                </DrawerBody>
                <DrawerFooter className='justify-end gap-2'>
                    <DrawerClose render={<Button variant='secondary' />}>Cancel</DrawerClose>
                    <Button>Create Item</Button>
                </DrawerFooter>
            </DrawerContent>
        </Drawer>
    ),
};

export const CustomWidth: Story = {
    args: { className: 'w-[860px] max-w-[100vw]' },
    parameters: {
        docs: {
            description: {
                story: 'Apply width classes to DrawerContent to make a side drawer wider. The max-width keeps this example within the viewport on smaller screens.',
            },
        },
    },
    render: Default.render,
};

export const ScrollingList: Story = {
    parameters: {
        docs: {
            description: {
                story: 'The default narrow side drawer with enough items to make DrawerBody scroll. The header and Close footer stay visible while the list moves between them.',
            },
        },
    },
    render: ({ className, ...drawerProps }) => (
        <Drawer {...drawerProps}>
            <DrawerTrigger render={<Button variant='secondary' />}>Open scrolling drawer</DrawerTrigger>
            <DrawerContent className={className}>
                <DrawerHeader>
                    <DrawerTitle>Scrollable Content</DrawerTitle>
                    <DrawerClose
                        aria-label='Close drawer'
                        className='rounded-full p-2 focus:outline-none focus-visible:focus-ring'>
                        <X aria-hidden='true' size={20} />
                    </DrawerClose>
                </DrawerHeader>
                <DrawerBody>
                    <DrawerDescription className='mb-6'>
                        Scroll this list inside the narrow drawer. The header and footer stay in place.
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

const DirectionalDrawer = ({
    swipeDirection,
    className,
    modal,
    showSwipeHandle,
    snapPoints,
}: Pick<DrawerStoryArgs, 'swipeDirection' | 'className' | 'modal' | 'showSwipeHandle' | 'snapPoints'>) => (
    <Drawer swipeDirection={swipeDirection} modal={modal} showSwipeHandle={showSwipeHandle} snapPoints={snapPoints}>
        <DrawerTrigger render={<Button variant='secondary' />}>Open {swipeDirection} drawer</DrawerTrigger>
        <DrawerContent className={className}>
            <DrawerHeader>
                <DrawerTitle>{swipeDirection} drawer</DrawerTitle>
                <DrawerClose
                    aria-label='Close drawer'
                    className='rounded-full p-2 focus:outline-none focus-visible:focus-ring'>
                    <X aria-hidden='true' size={20} />
                </DrawerClose>
            </DrawerHeader>
            <DrawerBody>
                <DrawerDescription>Content can come from any edge of the screen.</DrawerDescription>
            </DrawerBody>
        </DrawerContent>
    </Drawer>
);

export const Left: Story = {
    args: { swipeDirection: 'left' },
    parameters: {
        docs: { description: { story: 'Set swipeDirection to left to open the drawer from the left edge.' } },
    },
    render: (args) => (
        <DirectionalDrawer
            swipeDirection={args.swipeDirection ?? 'left'}
            className={args.className}
            modal={args.modal}
            showSwipeHandle={args.showSwipeHandle}
            snapPoints={args.snapPoints}
        />
    ),
};

export const Top: Story = {
    args: { swipeDirection: 'up' },
    parameters: {
        docs: { description: { story: 'Set swipeDirection to up to open a full-width drawer from the top edge.' } },
    },
    render: (args) => (
        <DirectionalDrawer
            swipeDirection={args.swipeDirection ?? 'up'}
            className={args.className}
            modal={args.modal}
            showSwipeHandle={args.showSwipeHandle}
            snapPoints={args.snapPoints}
        />
    ),
};

export const Bottom: Story = {
    args: { swipeDirection: 'down' },
    parameters: {
        docs: {
            description: { story: 'Set swipeDirection to down to open a full-width drawer from the bottom edge.' },
        },
    },
    render: (args) => (
        <DirectionalDrawer
            swipeDirection={args.swipeDirection ?? 'down'}
            className={args.className}
            modal={args.modal}
            showSwipeHandle={args.showSwipeHandle}
            snapPoints={args.snapPoints}
        />
    ),
};

export const SnapPoints: Story = {
    args: { swipeDirection: 'down', snapPoints: [0.4, 1], showSwipeHandle: true },
    parameters: {
        docs: {
            description: {
                story: 'Drag the handle on this bottom drawer to switch between 40% and full viewport height.',
            },
        },
    },
    render: ({ className, ...drawerProps }) => (
        <Drawer {...drawerProps}>
            <DrawerTrigger render={<Button variant='secondary' />}>Open snap drawer</DrawerTrigger>
            <DrawerContent className={className}>
                <DrawerHeader>
                    <div>
                        <DrawerTitle>Snap points</DrawerTitle>
                        <DrawerDescription>Drag the drawer to expand it.</DrawerDescription>
                    </div>
                    <DrawerClose
                        aria-label='Close drawer'
                        className='rounded-full p-2 focus:outline-none focus-visible:focus-ring'>
                        <X aria-hidden='true' size={20} />
                    </DrawerClose>
                </DrawerHeader>
                <DrawerBody>
                    <ul className='divide-y'>
                        {Array.from({ length: 20 }, (_, index) => (
                            <li className='py-3' key={index}>
                                Item {index + 1}
                            </li>
                        ))}
                    </ul>
                </DrawerBody>
            </DrawerContent>
        </Drawer>
    ),
};
