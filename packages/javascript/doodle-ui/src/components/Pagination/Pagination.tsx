import { forwardRef } from 'react';

import { faChevronLeft, faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { SelectValueProps } from '@radix-ui/react-select';
import { Slot } from '@radix-ui/react-slot';
import { Select, SelectContent, SelectItem, SelectPortal, SelectTrigger, SelectValue } from 'components/Select';
import { cn } from 'components/utils';

const PaginationNav = ({ className, ...props }: React.ComponentProps<'nav'>) => (
    <nav
        role='navigation'
        aria-label='pagination'
        className={cn('mx-auto flex w-full justify-center', className)}
        {...props}
    />
);
PaginationNav.displayName = 'PaginationNav';

const PaginationContent = forwardRef<HTMLDivElement, React.ComponentProps<'div'>>(({ className, ...props }, ref) => (
    <div ref={ref} className={cn('flex flex-row items-center gap-2 dark:text-white', className)} {...props} />
));
PaginationContent.displayName = 'PaginationContent';

type PaginationLinkProps = {
    isActive?: boolean;
    asChild?: boolean;
} & React.ComponentProps<'button'>;

const PaginationItem = forwardRef<HTMLDivElement, React.ComponentProps<'div'>>(({ className, ...props }, ref) => (
    <div ref={ref} className={cn('', className)} {...props} />
));
PaginationItem.displayName = 'PaginationItem';

const PaginationLink = ({ className, isActive, asChild, ...props }: PaginationLinkProps) => {
    const Comp = asChild ? Slot : 'button';

    return <Comp aria-current={isActive ? 'page' : undefined} className={cn(className)} {...props} />;
};
PaginationLink.displayName = 'PaginationLink';

const PaginationPrevious = ({ onClick, className, ...props }: React.ComponentProps<typeof PaginationLink>) => (
    <PaginationLink
        onClick={onClick}
        aria-label='Go to previous page'
        className={cn('gap-1 pl-2.5', className)}
        {...props}>
        <FontAwesomeIcon icon={faChevronLeft} />
    </PaginationLink>
);
PaginationPrevious.displayName = 'PaginationPrevious';

const PaginationNext = ({ onClick, className, ...props }: React.ComponentProps<typeof PaginationLink>) => (
    <PaginationLink onClick={onClick} aria-label='Go to next page' className={cn('gap-1 pr-2.5', className)} {...props}>
        <FontAwesomeIcon icon={faChevronRight} />
    </PaginationLink>
);
PaginationNext.displayName = 'PaginationNext';

const PaginationEllipsis = ({ className, ...props }: React.ComponentProps<'span'>) => (
    <span aria-hidden className={cn('flex h-9 w-9 items-center justify-center', className)} {...props}></span>
);
PaginationEllipsis.displayName = 'PaginationEllipsis';

interface PaginationSelectProps extends React.ComponentProps<typeof Select> {
    options: string[];
    placeholder?: SelectValueProps['placeholder'];
}
const PaginationSelect = (props: PaginationSelectProps) => {
    const { options, placeholder, ...rest } = props;

    const selectItems = options.map((value) => ({ value, display: value }));

    return (
        <Select {...rest}>
            <SelectTrigger className='w-fit'>
                <SelectValue placeholder={placeholder} />
            </SelectTrigger>
            <SelectPortal>
                <SelectContent>
                    {selectItems.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                            {option.display}
                        </SelectItem>
                    ))}
                </SelectContent>
            </SelectPortal>
        </Select>
    );
};

export interface PaginationProps extends React.ComponentProps<'nav'> {
    page: number;
    rowsPerPage: number;
    count: number;
    paginationOptions?: string[];
    defaultPaginationValue?: string;
    onPageChange: (page: number) => void;
    onRowsPerPageChange: (rowsPerPage: number) => void;
}

const Pagination: React.FC<PaginationProps> = (props) => {
    const {
        page,
        rowsPerPage,
        count,
        paginationOptions,
        defaultPaginationValue,
        onPageChange,
        onRowsPerPageChange,
        className,
        ...rest
    } = props;

    const handlePageChange = (incrementor: number) => () => {
        onPageChange(page + incrementor);
    };

    const handleValueChange = (value: string) => {
        if (isNaN(Number(value))) return;
        onPageChange(1);
        onRowsPerPageChange(Number(value));
    };

    const lowerBoundary = rowsPerPage * page + 1;
    const upperBoundary = rowsPerPage * (page + 1) > count ? count : rowsPerPage * (page + 1);

    const selectItems = paginationOptions ?? ['5', '10', '25', '100'];

    return (
        <PaginationNav className={cn('text-sm tabular-nums', className)} {...rest}>
            <PaginationContent>
                <PaginationItem className='flex justify-center items-center gap-2 px-1 py-2 mx-2'>
                    Rows per page:
                    <PaginationSelect
                        options={selectItems}
                        defaultValue={defaultPaginationValue ?? selectItems[0]}
                        onValueChange={handleValueChange}
                    />
                </PaginationItem>
                <PaginationItem className='px-1 py-2 mx-2'>
                    {lowerBoundary} - {upperBoundary} of {count}
                </PaginationItem>
                <PaginationItem>
                    <PaginationPrevious
                        onClick={handlePageChange(-1)}
                        disabled={lowerBoundary <= 1}
                        className='disabled:text-gray-400 px-1 py-2 mx-2'
                    />
                    <PaginationNext
                        onClick={handlePageChange(1)}
                        disabled={upperBoundary >= count}
                        className='disabled:text-gray-400 px-1 py-2 mx-2'
                    />
                </PaginationItem>
            </PaginationContent>
        </PaginationNav>
    );
};

export {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationLink,
    PaginationNav,
    PaginationNext,
    PaginationPrevious,
};
