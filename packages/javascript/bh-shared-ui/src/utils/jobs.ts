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

import type { IndicatorType } from '../types';

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

type JobsFilterParams = {
    client_id?: string;
    end_time?: string;
    start_time?: string;
    status?: JobStatusCode;
};

export type FileIngestFilterParams = {
    user_id?: string;
    end_time?: string;
    start_time?: string;
    status?: JobStatusCode;
};

export type FinishedJobsFilter = EnabledCollections & JobsFilterParams;

export interface FinishedJobParams {
    filters?: FinishedJobsFilter;
    page: number;
    rowsPerPage: number;
}

export interface FileUploadParams {
    filters?: FileIngestFilterParams;
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

export type JobStatusCode = keyof typeof JOB_STATUS_MAP;

export const JOB_STATUS_INDICATORS: Record<JobStatusCode, { status: IndicatorType; pulse?: boolean }> = {
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
} as const satisfies Record<JobStatusCode, { status: IndicatorType; pulse?: boolean }>;

export const COLLECTION_MAP: Record<JobCollectionKey, string> = {
    session_collection: 'Sessions',
    local_group_collection: 'Local Groups',
    ad_structure_collection: 'AD Structure',
    cert_services_collection: 'Certificate Services',
    ca_registry_collection: 'CA Registry',
    dc_registry_collection: 'DC Registry',
} as const satisfies Record<JobCollectionKey, string>;

/** Given a FinishedJobsFilter state, return an object containing just the enabled collections */
export const getCollectionState = (state: FinishedJobsFilter): EnabledCollections =>
    jobCollectionKeys.reduce<EnabledCollections>((acc, key) => {
        if (state[key] === true) acc[key] = true;
        return acc;
    }, {});

/** Given a string, return true if it is a valid job collection key */
export const isCollectionKey = (key: string): key is JobCollectionKey =>
    (jobCollectionKeys as readonly string[]).includes(key);

export const NO_PERMISSION_MESSAGE =
    'Your role does not permit viewing finished job details. Please contact your administrator for assistance.';
export const NO_PERMISSION_KEY = 'finished-jobs-permission';

export const FETCH_ERROR_MESSAGE = 'Unable to fetch finished jobs. Please try again.';
export const FETCH_ERROR_KEY = 'finished-jobs-error';

export const FILE_INGEST_NO_PERMISSION_MESSAGE = `Your user role does not grant permission to view the file ingest jobs details. Please
    contact your administrator for details.`;
export const FILE_INGEST_NO_PERMISSION_KEY = 'file-upload-permission';

export const FILE_INGEST_FETCH_ERROR_MESSAGE = 'Unable to fetch file upload jobs. Please try again.';
export const FILE_INGEST_FETCH_ERROR_KEY = 'file-upload-error';

/** Returns a string listing all the collections methods for the given job */
export const toCollected = (job: Pick<ScheduledJobDisplay, JobCollectionKey>) =>
    jobCollectionKeys
        .filter((key) => job[key] === true)
        .map((key) => COLLECTION_MAP[key])
        .join(', ');
