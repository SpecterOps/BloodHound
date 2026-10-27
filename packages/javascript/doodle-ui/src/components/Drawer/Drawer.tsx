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
import { faTimes, type IconDefinition } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { createContext, useContext, useMemo, type ComponentProps } from 'react';
import { IconButton } from '../IconButton';
import { TypographyVariants } from '../Typography';
import { cn, cnWithState } from '../utils';
import {
    drawerBodyClasses,
    drawerDescriptionClasses,
    drawerFooterClasses,
    drawerHeaderClasses,
    drawerInnerContentClasses,
    drawerOverlayClasses,
    drawerPopupClasses,
    drawerSwipeHandleClasses,
    drawerViewportClasses,
} from './Drawer.styles';

type DrawerContextProps = {
    hasSnapPoints: boolean;
    modal: DrawerPrimitive.Root.Props['modal'];
    showSwipeHandle: boolean;
    swipeDirection: NonNullable<DrawerPrimitive.Root.Props['swipeDirection']>;
};

const DrawerContext = createContext<DrawerContextProps | null>(null);

function useDrawer() {
    const context = useContext(DrawerContext);

    if (!context) {
        throw new Error('useDrawer must be used within a Drawer.');
    }

    return context;
}

function Drawer({
    modal = true,
    showSwipeHandle = false,
    snapPoints,
    swipeDirection = 'right',
    ...props
}: DrawerPrimitive.Root.Props & {
    showSwipeHandle?: boolean;
}) {
    const hasSnapPoints = snapPoints != null && snapPoints.length > 0;
    const contextValue = useMemo(
        () => ({ hasSnapPoints, modal, showSwipeHandle, swipeDirection }),
        [hasSnapPoints, modal, showSwipeHandle, swipeDirection]
    );

    return (
        <DrawerContext.Provider value={contextValue}>
            <DrawerPrimitive.Root
                data-slot='drawer'
                modal={modal}
                snapPoints={snapPoints}
                swipeDirection={swipeDirection}
                {...props}
            />
        </DrawerContext.Provider>
    );
}

function DrawerTrigger({ ...props }: DrawerPrimitive.Trigger.Props) {
    return <DrawerPrimitive.Trigger data-slot='drawer-trigger' {...props} />;
}

function DrawerPortal({ ...props }: DrawerPrimitive.Portal.Props) {
    return <DrawerPrimitive.Portal data-slot='drawer-portal' {...props} />;
}

function DrawerClose({ ...props }: DrawerPrimitive.Close.Props) {
    return <DrawerPrimitive.Close data-slot='drawer-close' {...props} />;
}

type DrawerCloseIconProps = Omit<
    DrawerPrimitive.Close.Props & { iconDefinition?: IconDefinition },
    'children' | 'render'
>;
function DrawerCloseIcon({ iconDefinition = faTimes, ...props }: DrawerCloseIconProps) {
    return (
        <DrawerPrimitive.Close
            data-slot='drawer-close'
            render={
                <IconButton size={20} aria-label={props['aria-label'] ?? 'Close drawer'}>
                    <FontAwesomeIcon aria-hidden='true' icon={iconDefinition} />
                </IconButton>
            }
            {...props}
        />
    );
}

function DrawerOverlay({ className, ...props }: DrawerPrimitive.Backdrop.Props) {
    return (
        <DrawerPrimitive.Backdrop
            data-slot='drawer-overlay'
            className={cnWithState(drawerOverlayClasses, className)}
            {...props}
        />
    );
}

function DrawerSwipeHandle({ className, ...props }: ComponentProps<'div'>) {
    return (
        <div
            data-slot='drawer-swipe-handle'
            aria-hidden='true'
            className={cn(drawerSwipeHandleClasses, className)}
            {...props}
        />
    );
}

function DrawerContent({ className, children, ...props }: DrawerPrimitive.Popup.Props) {
    const { hasSnapPoints, modal, showSwipeHandle, swipeDirection } = useDrawer();
    const swipeAxis = swipeDirection === 'down' || swipeDirection === 'up' ? 'y' : 'x';

    return (
        <DrawerPortal data-slot='drawer-portal'>
            {modal === true && <DrawerOverlay data-snap-points={hasSnapPoints ? '' : undefined} />}
            <DrawerPrimitive.Viewport data-slot='drawer-viewport' data-modal={modal} className={drawerViewportClasses}>
                <DrawerPrimitive.Popup
                    data-slot='drawer-popup'
                    data-swipe-axis={swipeAxis}
                    data-snap-points={hasSnapPoints ? '' : undefined}
                    className={cnWithState(drawerPopupClasses, className)}
                    {...props}>
                    {showSwipeHandle && <DrawerSwipeHandle />}
                    <DrawerPrimitive.Content data-slot='drawer-content' className={drawerInnerContentClasses}>
                        {children}
                    </DrawerPrimitive.Content>
                </DrawerPrimitive.Popup>
            </DrawerPrimitive.Viewport>
        </DrawerPortal>
    );
}

function DrawerHeader({ className, ...props }: ComponentProps<'div'>) {
    return <div data-slot='drawer-header' className={cn(drawerHeaderClasses, className)} {...props} />;
}

function DrawerFooter({ className, ...props }: ComponentProps<'div'>) {
    return <div data-slot='drawer-footer' className={cn(drawerFooterClasses, className)} {...props} />;
}

function DrawerTitle({ className, ...props }: DrawerPrimitive.Title.Props) {
    return (
        <DrawerPrimitive.Title
            data-slot='drawer-title'
            className={cnWithState('cn-drawer-title', TypographyVariants({ variant: 'h2' }), className)}
            {...props}
        />
    );
}

function DrawerDescription({ className, ...props }: DrawerPrimitive.Description.Props) {
    return (
        <DrawerPrimitive.Description
            data-slot='drawer-description'
            className={cnWithState(drawerDescriptionClasses, className)}
            {...props}
        />
    );
}

function DrawerBody({ className, ...props }: ComponentProps<'div'>) {
    return (
        <div
            data-slot='drawer-body'
            // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
            tabIndex={0} // Make the body focusable to allow keyboard scrolling
            className={cn(drawerBodyClasses, className)}
            {...props}
        />
    );
}

export {
    Drawer,
    DrawerBody,
    DrawerClose,
    DrawerCloseIcon,
    DrawerContent,
    DrawerDescription,
    DrawerFooter,
    DrawerHeader,
    DrawerOverlay,
    DrawerPortal,
    DrawerSwipeHandle,
    DrawerTitle,
    DrawerTrigger,
};
