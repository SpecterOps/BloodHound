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

import { AxiosRequestConfig } from 'axios';
import { AssetGroupTagSelector, AssetGroupTagSelectorSeed, SSOProviderConfiguration } from './types';
import { ConfigurationPayload } from './utils';

export type RequestOptions = AxiosRequestConfig;

export interface LoginRequest {
    login_method: string;
    secret: string;
    username: string;
    otp?: string;
}

export type PreviewSelectorsRequest = { seeds: SelectorSeedRequest[] };

// This type makes it so that `selector_id` is optional in the selector seed request shape.
// The `selector_id` will only be available when updating an already existing selector.
export type SelectorSeedRequest = Omit<AssetGroupTagSelectorSeed, 'selector_id'> & Partial<AssetGroupTagSelectorSeed>;

export type CreateSelectorRequest = Partial<Omit<AssetGroupTagSelector, 'seeds' | 'id'> & SelectorSeedRequest>;

export type UpdateSelectorRequest = Partial<
    Omit<CreateSelectorRequest, 'id | disabled_at'> & { disabled: boolean | string }
>;

export interface CreateAssetGroupRequest {
    name: string;
    tag: string;
}

export interface UpdateAssetGroupRequest {
    name: string;
}

export interface CreateAssetGroupSelectorRequest {
    node_label: string;
    selector_name: string;
    sid: string;
}

export interface UpdateAssetGroupSelectorRequest {
    selector_name: string;
    sid: string;
    action: 'add' | 'remove';
}

export interface CreateSharpHoundClientRequest {
    domain_controller: string;
    name: string;
    events?: any[];
    type: 'sharphound';
}

export interface CreateAzureHoundClientRequest {
    name: string;
    events?: any[];
    type: 'azurehound';
}

export interface UpdateSharpHoundClientRequest {
    domain_controller: string;
    name: string;
}

export interface UpdateAzureHoundClientRequest {
    name: string;
}

export interface CreateScheduledSharpHoundJobRequest {
    session_collection: boolean;
    ad_structure_collection: boolean;
    local_group_collection: boolean;
    cert_services_collection: boolean;
    ca_registry_collection: boolean;
    dc_registry_collection: boolean;
    domain_controller?: string;
    ous: string[];
    domains: string[];
    all_trusted_domains: boolean;
}

export type CreateScheduledAzureHoundJobRequest = Record<string, never>;

export type CreateScheduledJobRequest = CreateScheduledSharpHoundJobRequest | CreateScheduledAzureHoundJobRequest;

export interface ClientStartJobRequest {
    id: number;
    start_time: string;
}

export interface ClientEndJobRequest {
    end_time: string;
    id: number;
    log: string;
}

export interface CreateSharpHoundEventRequest {
    client_id: string;
    rrule: string;
    session_collection: boolean;
    ad_structure_collection: boolean;
    local_group_collection: boolean;
    cert_services_collection: boolean;
    ca_registry_collection: boolean;
    dc_registry_collection: boolean;
    ous: string[];
    domains: string[];
    all_trusted_domains: boolean;
}

export interface CreateAzureHoundEventRequest {
    client_id: string;
    rrule: string;
}

export interface UpdateSharpHoundEventRequest {
    client_id: string;
    rrule: string;
    session_collection: boolean;
    ad_structure_collection: boolean;
    local_group_collection: boolean;
    cert_services_collection: boolean;
    ca_registry_collection: boolean;
    dc_registry_collection: boolean;
    ous: string[];
    domains: string[];
    all_trusted_domains: boolean;
}

export interface UpdateAzureHoundEventRequest {
    client_id: string;
    rrule: string;
}

export interface PutUserAuthSecretRequest {
    currentSecret?: string;
    secret: string;
    needsPasswordReset: boolean;
}

export interface CreateOIDCProviderRequest extends SSOProviderConfiguration {
    name: string;
    client_id: string;
    issuer: string;
}
export type UpdateOIDCProviderRequest = Partial<CreateOIDCProviderRequest>;
export type UpsertOIDCProviderRequest = CreateOIDCProviderRequest | UpdateOIDCProviderRequest;

export type PostureRequest = {
    from: string;
    to: string;
    domain_sid?: string;
    sort_by?: string;
};

export type RiskDetailsRequest = {
    finding: string;
    skip: number;
    limit: number;
    sort_by?: string;
    Accepted?: string;
};

export interface CreateUserQueryRequest {
    name: string;
    query: string;
}

export interface ClearDatabaseRequest {
    deleteCollectedGraphData: boolean;
    deleteFileIngestHistory: boolean;
    deleteDataQualityHistory: boolean;
    deleteAssetGroupSelectors: number[];
}

export interface UpdateUserRequest {
    firstName: string;
    lastName: string;
    emailAddress: string;
    principal: string;
    roles: number[];
    SSOProviderId?: number;
    is_disabled?: boolean;
    /** @deprecated: this is left to maintain backwards compatability, please use SSOProviderId instead */
    SAMLProviderId?: string;
}

export interface CreateUserRequest extends Omit<UpdateUserRequest, 'is_disabled'> {
    password?: string;
    needsPasswordReset?: boolean;
}

export type UpdateConfigurationRequest = ConfigurationPayload;
