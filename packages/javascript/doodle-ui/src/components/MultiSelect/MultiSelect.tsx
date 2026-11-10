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

import { Search } from 'lucide-react';
import * as React from 'react';
import { Checkbox, CheckboxWithLabel } from '../Checkbox';
import { Input } from '../Input';
import { Popover, PopoverContent, PopoverTrigger } from '../Popover';
import { ScrollArea } from '../ScrollArea';
import { Skeleton } from '../Skeleton';
import { Typography } from '../Typography';
import { cn } from '../utils';
import {
    MultiSelectTriggerVariants,
    multiSelectCaretClasses,
    multiSelectCheckboxStyles,
    multiSelectContentClasses,
    multiSelectIndeterminateCheckboxStyles,
    multiSelectInteractiveRowStyles,
    multiSelectLoadingRowsClasses,
    multiSelectRowStyles,
    multiSelectSearchIconClasses,
    multiSelectSearchInputClasses,
    multiSelectSearchRowClasses,
    multiSelectStateRowClasses,
} from './MultiSelect.styles';

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

type MultiSelectVariant = 'outlined' | 'filled';

interface MultiSelectTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    open?: boolean;
    variant?: MultiSelectVariant;
}

const MultiSelectTrigger = React.forwardRef<HTMLButtonElement, MultiSelectTriggerProps>(
    ({ className, open, children, variant, ...props }, ref) => (
        <button ref={ref} type='button' className={cn(MultiSelectTriggerVariants({ className, variant }))} {...props}>
            <span className='truncate'>{children}</span>
            <CaretDown className={cn(multiSelectCaretClasses, open && 'rotate-180')} />
        </button>
    )
);
MultiSelectTrigger.displayName = 'MultiSelectTrigger';

interface MultiSelectProps {
    'data-analytics'?: string;
    options: MultiSelectOption[];
    value: string[];
    onValueChange: (values: string[]) => void;
    onSearchValueChange?: (value: string) => void;
    /** Set to false when options are already filtered by an external search. Defaults to true. */
    shouldFilter?: boolean;
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
    /** Displays the no-results message alongside retained options for an external search. */
    showNoResults?: boolean;
    /** Displays a search failure while keeping the supplied options available. */
    searchError?: string;
    /**
     * Controls the collapsed trigger appearance independently of its selected values.
     * Use `outlined` when a “Select All” is defaulted state. Example: Findings Table Trigger defaults to all selected.
     */
    variant?: MultiSelectVariant;
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

const MultiSelectOptionRow = ({ option, checked, onSelect }: MultiSelectOptionRowProps) => (
    <div className='p-1'>
        <CheckboxWithLabel
            label={option.label}
            checked={checked}
            disabled={option.disabled}
            onCheckedChange={() => onSelect(option.value)}
            className={multiSelectCheckboxStyles}
            fieldClassName={cn(
                multiSelectRowStyles,
                'text-base font-normal leading-4',
                option.disabled
                    ? 'cursor-not-allowed bg-[color:var(--button-disabled-fill)] text-disabled'
                    : multiSelectInteractiveRowStyles
            )}
            labelClassName='min-w-0 flex-1 truncate text-inherit'
        />
    </div>
);

const MultiSelectAllOptionsRow = ({ checked, label, onSelect }: MultiSelectAllOptionsRowProps) => (
    <div className='p-1'>
        <CheckboxWithLabel
            label={label}
            checked={checked}
            onCheckedChange={onSelect}
            className={cn(multiSelectCheckboxStyles, multiSelectIndeterminateCheckboxStyles)}
            fieldClassName={cn(
                multiSelectRowStyles,
                multiSelectInteractiveRowStyles,
                'text-base font-normal leading-4'
            )}
            labelClassName='min-w-0 flex-1 truncate text-inherit'
        />
    </div>
);

const MultiSelectLoadingRows = ({ loadingText }: { loadingText: string }) => (
    <div role='status' aria-label={loadingText}>
        {Array.from({ length: 6 }).map((_, index) => (
            <div key={index} className={multiSelectLoadingRowsClasses}>
                <Checkbox tabIndex={-1} disabled />
                <Skeleton className='h-4 flex-1' />
            </div>
        ))}
    </div>
);

const MultiSelectStateRow = ({ children }: { children: React.ReactNode }) => (
    <Typography variant='body1' component='div' className={multiSelectStateRowClasses}>
        {children}
    </Typography>
);

/**
 * MultiSelect Doodle Component
 */
const MultiSelect = ({
    'data-analytics': dataAnalytics,
    options,
    value,
    onValueChange,
    onSearchValueChange,
    shouldFilter = true,
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
    showNoResults = false,
    searchError,
    variant,
}: MultiSelectProps) => {
    const [open, setOpen] = React.useState(false);
    const [searchValue, setSearchValue] = React.useState('');

    const handleSearchChange = (value: string) => {
        setSearchValue(value);
        onSearchValueChange?.(value);
    };

    const handleOpenChange = (shouldOpen: boolean) => {
        setOpen(shouldOpen);

        if (!shouldOpen && searchValue !== '') {
            handleSearchChange('');
        }
    };

    let triggerVariant: MultiSelectVariant;

    if (variant !== undefined) {
        triggerVariant = variant;
    } else if (value.length === 0 && !disabled && !error) {
        triggerVariant = 'outlined';
    } else {
        triggerVariant = 'filled';
    }

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

    // select all/clear
    const enabledOptionValues = options.filter((option) => !option.disabled).map((option) => option.value);

    const selectedEnabledOptionCount = enabledOptionValues.filter((optionValue) => value.includes(optionValue)).length;

    const hasSelectedAllEnabledOptions =
        enabledOptionValues.length > 0 && selectedEnabledOptionCount === enabledOptionValues.length;

    const triggerText =
        hasSelectedAllEnabledOptions && selectAllLabel
            ? selectAllLabel
            : getMultiSelectTriggerText(options, value, placeholder);

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

    const filteredOptions = React.useMemo(() => {
        if (!shouldFilter || !searchTerm || searchError) {
            return options;
        }
        return options.filter((option) => option.label.toLowerCase().includes(searchTerm));
    }, [options, searchTerm, searchError, shouldFilter]);

    const hasOptions = options.length > 0;
    const hasSearchText = searchTerm.length > 0;
    const hasSearchResults = filteredOptions.length > 0;

    const isSearchResultEmpty = isSearchable && hasSearchText && !hasSearchResults;

    let listContent: React.ReactNode;

    if (isLoading && !hasSearchText) {
        listContent = <MultiSelectLoadingRows loadingText={loadingText} />;
    } else if (!hasOptions && searchError) {
        listContent = null;
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
                    data-analytics={dataAnalytics}
                    variant={triggerVariant}
                    open={open}
                    disabled={disabled}
                    aria-invalid={error || undefined}
                    className={className}>
                    {triggerText}
                </MultiSelectTrigger>
            </PopoverTrigger>
            <PopoverContent className={multiSelectContentClasses} align='start'>
                {isSearchable && (
                    <div className={multiSelectSearchRowClasses}>
                        <Search className={multiSelectSearchIconClasses} aria-hidden='true' />
                        <Input
                            aria-label={searchPlaceholder}
                            value={searchValue}
                            onChange={(e) => handleSearchChange(e.target.value)}
                            placeholder={searchPlaceholder}
                            className={multiSelectSearchInputClasses}
                        />
                    </div>
                )}
                {searchError && (
                    <div role='alert' className='px-3 py-2 text-status-error-main'>
                        {searchError}
                    </div>
                )}

                <ScrollArea
                    thumbColor='var(--multiselect-scrollbar-thumb)'
                    className='max-h-60 [&_[data-radix-scroll-area-viewport]]:max-h-60'>
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

                        {showNoResults && hasOptions && !isSearchResultEmpty && !isLoading && !searchError && (
                            <MultiSelectStateRow>{noResultsText}</MultiSelectStateRow>
                        )}
                        {listContent}
                    </div>
                </ScrollArea>
            </PopoverContent>
        </Popover>
    );
};

export { MultiSelect, MultiSelectOptionRow, MultiSelectTrigger };
export type { MultiSelectOption, MultiSelectProps, MultiSelectVariant };
