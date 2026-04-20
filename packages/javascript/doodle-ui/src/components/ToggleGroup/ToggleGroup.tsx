import * as ToggleGroupPrimitive from '@radix-ui/react-toggle-group';
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { ToggleVariants } from '../Toggle';
import { cn } from '../utils';

const ToggleGroupContext = React.createContext<VariantProps<typeof ToggleVariants>>({
    size: 'lg',
});

const ToggleGroup = React.forwardRef<
    React.ElementRef<typeof ToggleGroupPrimitive.Root>,
    React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Root> & VariantProps<typeof ToggleVariants>
>(({ className, size, children, ...props }, ref) => (
    <ToggleGroupPrimitive.Root ref={ref} className={cn('flex items-center justify-center gap-1', className)} {...props}>
        <ToggleGroupContext.Provider value={{ size }}>{children}</ToggleGroupContext.Provider>
    </ToggleGroupPrimitive.Root>
));

ToggleGroup.displayName = ToggleGroupPrimitive.Root.displayName;

const ToggleGroupItem = React.forwardRef<
    React.ElementRef<typeof ToggleGroupPrimitive.Item>,
    React.ComponentPropsWithoutRef<typeof ToggleGroupPrimitive.Item> & VariantProps<typeof ToggleVariants>
>(({ className, children, size, ...props }, ref) => {
    const context = React.useContext(ToggleGroupContext);

    return (
        <ToggleGroupPrimitive.Item
            ref={ref}
            className={cn(
                ToggleVariants({
                    size: context.size || size,
                }),
                className
            )}
            {...props}>
            {children}
        </ToggleGroupPrimitive.Item>
    );
});

ToggleGroupItem.displayName = ToggleGroupPrimitive.Item.displayName;

export { ToggleGroup, ToggleGroupItem };
