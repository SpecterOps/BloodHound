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

import { DatePicker } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { FC, useEffect, useId, useState } from 'react';
import { LuxonFormat, fiveYearsAgo, now } from '../../utils';

const DATE_FORMAT_MASK = 'yyyy-mm-dd';
const INVALID_DATE_MESSAGE = 'Input is not a valid date.';

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
const isDateValid = (
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
    } else {
        // Clear any previous errors
        onInvalid(['']);
    }

    return !hasErrors;
};

type Props = {
    /** The earliest selectable date */
    fromDate?: Date;

    /** Text to display in input when not focused */
    hint?: string;

    /** Callback executed when the date changes */
    onDateChange: (date?: Date) => void;

    /** Callback executed when the date is validated */
    onValidation?: (isValid: boolean) => void;

    /** The latest selectable date */
    toDate?: Date;

    /** Array of validation rules to apply to the date. If not provided, the date is always valid. */
    validations?: {
        /** Function that returns true when the date is valid */
        rule: (date: Date) => boolean;

        /** Message to display when the date is invalid */
        errorMessage: string;
    }[];

    /** Value of the date picker */
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

    // Validation produces an array of errors. Only one (the first) is displayed.
    const [validationError, setValidationError] = useState('');

    const setNextError = (errors: string[]) => setValidationError(errors[0]);

    // Update input as user types; validation will happen on blur
    const setInputFromEvent = (event: React.ChangeEvent<HTMLInputElement>) => {
        setValidationError('');
        setInputDateString(event.target.value);
    };

    // Update input as user types; validation will happen on blur
    const setInputFromCalendar = (date?: Date) => {
        setValidationError('');

        // Technically, not possible for calendar to select undefined, but this keeps TS happy
        if (date === undefined) return;

        validateAndUpdateValue(date);
    };

    // If input string is a date, call onDateChange with date
    const validateDateString = () => {
        setPlaceholder(hint);

        if (inputDateString === '') {
            onDateChange(undefined);
            return;
        }

        const dateTime = DateTime.fromFormat(inputDateString, LuxonFormat.ISO_8601);

        // Set a validation error if string is not a valid date
        if (!dateTime.isValid) {
            onDateChange(undefined);
            setNextError([INVALID_DATE_MESSAGE]);
            onValidation?.(false);
            return;
        }

        validateAndUpdateValue(dateTime.toJSDate());
    };

    const validateAndUpdateValue = (date: Date) => {
        if (validations.length === 0 || isDateValid(date, validations, setNextError)) {
            onValidation?.(true);
        } else {
            onValidation?.(false);
        }

        setInputDateString(DateTime.fromJSDate(date).toFormat(LuxonFormat.ISO_8601));
        onDateChange(date);
    };

    // Revalidate when value is changed from outside component
    useEffect(() => {
        if (validationError === '' || value === undefined) return;

        if (validations.length === 0 || isDateValid(value, validations, setNextError)) {
            onValidation?.(true);
        } else {
            onValidation?.(false);
        }
    }, [onValidation, validationError, validations, value]);

    const errorId = useId();
    return (
        <>
            <DatePicker
                className='bg-transparent dark:bg-transparent pl-2'
                onBlur={validateDateString}
                onChange={setInputFromEvent}
                onFocus={() => setPlaceholder(DATE_FORMAT_MASK)}
                placeholder={placeholder}
                // `value` only represents input buffer
                value={inputDateString}
                variant={'underlined'}
                aria-invalid={Boolean(validationError)}
                aria-describedby={validationError ? errorId : undefined}
                calendarProps={{
                    fromDate: fromDate,
                    mode: 'single',
                    onSelect: setInputFromCalendar,
                    // `selected` represents the value of the component
                    selected: value,
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
