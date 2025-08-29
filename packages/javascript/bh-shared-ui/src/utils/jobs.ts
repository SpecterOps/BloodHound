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

import { IndicatorType } from '../types';
import { typedEntries } from './object';

const jobCollectionKeys = [
    'session_collection',
    'local_group_collection',
    'ad_structure_collection',
    'cert_services_collection',
    'ca_registry_collection',
    'dc_registry_collection',
] as const;

type JobCollectionKey = (typeof jobCollectionKeys)[number];

export type EnabledCollections = Partial<Record<JobCollectionKey, boolean>>;

const jobParamsKeys = ['client_name', 'end_time', 'start_time', 'status'] as const;

type JobsParamsKey = (typeof jobParamsKeys)[number];

type JobsFilterParams = Partial<Record<JobsParamsKey, string>>;

export type FinishedJobsFilter = EnabledCollections & JobsFilterParams;

export interface FinishedJobParams {
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
} as const satisfies Record<number, string>;

export const JOB_STATUS_INDICATORS: Record<number, { status: IndicatorType; pulse?: boolean }> = {
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
} as const satisfies Record<number, { status: IndicatorType; pulse?: boolean }>;

export const COLLECTION_MAP: Record<JobCollectionKey, string> = {
    session_collection: 'Sessions',
    local_group_collection: 'Local Groups',
    ad_structure_collection: 'AD Structure',
    cert_services_collection: 'Certificate Services',
    ca_registry_collection: 'CA Registry',
    dc_registry_collection: 'DC Registry',
} as const satisfies Record<JobCollectionKey, string>;

/** Given a FinishedJobsFilters state, return an object containing just the enabled collections */
export const getCollectionState = (state: FinishedJobsFilter): EnabledCollections => {
    const entries = typedEntries(state).filter(([key, value]) => isCollectionKey(key) && value === true);
    return Object.fromEntries(entries) as EnabledCollections;
};

/** Given a string, return true if it is a valid job collection key */
export const isCollectionKey = (key: string): key is JobCollectionKey =>
    (jobCollectionKeys as readonly string[]).includes(key);

export const NO_PERMISSION_MESSAGE =
    'Your role does not permit viewing finished job details. Please contact your administrator for assistance.';
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch finished jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

/** Returns a string listing all the collections methods for the given job */
export const toCollected = (job: Pick<ScheduledJobDisplay, JobCollectionKey>) =>
    typedEntries(job)
        .filter(([key, value]) => isCollectionKey(key) && value)
        .map(([key]) => COLLECTION_MAP[key])
        .join(', ');
