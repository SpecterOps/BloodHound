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

import * as axios from 'axios';
import { ConfigurationPayload } from './utils';

export type RequestOptions = axios.AxiosRequestConfig;

export interface Serial {
    id: number;
    created_at: string;
    updated_at: string;
}

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

export interface AssetGroupMemberParams {
    environment_kind?: string;
    environment_id?: string;
    primary_kind?: string;
    custom_member?: string;
    skip?: number;
    limit?: number;
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

export interface CreateSAMLProviderFormInputs extends SSOProviderConfiguration {
    name: string;
    metadata: FileList;
}
export type UpdateSAMLProviderFormInputs = Partial<CreateSAMLProviderFormInputs>;
export type UpsertSAMLProviderFormInputs = CreateSAMLProviderFormInputs | UpdateSAMLProviderFormInputs;

export interface CreateOIDCProviderRequest extends SSOProviderConfiguration {
    name: string;
    client_id: string;
    issuer: string;
}
export type UpdateOIDCProviderRequest = Partial<CreateOIDCProviderRequest>;
export type UpsertOIDCProviderRequest = CreateOIDCProviderRequest | UpdateOIDCProviderRequest;

export interface SAMLProviderInfo extends Serial {
    name: string;
    display_name: string;
    idp_issuer_uri: string;
    idp_sso_uri: string;
    principal_attribute_mappings: string[] | null;
    sp_issuer_uri: string;
    sp_sso_uri: string;
    sp_metadata_uri: string;
    sp_acs_uri: string;
    sso_provider_id: number;
}

export interface OIDCProviderInfo extends Serial {
    client_id: string;
    issuer: string;
    sso_provider_id: number;
}

export interface SSOProviderConfiguration {
    config: {
        auto_provision: {
            enabled: boolean;
            default_role_id: number;
            role_provision: boolean;
        };
    };
}

export interface SSOProvider extends Serial, SSOProviderConfiguration {
    name: string;
    slug: string;
    type: 'OIDC' | 'SAML';
    details: SAMLProviderInfo | OIDCProviderInfo;
    login_uri: string;
    callback_uri: string;
}

export interface ListSSOProvidersResponse {
    data: SSOProvider[];
}

export interface User {
    id: string;
    sso_provider_id: number | null;
    AuthSecret: any;
    roles: Role[];
    first_name: string | null;
    last_name: string | null;
    email_address: string | null;
    principal_name: string;
    last_login: string;
    created_at: string;
    updated_at: string;
    is_disabled: boolean;
    eula_accepted: boolean;
}

interface Permission {
    id: number;
    name: string;
    authority: string;
}

export interface Role {
    id: number;
    name: string;
    description: string;
    permissions: Permission[];
}

export interface ListRolesResponse {
    data: {
        roles: Role[];
    };
}

export interface ListUsersResponse {
    data: {
        users: User[];
    };
}

export interface LoginRequest {
    login_method: string;
    secret: string;
    username: string;
    otp?: string;
}

export interface LoginResponse {
    data: {
        user_id: string;
        auth_expired: boolean;
        eula_accepted: boolean;
        session_token: string;
        user_name: string;
    };
}

export type CollectorType = 'sharphound_enterprise' | 'sharphound' | 'azurehound';
export type CommunityCollectorType = Exclude<CollectorType, 'sharphound_enterprise'>;

export interface GetCollectorsResponse {
    data: {
        latest: string;
        versions: {
            version: string;
            sha256sum: string;
            deprecated: boolean;
        }[];
    };
}

export interface GetEnterpriseCollectorsResponse {
    data: {
        sharphound: EnterpriseCollector[];
        azurehound: EnterpriseCollector[];
    }
}

interface EnterpriseCollector {
    version: string;
    version_meta: VersionMeta;
    release_date: string;
    release_assets: EnterpriseCollectorAsset[];
}

interface VersionMeta {
    major: number;
    minor: number;
    patch: number;
    prerelease: string;
}

interface EnterpriseCollectorAsset {
    name: string;
    download_url: string;
    checksum_download_url: string;
    os: string;
    arch: string;
}

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

export type GraphNode = {
    label: string;
    kind: string;
    objectId: string;
    lastSeen: string;
    isTierZero: boolean;
    isOwnedObject: boolean;
    descendent_count?: number | null;
};

export type GraphNodes = Record<string, GraphNode>;

export type GraphEdge = {
    source: string;
    target: string;
    label: string;
    kind: string;
    lastSeen: string;
    impactPercent?: number;
    exploreGraphId?: string;
    data?: Record<string, any>;
};

export type GraphEdges = GraphEdge[];

export type GraphData = { nodes: GraphNodes; edges: GraphEdges };

export type GraphResponse = {
    data: GraphData;
};

export type StyledGraphNode = {
    color: string;
    data: Record<string, any>;
    border: {
        color: string;
    };
    fontIcon: {
        text: string;
    };
    label: {
        backgroundColor: string;
        center: boolean;
        fontSize: number;
        text: string;
    };
    size: number;
};

export type StyledGraphEdge = {
    color: string;
    data: Record<string, any>;
    end1?: {
        arrow: boolean;
    };
    end2?: {
        arrow: boolean;
    };
    id1: string;
    id2: string;
    label: {
        text: string;
    };
};

export type FlatGraphResponse = Record<string, StyledGraphNode | StyledGraphEdge>;

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
    SAMLProviderId?: string; // deprecated: this is left to maintain backwards compatability, please use SSOProviderId instead
    SSOProviderId?: number;
    is_disabled?: boolean;
}

export interface CreateUserRequest extends Omit<UpdateUserRequest, 'is_disabled'> {
    password?: string;
    needsPasswordReset?: boolean;
}

export type UpdateConfigurationRequest = ConfigurationPayload;
