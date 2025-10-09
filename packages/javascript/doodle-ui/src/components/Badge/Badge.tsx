import * as React from 'react';
import { cn } from 'components/utils';

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
    label: string;
    icon?: React.ReactNode;
    color?: string;
    backgroundColor?: string;
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ label, icon, color, backgroundColor, className, ...rest }, ref) => {
        return (
            <div
                ref={ref}
                {...rest}
                className={cn([
                    'inline-flex items-center justify-center rounded min-w-16 h-8 p-2 bg-neutral-light-3 dark:bg-neutral-dark-3 text-neutral-dark-1 dark:text-white border border-neutral-light-5 dark:border-neutral-dark-5',
                    icon && 'pr-3',
                    className,
                ])}
                style={{
                    borderColor: color,
                    backgroundColor: backgroundColor,
                }}>
                {icon && <span style={{ color }}>{icon}</span>}
                {label}
            </div>
        );
    }
);

export { Badge };
