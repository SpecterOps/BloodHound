// Copyright 2025 Specter Ops, Inc.
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
import {
    Button,
    Calendar,
    Input,
    Label,
    Popover,
    PopoverContent,
    PopoverTrigger,
} from '@bloodhoundenterprise/doodleui';
import { faCalendarDay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import InputMask from '@mona-health/react-input-mask';
import { DateTime } from 'luxon';
import { HTMLAttributes, InputHTMLAttributes, forwardRef, useState } from 'react';
import { LuxonFormat } from '../../utils';

interface DatePickerProps extends InputHTMLAttributes<HTMLInputElement> {
    setValue: (date: string) => void;
    label?: string;
    labelClasses?: HTMLAttributes<'div'>['className'];
    error?: string;
    clearError?: () => void;
    onOpenChange?: (open: boolean) => void;
    fromDate?: DateTime;
    toDate?: DateTime;
}

export const DatePicker = forwardRef<HTMLInputElement, DatePickerProps>(
    ({ setValue, label, labelClasses, error, clearError, onOpenChange, fromDate, toDate, ...props }, ref) => {
        const [calendarOpen, setCalendarOpen] = useState<boolean>(false);

        const handleSelect = (date: Date | undefined) => {
            if (date) {
                setValue(DateTime.fromJSDate(date).toFormat(LuxonFormat.ISO_8601));
                setCalendarOpen(false);
            }
            if (error && clearError) clearError();
        };

        const handleOpenChange = (open: boolean) => {
            setCalendarOpen(open);
            onOpenChange?.(open);
        };

        const getPropsAsJSDate = () => {
            return props.value
                ? DateTime.fromFormat(props.value.toString(), LuxonFormat.ISO_8601).toJSDate()
                : undefined;
        };

        return (
            <Popover onOpenChange={handleOpenChange} open={calendarOpen}>
                <div>
                    {label && (
                        <Label htmlFor={props.name} size='small' className={labelClasses}>
                            {label}
                        </Label>
                    )}
                    <div className='relative'>
                        <InputMask
                            {...props}
                            id={props.name}
                            mask='9999-99-99'
                            maskPlaceholder=''
                            placeholder='yyyy-mm-dd'
                            ref={ref}>
                            <Input
                                id={props.name}
                                variant={'outlined'}
                                className='rounded bg-neutral-light-1 dark:bg-neutral-dark-1'
                            />
                        </InputMask>
                        <PopoverTrigger asChild>
                            <Button variant={'text'} className='absolute right-2 top-2 p-0 h-6'>
                                <FontAwesomeIcon
                                    aria-describedby='Chooose Date'
                                    size='xl'
                                    icon={faCalendarDay}
                                    className='opacity-50 hover:opacity-100 hover:cursor-pointer transition-all'
                                />
                            </Button>
                        </PopoverTrigger>
                    </div>
                    {error && (
                        <span role='alert' className='px-2 pt-1 block text-xs text-red'>
                            {error}
                        </span>
                    )}
                </div>
                <PopoverContent
                    align='end'
                    alignOffset={-7}
                    className='w-auto p-3 border border-neutral-light-5 dark:border-neutral-light-1'>
                    <Calendar
                        mode='single'
                        selected={getPropsAsJSDate()}
                        defaultMonth={getPropsAsJSDate()}
                        onSelect={handleSelect}
                        fromDate={fromDate?.toJSDate()}
                        toDate={toDate?.toJSDate()}
                    />
                </PopoverContent>
            </Popover>
        );
    }
);

DatePicker.displayName = 'DatePicker';
