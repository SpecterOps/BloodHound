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
import { Button } from '../Button';
import { Calendar, CalendarProps } from '../Calendar';
import { Input, InputProps } from '../Input';
import { Popover, PopoverContent, PopoverTrigger } from '../Popover';
import { cn } from '../utils';

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
                <div className='relative'>
                    {InputElement ? (
                        <InputElement
                            {...props}
                            className={cn('bg-neutral-light-1 dark:bg-neutral-dark-1 peer', props.className)}
                            ref={ref}
                        />
                    ) : (
                        <Input
                            variant={'outlined'}
                            {...props}
                            className={cn('bg-neutral-light-1 dark:bg-neutral-dark-1 peer', props.className)}
                        />
                    )}
                    <PopoverTrigger asChild>
                        <Button
                            variant={'text'}
                            className='absolute right-2 top-2 p-0 h-6 opacity-50 peer-hover:opacity-100'
                            aria-label='Choose Date'>
                            <FontAwesomeIcon aria-hidden='true' size='xl' icon={faCalendarDay} />
                        </Button>
                    </PopoverTrigger>
                </div>
                <PopoverContent
                    align='end'
                    alignOffset={-7}
                    className='w-auto p-3 border border-neutral-light-5 dark:border-neutral-light-1 z-[1500]'>
                    <Calendar classNames={{ day_selected: 'bg-primary text-neutral-light-1' }} {...calendarProps} />
                </PopoverContent>
            </Popover>
        );
    }
);

DatePicker.displayName = 'DatePicker';
