import { DatePicker } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { FC, useEffect, useState } from 'react';
import { FIVE_YEARS_AGO, LuxonFormat, NOW } from '../../utils';

const DATE_FORMAT = 'yyyy-mm-dd';

export const VALIDATIONS = {
    isAfterDate: (after: Date | undefined, errorMessage: string) => ({
        rule: (date?: Date) => {
            if (!after || !date) {
                return true;
            }
            const [first, second] = [DateTime.fromJSDate(date), DateTime.fromJSDate(after)];
            return first >= second;
        },
        errorMessage,
    }),
    isBeforeDate: (before: Date | undefined, errorMessage: string) => ({
        rule: (date?: Date) => {
            if (!before || !date) {
                return true;
            }
            const [first, second] = [DateTime.fromJSDate(before), DateTime.fromJSDate(date)];
            return first >= second;
        },
        errorMessage,
    }),
};

/** Returns true if date passes validations, otherwise return false and runs onInvalid */
const validateDate = (date: Date, validations: Props['validations'] = [], onInvalid: (errors: string[]) => void) => {
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
    fromDate = FIVE_YEARS_AGO,
    hint = DATE_FORMAT,
    onDateChange,
    onValidation,
    toDate = NOW,
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
        onValidation?.(true);

        if (dateString === '') {
            onDateChange();
            setCalendarDate(undefined);
        }

        const dateTime = DateTime.fromFormat(dateString, LuxonFormat.ISO_8601);

        if (!dateTime.isValid) {
            return;
        }

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

    return (
        <>
            <DatePicker
                className='bg-transparent dark:bg-transparent pl-2'
                onChange={syncDateInput}
                onFocus={() => setPlaceholder(DATE_FORMAT)}
                onBlur={updateHintAndValidate}
                placeholder={placeholder}
                // `value` only represents input buffer
                value={inputDateString}
                variant={'underlined'}
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
            {validationError && <span className='text-error text-sm'>{validationError}</span>}
        </>
    );
};
