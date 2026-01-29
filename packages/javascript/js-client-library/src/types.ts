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

import { EnvironmentRequest } from './requests';

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

export const BloodHoundString = 'BloodHound' as const;

type BloodHound = typeof BloodHoundString;

type ISO_DATE_STRING = string;

export type TimestampFields = {
    created_at: string;
    updated_at: string;
    deleted_at: {
        Time: string;
        Valid: boolean;
    };
};

interface Created {
    created_at: ISO_DATE_STRING;
    created_by: string | BloodHound;
}

interface Updated {
    updated_at: ISO_DATE_STRING;
    updated_by: string | BloodHound;
}

interface Deleted {
    deleted_at: ISO_DATE_STRING | null;
    deleted_by: string | null;
}

interface Disabled {
    disabled_at: ISO_DATE_STRING | null;
    disabled_by: string | null;
}

export interface AssetGroupTagHistoryRecord {
    id: number;
    created_at: string;
    actor: string;
    email: string | null;
    action: string;
    target: string;
    asset_group_tag_id: number;
    environment_id: string | null;
    note: string | null;
}

export interface AssetGroupTagCertificationRecord {
    id: number;
    object_id: string;
    environment_id: string;
    primary_kind: string;
    name: string;
    created_at: string;
    asset_group_tag_id: number;
    certified_by: string;
    certified: number;
}

export const CertificationPending = 0 as const;
export const CertificationRevoked = 1 as const;
export const CertificationManual = 2 as const;
export const CertificationAuto = 3 as const;

export type CertificationType =
    | typeof CertificationPending
    | typeof CertificationRevoked
    | typeof CertificationManual
    | typeof CertificationAuto;

export const CertificationTypeMap: Record<CertificationType, string> = {
    [CertificationPending]: 'Pending',
    [CertificationRevoked]: 'Rejected',
    [CertificationManual]: 'User Certified',
    [CertificationAuto]: 'Automatic',
};

export type AssetGroupTagCertificationParams = {
    certified?: CertificationType;
    certified_by?: string;
    primary_kind?: string;
    created_at?: string;
};

export const HighestPrivilegePosition = 1 as const;

export const AssetGroupTagTypeZone = 1 as const;
export const AssetGroupTagTypeLabel = 2 as const;
export const AssetGroupTagTypeOwned = 3 as const;

export type AssetGroupTagType =
    | typeof AssetGroupTagTypeZone
    | typeof AssetGroupTagTypeLabel
    | typeof AssetGroupTagTypeOwned;

export const AssetGroupTagTypeMap = {
    [AssetGroupTagTypeZone]: 'zone',
    [AssetGroupTagTypeLabel]: 'label',
    [AssetGroupTagTypeOwned]: 'owned',
} as const;

export const RuleKey = 'selector' as const;
export const RulesKey = 'selectors' as const;
export const CustomRulesKey = 'custom_selectors' as const;
export const DefaultRulesKey = 'default_selectors' as const;
export const DisabledRulesKey = 'disabled_selectors' as const;

export const ObjectKey = 'member' as const;
export const ObjectsKey = 'members' as const;

export interface AssetGroupTagCounts {
    [RulesKey]: number;
    [CustomRulesKey]: number;
    [DefaultRulesKey]: number;
    [DisabledRulesKey]: number;
    [ObjectsKey]: number;
}

export interface AssetGroupTag extends Created, Updated, Deleted {
    id: number;
    name: string;
    kind_id: number;
    type: AssetGroupTagType;
    position: number | null;
    require_certify: boolean | null;
    description: string;
    analysis_enabled: boolean | null;
    glyph: string | null;
    counts?: AssetGroupTagCounts;
}

export const SeedTypeObjectId = 1 as const;
export const SeedTypeCypher = 2 as const;

export const SeedExpansionMethodNone = 0 as const;
export const SeedExpansionMethodAll = 1 as const;
export const SeedExpansionMethodChild = 2 as const;
export const SeedExpansionMethodParent = 3 as const;
export type SeedExpansionMethod =
    | typeof SeedExpansionMethodNone
    | typeof SeedExpansionMethodAll
    | typeof SeedExpansionMethodChild
    | typeof SeedExpansionMethodParent;

export type SeedTypes = typeof SeedTypeObjectId | typeof SeedTypeCypher;

export const SeedTypesMap = {
    [SeedTypeObjectId]: 'Object ID',
    [SeedTypeCypher]: 'Cypher',
} as const;

export const AssetGroupTagSelectorAutoCertifyDisabled = 0 as const;
export const AssetGroupTagSelectorAutoCertifySeedsOnly = 2 as const;
export const AssetGroupTagSelectorAutoCertifyAllMembers = 1 as const;

export type AssetGroupTagSelectorAutoCertifyType =
    | typeof AssetGroupTagSelectorAutoCertifyDisabled
    | typeof AssetGroupTagSelectorAutoCertifySeedsOnly
    | typeof AssetGroupTagSelectorAutoCertifyAllMembers;

export const AssetGroupTagSelectorAutoCertifyMap = {
    [AssetGroupTagSelectorAutoCertifyDisabled]: 'Off',
    [AssetGroupTagSelectorAutoCertifySeedsOnly]: 'Direct Objects',
    [AssetGroupTagSelectorAutoCertifyAllMembers]: 'All Objects',
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
    auto_certify: AssetGroupTagSelectorAutoCertifyType;
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
export interface AssetGroupTagMember {
    asset_group_tag_id: number;
    id: number; // uint64 graphID
    primary_kind: string;
    object_id: string;
    name: string;
    source: number;
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
    all_environments?: boolean;
    environment_targeted_access_control?: EnvironmentRequest[] | null;
}

export interface UserMinimal {
    id: string;
    first_name: string | null;
    last_name: string | null;
    email_address: string | null;
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

export interface ListUsersMinimalResponse {
    data: {
        users: UserMinimal[];
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

export interface GraphNodeProperties {
    nodetype?: string;
    displayname?: string;
    enabled?: boolean;
    pwdlastset?: number;
    lastlogontimestamp?: number;
    descendent_count?: number | null;
    [key: string]: any;
}

export type GraphNode = {
    label: string;
    kind: string;
    kinds: string[];
    objectId: string;
    lastSeen: string;
    isTierZero: boolean;
    isOwnedObject: boolean;
    properties?: GraphNodeProperties;
};

export type GraphNodeSpreadWithProperties = Partial<Omit<GraphNode, 'properties'> & GraphNodeProperties>;

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

export type GraphData = { nodes: GraphNodes; edges: GraphEdges; node_keys?: string[] };

export type StyledGraphNode = {
    color: string;
    data: GraphNodeSpreadWithProperties;
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

export type OuDetails = {
    objectid: string;
    name?: string;
    exists?: boolean;
    distinguishedname?: string;
    type?: string;
};

export type DomainDetails = {
    objectid: string;
    name: string;
    exists: boolean;
    type: string;
};

export type DomainResult = {
    job_id: number;
    domain_name: string;
    success: boolean;
    message: string;
    user_count: number;
    group_count: number;
    computer_count: number;
    gpo_count: number;
    ou_count: number;
    container_count: number;
    aiaca_count: number;
    rootca_count: number;
    enterpriseca_count: number;
    ntauthstore_count: number;
    certtemplate_count: number;
    deleted_count: number;
    id: number;
    created_at: string;
    updated_at: string;
    deleted_at: {
        Time: string;
        Valid: boolean;
    };
};

export type ScheduledJobDisplay = {
    id: number;
    client_id: string;
    client_name: string;
    event_id: number;
    execution_time: string;
    start_time: string;
    end_time: string;
    status: number;
    status_message: string;
    session_collection: boolean;
    local_group_collection: boolean;
    ad_structure_collection: boolean;
    cert_services_collection: boolean;
    ca_registry_collection: boolean;
    dc_registry_collection: boolean;
    all_trusted_domains: boolean;
    domain_controller: string;
    ous: OuDetails[];
    domains: DomainDetails[];
    domain_results: DomainResult[];
};

export type Client = {
    configured_user: string;
    events: {
        id: number;
        client_id: string;
        session_collection: boolean;
        local_group_collection: boolean;
        ad_structure_collection: boolean;
        cert_services_collection: boolean;
        ca_registry_collection: boolean;
        dc_registry_collection: boolean;
        all_trusted_domains: boolean;
        ous: OuDetails[];
        domains: DomainDetails[];
        rrule: string;
    }[];
    hostname: string;
    id: string;
    ip_address: string;
    last_checkin: string;
    name: string;
    token: unknown;
    current_job_id: number | null;
    current_task_id: number | null;
    current_job: ScheduledJobDisplay;
    current_task: ScheduledJobDisplay;
    completed_job_count: number;
    completed_task_count: number;
    domain_controller: unknown;
    version: string;
    user_sid: string;
    type: string;
    issuer_address: string;
    issuer_address_override: string;
};

export type FileIngestJob = TimestampFields & {
    end_time: string;
    failed_files: number;
    id: number;
    last_ingest: string;
    start_time: string;
    status_message: string;
    status: number;
    total_files: number;
    user_email_address: string;
    user_id: string;
};

export type FileIngestCompletedTask = TimestampFields & {
    errors: string[];
    warnings: string[];
    file_name: string;
    id: number;
    parent_file_name: string;
};

export const WindowsAuth = 'windows' as const;
export const BloodHoundAuth = 'bloodhound' as const;

export type AuthenticationMethod = typeof BloodHoundAuth | typeof WindowsAuth;

export type FindingAssetsResponse = {
    long_description: string;
    long_remediation: string;
    references: string;
    short_description: string;
    short_remediation: string;
    title: string;
    type: string;
};
