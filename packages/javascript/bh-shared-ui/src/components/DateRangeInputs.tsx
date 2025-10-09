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

import { Label } from 'doodle-ui';
import { DateTime } from 'luxon';
import { useEffect, useState, type FC } from 'react';
import { FinishedJobsFilter } from '../utils';
import { ManagedDatePicker, VALIDATIONS } from './ManagedDatePicker';

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
