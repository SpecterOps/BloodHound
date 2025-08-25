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
    filters: FinishedJobsFilters;
    page: number;
    rowsPerPage: number;
}

export const JOB_STATUS_MAP: Record<number, string> = {
    [-1]: 'Invalid',
    0: 'Ready',
    1: 'Running',
    2: 'Complete',
    3: 'Canceled',
    4: 'Timed Out',
    5: 'Failed',
    6: 'Ingesting',
    7: 'Analyzing',
    8: 'Partially Completed',
} as const;

export const JOB_STATUS_INDICATORS: Record<number, { status: StatusType; pulse?: boolean }> = {
    [-1]: { status: 'bad' },
    0: { status: 'good' },
    1: { status: 'pending', pulse: true },
    2: { status: 'good' },
    3: { status: 'bad' },
    4: { status: 'bad' },
    5: { status: 'bad' },
    6: { status: 'pending', pulse: true },
    7: { status: 'pending' },
    8: { status: 'pending' },
} as const;

export const FINISHED_JOBS_LOG_HEADERS = [
    { label: 'ID / Client / Status', width: '240px' },
    { label: 'Status Message', width: '240px' },
    { label: 'Start Time', width: '110px' },
    { label: 'Duration', width: '85px' },
    { label: 'Data Collected', width: '240px' },
] as const;

export type JobCollectionType =
    | 'session_collection'
    | 'local_group_collection'
    | 'ad_structure_collection'
    | 'cert_services_collection'
    | 'ca_registry_collection'
    | 'dc_registry_collection';

export type FinishedJobsFilters = Record<string, boolean | string> & {
    client_name?: string;
    end_time?: string;
    start_time?: string;
    status?: string;
    session_collection?: boolean;
    local_group_collection?: boolean;
    ad_structure_collection?: boolean;
    cert_services_collection?: boolean;
    ca_registry_collection?: boolean;
    dc_registry_collection?: boolean;
};

export type EnabledCollections = {
    session_collection?: boolean;
    local_group_collection?: boolean;
    ad_structure_collection?: boolean;
    cert_services_collection?: boolean;
    ca_registry_collection?: boolean;
    dc_registry_collection?: boolean;
};

export const COLLECTION_MAP: Map<JobCollectionType, string> = new Map();
COLLECTION_MAP.set('session_collection', 'Sessions');
COLLECTION_MAP.set('local_group_collection', 'Local Groups');
COLLECTION_MAP.set('ad_structure_collection', 'AD Structure');
COLLECTION_MAP.set('cert_services_collection', 'Certificate Services');
COLLECTION_MAP.set('ca_registry_collection', 'CA Registry');
COLLECTION_MAP.set('dc_registry_collection', 'DC Registry');

export const PERSIST_NOTIFICATION: OptionsObject = {
    persist: true,
    anchorOrigin: { vertical: 'top', horizontal: 'right' },
};

export const NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the finished jobs details. Please
    contact your administrator for details.`;
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

export const getCollectionState = (state: Record<string, unknown>) =>
    Object.keys(state)
        .filter(isCollectionKey)
        .reduce((collections, key) => {
            collections[key] = state[key] as boolean;
            return collections;
        }, {} as EnabledCollections);

export const isCollectionKey = (key: string): key is JobCollectionType => key.endsWith('_collection');

/** Returns a string listing all the collections methods for the given job */
export const toCollected = (job: ScheduledJobDisplay) =>
    Object.entries(job)
        .filter(([key, value]) => COLLECTION_MAP.has(key as JobCollectionType) && value)
        .map(([key]) => COLLECTION_MAP.get(key as JobCollectionType))
        .join(', ');

/** Returns the duration, in mins, between 2 given ISO datetime strings */
export const toMins = (start: string, end: string) =>
    Math.floor(Interval.fromDateTimes(DateTime.fromISO(start), DateTime.fromISO(end)).length('minutes')) + ' Min';

/** Returns the given ISO datetime string formatted with the the timezone */
export const toFormatted = (dateStr: string) => DateTime.fromISO(dateStr).toFormat(LuxonFormat.DATE_WITHOUT_GMT);
