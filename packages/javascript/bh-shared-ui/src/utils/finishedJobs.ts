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

import type { ScheduledJobDisplay } from 'js-client-library';
import { DateTime, Interval } from 'luxon';
import type { OptionsObject } from 'notistack';
import type { StatusType } from '../components';
import { LuxonFormat } from './datetime';

export interface FinishedJobParams {
    page: number;
    rowsPerPage: number;
}

export const JOB_STATUS_MAP: Record<number, { label: string; status: StatusType; pulse?: boolean }> = {
    [-1]: { label: 'Invalid', status: 'bad' },
    0: { label: 'Ready', status: 'good' },
    1: { label: 'Running', status: 'pending', pulse: true },
    2: { label: 'Complete', status: 'good' },
    3: { label: 'Canceled', status: 'bad' },
    4: { label: 'Timed Out', status: 'bad' },
    5: { label: 'Failed', status: 'bad' },
    6: { label: 'Ingesting', status: 'pending', pulse: true },
    7: { label: 'Analyzing', status: 'pending' },
    8: { label: 'Partially Completed', status: 'pending' },
};

export const FINISHED_JOBS_LOG_HEADERS = [
    { label: 'ID / Client / Status', width: '240px' },
    { label: 'Status Message', width: '240px' },
    { label: 'Start Time', width: '110px' },
    { label: 'Duration', width: '85px' },
    { label: 'Data Collected', width: '240px' },
];

export const COLLECTION_MAP = new Map(
    Object.entries({
        session_collection: 'Sessions',
        local_group_collection: 'Local Groups',
        ad_structure_collection: 'AD Structure',
        cert_services_collection: 'Certificate Services',
        ca_registry_collection: 'CA Registry',
        dc_registry_collection: 'DC Registry',
    })
);

export const PERSIST_NOTIFICATION: OptionsObject = {
    persist: true,
    anchorOrigin: { vertical: 'top', horizontal: 'right' },
};

export const NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the finished jobs details. Please
    contact your administrator for details.`;
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

/** Returns the duration, in mins, between 2 given ISO datetime strings */
export const toMins = (start: string, end: string) =>
    Math.floor(Interval.fromDateTimes(DateTime.fromISO(start), DateTime.fromISO(end)).length('minutes')) + ' Min';

/** Returns a string listing all the collections methods for the given job */
export const toCollected = (job: ScheduledJobDisplay) =>
    Object.entries(job)
        .filter(([key, value]) => COLLECTION_MAP.has(key) && value)
        .map(([key]) => COLLECTION_MAP.get(key))
        .join(', ');

/** Returns the given ISO datetime string formatted with the the timezone */
export const toFormatted = (dateStr: string) => DateTime.fromISO(dateStr).toFormat(LuxonFormat.DATE_WITHOUT_GMT);
