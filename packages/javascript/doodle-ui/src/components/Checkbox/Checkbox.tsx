import type { CheckedState } from '@radix-ui/react-checkbox';
import * as CheckboxPrimitive from '@radix-ui/react-checkbox';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import { Check } from 'lucide-react';
import * as React from 'react';

const CheckboxVariants = cva(
    'peer shrink-0 rounded-sm border-2 border-neutral-dark-1 dark:border-neutral-light-1 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-neutral-dark-1 data-[state=checked]:text-neutral-light-1 dark:data-[state=checked]:bg-neutral-light-1 dark:data-[state=checked]:text-neutral-dark-1',
    {
        variants: {
            size: {
                lg: 'h-[24px] w-[24px]',
                md: 'h-[18px] w-[18px]',
                sm: 'h-[12px] w-[12px]',
            },
        },
    }
);

interface CheckboxProps
    extends React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>,
        VariantProps<typeof CheckboxVariants> {
    icon?: React.ReactNode;
}

const Checkbox = React.forwardRef<React.ElementRef<typeof CheckboxPrimitive.Root>, CheckboxProps>(
    ({ size = 'md', icon, className, ...props }, ref) => (
        <CheckboxPrimitive.Root ref={ref} className={cn(CheckboxVariants({ size, className }))} {...props}>
            <CheckboxPrimitive.Indicator className={cn('flex items-center justify-center text-current')}>
                {icon ? icon : <Check className='h-full w-full' absoluteStrokeWidth={true} strokeWidth={3} />}
            </CheckboxPrimitive.Indicator>
        </CheckboxPrimitive.Root>
    )
);
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox, CheckedState };
