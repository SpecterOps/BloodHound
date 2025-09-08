import { Label } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { useEffect, useState, type FC } from 'react';
import { FinishedJobsFilter } from '../../utils';
import { ManagedDatePicker, VALIDATIONS } from '../ManagedDatePicker/ManagedDatePicker';

export type DateRangeChange = Pick<FinishedJobsFilter, 'start_time' | 'end_time'>;

type DateRangeInputsProps = {
    end?: string;
    onChange: (changed: DateRangeChange) => void;
    onValidation?: (isValid: boolean) => void;
    start?: string;
};

export const DateRangeInputs: FC<DateRangeInputsProps> = ({ end, onChange, onValidation, start }) => {
    const endDate = end ? new Date(end) : undefined;
    const startDate = start ? new Date(start) : undefined;

    const [isEndValid, setIsEndValid] = useState(true);
    const [isStartValid, setIsStartValid] = useState(true);

    useEffect(() => {
        onValidation?.(isEndValid && isStartValid);
    }, [isEndValid, isStartValid, onValidation]);

    return (
        <div className='flex flex-col gap-2 w-56'>
            <Label>Date Range</Label>

            <ManagedDatePicker
                hint='Start Date'
                onDateChange={(date) => onChange({ start_time: date ? date.toISOString() : undefined })}
                onValidation={setIsStartValid}
                value={startDate}
                validations={[VALIDATIONS.isBeforeDate(endDate, 'Start time must come before end.')]}
            />

            <ManagedDatePicker
                hint='End Date'
                // Convert to the end of the day
                onDateChange={(date) =>
                    onChange({ end_time: date ? DateTime.fromJSDate(date).endOf('day').toISO()! : undefined })
                }
                onValidation={setIsEndValid}
                value={endDate}
                validations={[VALIDATIONS.isAfterDate(startDate, 'End time must be after start.')]}
            />
        </div>
    );
};
