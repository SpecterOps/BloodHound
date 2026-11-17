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

import * as DialogPrimitive from '@radix-ui/react-dialog';
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { TypographyVariants } from '../Typography';
import { cn } from '../utils';
import {
    dialogActionsClasses,
    dialogBodyClasses,
    dialogContentClasses,
    dialogMaxWidthClasses,
    DialogOverlayVariants,
    dialogTriggerClasses,
} from './Dialog.styles';

/**
 * See documentation: [Dialog](https://www.radix-ui.com/primitives/docs/components/dialog#root)
 */
const Dialog = DialogPrimitive.Root;

/**
 * See documentation: [DialogTrigger](https://www.radix-ui.com/primitives/docs/components/dialog#trigger)
 */
const DialogTrigger = React.forwardRef<
    React.ElementRef<typeof DialogPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof DialogPrimitive.Trigger>
>(({ className, ...props }, ref) => (
    <DialogPrimitive.Trigger ref={ref} className={cn(dialogTriggerClasses, className)} {...props} />
));
DialogTrigger.displayName = DialogPrimitive.Trigger.displayName;

/**
 * See documentation: [DialogPortal](https://www.radix-ui.com/primitives/docs/components/dialog#overlay)
 */
const DialogPortal = DialogPrimitive.Portal;

/**
 * See documentation: [DialogClose](https://www.radix-ui.com/primitives/docs/components/dialog#close)
 */
const DialogClose = React.forwardRef<
    React.ElementRef<typeof DialogPrimitive.Close>,
    React.ComponentPropsWithoutRef<typeof DialogPrimitive.Close>
>(({ className, ...props }, ref) => (
    <DialogPrimitive.Close ref={ref} className={cn(dialogTriggerClasses, className)} {...props} />
));
DialogClose.displayName = DialogPrimitive.Close.displayName;

interface DialogOverlayProps
    extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>,
        VariantProps<typeof DialogOverlayVariants> {}

/**
 * See documentation: [DialogOverlay](https://www.radix-ui.com/primitives/docs/components/dialog#overlay)
 */
const DialogOverlay = React.forwardRef<React.ElementRef<typeof DialogPrimitive.Overlay>, DialogOverlayProps>(
    ({ blurBackground, className, allowNav = false, ...props }, ref) => (
        <DialogPrimitive.Overlay
            ref={ref}
            className={cn(DialogOverlayVariants({ blurBackground, allowNav, className }))}
            {...props}
        />
    )
);
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName;

export const DialogMaxWidth = ['xl', 'lg', 'md', 'sm', 'xs'] as const;
interface DialogContentProps extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> {
    allowNav?: boolean;
    DialogOverlayProps?: DialogOverlayProps;
    maxWidth?: (typeof DialogMaxWidth)[number];
    disableStickyLayout?: boolean;
}

/**
 * See documentation: [DialogContent](https://www.radix-ui.com/primitives/docs/components/dialog#content)
 */
const DialogContent = React.forwardRef<React.ElementRef<typeof DialogPrimitive.Content>, DialogContentProps>(
    (
        {
            allowNav = false,
            children,
            className,
            DialogOverlayProps,
            disableStickyLayout = false,
            maxWidth = 'sm',
            ...props
        },
        ref
    ) => {
        const maxWidthClass = maxWidth ? dialogMaxWidthClasses[maxWidth] : '';

        const titleChildren: React.ReactNode[] = [];
        const actionsChildren: React.ReactNode[] = [];
        const middleChildren: React.ReactNode[] = [];

        // Title and Actions are separated from other content so that they may be "stickied" to top/bottom.
        // All other content is added to its own container so that if it causes overflow, the content is
        // scrollable, but Title and Actions are always visible.
        !disableStickyLayout &&
            React.Children.toArray(children).forEach((child) => {
                if (React.isValidElement(child) && child.type === DialogTitle) {
                    titleChildren.push(child);
                } else if (React.isValidElement(child) && child.type === DialogActions) {
                    actionsChildren.push(child);
                } else {
                    middleChildren.push(child);
                }
            });

        return (
            <>
                <DialogOverlay {...DialogOverlayProps} allowNav={allowNav} />
                <DialogPrimitive.Content
                    ref={ref}
                    className={cn(
                        dialogContentClasses,
                        maxWidthClass,
                        { 'overflow-y-auto': disableStickyLayout },
                        className
                    )}
                    onOpenAutoFocus={() => {
                        if (allowNav) {
                            // unblocks the body from being clickable so the user can go to another tab
                            document.body.style.pointerEvents = '';
                        }
                    }}
                    {...props}>
                    {disableStickyLayout ? (
                        children
                    ) : (
                        <>
                            {titleChildren}
                            {middleChildren.length > 0 && <div className={dialogBodyClasses}>{middleChildren}</div>}
                            {actionsChildren}
                        </>
                    )}
                </DialogPrimitive.Content>
            </>
        );
    }
);
DialogContent.displayName = DialogPrimitive.Content.displayName;

const DialogActions = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div className={cn(dialogActionsClasses, className)} {...props} />
);
DialogActions.displayName = 'DialogActions';

/**
 * See documentation: [DialogTitle](https://www.radix-ui.com/primitives/docs/components/dialog#title)
 */
const DialogTitle = React.forwardRef<
    React.ElementRef<typeof DialogPrimitive.Title>,
    React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
    <DialogPrimitive.Title
        ref={ref}
        className={cn('flex text-left', TypographyVariants({ variant: 'h3' }), className)}
        {...props}
    />
));
DialogTitle.displayName = DialogPrimitive.Title.displayName;

/**
 * See documentation: [DialogDescription](https://www.radix-ui.com/primitives/docs/components/dialog#description)
 */
const DialogDescription = React.forwardRef<
    React.ElementRef<typeof DialogPrimitive.Description>,
    React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>((props, ref) => <DialogPrimitive.Description ref={ref} {...props} />);
DialogDescription.displayName = DialogPrimitive.Description.displayName;

export {
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogPortal,
    DialogTitle,
    DialogTrigger,
};
