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
import * as React from 'react';
import { Checkbox } from '../Checkbox';
import { Popover, PopoverContent, PopoverTrigger } from '../Popover';
import { ScrollArea } from '../ScrollArea';
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
    'flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm border-input-border-default bg-input-fill text-input-placeholder-text focus:outline-none focus-visible:focus-ring data-[state=open]:focus-ring enabled:hover:border-input-border-hover enabled:hover:bg-secondary enabled:hover:text-white disabled:cursor-not-allowed disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-icon-disabled aria-[invalid=true]:border-status-error-main'
);

const multiSelectRowStyles =
    'flex items-center gap-2 mx-1 p-2 py-2 rounded-lg cursor-pointer hover:bg-secondary hover:text-white';

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
    checked: boolean;
    label: string;
    onSelect: () => void;
}

const MultiSelectOptionRow = ({ option, checked, onSelect }: MultiSelectOptionRowProps) => (
    <div
        role='option'
        aria-selected={checked}
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
        aria-selected={checked}
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
        <span className='truncate text-sm'>{label}</span>
    </div>
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
}: MultiSelectProps) => {
    const [open, setOpen] = React.useState(false);

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

    const selectableValues = options.filter((option) => !option.disabled).map((option) => option.value);

    const hasSelectedAllOptions =
        selectableValues.length > 0 && selectableValues.every((optionValue) => value.includes(optionValue));

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
            <PopoverContent className='p-0 w-[var(--radix-popover-trigger-width)]' align='start'>
                <ScrollArea className='max-h-60'>
                    <div role='listbox' aria-multiselectable='true' className='py-1'>
                        {selectAllLabel && (
                            <MultiSelectActionRow
                                checked={hasSelectedAllOptions}
                                label={selectAllLabel}
                                onSelect={handleSelectAll}
                            />
                        )}
                        {options.map((option) => (
                            <MultiSelectOptionRow
                                key={option.value}
                                option={option}
                                checked={value.includes(option.value)}
                                onSelect={handleSelect}
                            />
                        ))}
                    </div>
                </ScrollArea>
            </PopoverContent>
        </Popover>
    );
};

export { MultiSelect, MultiSelectOptionRow, MultiSelectTrigger };
export type { MultiSelectOption, MultiSelectProps };
