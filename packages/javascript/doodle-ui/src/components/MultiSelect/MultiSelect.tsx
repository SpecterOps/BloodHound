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
import { Label } from '../Label';
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
    'flex h-10 w-full items-center justify-between rounded bg-primary px-[14px] py-2 text-base font-normal leading-6 tracking-[0.15px] text-text-contrast focus:outline-none focus-visible:focus-ring data-[state=open]:bg-primary enabled:hover:bg-secondary disabled:cursor-not-allowed disabled:border disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-text-disabled aria-[invalid=true]:[&>svg]:text-text-main aria-[invalid=true]:border aria-[invalid=true]:border-status-error-main aria-[invalid=true]:bg-select-trigger-outlined-fill aria-[invalid=true]:text-input-placeholder-text  aria-[invalid=true]:enabled:hover:border-status-error-main aria-[invalid=true]:enabled:hover:bg-select-trigger-outlined-fill aria-[invalid=true]:data-[state=open]:bg-select-trigger-outlined-fill'
);
const multiSelectRowStyles = 'flex w-full items-center gap-2 rounded-lg p-2';

const multiSelectInteractiveRowStyles = 'group cursor-pointer hover:bg-secondary hover:text-text-contrast';

const multiSelectCheckboxStyles = 'enabled:data-[state=unchecked]:!border-current';
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

interface MultiSelectAllOptionsRowProps {
    checked: React.ComponentProps<typeof Checkbox>['checked'];
    label: string;
    onSelect: () => void;
}

const MultiSelectOptionRow = ({ option, checked, onSelect }: MultiSelectOptionRowProps) => {
    const checkboxId = React.useId();

    return (
        <div className='p-1'>
            <Label
                htmlFor={checkboxId}
                className={cn(
                    multiSelectRowStyles,
                    'text-base font-normal leading-4',
                    option.disabled
                        ? 'cursor-not-allowed bg-btn-disabled-fill text-text-disabled'
                        : multiSelectInteractiveRowStyles
                )}>
                <Checkbox
                    id={checkboxId}
                    checked={checked}
                    disabled={option.disabled}
                    onCheckedChange={() => onSelect(option.value)}
                    className={multiSelectCheckboxStyles}
                />
                <span className='min-w-0 flex-1 truncate'>{option.label}</span>
            </Label>
        </div>
    );
};

const MultiSelectAllOptionsRow = ({ checked, label, onSelect }: MultiSelectAllOptionsRowProps) => {
    const checkboxId = React.useId();

    return (
        <div className='p-1'>
            <Label
                htmlFor={checkboxId}
                className={cn(
                    multiSelectRowStyles,
                    multiSelectInteractiveRowStyles,
                    'text-base font-normal leading-4'
                )}>
                <Checkbox
                    id={checkboxId}
                    checked={checked}
                    onCheckedChange={onSelect}
                    className={multiSelectCheckboxStyles}
                />
                <span className='min-w-0 flex-1 truncate'>{label}</span>
            </Label>
        </div>
    );
};

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
    <Typography variant='body1' component='div' className='px-3 py-2 text-center text-text-main'>
        {children}
    </Typography>
);

/**
 * MultiSelect Doodle Component
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
    emptyText = 'No options available',
    noResultsText = 'No matches',
}: MultiSelectProps) => {
    const [open, setOpen] = React.useState(false);
    const [searchValue, setSearchValue] = React.useState('');

    const handleOpenChange = (shouldOpen: boolean) => {
        setOpen(shouldOpen);

        if (!shouldOpen) {
            setSearchValue('');
        }
    };

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
    const enabledOptionValues = options.filter((option) => !option.disabled).map((option) => option.value);

    const selectedEnabledOptionCount = enabledOptionValues.filter((optionValue) => value.includes(optionValue)).length;

    const hasSelectedAllEnabledOptions =
        enabledOptionValues.length > 0 && selectedEnabledOptionCount === enabledOptionValues.length;

    const hasSelectedSomeOptions = selectedEnabledOptionCount > 0 && !hasSelectedAllEnabledOptions;

    const selectAllChecked = hasSelectedAllEnabledOptions ? true : hasSelectedSomeOptions ? 'indeterminate' : false;

    const handleSelectAll = () => {
        if (hasSelectedAllEnabledOptions) {
            const updatedValues = value.filter((selectedValue) => !enabledOptionValues.includes(selectedValue));
            onValueChange(updatedValues);
            return;
        }

        const unselectedValues = enabledOptionValues.filter((optionValue) => !value.includes(optionValue));
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
        <Popover open={open} onOpenChange={handleOpenChange}>
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
                className='w-[var(--radix-popover-trigger-width)] rounded-lg bg-select-content-fill p-0 text-text-main'
                align='start'>
                {isSearchable && (
                    <div className='flex items-center gap-2 p-2'>
                        <Search className='size-4 shrink-0 text-text-main' aria-hidden='true' />
                        <Input
                            aria-label={searchPlaceholder}
                            value={searchValue}
                            onChange={(e) => setSearchValue(e.target.value)}
                            placeholder={searchPlaceholder}
                            className='h-6 border-none bg-transparent px-0 text-text-main leading-4 placeholder:text-text-main focus-visible:outline-none focus-visible:ring-0 focus-visible:ring-offset-0'
                        />
                    </div>
                )}

                <ScrollArea className='max-h-60 [&_[data-radix-scroll-area-viewport]]:max-h-60'>
                    <div
                        role='group'
                        aria-label={placeholder ?? 'Select options'}
                        aria-busy={isLoading || undefined}
                        className='p-1'>
                        {selectAllLabel &&
                            !isLoading &&
                            !hasSearchText &&
                            !isSearchResultEmpty &&
                            enabledOptionValues.length > 0 && (
                                <MultiSelectAllOptionsRow
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
