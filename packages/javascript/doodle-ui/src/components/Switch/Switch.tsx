import { cva, VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import { forwardRef } from 'react';
import * as SwitchPrimitives from '@radix-ui/react-switch';

type sizeOptions = 'small' | 'medium' | 'large';

const defaultSize: Record<'size', sizeOptions> = {
    size: 'medium',
};

const SwitchVariants = cva(
    'flex items-center group rounded-3xl transition-all ease-in-out bg-neutral-dark-5 dark:bg-neutral-light-5 disabled:bg-neutral-light-5 disabled:opacity-50 data-[state=checked]:bg-primary dark:data-[state=checked]:bg-primary disabled:cursor-not-allowed focus-visible:outline-none',
    {
        variants: {
            size: {
                small: 'h-2 w-4',
                medium: 'h-3 w-6',
                large: 'h-4 w-8',
            },
        },
        defaultVariants: defaultSize,
    }
);

const ThumbVariants = cva(
    'transition-all ease-in-out rounded-full bg-neutral-light-2 ring-primary shadow-outer-1 group-hover:group-enabled:ring-1 group-focus:ring-1 group-focus-visible:ring-1 data-[state=checked]:shadow-outer-2 data-[state=checked]:ring-1 data-[state=checked]:hover:ring-1 translate-x-px',
    {
        variants: {
            size: {
                small: 'h-1.5 w-1.5 data-[state=checked]:translate-x-[9px]',
                medium: 'h-2.5 w-2.5 data-[state=checked]:translate-x-[13px]',
                large: 'h-3.5 w-3.5 data-[state=checked]:translate-x-[17px]',
            },
        },
        defaultVariants: defaultSize,
    }
);

const LabelVariants = cva('transition-all ease-in-out text-neutral-dark-5 dark:text-white hover:cursor-pointer', {
    variants: {
        size: {
            small: 'text-sm',
            medium: 'text-base',
            large: 'text-lg',
        },
        position: {
            left: 'mr-3',
            right: 'ml-3',
        },
    },
    defaultVariants: {
        size: 'medium',
        position: 'right',
    },
    compoundVariants: [
        {
            size: 'small',
            position: 'left',
            className: 'mr-2',
        },
        {
            size: 'medium',
            position: 'left',
            className: 'mr-3',
        },
        {
            size: 'large',
            position: 'left',
            className: 'mr-4',
        },
        {
            size: 'small',
            position: 'right',
            className: 'ml-2',
        },
        {
            size: 'medium',
            position: 'left',
            className: 'ml-3',
        },
        {
            size: 'large',
            position: 'left',
            className: 'ml-4',
        },
    ],
});

const Switch = forwardRef<
    React.ElementRef<typeof SwitchPrimitives.Root>,
    React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root> &
        VariantProps<typeof SwitchVariants> & {
            label?: string;
            labelPosition?: 'left' | 'right';
        }
>(({ className, size, label, labelPosition, ...props }, ref) => {
    const ariaLabel = label || 'switch';

    return (
        <div className={cn('flex items-center')}>
            {labelPosition && labelPosition === 'left' && (
                <label
                    className={cn(LabelVariants({ size: size, position: labelPosition }), label ? 'visible' : 'hidden')}
                    aria-hidden={!label}
                    htmlFor={ariaLabel}>
                    {ariaLabel}
                </label>
            )}
            <SwitchPrimitives.Root
                className={cn(SwitchVariants({ size: size }), className)}
                aria-label={ariaLabel}
                id={ariaLabel}
                {...props}
                ref={ref}>
                <SwitchPrimitives.Thumb className={cn(ThumbVariants({ size: size }))} />
            </SwitchPrimitives.Root>
            {labelPosition !== 'left' && (
                <label
                    className={cn(LabelVariants({ size: size, position: labelPosition }), label ? 'visible' : 'hidden')}
                    aria-hidden={!label}
                    htmlFor={ariaLabel}>
                    {ariaLabel}
                </label>
            )}
        </div>
    );
});

export { Switch };
