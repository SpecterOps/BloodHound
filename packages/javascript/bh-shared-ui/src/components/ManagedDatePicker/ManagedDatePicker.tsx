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
import { FC, useEffect, useId, useRef, useState } from 'react';
import { LuxonFormat, fiveYearsAgo, now } from '../../utils';

const DATE_FORMAT_MASK = 'yyyy-mm-dd';
const INVALID_DATE_MESSAGE = 'Input is not a valid date.';

type Props = {
    /** The earliest selectable date */
    fromDate?: Date;

    /** Text to display in input when not focused */
    hint?: string;

    /** Callback executed when the date changes; `isValid` is false if `dateStr` is not a valid date */
    onDateChange: (dateStr: string, isValid: boolean) => void;

    /** The latest selectable date */
    toDate?: Date;

    /** If provided, will be displayed as an error */
    validationError?: string;

    /** The initial value of the component */
    value?: string;
};

/**
 * A controlled, extended version of Doodle UI's DatePicker providing a simpler
 * API, calendar/input synchronization, validation, and input hints.
 */
export const ManagedDatePicker: FC<Props> = ({
    fromDate = fiveYearsAgo(),
    hint = DATE_FORMAT_MASK,
    onDateChange,
    toDate = now(),
    validationError: validationErrorProp = '',
    value,
}) => {
    // `hint` shows (ex. 'Start Date') as placeholder while input is unfocused.
    // When focused, switch to mask placeholder ('yyyy-mm-dd').
    const [placeholder, setPlaceholder] = useState(hint);

    // Value for text input. Will be invalid as date during keystrokes.
    const [inputDateString, setInputDateString] = useState<string>(value?.slice(0, 10) || '');

    const calendarValue = value ? DateTime.fromISO(value) : undefined;

    // Validation produces an array of errors. Only one (the first) is displayed.
    const [validationError, setValidationError] = useState(validationErrorProp);

    // Used to debounce execution of validation
    const lastKeyEvent = useRef<NodeJS.Timeout | null>(null);

    // Update input as user types
    const setInputFromEvent = (event: React.ChangeEvent<HTMLInputElement>) => {
        const dateStr = event.target.value;
        setInputDateString(dateStr);
        setValidationError(validationErrorProp);

        if (dateStr === '') {
            onDateChange(dateStr, true);
            lastKeyEvent.current && clearTimeout(lastKeyEvent.current);
            return;
        }

        // Debounce execution of validation
        // Can't use lodash debounce because validateInput needs updated onDateChange
        lastKeyEvent.current && clearTimeout(lastKeyEvent.current);
        lastKeyEvent.current = setTimeout(() => validateInput(dateStr), 300);
    };

    // Check if `inputDateString` is a valid date
    const validateInput = (dateStr: string) => {
        if (dateStr === '') {
            return;
        }

        const dateTime = DateTime.fromFormat(dateStr, LuxonFormat.ISO_8601);

        // Date string doesn't match format
        if (dateTime.invalidReason === 'unparsable') {
            setValidationError(INVALID_DATE_MESSAGE);
            onDateChange(dateStr, false);
            return;
        }

        // Date string is not a valid date
        if (!dateTime.isValid) {
            setValidationError(INVALID_DATE_MESSAGE);
            onDateChange(dateStr, false);
            return;
        }

        // Date string is a valid date
        onDateChange(dateStr, true);
    };

    // Update input when user selects a date from the calendar
    const setInputFromCalendar = (date?: Date) => {
        setValidationError(validationErrorProp);

        // Technically, not possible for calendar to select undefined, but this keeps TS happy
        if (date === undefined) return;

        const dateStr = DateTime.fromJSDate(date).toFormat(LuxonFormat.ISO_8601);
        setInputDateString(dateStr);
        onDateChange(dateStr, true);
    };

    // Clear input if value is changed to undefined
    useEffect(() => {
        if (value === undefined) {
            setInputDateString('');
            setValidationError(validationErrorProp);
        }
    }, [validationErrorProp, value]);

    // Update validation error if prop changes
    useEffect(() => {
        setValidationError(validationErrorProp);
    }, [validationErrorProp]);

    const errorId = useId();
    return (
        <>
            <DatePicker
                className='bg-transparent dark:bg-transparent pl-2'
                onBlur={() => {
                    setPlaceholder(hint);
                    validateInput(inputDateString);
                }}
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
                    selected: calendarValue?.isValid ? calendarValue.toJSDate() : undefined,
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
