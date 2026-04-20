import * as TogglePrimitive from '@radix-ui/react-toggle';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

// TODO: Replace hardcoded hex colors with design token CSS variables once the token system is ready.
const ToggleVariants = cva(
    'inline-flex items-center justify-center rounded-lg text-sm transition-colors gap-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 border border-[#C0C6CB] bg-neutral-1 hover:border-[#4A3BD7] hover:bg-[#4A3BD7] hover:text-neutral-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-[#4A3BD7] focus-visible:ring-offset-neutral-1 focus-visible:border-[#4A3BD7] focus-visible:bg-[#4A3BD7] focus-visible:text-neutral-1 active:border-[#C0C6CB] active:bg-[#C0C6CB] active:text-[#121212] data-[state=on]:border-[#2C2677] data-[state=on]:bg-[#2C2677] data-[state=on]:text-neutral-1 disabled:cursor-not-allowed disabled:border-[#E3E7EA] disabled:bg-[#E3E7EA] disabled:text-[#616161] dark:border-[#121212] dark:bg-[#121212] dark:hover:border-[#66A3FF] dark:hover:bg-[#66A3FF] dark:hover:text-neutral-dark-1 dark:focus-visible:ring-[#66A3FF] dark:focus-visible:ring-offset-neutral-dark-1 dark:focus-visible:border-[#66A3FF] dark:focus-visible:bg-[#66A3FF] dark:focus-visible:text-neutral-dark-1 dark:active:border-[#2C2C2C] dark:active:bg-[#2C2C2C] dark:active:text-white dark:data-[state=on]:border-[#A1A0FF] dark:data-[state=on]:bg-[#A1A0FF] dark:data-[state=on]:text-neutral-dark-1 dark:disabled:border-[#2E2E2E] dark:disabled:bg-[#2E2E2E] dark:disabled:text-[#A6A6A6]',
    {
        variants: {
            size: {
                sm: 'h-6 px-2 py-1',
                lg: 'h-10 px-2',
            },
        },
        defaultVariants: {
            size: 'sm',
        },
    }
);

const Toggle = React.forwardRef<
    React.ElementRef<typeof TogglePrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof TogglePrimitive.Root> & VariantProps<typeof ToggleVariants>
>(({ className, size, ...props }, ref) => (
    <TogglePrimitive.Root ref={ref} className={cn(ToggleVariants({ size, className }))} {...props} />
));

Toggle.displayName = TogglePrimitive.Root.displayName;

export { Toggle, ToggleVariants };
