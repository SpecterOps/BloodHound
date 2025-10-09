import * as LabelPrimitive from '@radix-ui/react-label';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import * as React from 'react';

const LabelVariants = cva('leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70', {
    variants: {
        size: {
            small: 'text-sm font-medium',
            medium: 'text-base font-bold',
            large: 'text-lg font-bold',
        },
    },
    defaultVariants: {
        size: 'medium',
    },
});

const Label = React.forwardRef<
    React.ElementRef<typeof LabelPrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root> & VariantProps<typeof LabelVariants>
>(({ className, size, ...props }, ref) => (
    <LabelPrimitive.Root ref={ref} className={cn(LabelVariants({ size, className }))} {...props} />
));

Label.displayName = LabelPrimitive.Root.displayName;

export { Label };
