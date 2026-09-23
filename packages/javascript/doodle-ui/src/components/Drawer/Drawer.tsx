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
import { Drawer as DrawerPrimitive } from '@base-ui/react/drawer';
import * as React from 'react';
import { TypographyVariants } from '../Typography';
import { cn } from '../utils';

type DrawerProps = Omit<DrawerPrimitive.Root.Props, 'swipeDirection'>;

const Drawer = (props: DrawerProps) => <DrawerPrimitive.Root swipeDirection='right' {...props} />;

const DrawerTrigger = DrawerPrimitive.Trigger;
const DrawerClose = DrawerPrimitive.Close;

const DrawerContent = React.forwardRef<React.ComponentRef<typeof DrawerPrimitive.Popup>, DrawerPrimitive.Popup.Props>(
    ({ children, className, ...props }, ref) => (
        <DrawerPrimitive.Portal>
            <DrawerPrimitive.Backdrop className='fixed inset-0 z-[1410] bg-black/40 opacity-[calc(1-var(--drawer-swipe-progress))] transition-opacity duration-300 data-[starting-style]:opacity-0 data-[ending-style]:opacity-0 data-[swiping]:transition-none' />
            <DrawerPrimitive.Viewport className='pointer-events-none fixed inset-0 z-[1500] flex justify-end'>
                <DrawerPrimitive.Popup
                    ref={ref}
                    className={cn(
                        'pointer-events-auto flex h-dvh w-full max-w-[860px] flex-col gap-4 overflow-hidden bg-neutral-light-1 pl-3 py-4 text-main shadow-xl dark:bg-neutral-dark-1',
                        '[transform:translateX(var(--drawer-swipe-movement-x))] transition-transform duration-300 ease-out data-[starting-style]:[transform:translateX(100%)] data-[ending-style]:[transform:translateX(100%)] data-[swiping]:transition-none',
                        'focus:outline-none focus-visible:focus-ring',
                        className
                    )}
                    {...props}>
                    <DrawerPrimitive.Content className='contents'>{children}</DrawerPrimitive.Content>
                </DrawerPrimitive.Popup>
            </DrawerPrimitive.Viewport>
        </DrawerPrimitive.Portal>
    )
);
DrawerContent.displayName = 'DrawerContent';

const DrawerHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => (
        <div ref={ref} className={cn('flex shrink-0 items-start justify-between gap-4', className)} {...props} />
    )
);
DrawerHeader.displayName = 'DrawerHeader';

const DrawerTitle = React.forwardRef<React.ComponentRef<typeof DrawerPrimitive.Title>, DrawerPrimitive.Title.Props>(
    ({ className, ...props }, ref) => (
        <DrawerPrimitive.Title ref={ref} className={cn(TypographyVariants({ variant: 'h2' }), className)} {...props} />
    )
);
DrawerTitle.displayName = 'DrawerTitle';

const DrawerDescription = DrawerPrimitive.Description;

const DrawerBody = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => (
        <div ref={ref} className={cn('min-h-0 flex-1 overflow-y-auto pr-3', className)} {...props} />
    )
);
DrawerBody.displayName = 'DrawerBody';

const DrawerFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => <div ref={ref} className={cn('flex shrink-0', className)} {...props} />
);
DrawerFooter.displayName = 'DrawerFooter';

export {
    Drawer,
    DrawerBody,
    DrawerClose,
    DrawerContent,
    DrawerDescription,
    DrawerFooter,
    DrawerHeader,
    DrawerTitle,
    DrawerTrigger,
};
