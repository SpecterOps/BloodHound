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

import { faCalendarDay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import { InputHTMLAttributes, forwardRef, useState } from 'react';
import { Calendar, CalendarProps } from '../Calendar';
import { IconButton } from '../IconButton';
import { Input, InputProps } from '../Input';
import { Popover, PopoverContent, PopoverTrigger } from '../Popover';
import { cn } from '../utils';
import {
    datePickerContentClasses,
    datePickerInputClasses,
    datePickerStyles,
    datePickerTriggerClasses,
} from './DatePicker.styles';

interface DatePickerProps extends InputHTMLAttributes<HTMLInputElement>, InputProps {
    calendarProps?: CalendarProps;
    InputElement?: React.ForwardRefExoticComponent<InputProps & React.RefAttributes<HTMLInputElement>>;
    onOpenChange?: (open: boolean) => void;
}

export const DatePicker = forwardRef<HTMLInputElement, DatePickerProps>(
    ({ calendarProps, InputElement, onOpenChange, ...props }, ref) => {
        const [calendarOpen, setCalendarOpen] = useState<boolean>(false);

        const handleOpenChange = (open: boolean) => {
            setCalendarOpen(open);
            onOpenChange?.(open);
        };

        return (
            <Popover onOpenChange={handleOpenChange} open={calendarOpen}>
                <div className={datePickerStyles.wrapper}>
                    {InputElement ? (
                        <InputElement
                            {...props}
                            className={cn(datePickerInputClasses({ variant: props.variant }), props.className)}
                            ref={ref}
                        />
                    ) : (
                        <Input
                            variant='outlined'
                            {...props}
                            className={cn(datePickerInputClasses({ variant: props.variant }), props.className)}
                        />
                    )}
                    <PopoverTrigger asChild>
                        <IconButton
                            triggerClassName={datePickerStyles.iconTrigger}
                            className={datePickerTriggerClasses}
                            aria-label='Choose Date'>
                            <FontAwesomeIcon aria-hidden='true' size='lg' icon={faCalendarDay} />
                        </IconButton>
                    </PopoverTrigger>
                </div>
                <PopoverContent align='end' alignOffset={-7} className={datePickerContentClasses}>
                    <Calendar classNames={{ day_selected: datePickerStyles.selectedDay }} {...calendarProps} />
                </PopoverContent>
            </Popover>
        );
    }
);

DatePicker.displayName = 'DatePicker';
