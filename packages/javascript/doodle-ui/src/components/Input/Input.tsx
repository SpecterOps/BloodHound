import { cva, VariantProps } from 'class-variance-authority';
import { cn } from 'components/utils';
import * as React from 'react';

export const InputVariants = cva(
    'flex h-10 w-full text-base text-neutral-dark-1 dark:text-neutral-light-1 disabled:cursor-not-allowed disabled:opacity-50 file:border-0 file:bg-transparent file:pr-3 file:text-sm file:font-medium file:text-neutral-dark-0 dark:file:text-neutral-light-1 file:cursor-pointer',
    {
        variants: {
            variant: {
                outlined:
                    'rounded-md ring-1 ring-neutral-dark-5 dark:ring-neutral-light-5 bg-neutral-2 px-3 py-2 text-sm ring-offset-secondary dark:ring-offset-secondary-variant-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary dark:focus-visible:ring-secondary-variant-2 focus-visible:ring-offset-2 hover:ring-2',
                underlined:
                    'rounded-none bg-transparent border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b focus-visible:outline-none focus:border-t-0 focus:border-x-0 focus-visible:ring-offset-0 focus-visible:ring-transparent focus-visible:border-secondary focus-visible:border-b-2 focus:border-secondary focus:border-b-2 dark:focus-visible:outline-none dark:focus:border-t-0 dark:focus:border-x-0 dark:focus-visible:ring-offset-0 dark:focus-visible:ring-transparent dark:focus-visible:border-secondary-variant-2 dark:focus-visible:border-b-2 dark:focus:border-secondary-variant-2 dark:focus:border-b-2 hover:border-b-2',
            },
        },
        defaultVariants: {
            variant: 'underlined',
        },
    }
);

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement>, VariantProps<typeof InputVariants> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(({ className, type, variant, ...props }, ref) => {
    return <input type={type} className={cn(InputVariants({ variant, className }))} ref={ref} {...props} />;
});
Input.displayName = 'Input';

export { Input };
