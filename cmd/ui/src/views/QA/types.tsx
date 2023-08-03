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

export interface ActiveDirectoryQualityStat {
    groups: number;
    ous: number;
    gpos: number;
    acls: number;
    relationships: number;
    users: number;
    containers?: number;
    computers: number;
    domains?: number;
    sessions: number;
    local_group_completeness: number;
    session_completeness: number;
    created_at: string;
}

export interface ActiveDirectoryDataQualityResponse {
    start: string;
    end: string;
    limit: number;
    data: ActiveDirectoryQualityStat[];
}

export interface AzureDataQualityStat {
    tenantid: string;
    users: number;
    groups: number;
    apps: number;
    service_principals: number;
    devices: number;
    management_groups: number;
    subscriptions: number;
    tenants?: number;
    resource_groups: number;
    vms: number;
    key_vaults: number;
    relationships: number;
    run_id: string;
}

export interface AzureDataQualityResponse {
    start: string;
    end: string;
    limit: number;
    data: AzureDataQualityStat[];
}
