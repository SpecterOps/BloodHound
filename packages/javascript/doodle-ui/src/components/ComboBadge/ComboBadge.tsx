import { cva } from 'class-variance-authority';
import { cn } from 'components/utils';

const ComboBadgeVariants = cva('relative leading-[1] z-10 h-8 text-nowrap', {
    variants: {
        type: {
            slideLeft: 'clip-right-rounded flex flex-row-reverse',
            slideRight: 'clip-left-rounded',
            inlineSlideLeft: 'clip-right-rounded flex items-center flex-row flex-row-reverse',
            inlineSlideRight: 'clip-left-rounded flex items-center flex-row',
        },
    },
});

const AdornmentVariants = cva(
    'bg-primary text-neutral-light-1 rounded px-4 py-2 z-10 transition-[transform] duration-300',
    {
        variants: {
            type: {
                slideLeft: 'absolute left-2 bottom-0 top-0 peer-hover:-translate-x-full',
                slideRight: 'absolute right-2 bottom-0 top-0 peer-hover:translate-x-full',
                inlineSlideLeft: 'inline-block peer-hover:translate-x-0 translate-x-full -mr-2',
                inlineSlideRight: 'inline-block peer-hover:translate-x-0 -translate-x-full -ml-2',
            },
            displayAdornment: {
                true: 'translate-x-0 peer-hover:translate-x-0',
            },
        },
        compoundVariants: [
            {
                type: ['slideLeft'],
                displayAdornment: true,
                className: 'inline-block relative left-0 -mr-2',
            },
            {
                type: ['slideRight'],
                displayAdornment: true,
                className: 'inline-block relative right-0 -ml-2',
            },
        ],
    }
);

interface ComboBadgeProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'aria-label'> {
    label: string | JSX.Element;
    adornment?: string | JSX.Element;
    /** An arial label is required because of the hidden adornment. Please provide context as to what this badge is conveying to a visual user  */
    ariaLabel: string;
    /** slideLeft by default */
    type?: 'slideLeft' | 'slideRight' | 'inlineSlideLeft' | 'inlineSlideRight';
    /** false by default, displays the adornment without hovering over its peer */
    displayAdornment?: boolean;
    disableAdornment?: boolean;
}

const ComboBadge = (props: ComboBadgeProps) => {
    const {
        label,
        adornment,
        ariaLabel,
        type = 'slideLeft',
        displayAdornment = false,
        disableAdornment = false,
        className,
        ...rest
    } = props;

    return (
        <div role='status' aria-label={ariaLabel} className={cn(ComboBadgeVariants({ type }), className)} {...rest}>
            <div
                className={cn(
                    'bg-neutral-light-5 text-primary dark:bg-neutral-dark-5 dark:text-neutral-light-1 rounded inline-block px-4 py-2 relative z-20',
                    { peer: !disableAdornment }
                )}>
                {label}
            </div>
            {adornment ? <div className={cn(AdornmentVariants({ type, displayAdornment }))}>{adornment}</div> : null}
        </div>
    );
};

export { ComboBadge };
