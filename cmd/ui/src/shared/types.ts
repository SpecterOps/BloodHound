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

import { ISO_DATE_STRING } from 'bh-shared-ui';

export interface ClientToken {
    id: string;
    key: string;
}

export interface BaseClient {
    id: string;
    name: string;
    ip_address: string;
    hostname: string;
    version: string;
    configured_user: string;
    last_checkin: ISO_DATE_STRING;
    completed_task_count: number;
    completed_job_count: number;
    token: ClientToken | null;
    current_task_id: string | null;
    current_job_id: string | null;
    current_task?: Job;
    current_job?: Job;
    events: Event[];
}

export interface SharpHoundClient extends BaseClient {
    domain_controller: string;
    current_task?: SharpHoundJob;
    current_job?: SharpHoundJob;
    events: SharpHoundEvent[];
    type: 'sharphound';
}

export interface AzureHoundClient extends BaseClient {
    current_task?: AzureHoundJob;
    current_job?: AzureHoundJob;
    events: AzureHoundEvent[];
    type: 'azurehound';
}

export type Client = SharpHoundClient | AzureHoundClient;

export interface BaseJob {
    client_name: string;
    client_id: string;
    id: number;
    event_id: number;
    event_title: string;
    execution_time: string;
    status: number;
    start_time: string;
    end_time: string;
    log?: string;
    log_path: any;
    last_ingest: string;
    created_at: string;
    updated_at?: string;
    deleted_at?: {
        Time: string;
        Valid: boolean;
    };
}

export interface SharpHoundJob extends BaseJob {
    domain_controller: string | null;
    session_collection: boolean;
    ad_structure_collection: boolean;
    local_group_collection: boolean;
    ous: any[];
    domains: any[];
    all_trusted_domains: boolean;
}

export type AzureHoundJob = BaseJob;

export type Job = SharpHoundJob | AzureHoundJob;

export interface BaseEvent {
    id: string;
    client_id: string;
    rrule: string;
}

export interface SharpHoundEvent extends BaseEvent {
    session_collection: boolean;
    ad_structure_collection: boolean;
    local_group_collection: boolean;
    ous: any[];
    domains: any[];
    all_trusted_domains: boolean;
}

export type AzureHoundEvent = BaseEvent;

export type Event = SharpHoundEvent | AzureHoundEvent;

export const baseSharpHoundClient: SharpHoundClient = {
    id: '',
    name: '',
    version: '',
    completed_task_count: 0,
    completed_job_count: 0,
    configured_user: '',
    current_task_id: null,
    current_job_id: null,
    events: [],
    hostname: '',
    ip_address: '',
    last_checkin: '',
    domain_controller: '',
    token: null,
    type: 'sharphound',
};

export const baseSharpHoundEvent: SharpHoundEvent = {
    id: '',
    client_id: '',
    ad_structure_collection: false,
    local_group_collection: false,
    session_collection: false,
    ous: [],
    domains: [],
    all_trusted_domains: false,
    rrule: 'RRULE:FREQ=DAILY;INTERVAL=1',
};

export const baseAzureHoundClient: AzureHoundClient = {
    id: '',
    name: '',
    events: [],
    type: 'azurehound',
    completed_task_count: 0,
    completed_job_count: 0,
    configured_user: '',
    token: null,
    hostname: '',
    version: '',
    ip_address: '',
    last_checkin: '',
    current_task_id: null,
    current_job_id: null,
};

export const baseAzureHoundEvent: AzureHoundEvent = {
    id: '',
    client_id: '',
    rrule: 'RRULE:FREQ=DAILY;INTERVAL=1',
};

export interface DataPoint {
    x: number;
    y: number;
    name?: string;
}

interface Color {
    color: string;
}

export type FindingSeverityDataPoint = DataPoint & Color;
