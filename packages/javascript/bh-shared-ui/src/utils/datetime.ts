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

import { DateTime, Interval } from 'luxon';

export const LUXON_DATETIME_REGEX = /(\d\d\d\d)-(\d||\d\d)-(\d||\d\d) (\d||\d\d):\d\d ..T \(GMT-\d\d\d\d\)/;

export enum LuxonFormat {
    DATETIME = "yyyy-MM-dd T ZZZZ '(GMT'ZZZ')'",
    DATE_WITHOUT_GMT = 'yyyy-MM-dd HH:mm ZZZZ',
    DATETIME_WITHOUT_TIMEZONE = 'yyyy-MM-dd T',
    ISO_8601 = 'yyyy-MM-dd',
    YEAR_MONTH_DAY_SLASHES = 'yyyy/MM/dd',
    YEAR_MONTH_DAY_DOTS = 'yyyy.MM.dd',
    TIMEZONE_AND_GMT_OFFSET = "ZZZZ '(GMT'ZZZ')'",
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

/** Returns the given ISO datetime string formatted with the the timezone */
export const toFormatted = (dateStr: string) => DateTime.fromISO(dateStr).toFormat(LuxonFormat.DATE_WITHOUT_GMT);

/** Returns the duration, in mins, between 2 given ISO datetime strings */
export const toMins = (start: string, end: string) => {
    const interval = Interval.fromDateTimes(DateTime.fromISO(start), DateTime.fromISO(end));

    if (!interval.isValid) {
        return '';
    }

    return Math.floor(interval.length('minutes')) + ' Min';
};
