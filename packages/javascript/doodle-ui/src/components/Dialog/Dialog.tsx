import * as DialogPrimitive from '@radix-ui/react-dialog';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import * as React from 'react';

/**
 * See documentation: [Dialog](https://www.radix-ui.com/primitives/docs/components/dialog#root)
 */
const Dialog = DialogPrimitive.Root;

/**
 * See documentation: [DialogTrigger](https://www.radix-ui.com/primitives/docs/components/dialog#trigger)
 */
const DialogTrigger = DialogPrimitive.Trigger;

/**
 * See documentation: [DialogPortal](https://www.radix-ui.com/primitives/docs/components/dialog#overlay)
 */
const DialogPortal = DialogPrimitive.Portal;

/**
 * See documentation: [DialogClose](https://www.radix-ui.com/primitives/docs/components/dialog#close)
 */
const DialogClose = DialogPrimitive.Close;

const DialogOverlayVariants = cva(
    'fixed inset-0 z-50 bg-black/40 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
    { variants: { blurBackground: { true: 'backdrop-blur-sm' } } }
);

interface DialogOverlayProps
    extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>,
        VariantProps<typeof DialogOverlayVariants> {}

/**
 * See documentation: [DialogOverlay](https://www.radix-ui.com/primitives/docs/components/dialog#overlay)
 */
const DialogOverlay = React.forwardRef<React.ElementRef<typeof DialogPrimitive.Overlay>, DialogOverlayProps>(
    ({ blurBackground, className, ...props }, ref) => (
        <DialogPrimitive.Overlay
            ref={ref}
            className={cn(DialogOverlayVariants({ blurBackground, className }))}
            {...props}
        />
    )
);
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName;

export const DialogMaxWidth = ['xl', 'lg', 'md', 'sm', 'xs'] as const;
interface DialogContentProps extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> {
    maxWidth?: (typeof DialogMaxWidth)[number];
    DialogOverlayProps?: DialogOverlayProps;
}

/**
 * See documentation: [DialogContent](https://www.radix-ui.com/primitives/docs/components/dialog#content)
 */
const DialogContent = React.forwardRef<React.ElementRef<typeof DialogPrimitive.Content>, DialogContentProps>(
    ({ DialogOverlayProps, maxWidth = 'sm', className, children, ...props }, ref) => {
        // Where do these magic values come from? They match MUIs maxWidth values
        const maxWidthMap: Record<(typeof DialogMaxWidth)[number], string> = {
            xl: 'max-w-[1536px]',
            lg: 'max-w-[1200px]',
            md: 'max-w-[900px]',
            sm: 'max-w-[600px]',
            xs: 'max-w-[444px]',
        };
        const maxWidthClass = maxWidth ? maxWidthMap[maxWidth] : '';

        return (
            <>
                <DialogOverlay {...DialogOverlayProps} />
                <DialogPrimitive.Content
                    ref={ref}
                    className={cn(
                        'fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] rounded-md bg-neutral-light-2 dark:bg-neutral-dark-2 dark:text-neutral-light-1',
                        maxWidthClass,
                        className
                    )}
                    {...props}>
                    {children}
                </DialogPrimitive.Content>
            </>
        );
    }
);
DialogContent.displayName = DialogPrimitive.Content.displayName;

const DialogActions = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
    <div className={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)} {...props} />
);
DialogActions.displayName = 'DialogActions';

/**
 * See documentation: [DialogTitle](https://www.radix-ui.com/primitives/docs/components/dialog#title)
 */
const DialogTitle = React.forwardRef<
    React.ElementRef<typeof DialogPrimitive.Title>,
    React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
    <DialogPrimitive.Title ref={ref} className={cn('flex text-left font-bold', className)} {...props} />
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
