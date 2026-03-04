import { cn } from 'components/utils';
import * as React from 'react';

const Skeleton = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => (
        <div
            ref={ref}
            className={cn('animate-pulse bg-neutral-light-5 dark:bg-neutral-dark-5 rounded-md', className)}
            {...props}
        />
    )
);

Skeleton.displayName = 'Skeleton';

export { Skeleton };
