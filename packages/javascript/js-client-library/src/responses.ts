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

import type { AxiosResponse } from 'axios';
import {
    AssetGroupTag,
    AssetGroupTagCertificationRecord,
    AssetGroupTagHistoryRecord,
    AssetGroupTagMember,
    AssetGroupTagSelector,
    Client,
    CollectorManifest,
    CommunityCollectorType,
    CustomNodeKindType,
    EnterpriseCollectorType,
    FileIngestCompletedTask,
    FileIngestJob,
    GraphData,
    NodeSourceTypes,
    ScheduledJobDisplay,
    TimestampFields,
} from './types';
import { ConfigurationPayload } from './utils/config';

export interface BasicResponse<T> {
    data: T;
}

export interface TimeWindowedResponse<T> extends BasicResponse<T> {
    start: string;
    end: string;
}

export type PaginatedResponse<T> = Partial<TimeWindowedResponse<T>> &
    Required<BasicResponse<T>> & {
        count: number;
        limit: number;
        skip: number;
    };

export type Environment = {
    type: 'active-directory' | 'azure';
    impactValue: number;
    name: string;
    id: string;
    collected: boolean;
};

export type GraphResponse = BasicResponse<GraphData>;

export type ActiveDirectoryQualityStat = TimestampFields & {
    users: number;
    computers: number;
    groups: number;
    ous: number;
    gpos: number;
    aiacas: number;
    rootcas: number;
    enterprisecas: number;
    ntauthstores: number;
    certtemplates: number;
    issuancepolicies: number;
    acls: number;
    relationships: number;
    sessions: number;
    local_group_completeness: number;
    session_completeness: number;
    containers?: number;
    domains?: number;
};

export type ActiveDirectoryDataQualityResponse = PaginatedResponse<ActiveDirectoryQualityStat[]>;

export type AzureDataQualityStat = TimestampFields & {
    run_id: string;
    relationships: number;
    users: number;
    groups: number;
    apps: number;
    service_principals: number;
    devices: number;
    management_groups: number;
    subscriptions: number;
    resource_groups: number;
    vms: number;
    key_vaults: number;
    automation_accounts: number;
    container_registries: number;
    function_apps: number;
    logic_apps: number;
    managed_clusters: number;
    vm_scale_sets: number;
    web_apps: number;
    tenants?: number;
    tenantid?: string;
};

export type AzureDataQualityResponse = PaginatedResponse<AzureDataQualityStat[]>;

type PostureStat = TimestampFields & {
    domain_sid: string;
    exposure_index: number;
    tier_zero_count: number;
    critical_risk_count: number;
    id: number;
};

export type PostureResponse = PaginatedResponse<PostureStat[]>;

type PostureFindingTrend = {
    environment_id: string;
    finding: string;
    finding_count_start: number;
    finding_count_end: number;
    finding_count_increase: number;
    finding_count_decrease: number;
    composite_risk: number;
    display_title: string;
    display_type: string;
};

export type PostureFindingTrendsResponse = TimeWindowedResponse<{
    findings: PostureFindingTrend[];
    total_finding_count_start: number;
    total_finding_count_end: number;
}>;

export type PostureHistoryData = {
    date: string;
    value: number;
};

export type PostureHistoryResponse = TimeWindowedResponse<PostureHistoryData[]> & {
    data_type: string;
};

type DatapipeStatus = {
    status: 'idle' | 'ingesting' | 'analyzing' | 'purging';
    last_complete_analysis_at: string;
    updated_at: string;
};

export type DatapipeStatusResponse = BasicResponse<DatapipeStatus>;

export type AuthToken = TimestampFields & {
    hmac_method: string;
    id: string;
    last_access: string;
    name: string;
    user_id: string;
};

export type ListAuthTokensResponse = BasicResponse<{ tokens: AuthToken[] }>;

export type NewAuthToken = AuthToken & {
    key: string;
};

export type CreateAuthTokenResponse = BasicResponse<NewAuthToken>;

export type AssetGroupTagsHistory = PaginatedResponse<{ records: AssetGroupTagHistoryRecord[] }>;

export type PreviewSelectorsResponse = BasicResponse<{ members: AssetGroupTagMember[] }>;
export type AssetGroupTagsCertification = PaginatedResponse<{ members: AssetGroupTagCertificationRecord[] }>;

export interface AssetGroupTagMemberListItem extends AssetGroupTagMember {
    source: NodeSourceTypes;
}

export interface AssetGroupTagMemberInfo extends AssetGroupTagMember {
    properties: Record<string, any>;
    selectors: AssetGroupTagSelector[];
}

export type AssetGroupTagSearchResponse = BasicResponse<{
    tags: AssetGroupTag[];
    selectors: AssetGroupTagSelector[];
    members: AssetGroupTagMember[];
}>;

export type AssetGroupTagResponse = BasicResponse<{ tag: AssetGroupTag }>;
export type AssetGroupTagsResponse = BasicResponse<{ tags: AssetGroupTag[] }>;
export type AssetGroupTagSelectorResponse = BasicResponse<{ selector: AssetGroupTagSelector }>;
export type AssetGroupTagSelectorsResponse = PaginatedResponse<{ selectors: AssetGroupTagSelector[] }>;
export type AssetGroupTagMembersResponse = PaginatedResponse<{ members: AssetGroupTagMemberListItem[] }>;
export type AssetGroupTagMemberInfoResponse = BasicResponse<{
    member: AssetGroupTagMemberInfo;
}>;

export type AssetGroupSelector = TimestampFields & {
    id: number;
    asset_group_id: number;
    name: string;
    selector: string;
    system_selector: boolean;
};

export type AssetGroup = TimestampFields & {
    id: number;
    name: string;
    tag: string;
    member_count: number;
    system_group: boolean;
    Selectors: AssetGroupSelector[];
};

export type AssetGroupMember = {
    asset_group_id: number;
    custom_member: boolean;
    environment_id: string;
    environment_kind: string;
    kinds: string[];
    name: string;
    object_id: string;
    primary_kind: string;
};

export type AssetGroupMemberCounts = {
    total_count: number;
    counts: Record<AssetGroupMember['primary_kind'], number>;
};

export type AssetGroupResponse = BasicResponse<{ asset_groups: AssetGroup[] }>;

export type AssetGroupMembersResponse = PaginatedResponse<{ members: AssetGroupMember[] }>;

export type AssetGroupMemberCountsResponse = BasicResponse<AssetGroupMemberCounts>;

export type SavedQuery = {
    id: number;
    name: string;
    description: string;
    query: string;
    user_id: string;
};

export type SavedQueryPermissionsResponse = {
    shared_to_user_ids: string[];
    query_id: number | undefined;
    public: boolean;
};

export type ListFileIngestJobsResponse = PaginatedResponse<FileIngestJob[]>;

export type ListFileTypesForIngestResponse = BasicResponse<string[]>;

export type StartFileIngestResponse = BasicResponse<FileIngestJob>;

export type UploadFileToIngestResponse = null;

export type FileIngestCompletedTasksResponse = BasicResponse<FileIngestCompletedTask[] | null>;

export type EndFileIngestResponse = null;

export type ConfigurationWithMetadata<T> = TimestampFields &
    T & {
        name: string;
        description: string;
        id: number;
    };

export type GetConfigurationResponse = BasicResponse<ConfigurationWithMetadata<ConfigurationPayload>[]>;

export type UpdateConfigurationResponse = BasicResponse<ConfigurationPayload>;

export type GetCollectorsResponse = BasicResponse<{
    latest: string;
    versions: {
        version: string;
        sha256sum: string;
        deprecated: boolean;
    }[];
}>;

export type GetCommunityCollectorsResponse = BasicResponse<Record<CommunityCollectorType, CollectorManifest[]>>;

export type GetEnterpriseCollectorsResponse = BasicResponse<Record<EnterpriseCollectorType, CollectorManifest[]>>;

export type GetCustomNodeKindsResponse = BasicResponse<CustomNodeKindType[]>;

export type GetScheduledJobDisplayResponse = PaginatedResponse<ScheduledJobDisplay[]>;

export type GetExportQueryResponse = AxiosResponse<Blob>;

export type GetClientResponse = PaginatedResponse<Client[]>;
