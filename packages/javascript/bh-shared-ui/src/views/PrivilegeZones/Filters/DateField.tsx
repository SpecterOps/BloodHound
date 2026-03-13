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
import { faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, DatePicker, FormField, FormItem, FormLabel, FormMessage } from 'doodle-ui';
import { DateTime } from 'luxon';
import { FC } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { MaskedInput } from '../../../components';
import { END_DATE, LuxonFormat, START_DATE, cn } from '../../../utils';

const toDate = DateTime.local().toJSDate();
const fromDate = DateTime.fromJSDate(toDate).minus({ years: 1 }).toJSDate();

const DateField: FC<{
    form: UseFormReturn;
    name: typeof START_DATE | typeof END_DATE;
}> = ({ form, name }) => {
    return (
        <FormField
            name={name}
            control={form.control}
            render={({ field }) => (
                <FormItem className='w-40 flex flex-col gap-2 justify-start'>
                    <FormLabel htmlFor={field.name}>{name === START_DATE ? 'Start Date' : 'End Date'}</FormLabel>
                    <div className='flex gap-2'>
                        <DatePicker
                            {...field}
                            InputElement={MaskedInput}
                            calendarProps={{
                                className: 'w-11/12',
                                mode: 'single',
                                fromDate,
                                toDate,
                                selected: DateTime.fromFormat(field.value, LuxonFormat.ISO_8601).toJSDate(),
                                onSelect: (value: Date | undefined) => {
                                    form.setValue(
                                        field.name,
                                        value ? DateTime.fromJSDate(value).toFormat(LuxonFormat.ISO_8601) : ''
                                    );
                                },
                                disabled: (date: Date) => {
                                    return DateTime.fromJSDate(date) > DateTime.local();
                                },
                            }}
                        />
                        <Button
                            variant={'text'}
                            disabled={!field.value}
                            className={cn('w-1/12 p-0', { invisible: !field.value })}
                            onClick={() => {
                                form.setValue(field.name, '');
                                form.clearErrors();
                            }}>
                            <FontAwesomeIcon icon={faClose} />
                        </Button>
                    </div>
                    <FormMessage />
                </FormItem>
            )}
        />
    );
};

export const StartDateField: FC<{ form: UseFormReturn }> = ({ form }) => {
    return <DateField form={form} name={START_DATE} />;
};

export const EndDateField: FC<{ form: UseFormReturn }> = ({ form }) => {
    return <DateField form={form} name={END_DATE} />;
};
