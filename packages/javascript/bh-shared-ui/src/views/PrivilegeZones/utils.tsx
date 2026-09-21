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

import { AssetGroupTag, AssetGroupTagSelector, SeedTypes } from 'js-client-library';
import { DateTime } from 'luxon';
import { ErrorOption, FieldValues, UseFormReturn } from 'react-hook-form';
import { CustomRangeError, END_DATE, LuxonFormat, START_DATE } from '../../utils';

export const isTag = (data: any): data is AssetGroupTag => {
    return 'kind_id' in data;
};

export const isRule = (data: any): data is AssetGroupTagSelector => {
    return 'is_default' in data;
};

export const getRuleSeedType = (rule: AssetGroupTagSelector): SeedTypes => {
    const firstSeed = rule.seeds[0];

    return firstSeed.type;
};

export const TagTabValue = 'tag' as const;
export const RuleTabValue = 'rule' as const;
export const ObjectTabValue = 'object' as const;

export type DetailsTabOption = typeof TagTabValue | typeof RuleTabValue | typeof ObjectTabValue;

export const getListHeight = (windoHeight: number) => {
    if (windoHeight > 1080) return 760;
    if (1080 >= windoHeight && windoHeight > 900) return 640;
    if (900 >= windoHeight) return 436;
    return 436;
};

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

export const measureElement: ((element: Element) => number) | undefined =
    typeof window !== 'undefined' && navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined;
