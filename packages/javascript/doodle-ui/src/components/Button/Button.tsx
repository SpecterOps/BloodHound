import { Slot, SlotProps } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import * as React from 'react';

export const ButtonVariants = cva(
    'inline-flex items-center justify-center whitespace-nowrap h-10 px-6 py-2 rounded-3xl text-sm ring-offset-background transition-colors hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:no-underline',
    {
        variants: {
            variant: {
                primary: 'bg-primary text-white shadow-outer-1 hover:bg-secondary',
                secondary: 'bg-neutral-light-4 text-neutral-dark-1 shadow-outer-1 hover:bg-secondary hover:text-white',
                tertiary:
                    'bg-transparent border-2 border-neutral-light-5 shadow-outer-1 hover:bg-tertiary dark:hover:text-neutral-dark-1 hover:border-tertiary',
                transparent:
                    'bg-transparent border border-neutral-5 dark:text-white hover:bg-primary hover:text-white hover:border-primary hover:no-underline',
                text: 'text-neutral-dark-5 dark:text-neutral-light-5',
                icon: 'rounded-full text-neutral-dark-1 bg-neutral-light-5 p-0 size-10 shadow-outer-1 hover:border-2 hover:border-secondary active:border-none',
            },
            fontColor: {
                primary: 'text-primary dark:text-secondary-variant-2',
            },
            size: {
                small: 'h-9 px-4 py-1 text-xs',
                medium: 'h-10 px-6 py-2',
                large: 'h-11 px-8 py-3 text-base',
            },
        },
        defaultVariants: {
            variant: 'primary',
        },
    }
);

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
        VariantProps<typeof ButtonVariants> {
    asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, fontColor, asChild = false, ...props }, ref) => {
        const defaultType = 'button';

        let Comp: 'button' | React.ForwardRefExoticComponent<SlotProps & React.RefAttributes<HTMLElement>> =
            defaultType;

        if (asChild) {
            Comp = Slot;
        } else {
            if (!props.type) props.type = defaultType;
        }

        return <Comp className={cn(ButtonVariants({ variant, size, fontColor, className }))} ref={ref} {...props} />;
    }
);
Button.displayName = 'Button';

export { Button };
