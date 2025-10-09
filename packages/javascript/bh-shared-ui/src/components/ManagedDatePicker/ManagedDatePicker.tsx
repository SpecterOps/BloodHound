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

import { DatePicker } from 'doodle-ui';
import { DateTime } from 'luxon';
import { FC, useEffect, useId, useState } from 'react';
import { LuxonFormat, fiveYearsAgo, now } from '../../utils';

const DATE_FORMAT_MASK = 'yyyy-mm-dd';

export const VALIDATIONS = {
    isAfterDate: (after: Date | undefined, errorMessage: string) => ({
        rule: (date?: Date) => {
            if (!after || !date) {
                return true;
            }
            return date.getTime() >= after.getTime();
        },
        errorMessage,
    }),
    isBeforeDate: (before: Date | undefined, errorMessage: string) => ({
        rule: (date?: Date) => {
            if (!before || !date) {
                return true;
            }
            return before.getTime() >= date.getTime();
        },
        errorMessage,
    }),
};

/** Returns true if date passes validations, otherwise return false and runs onInvalid */
const validateDate = (
    date: Date,
    validations: NonNullable<Props['validations']> = [],
    onInvalid: (errors: string[]) => void
) => {
    const errors = validations.reduce((agg: string[], { rule, errorMessage }) => {
        if (!rule(date)) {
            agg.push(errorMessage);
        }
        return agg;
    }, []);

    const hasErrors = errors.length > 0;

    if (hasErrors) {
        onInvalid(errors);
    }

    return !hasErrors;
};

type Props = {
    fromDate?: Date;
    hint?: string;
    onDateChange: (date?: Date) => void;
    onValidation?: (isValid: boolean) => void;
    toDate?: Date;
    validations?: {
        rule: (date: Date) => boolean;
        errorMessage: string;
    }[];
    value?: Date;
};

/**
 * A controlled, extended version of Doodle UI's DatePicker providing a simpler
 * API, calendar/input synchronization, validation, and input hints.
 */
export const ManagedDatePicker: FC<Props> = ({
    fromDate = fiveYearsAgo(),
    hint = DATE_FORMAT_MASK,
    onDateChange,
    onValidation,
    toDate = now(),
    validations = [],
    value,
}) => {
    // `hint` shows (ex. 'Start Date') as placeholder while input is unfocused.
    // When focused, switch to mask placeholder ('yyyy-mm-dd').
    const [placeholder, setPlaceholder] = useState(hint);

    // Value for text input. Will be invalid as date during keystrokes.
    const [inputDateString, setInputDateString] = useState<string>(
        value ? DateTime.fromJSDate(value).toFormat(LuxonFormat.ISO_8601) : ''
    );

    // When inputDateString is valid, this will hold the value as a JS Date
    const [calendarDate, setCalendarDate] = useState<Date | undefined>(value);

    // Validation produces an array of errors. Only one (the first) is displayed.
    const [validationError, setValidationError] = useState('');

    // Reset inputDateString when value goes undefiend
    useEffect(() => {
        if (value === undefined) {
            setInputDateString('');
            setCalendarDate(undefined);
        }
    }, [value]);

    const setNextError = (errors: string[]) => setValidationError(errors[0]);

    // Checks if typed date is valid. Updates calendarDate state when valid.
    const syncDateInput = (event: React.ChangeEvent<HTMLInputElement>) => {
        const dateString = event.target.value;
        setInputDateString(dateString);
        setValidationError('');

        if (dateString === '') {
            onDateChange();
            setCalendarDate(undefined);
            onValidation?.(true);
            return;
        }

        const dateTime = DateTime.fromFormat(dateString, LuxonFormat.ISO_8601);

        if (!dateTime.isValid) {
            return;
        }

        onValidation?.(true);
        setCalendarDate(dateTime.toJSDate());
    };

    const updateHintAndValidate = () => {
        if (inputDateString !== '' && !DateTime.fromFormat(inputDateString, LuxonFormat.ISO_8601).isValid) {
            setValidationError('Input is not a valid date.');
            onValidation?.(false);
            return;
        }

        setPlaceholder(hint);
        validateCalendarDate(calendarDate);
    };

    // Apply validation props and update text input when Calendar date is clicked
    const validateCalendarDate = (selectedDay: Date | undefined) => {
        setValidationError('');

        if (selectedDay === undefined) {
            return;
        }

        setCalendarDate(selectedDay);
        setInputDateString(DateTime.fromJSDate(selectedDay).toFormat(LuxonFormat.ISO_8601));

        if (validateDate(selectedDay, validations, setNextError)) {
            onDateChange(selectedDay);
            onValidation?.(true);
        } else {
            onValidation?.(false);
        }
    };

    const errorId = useId();
    return (
        <>
            <DatePicker
                className='bg-transparent dark:bg-transparent pl-2'
                onChange={syncDateInput}
                onFocus={() => setPlaceholder(DATE_FORMAT_MASK)}
                onBlur={updateHintAndValidate}
                placeholder={placeholder}
                // `value` only represents input buffer
                value={inputDateString}
                variant={'underlined'}
                aria-invalid={Boolean(validationError)}
                aria-describedby={validationError ? errorId : undefined}
                calendarProps={{
                    fromDate: fromDate,
                    mode: 'single',
                    // Validation happen on calendar date select
                    onSelect: validateCalendarDate,
                    // `selected` represents the value of the component
                    selected: calendarDate,
                    toDate: toDate,
                }}
            />
            {validationError && (
                <span className='text-error text-sm' id={errorId}>
                    {validationError}
                </span>
            )}
        </>
    );
};
