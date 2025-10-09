import { cn } from 'components/utils';
import * as React from 'react';

function Textarea({ className, ...props }: React.ComponentProps<'textarea'>) {
    return (
        <textarea
            data-slot='textarea'
            className={cn(
                'resize-none rounded-md border border-neutral-5 dark:text-neutral-light-1 dark:bg-neutral-dark-5 pl-2 w-full hover:border-secondary dark:hover:border-secondary-variant-2 hover:focus-visible:border-transparent dark:hover:focus-visible:border-transparent focus-visible:border-transparent focus-visible:outline-none focus:ring-secondary focus-visible:ring-secondary focus:outline-secondary focus-visible:outline-secondary dark:focus:ring-secondary-variant-2 dark:focus-visible:ring-secondary-variant-2 dark:focus:outline-secondary-variant-2 dark:focus-visible:outline-secondary-variant-2 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm transition-[color,box-shadow] aria-invalid:ring-error/20 dark:aria-invalid:ring-error/40 aria-invalid:border-error',
                className
            )}
            {...props}
        />
    );
}

export { Textarea };
