import * as TogglePrimitive from '@radix-ui/react-toggle';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

const toggleVariants = cva(
    'inline-flex items-center justify-center rounded-lg text-sm ring-offset-background transition-colors active:border-[#C0C6CB] active:bg-[#C0C6CB] active:text-[#121212] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:border-[#E3E7EA] disabled:bg-[#E3E7EA] disabled:text-[#616161] [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 gap-2',
    {
        variants: {
            variant: {
                default:
                    'border border-[#C0C6CB] bg-white hover:border-[#4A3BD7] hover:bg-[#4A3BD7] hover:text-white data-[state=on]:border-[#2C2677] data-[state=on]:bg-[#2C2677] data-[state=on]:text-white',
                // outline: 'border border-input bg-transparent hover:bg-accent hover:text-accent-foreground',
            },
            size: {
                default: 'h-10 px-2 py-1 min-w-12',
                sm: 'h-8 px-2 min-w-8',
                lg: 'h-10 px-2 min-w-12',
            },
        },
        defaultVariants: {
            variant: 'default',
            size: 'lg',
        },
    }
);

const Toggle = React.forwardRef<
    React.ElementRef<typeof TogglePrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof TogglePrimitive.Root> & VariantProps<typeof toggleVariants>
>(({ className, variant, size, ...props }, ref) => (
    <TogglePrimitive.Root ref={ref} className={cn(toggleVariants({ variant, size, className }))} {...props} />
));

Toggle.displayName = TogglePrimitive.Root.displayName;

export { Toggle, toggleVariants };
