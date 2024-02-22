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

export type RequestOptions = axios.AxiosRequestConfig;

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
    secret: string;
    needs_password_reset: boolean;
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
    Accepted?: string;
};

export type GraphNode = {
    label: string;
    kind: string;
    objectId: string;
    lastSeen: string;
    isTierZero: boolean;
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
