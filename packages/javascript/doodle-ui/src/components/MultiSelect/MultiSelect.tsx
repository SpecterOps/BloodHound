// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// SPDX-License-Identifier: Apache-2.0
import { cva } from 'class-variance-authority';
import { Search } from 'lucide-react';
import * as React from 'react';
import { Checkbox } from '../Checkbox';
import { Input } from '../Input';
import { Popover, PopoverContent, PopoverTrigger } from '../Popover';
import { ScrollArea } from '../ScrollArea';
import { Skeleton } from '../Skeleton';
import { Typography } from '../Typography';
import { cn } from '../utils';

// TODO: move to doodle-ui icons system in follow-up ticket
const CaretDown = ({ className, size = 12 }: { className?: string; size?: number }) => (
    <svg
        width={size}
        height={size}
        viewBox='0 0 10 6'
        fill='currentColor'
        xmlns='http://www.w3.org/2000/svg'
        className={className}>
        <path d='M0 0.977539L5 5.97754L10 0.977539H0Z' />
    </svg>
);

const MultiSelectTriggerVariants = cva(
    'flex h-10 w-full items-center justify-between rounded-lg bg-primary px-4 py-2 text-base font-normal leading-6 tracking-[0.15px] text-common-white dark:text-neutral-dark-1 focus:outline-none focus-visible:focus-ring data-[state=open]:bg-primary enabled:hover:bg-secondary disabled:cursor-not-allowed disabled:rounded disabled:border disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-text-disabled aria-[invalid=true]:border-status-error-main'
);

const multiSelectRowStyles =
    'flex w-full items-center gap-2 rounded-lg p-2 cursor-pointer hover:bg-secondary hover:text-common-white dark:hover:text-neutral-dark-1';

interface MultiSelectTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    open?: boolean;
}

const MultiSelectTrigger = React.forwardRef<HTMLButtonElement, MultiSelectTriggerProps>(
    ({ className, open, children, ...props }, ref) => (
        <button ref={ref} type='button' className={cn(MultiSelectTriggerVariants({ className }))} {...props}>
            <span className='truncate'>{children}</span>
            <CaretDown className={cn('ml-2 shrink-0 transition-transform duration-200', open && 'rotate-180')} />
        </button>
    )
);
MultiSelectTrigger.displayName = 'MultiSelectTrigger';

interface MultiSelectProps {
    options: MultiSelectOption[];
    value: string[];
    onValueChange: (values: string[]) => void;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean;
    className?: string;
    selectAllLabel?: string;
    isSearchable?: boolean;
    isLoading?: boolean;
    searchPlaceholder?: string;
    loadingText?: string;
    emptyText?: string;
    noResultsText?: string;
}

interface MultiSelectOption {
    value: string;
    label: string;
    disabled?: boolean;
}

interface MultiSelectOptionRowProps {
    option: MultiSelectOption;
    checked: boolean;
    onSelect: (value: string) => void;
}

interface MultiSelectActionRowProps {
    checked: React.ComponentProps<typeof Checkbox>['checked'];
    label: string;
    onSelect: () => void;
}

const MultiSelectOptionRow = ({ option, checked, onSelect }: MultiSelectOptionRowProps) => (
    <div
        role='option'
        aria-selected={checked === true}
        aria-disabled={option.disabled}
        tabIndex={option.disabled ? -1 : 0}
        className={cn(
            multiSelectRowStyles,
            option.disabled && 'cursor-not-allowed opacity-50 hover:bg-transparent hover:text-inherit'
        )}
        onClick={() => !option.disabled && onSelect(option.value)}
        onKeyDown={(event) => {
            if (!option.disabled && (event.key === 'Enter' || event.key === ' ')) {
                event.preventDefault();
                onSelect(option.value);
            }
        }}>
        <Checkbox
            tabIndex={-1}
            checked={checked}
            disabled={option.disabled}
            onClick={(event) => event.stopPropagation()}
            onCheckedChange={() => !option.disabled && onSelect(option.value)}
        />
        <span className='truncate text-sm'>{option.label}</span>
    </div>
);

const MultiSelectActionRow = ({ checked, label, onSelect }: MultiSelectActionRowProps) => (
    <div
        role='option'
        aria-selected={checked === true}
        tabIndex={0}
        className={multiSelectRowStyles}
        onClick={onSelect}
        onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault();
                onSelect();
            }
        }}>
        <Checkbox
            tabIndex={-1}
            checked={checked}
            onClick={(event) => event.stopPropagation()}
            onCheckedChange={onSelect}
        />
        <span className='truncate text-sm min-w-0'>{label}</span>
    </div>
);

const MultiSelectLoadingRows = ({ loadingText }: { loadingText: string }) => (
    <div role='status' aria-label={loadingText}>
        {Array.from({ length: 6 }).map((_, index) => (
            <div key={index} className='flex items-center gap-2 mx-1 p-2 py-2'>
                <Checkbox tabIndex={-1} disabled />
                <Skeleton className='h-4 flex-1' />
            </div>
        ))}
    </div>
);

const MultiSelectStateRow = ({ children }: { children: React.ReactNode }) => (
    <Typography variant='body2' component='div' className='px-3 py-2 text-neutral-5'>
        {children}
    </Typography>
);

/**
 * Description for MultiSelect
 */
const MultiSelect = ({
    options,
    value,
    onValueChange,
    placeholder,
    disabled,
    error,
    className,
    selectAllLabel,

    isSearchable = false,
    isLoading = false,
    searchPlaceholder = 'Search',
    loadingText = 'Loading options',
    emptyText = 'No options available.',
    noResultsText = 'No matches',
}: MultiSelectProps) => {
    const [open, setOpen] = React.useState(false);
    const [searchValue, setSearchValue] = React.useState('');

    const handleSelect = (selectedValue: string) => {
        const isSelected = value.includes(selectedValue);

        if (isSelected) {
            const updatedValues = value.filter((currentValue) => currentValue !== selectedValue);
            onValueChange(updatedValues);
            return;
        }

        const updatedValues = [...value, selectedValue];
        onValueChange(updatedValues);
    };

    const getMultiSelectTriggerText = (
        options: MultiSelectOption[],
        value: string[],
        placeholder = 'Select options'
    ) => {
        if (value.length === 0) return placeholder;

        if (value.length === 1) {
            const selectedOption = options.find((option) => option.value === value[0]);
            return selectedOption?.label ?? value[0];
        }

        return `${value.length} Selected`;
    };

    const triggerText = getMultiSelectTriggerText(options, value, placeholder);

    // select all/clear
    const selectableValues = options.filter((option) => !option.disabled).map((option) => option.value);

    const selectedValueCount = selectableValues.filter((optionValue) => value.includes(optionValue)).length;

    const hasSelectedAllOptions = selectableValues.length > 0 && selectedValueCount === selectableValues.length;

    const hasSelectedSomeOptions = selectedValueCount > 0 && !hasSelectedAllOptions;

    const selectAllChecked = hasSelectedAllOptions ? true : hasSelectedSomeOptions ? 'indeterminate' : false;

    const handleSelectAll = () => {
        if (hasSelectedAllOptions) {
            const updatedValues = value.filter((selectedValue) => !selectableValues.includes(selectedValue));
            onValueChange(updatedValues);
            return;
        }

        const unselectedValues = selectableValues.filter((optionValue) => !value.includes(optionValue));
        const updatedValues = [...value, ...unselectedValues];

        onValueChange(updatedValues);
    };

    //search
    const searchTerm = searchValue.trim().toLowerCase();

    const filteredOptions = searchTerm
        ? options.filter((option) => option.label.toLowerCase().includes(searchTerm))
        : options;

    const hasOptions = options.length > 0;
    const hasSearchText = searchTerm.length > 0;
    const hasSearchResults = filteredOptions.length > 0;

    const isSearchResultEmpty = isSearchable && hasSearchText && !hasSearchResults;

    let listContent: React.ReactNode;

    if (isLoading) {
        listContent = <MultiSelectLoadingRows loadingText={loadingText} />;
    } else if (!hasOptions) {
        listContent = <MultiSelectStateRow>{emptyText}</MultiSelectStateRow>;
    } else if (isSearchResultEmpty) {
        listContent = <MultiSelectStateRow>{noResultsText}</MultiSelectStateRow>;
    } else {
        listContent = filteredOptions.map((option) => (
            <MultiSelectOptionRow
                key={option.value}
                option={option}
                checked={value.includes(option.value)}
                onSelect={handleSelect}
            />
        ));
    }
    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <MultiSelectTrigger
                    open={open}
                    disabled={disabled}
                    aria-invalid={error || undefined}
                    className={className}>
                    {triggerText}
                </MultiSelectTrigger>
            </PopoverTrigger>
            <PopoverContent
                className='w-[var(--radix-popover-trigger-width)] rounded-lg bg-neutral-light-1 dark:bg-neutral-dark-2 p-0 text-text-main dark:text-white'
                align='start'>
                {isSearchable && (
                    <div className='flex items-center gap-2 border-b p-2'>
                        <Search className='size-5 shrink-0 text-neutral-dark-1 dark:text-white' aria-hidden='true' />
                        <Input
                            aria-label={searchPlaceholder}
                            value={searchValue}
                            onChange={(e) => setSearchValue(e.target.value)}
                            placeholder={searchPlaceholder}
                            className='h-8 border-none bg-transparent px-0 text-text-main placeholder:text-text-main dark:text-white dark:placeholder:text-white focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:outline-none'
                        />
                    </div>
                )}

                <ScrollArea className='h-60'>
                    <div role='listbox' aria-multiselectable='true' aria-busy={isLoading || undefined} className='p-1'>
                        {selectAllLabel && !isLoading && selectableValues.length > 0 && (
                            <MultiSelectActionRow
                                checked={selectAllChecked}
                                label={selectAllLabel}
                                onSelect={handleSelectAll}
                            />
                        )}

                        {listContent}
                    </div>
                </ScrollArea>
            </PopoverContent>
        </Popover>
    );
};

export { MultiSelect, MultiSelectOptionRow, MultiSelectTrigger };
export type { MultiSelectOption, MultiSelectProps };
