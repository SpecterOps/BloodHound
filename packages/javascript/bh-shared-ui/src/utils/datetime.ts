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

export const nowDateTime = () => DateTime.local();
export const now = () => nowDateTime().toJSDate();
export const fiveYearsAgo = () => nowDateTime().minus({ years: 5 }).toJSDate();

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
    LOCAL_TIME = 'HH:mm:ss',
    DATETIME_WITH_LINEBREAKS = "yyyy-MM-dd '\n'T ZZZZ\n'(GMT'ZZZ')'",
    TIME_WITH_LINEBREAKS = "T ZZZZ'\n(GMT'ZZZ')'",
}

export type ISO_DATE_STRING = string;

/** Returns the duration between 2 given ISO datetime strings in a simple format */
export const getSimpleDuration = (start: ISO_DATE_STRING, end: ISO_DATE_STRING): string => {
    const interval = Interval.fromISO(`${start}/${end}`);

    if (!interval.isValid) {
        return '';
    }

    const minutes = Math.floor(interval.length('minutes'));
    const days = Math.floor(interval.length('days'));

    if (days === 1) {
        return 'a day';
    } else if (days >= 2) {
        return `${days} days`;
    } else if (minutes === 1) {
        return `${minutes} min`;
    } else {
        return `${minutes} mins`;
    }
};

/** Returns the given ISO datetime string formatted with the timezone */
export const toFormatted = (dateStr: ISO_DATE_STRING): string =>
    DateTime.fromISO(dateStr).toFormat(LuxonFormat.DATE_WITHOUT_GMT);

export const floorToNearestMinute = (dateTime: DateTime) => {
    return dateTime.set({
        minute: Math.floor(dateTime.minute) * 1,
        second: 0,
        millisecond: 0,
    });
};
