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

import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { CSSProperties } from 'react';
import { ErrorOption, FieldValues, UseFormReturn } from 'react-hook-form';
import { CustomRangeError, END_DATE, LuxonFormat, START_DATE } from '../../utils';

export const ItemSkeleton = (title: string, key: number, height?: string, style?: CSSProperties) => {
    return (
        <li
            key={key}
            data-testid={`privilege-zones_${title.toLowerCase()}-list_loading-skeleton`}
            style={style}
            className='border-y border-neutral-light-3 dark:border-neutral-dark-3 relative w-full'>
            <Skeleton className={`${height ?? 'min-h-10'} rounded-none`} />
        </li>
    );
};

export const itemSkeletons = [ItemSkeleton, ItemSkeleton, ItemSkeleton];

export function validateFormDates<T extends FieldValues>(
    form: UseFormReturn<T>,
    startDate: DateTime,
    endDate: DateTime
) {
    return () => {
        form.clearErrors();
        const errors: { name: typeof START_DATE | typeof END_DATE; error: ErrorOption }[] = [];

        if (!startDate.isValid) {
            errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_DATE } });
        }
        if (!endDate.isValid) {
            errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_DATE } });
        }
        if (errors.length === 0 && startDate > endDate) {
            errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_RANGE_START } });
            errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_RANGE_END } });
        }

        if (errors.length > 0) {
            //@ts-ignore
            errors.forEach((error) => form.setError(error.name, error.error));
            return false;
        } else {
            form.clearErrors();
            return true;
        }
    };
}

export function getStartAndEndDateTimes(start: string | undefined = '', end: string | undefined = '') {
    // If the start date is empty use the start of epoch time
    const startDate = start !== '' ? DateTime.fromFormat(start, LuxonFormat.ISO_8601) : DateTime.fromMillis(0);
    // Use the client time if the end date is empty
    const endDate = end !== '' ? DateTime.fromFormat(end, LuxonFormat.ISO_8601) : DateTime.now();

    return { startDate, endDate };
}

export const toDate = DateTime.local().toJSDate();
export const fromDate = DateTime.fromJSDate(toDate).minus({ years: 1 }).toJSDate();
