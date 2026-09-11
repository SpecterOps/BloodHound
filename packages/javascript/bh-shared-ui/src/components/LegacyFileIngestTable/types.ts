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

export type FileUploadJob = {
    id: number;
    user_id: string;
    user_email_address: string;
    status: FileUploadJobStatus;
    status_message: string;
    start_time: string;
    end_time: string;
    last_ingest: string;
};

export enum FileUploadJobStatus {
    INVALID = -1,
    READY = 0,
    RUNNING = 1,
    COMPLETE = 2,
    CANCELED = 3,
    TIMED_OUT = 4,
    FAILED = 5,
    INGESTING = 6,
    ANALYZING = 7,
    PARTIALLY_COMPLETE = 8,
}

export const FileUploadJobStatusToString: Record<FileUploadJobStatus, string> = {
    [-1]: 'Invalid',
    0: 'Ready',
    1: 'Running',
    2: 'Complete',
    3: 'Canceled',
    4: 'Timed Out',
    5: 'Failed',
    6: 'Ingesting',
    7: 'Analyzing',
    8: 'Partially Complete',
};
