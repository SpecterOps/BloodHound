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
import { useCallback, useEffect, useState, type FC } from 'react';
import { FinishedJobsFilter, LuxonFormat } from '../utils';
import { ManagedDatePicker } from './ManagedDatePicker';

export type DateRangeChange = Pick<FinishedJobsFilter, 'start_time' | 'end_time'>;

type DateRangeInputsProps = {
    end?: string;
    onChange: (changed: DateRangeChange) => void;
    onValidation?: (isValid: boolean) => void;
    start?: string;
};

export const DateRangeInputs: FC<DateRangeInputsProps> = ({ end, onChange, onValidation, start }) => {
    const [isEndValid, setIsEndValid] = useState(true);
    const [isStartValid, setIsStartValid] = useState(true);

    const isCorrectOrder =
        !start || !end || (isStartValid && isEndValid && DateTime.fromISO(start) <= DateTime.fromISO(end));

    const onStartChange = useCallback(
        (dateStr: string, isValid: boolean) => {
            setIsStartValid(isValid);

            // Convert to the start of the day in ISO format if valid
            let start_time: string | undefined;

            if (dateStr === '') {
                start_time = undefined;
            } else if (isValid) {
                start_time = DateTime.fromFormat(dateStr, LuxonFormat.ISO_8601).startOf('day').toISO() ?? '';
            } else {
                start_time = dateStr;
            }

            onChange({ start_time });
        },
        [onChange]
    );

    const onEndChange = useCallback(
        (dateStr: string, isValid: boolean) => {
            setIsEndValid(isValid);

            // Convert to the end of the day in ISO format if valid
            let end_time: string | undefined;

            if (dateStr === '') {
                end_time = undefined;
            } else if (isValid) {
                end_time = DateTime.fromFormat(dateStr, LuxonFormat.ISO_8601).endOf('day').toISO() ?? '';
            } else {
                end_time = dateStr;
            }

            onChange({ end_time });
        },
        [onChange]
    );

    useEffect(() => {
        onValidation?.(isStartValid && isEndValid && isCorrectOrder);
    }, [end, isCorrectOrder, isEndValid, isStartValid, onValidation, start]);

    return (
        <div className='flex flex-col gap-2 w-56'>
            <Label>Date Range</Label>

            <ManagedDatePicker hint='Start Date' onDateChange={onStartChange} value={start} />
            <ManagedDatePicker
                hint='End Date'
                onDateChange={onEndChange}
                value={end}
                validationError={
                    !isCorrectOrder && isStartValid && isEndValid ? 'End date must be on or after start date' : ''
                }
            />
        </div>
    );
};
