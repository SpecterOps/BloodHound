import { cva } from 'class-variance-authority';
import { ElementType } from 'react';
import { cn } from '../utils';
import { DEFAULT_VARIANT, Variant, variantMapping } from './utils';

export const TypographyVariants = cva('', {
    variants: {
        variant: {
            h1: 'text-[1.8rem] font-normal leading-[1] tracking-normal',
            h2: 'text-2xl font-medium leading-[1.5] tracking-normal',
            h3: 'text-[1.2rem] font-medium leading-[1.25] tracking-normal',
            h4: 'text-xl font-medium leading-[1.5] tracking-normal',
            h5: 'text-xl font-bold leading-[1.5] tracking-[0.25em]',
            h6: 'text-base font-bold leading-[1.5] tracking-[0.25em]',
            body1: 'text-base font-normal leading-[1.5] tracking-[0.00938em',
            body2: 'text-sm font-normal leading-[1.43] tracking-[0.01071em',
            caption: 'text-xs font-normal leading-[1.77] tracking-[0.03333em',
            subtitle1: 'text-base font-normal leading-[1.75] tracking-[0.00938em',
            subtitle2: 'text-sm font-medium leading-[1.57] tracking-[0.00714em',
        },
    },
    defaultVariants: {
        variant: DEFAULT_VARIANT,
    },
});

interface TypographyProps extends React.HTMLAttributes<HTMLElement> {
    variant?: Variant;
    component?: ElementType;
}

const Typography = ({ variant, component, children, className, ...rest }: TypographyProps) => {
    const Tag = (component || variantMapping[variant ?? DEFAULT_VARIANT]) as ElementType;

    return (
        <Tag className={cn(TypographyVariants({ variant }), className)} {...rest}>
            {children}
        </Tag>
    );
};

Typography.displayName = 'Typography';

export { Typography };
