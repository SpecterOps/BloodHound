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

export interface Serial {
    id: number;
    created_at: string;
    updated_at: string;
}

export interface AssetGroupMemberParams {
    environment_kind?: string;
    environment_id?: string;
    primary_kind?: string;
    custom_member?: string;
    skip?: number;
    limit?: number;
}

type System = 'SYSTEM';

type ISO_DATE_STRING = string;

interface Created {
    created_at: ISO_DATE_STRING;
    created_by: string | System;
}

interface Updated {
    updated_at: ISO_DATE_STRING;
    updated_by: string | System;
}

interface Deleted {
    deleted_at: ISO_DATE_STRING | null;
    deleted_by: string | null;
}

interface Disabled {
    disabled_at: ISO_DATE_STRING | null;
    disabled_by: string | null;
}

export const AssetGroupTagTypeTier = 1 as const;
export const AssetGroupTagTypeLabel = 2 as const;
export const AssetGroupTagTypeOwned = 3 as const;

export type AssetGroupTagTypes =
    | typeof AssetGroupTagTypeTier
    | typeof AssetGroupTagTypeLabel
    | typeof AssetGroupTagTypeOwned;

export const AssetGroupTagTypesMap = {
    1: 'tier',
    2: 'label',
    3: 'owned',
} as const;

export interface AssetGroupTagCounts {
    selectors: number;
    members: number;
}

export interface AssetGroupTag extends Created, Updated, Deleted {
    id: number;
    name: string;
    kind_id: number;
    type: AssetGroupTagTypes;
    position: number | null;
    requireCertify: boolean | null;
    description: string;
    counts?: AssetGroupTagCounts;
    analysis_enabled?: boolean;
}

export const SeedTypeObjectId = 1 as const;
export const SeedTypeCypher = 2 as const;

export type SeedTypes = typeof SeedTypeObjectId | typeof SeedTypeCypher;

export const SeedTypesMap = {
    [SeedTypeObjectId]: 'Object ID',
    [SeedTypeCypher]: 'Cypher',
} as const;

export interface AssetGroupTagSelectorCounts {
    members: number;
}
export interface AssetGroupTagSelector extends Created, Updated, Disabled {
    id: number;
    asset_group_tag_id: number | null;
    name: string;
    description: string;
    is_default: boolean;
    allow_disable: boolean;
    auto_certify: boolean;
    seeds: AssetGroupTagSelectorSeed[];
    counts?: AssetGroupTagSelectorCounts;
}

export interface AssetGroupTagSelectorSeed {
    selector_id: number;
    type: SeedTypes;
    value: string;
}

export const AssetGroupTagCertifiedMap = {
    '-1': 'Manually not certified (revoked)',
    0: 'No certification (only automatically tagged if certify is enabled)',
    1: 'Manually certified',
    2: 'Auto certified (automatically tagged)',
} as const;

export type AssetGroupTagCertifiedValues = keyof typeof AssetGroupTagCertifiedMap;

export const NodeSourceSeed = 1 as const;
export const NodeSourceChild = 2 as const;
export const NodeSourceParent = 3 as const;

export type NodeSourceTypes = typeof NodeSourceSeed | typeof NodeSourceChild | typeof NodeSourceParent;
export interface AssetGroupTagNode {
    id: number; // uint64 graphID
    primary_kind: string;
    object_id: string;
    name: string;
}

export interface CreateSAMLProviderFormInputs extends SSOProviderConfiguration {
    name: string;
    metadata: FileList;
}
export type UpdateSAMLProviderFormInputs = Partial<CreateSAMLProviderFormInputs>;
export type UpsertSAMLProviderFormInputs = CreateSAMLProviderFormInputs | UpdateSAMLProviderFormInputs;

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

export interface LoginResponse {
    data: {
        user_id: string;
        auth_expired: boolean;
        eula_accepted: boolean;
        session_token: string;
        user_name: string;
    };
}

export type CommunityCollectorType = 'sharphound' | 'azurehound';
export type EnterpriseCollectorType = 'sharphound_enterprise' | 'azurehound_enterprise';
export type CollectorType = CommunityCollectorType | EnterpriseCollectorType;

export interface CollectorManifest {
    version: string;
    version_meta: VersionMeta;
    release_date: string;
    release_assets: CollectorAsset[];
}

interface VersionMeta {
    major: number;
    minor: number;
    patch: number;
    prerelease: string;
}

interface CollectorAsset {
    name: string;
    download_url: string;
    checksum_download_url: string;
    os: string;
    arch: string;
}

export type GraphNode = {
    label: string;
    kind: string;
    objectId: string;
    objectid?: string;
    lastSeen: string;
    isTierZero: boolean;
    isOwnedObject: boolean;
    nodetype?: string;
    displayname?: string;
    enabled?: boolean;
    pwdlastset?: number;
    lastlogontimestamp?: number;
    descendent_count?: number | null;
    properties?: Record<string, any>;
} & Record<string, any>;

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
} & Record<string, any>;

export type GraphEdges = GraphEdge[];

export type GraphData = { nodes: GraphNodes; edges: GraphEdges; node_keys?: string[] };

export type StyledGraphNode = {
    color: string;
    data: GraphNode;
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
    data: GraphEdge;
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

export type CustomNodeKindType = {
    id: number;
    kindName: string;
    config: {
        icon: {
            type: string;
            name: string;
            color: string;
        };
    };
};
