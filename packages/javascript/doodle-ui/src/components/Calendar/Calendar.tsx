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

import { ChevronLeft, ChevronRight } from 'lucide-react';
import * as React from 'react';
import { DayPicker } from 'react-day-picker';

import { TextButtonVariants } from '../Button/Button';
import { Select, SelectContent, SelectItem, SelectPortal, SelectTrigger, SelectValue } from '../Select';
import { cn } from '../utils';
import { calendarClassNames, calendarDropdownTriggerClasses, calendarSelectItemClasses } from './Calendar.styles';

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

function Calendar({
    className,
    classNames,
    showOutsideDays = true,
    ...props
}: CalendarProps & { onChange?: React.ChangeEventHandler<HTMLSelectElement> }) {
    const handleCalendarChange = (_value: string | number, e: React.ChangeEventHandler<HTMLSelectElement>) => {
        const event = {
            target: {
                value: String(_value),
            },
        } as React.ChangeEvent<HTMLSelectElement>;
        e(event);
    };

    return (
        <DayPicker
            showOutsideDays={showOutsideDays}
            className={cn('p-3', className)}
            classNames={{ ...calendarClassNames, ...classNames }}
            components={{
                IconLeft: () => <ChevronLeft className='h-4 w-4' />,
                IconRight: () => <ChevronRight className='h-4 w-4' />,
                Dropdown: ({ ...props }) => (
                    <Select
                        {...props}
                        onValueChange={(value) => {
                            if (props.onChange) {
                                handleCalendarChange(value, props.onChange);
                            }
                        }}
                        value={props.value as string}>
                        <SelectTrigger className={cn(TextButtonVariants(), calendarDropdownTriggerClasses)}>
                            <SelectValue placeholder={props?.caption}>{props?.caption}</SelectValue>
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {props.children &&
                                    React.Children.map(props.children, (child) => (
                                        <SelectItem
                                            value={(child as React.ReactElement)?.props?.value}
                                            className={calendarSelectItemClasses}>
                                            {(child as React.ReactElement)?.props?.children}
                                        </SelectItem>
                                    ))}
                            </SelectContent>
                        </SelectPortal>
                    </Select>
                ),
            }}
            {...props}
        />
    );
}
Calendar.displayName = 'Calendar';

export { Calendar };
