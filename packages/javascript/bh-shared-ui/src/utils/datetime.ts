// Copyright 2023 Specter Ops, Inc.
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

import { DateTime } from 'luxon';

export const LUXON_DATETIME_REGEX = /(\d\d\d\d)-(\d||\d\d)-(\d||\d\d) (\d||\d\d):\d\d ..T \(GMT-\d\d\d\d\)/;

export enum LuxonFormat {
    DATETIME = "yyyy-MM-dd T ZZZZ '(GMT'ZZZ')'",
    TIME = "T ZZZZ' (GMT'ZZZ')'",
    DATETIME_WITH_LINEBREAKS = "yyyy-MM-dd '\n'T ZZZZ\n'(GMT'ZZZ')'",
    TIME_WITH_LINEBREAKS = "T ZZZZ'\n(GMT'ZZZ')'",
}

export type ISO_DATE_STRING = string;

export const calculateJobDuration = (start: ISO_DATE_STRING, end: ISO_DATE_STRING): string => {
    const duration = DateTime.fromISO(end).diff(DateTime.fromISO(start), ['minutes', 'days']);

    const minutes = Math.floor(duration.minutes);
    const days = Math.floor(duration.days);

    if (days === 1) {
        return 'a day';
    }
    if (days >= 2) {
        return `${days} days`;
    }
    if (minutes === 1) {
        return `${minutes} minute`;
    }
    return `${minutes} minutes`;
};

const formatSimple = (value: any): string => {
    const type = typeof value;
    if (type === 'number') {
        const currentDate = Math.round(new Date().getTime() / 1000);

        //315536400 = January 1st, 1980. Seems like a safe bet
        if (value > 315536400 && value < currentDate) {
            return DateTime.fromSeconds(value).toFormat(LuxonFormat.DATETIME);
        } else {
            return `${value}`.toLocaleString();
        }
    }

    if (type === 'boolean') {
        return value.toString().toUpperCase();
    }

    const potentialDate: any = DateTime.fromISO(value);

    if (potentialDate.invalid === null) return potentialDate.toFormat(LuxonFormat.DATETIME);

    return value;
};

export const format = (value: any): string | string[] | null => {
    if (Array.isArray(value)) {
        const fields: string[] = [];
        value.forEach((val) => {
            fields.push(formatSimple(val));
        });
        return fields;
    } else {
        return formatSimple(value);
    }
};
