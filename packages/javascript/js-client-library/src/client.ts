// Copyright 2026 Specter Ops, Inc.
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

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import {
    ClearDatabaseRequest,
    CreateAssetGroupRequest,
    CreateAssetGroupTagRequest,
    CreateAzureHoundClientRequest,
    CreateAzureHoundEventRequest,
    CreateOIDCProviderRequest,
    CreateScheduledJobRequest,
    CreateSelectorRequest,
    CreateSharpHoundClientRequest,
    CreateSharpHoundEventRequest,
    CreateUserQueryRequest,
    CreateUserRequest,
    DeleteUserQueryPermissionsRequest,
    LoginRequest,
    PostureRequest,
    PreviewSelectorsRequest,
    PutUserAuthSecretRequest,
    QueryScope,
    RequestOptions,
    UpdateAssetGroupRequest,
    UpdateAssetGroupSelectorRequest,
    UpdateAssetGroupTagRequest,
    UpdateAzureHoundClientRequest,
    UpdateAzureHoundEventRequest,
    UpdateCertificationRequest,
    UpdateConfigurationRequest,
    UpdateOIDCProviderRequest,
    UpdateSelectorRequest,
    UpdateSharpHoundClientRequest,
    UpdateSharpHoundEventRequest,
    UpdateUserQueryPermissionsRequest,
    UpdateUserQueryRequest,
    UpdateUserRequest,
} from './requests';
import {
    ActiveDirectoryDataQualityResponse,
    AssetGroupMemberCountsResponse,
    AssetGroupMembersResponse,
    AssetGroupResponse,
    AssetGroupTagMemberInfoResponse,
    AssetGroupTagMembersResponse,
    AssetGroupTagResponse,
    AssetGroupTagSearchResponse,
    AssetGroupTagSelectorResponse,
    AssetGroupTagSelectorsResponse,
    AssetGroupTagsCertification,
    AssetGroupTagsHistory,
    AssetGroupTagsResponse,
    AzureDataQualityResponse,
    BasicResponse,
    CreateAuthTokenResponse,
    DatapipeStatusResponse,
    EndFileIngestResponse,
    Environment,
    FileIngestCompletedTasksResponse,
    GetClientResponse,
    GetCollectorsResponse,
    GetCommunityCollectorsResponse,
    GetConfigurationResponse,
    GetCustomNodeKindsResponse,
    GetEdgeTypesResponse,
    GetEnterpriseCollectorsResponse,
    GetExportQueryResponse,
    GetScheduledJobDisplayResponse,
    GetSelfResponse,
    GraphResponse,
    ListAuthTokensResponse,
    ListFileIngestJobsResponse,
    ListFileTypesForIngestResponse,
    PaginatedResponse,
    PostureFindingTrendsResponse,
    PostureHistoryResponse,
    PostureResponse,
    PreviewSelectorsResponse,
    SavedQuery,
    SavedQueryPermissionsResponse,
    StartFileIngestResponse,
    UpdateConfigurationResponse,
    UploadFileToIngestResponse,
} from './responses';
import * as types from './types';
import { FindingAssetsResponse } from './types';

/** Return the value as a string with the given prefix */
const prefixValue = (prefix: string, value: any) => (value ? `${prefix}:${value.toString()}` : undefined);

/** Return a copy of the object with all keys having undefined values have been stripped out  */
const omitUndefined = (obj: Record<string, unknown>) =>
    Object.fromEntries(Object.entries(obj).filter(([, value]) => value !== undefined));

class BHEAPIClient {
    baseClient: AxiosInstance;

    constructor(config: RequestOptions) {
        this.baseClient = axios.create(config);
    }

    /* health */
    health = (options?: RequestOptions) => this.baseClient.get('/health', options);

    /* version */
    version = (options?: RequestOptions) => this.baseClient.get('/api/version', options);

    /* datapipe status */
    getDatapipeStatus = (options?: RequestOptions) =>
        this.baseClient.get<DatapipeStatusResponse>('/api/v2/datapipe/status', options);

    /* search */
    searchHandler = (keyword: string, type?: string, options?: RequestOptions) => {
        return this.baseClient.get(
            '/api/v2/search',
            Object.assign(
                {
                    params: {
                        q: keyword,
                        type: type,
                    },
                },
                options
            )
        );
    };

    cypherSearch = (query: string, options?: RequestOptions, includeProperties?: boolean) => {
        return this.baseClient.post<GraphResponse>(
            '/api/v2/graphs/cypher',
            { query, include_properties: includeProperties || false },
            options
        );
    };

    getUserSavedQueries = (scope: QueryScope, options?: RequestOptions) => {
        return this.baseClient.get<PaginatedResponse<SavedQuery[]>>(
            '/api/v2/saved-queries',
            Object.assign(
                {
                    params: {
                        sort_by: 'name',
                        scope: scope,
                    },
                },
                options
            )
        );
    };

    createUserQuery = (payload: CreateUserQueryRequest, options?: RequestOptions) => {
        return this.baseClient.post<BasicResponse<SavedQuery>>('/api/v2/saved-queries', payload, options);
    };

    updateUserQuery = (payload: UpdateUserQueryRequest) => {
        const headers = {
            'Content-Type': 'application/json',
        };
        return this.baseClient.put<BasicResponse<SavedQuery>>(`/api/v2/saved-queries/${payload.id}`, payload, {
            headers,
        });
    };

    deleteUserQuery = (queryId: number, options?: RequestOptions) => {
        return this.baseClient.delete(`/api/v2/saved-queries/${queryId}`, options);
    };

    getExportCypherQueries = (): Promise<any> =>
        this.baseClient.get(
            `/api/v2/saved-queries/export?scope=all`,
            Object.assign({
                responseType: 'blob',
            })
        );

    getExportCypherQuery = (id: number, options?: RequestOptions): Promise<GetExportQueryResponse> =>
        this.baseClient.get(
            `/api/v2/saved-queries/${id}/export`,
            Object.assign(
                {
                    responseType: 'blob',
                },
                options
            )
        );

    importUserQuery = (payload: FormData | Blob | object, options?: RequestOptions) => {
        const cfg: AxiosRequestConfig = { ...(options ?? {}) };
        if (payload instanceof FormData) {
            // Let the browser set multipart/form-data with boundary
        } else if (payload instanceof Blob) {
            cfg.headers = { ...(options?.headers ?? {}), 'Content-Type': payload.type || 'application/octet-stream' };
        } else {
            cfg.headers = { ...(options?.headers ?? {}), 'Content-Type': 'application/json' };
        }
        return this.baseClient.post<BasicResponse<any>>('/api/v2/saved-queries/import', payload as any, cfg);
    };

    getUserQueryPermissions = (queryId: number, options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<SavedQueryPermissionsResponse>>(
            `/api/v2/saved-queries/${queryId}/permissions`,
            options
        );

    updateUserQueryPermissions = (
        queryId: number,
        queryPermissionsPayload: UpdateUserQueryPermissionsRequest,
        options?: RequestOptions
    ) => this.baseClient.put(`/api/v2/saved-queries/${queryId}/permissions`, queryPermissionsPayload, options);

    deleteUserQueryPermissions = (
        queryId: number,
        queryPermissionsPayload: DeleteUserQueryPermissionsRequest,
        options?: RequestOptions
    ) =>
        this.baseClient.delete(
            `/api/v2/saved-queries/${queryId}/permissions`,
            Object.assign(
                {
                    data: queryPermissionsPayload,
                },
                options
            )
        );

    getKinds = (options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<{ kinds: string[] }>>('/api/v2/graphs/kinds', options);

    getSourceKinds = (options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<{ kinds: { id: number; name: string }[] }>>(
            '/api/v2/graphs/source-kinds',
            options
        );

    clearDatabase = (payload: ClearDatabaseRequest, options?: RequestOptions) => {
        return this.baseClient.post('/api/v2/clear-database', payload, options);
    };

    getAvailableEnvironments = (options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<Environment[]>>('/api/v2/available-domains', options);

    /* audit */
    getAuditLogs = (options?: RequestOptions) => this.baseClient.get('/api/v2/audit', options);

    /* asset group tags (AGT) */

    getAssetGroupTagHistory = (options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagsHistory>(`/api/v2/asset-group-tags-history`, {
            paramsSerializer: {
                indexes: null,
            },
            ...options,
        });

    searchAssetGroupTagHistory = (query: string, options?: RequestOptions) =>
        this.baseClient.post<AssetGroupTagsHistory>(
            `/api/v2/asset-group-tags-history`,
            { query },
            {
                paramsSerializer: {
                    indexes: null,
                },
                ...options,
            }
        );

    getAssetGroupTagsCertifications = (options?: RequestOptions) => {
        return this.baseClient.get<AssetGroupTagsCertification>(`/api/v2/asset-group-tags/certifications`, {
            paramsSerializer: {
                indexes: null,
            },
            ...options,
        });
    };

    getAssetGroupTags = (options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagsResponse>(`/api/v2/asset-group-tags`, options);

    searchAssetGroupTags = (body: { query: string; tag_type: number }, options?: RequestOptions) =>
        this.baseClient.post<AssetGroupTagSearchResponse>(`/api/v2/asset-group-tags/search`, body, options);

    getAssetGroupTag = (tagId: number | string, options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagResponse>(`/api/v2/asset-group-tags/${tagId}`, options);

    createAssetGroupTag = (values: CreateAssetGroupTagRequest, options?: RequestOptions) =>
        this.baseClient.post<BasicResponse<types.AssetGroupTag>>(`/api/v2/asset-group-tags`, values, options);

    updateAssetGroupTag = (
        tagId: number | string,
        updatedValues: UpdateAssetGroupTagRequest,
        options?: RequestOptions
    ) =>
        this.baseClient.patch<BasicResponse<types.AssetGroupTag>>(
            `/api/v2/asset-group-tags/${tagId}`,
            updatedValues,
            options
        );

    deleteAssetGroupTag = (tagId: string | number, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-group-tags/${tagId}`, options);

    getAssetGroupTagMemberInfo = (tagId: number | string, memberId: number | string, options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagMemberInfoResponse>(
            `/api/v2/asset-group-tags/${tagId}/members/${memberId}`,
            options
        );

    getAssetGroupTagSelectors = (tagId: number | string, options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagSelectorsResponse>(`/api/v2/asset-group-tags/${tagId}/selectors`, {
            ...options,
            paramsSerializer: { indexes: null },
        });

    getAssetGroupTagSelector = (tagId: number | string, ruleId: number | string, options?: RequestOptions) =>
        this.baseClient.get<AssetGroupTagSelectorResponse>(
            `/api/v2/asset-group-tags/${tagId}/selectors/${ruleId}`,
            options
        );

    createAssetGroupTagSelector = (tagId: number | string, values: CreateSelectorRequest, options?: RequestOptions) =>
        this.baseClient.post(`/api/v2/asset-group-tags/${tagId}/selectors`, values, options);

    updateAssetGroupTagSelector = (
        tagId: number | string,
        ruleId: number | string,
        updatedValues: UpdateSelectorRequest,
        options?: RequestOptions
    ) => this.baseClient.patch(`/api/v2/asset-group-tags/${tagId}/selectors/${ruleId}`, updatedValues, options);

    deleteAssetGroupTagSelector = (tagId: string | number, ruleId: string | number, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-group-tags/${tagId}/selectors/${ruleId}`, options);

    getAssetGroupTagMembers = (
        assetGroupTagId: number | string,
        skip: number | string,
        limit: number,
        sort_by: string,
        environments?: string[],
        primary_kind?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get<AssetGroupTagMembersResponse>(`/api/v2/asset-group-tags/${assetGroupTagId}/members`, {
            ...options,
            params: {
                ...options?.params,
                environments,
                primary_kind: primary_kind ? `eq:${primary_kind}` : undefined,
                skip,
                limit,
                sort_by,
            },
            paramsSerializer: { indexes: null },
        });

    getAssetGroupTagSelectorMembers = (
        tagId: number | string,
        ruleId: number | string,
        skip: number,
        limit: number,
        sort_by: string,
        environments?: string[],
        primary_kind?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get<AssetGroupTagMembersResponse>(
            `/api/v2/asset-group-tags/${tagId}/selectors/${ruleId}/members`,
            {
                ...options,
                params: {
                    ...options?.params,
                    environments,
                    primary_kind: primary_kind ? `eq:${primary_kind}` : undefined,
                    skip,
                    limit,
                    sort_by,
                },
                paramsSerializer: { indexes: null },
            }
        );

    getAssetGroupTagMembersCount = (tagId: string, environments?: string[], options?: RequestOptions) =>
        this.baseClient.get<AssetGroupMemberCountsResponse>(`/api/v2/asset-group-tags/${tagId}/members/counts`, {
            ...options,
            params: { ...options?.params, environments },
            paramsSerializer: { indexes: null },
        });

    getAssetGroupTagRuleMembersCount = (
        tagId: string,
        ruleId: string,
        environments?: string[],
        options?: RequestOptions
    ) =>
        this.baseClient.get<AssetGroupMemberCountsResponse>(
            `/api/v2/asset-group-tags/${tagId}/selectors/${ruleId}/members/counts`,
            {
                ...options,
                params: { ...options?.params, environments },
                paramsSerializer: { indexes: null },
            }
        );

    assetGroupTagsPreviewSelectors = (payload: PreviewSelectorsRequest, options: RequestOptions) => {
        return this.baseClient.post<PreviewSelectorsResponse>(
            '/api/v2/asset-group-tags/preview-selectors',
            payload,
            options
        );
    };

    updateAssetGroupTagCertification = (requestBody: UpdateCertificationRequest) => {
        return this.baseClient.post('/api/v2/asset-group-tags/certifications', requestBody);
    };

    /* asset group isolation (AGI) */

    createAssetGroup = (assetGroup: CreateAssetGroupRequest, options?: RequestOptions) =>
        this.baseClient.post('/api/v2/asset-groups', assetGroup, options);

    getAssetGroup = (assetGroupId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/asset-groups/${assetGroupId}`, options);

    deleteAssetGroup = (assetGroupId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-groups/${assetGroupId}`, options);

    updateAssetGroup = (assetGroupId: string, assetGroup: UpdateAssetGroupRequest, options?: RequestOptions) =>
        this.baseClient.put(`/api/v2/asset-groups/${assetGroupId}`, assetGroup, options);

    updateAssetGroupSelector = (
        assetGroupId: number,
        selectorChangeset: UpdateAssetGroupSelectorRequest[],
        options?: RequestOptions
    ) => this.baseClient.put(`/api/v2/asset-groups/${assetGroupId}/selectors`, selectorChangeset, options);

    deleteAssetGroupSelector = (assetGroupId: string, selectorId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-groups/${assetGroupId}/selectors/${selectorId}`, options);

    listAssetGroupCollections = (assetGroupId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/asset-groups/${assetGroupId}/collections`, options);

    listAssetGroupMembers = (assetGroupId: number, params?: types.AssetGroupMemberParams, options?: RequestOptions) =>
        this.baseClient.get<AssetGroupMembersResponse>(
            `/api/v2/asset-groups/${assetGroupId}/members`,
            Object.assign({ params }, options)
        );

    getAssetGroupMembersCount = (
        assetGroupId: string,
        params?: Pick<types.AssetGroupMemberParams, 'environment_id' | 'environment_kind'>,
        options?: RequestOptions
    ) =>
        this.baseClient.get<AssetGroupMemberCountsResponse>(
            `/api/v2/asset-groups/${assetGroupId}/members/counts`,
            Object.assign({ params }, options)
        );

    listAssetGroups = (options?: RequestOptions) =>
        this.baseClient.get<AssetGroupResponse>('/api/v2/asset-groups', options);

    /* attack paths */

    /**
     * getAttackPathsGraph returns the graph data that is rendered in the attack paths page for a given environment
     */
    getAttackPathsGraph = (environmentId: string, nodeId: string | null = null, options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<types.FlatGraphResponse>>(
            `/api/v2/meta-trees/${environmentId}`,
            Object.assign({ params: nodeId && { node_id: nodeId } }, options)
        );

    /**
     * getLatestMetaNode returns Meta node data from the most recent analysis for a given environment
     */
    getLatestMetaNode = (environmentId: string, options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<types.FlatGraphResponse>>(`/api/v2/meta-nodes/${environmentId}`, options);

    getFindings = (key: string, options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<FindingAssetsResponse>>(`/api/v2/findings/${key}`, options);

    /**
     * getFindingDetails returns data associated with a finding for a given environment
     */
    getFindingDetails = <T = any>(
        environmentId: string,
        finding: string,
        skip: number,
        limit: number,
        filterAccepted?: boolean,
        sortBy?: string | string[],
        assetGroupTagId?: number,
        options?: RequestOptions
    ) => {
        const params = new URLSearchParams();
        params.append('finding', finding);

        params.append('skip', skip.toString());
        params.append('limit', limit.toString());
        if (sortBy) {
            if (typeof sortBy === 'string') {
                params.append('sort_by', sortBy);
            } else {
                sortBy.forEach((sort) => params.append('sort_by', sort));
            }
        }

        if (typeof assetGroupTagId === 'number') params.append('asset_group_tag_id', assetGroupTagId.toString());

        if (typeof filterAccepted === 'boolean') params.append('Accepted', `eq:${filterAccepted}`);

        return this.baseClient.get<T>(
            `/api/v2/domains/${environmentId}/details`,
            Object.assign(
                {
                    params: params,
                    headers: options?.headers,
                },
                options
            )
        );
    };

    /**
     * getFindingSparklineData returns the data that is used to plot finding exposure over time and count over time chart on the attack paths page for a given environment
     */
    getFindingSparklineData = <T = any>(environmentId: string, options?: RequestOptions) =>
        this.baseClient.get<T>(`/api/v2/domains/${environmentId}/sparkline`, options);

    changeFindingAcceptance = (
        attackPathId: string,
        findingType: string,
        accepted: boolean,
        acceptUntil?: Date,
        options?: RequestOptions
    ) =>
        this.baseClient.put(
            `/api/v2/attack-paths/${attackPathId}/acceptance`,
            {
                risk_type: findingType,
                accepted: accepted,
                accept_until: acceptUntil && acceptUntil.toISOString(),
            },
            options
        );

    requestAttackPaths = (options?: RequestOptions) => this.baseClient.put('/api/v2/attack-paths', options);

    requestAnalysis = (options?: RequestOptions) => this.baseClient.put('/api/v2/analysis', options);

    /**
     * getAvailableFindingTypes returns a list of findings that were discovered through the most recent analysis for a given environment
     */
    getAvailableFindingTypes = (environmentId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/domains/${environmentId}/available-types`, options);

    exportRiskFindings = (environmentId: string, findingType: string, accepted?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${environmentId}/attack-path-findings`,
            Object.assign(
                {
                    params: {
                        finding: findingType,
                        accepted: accepted,
                    },
                    responseType: 'blob',
                },
                options
            )
        );

    getAssetGroupComboNode = (assetGroupId: string, domainsid?: string, options?: RequestOptions) => {
        return this.baseClient.get<BasicResponse<types.FlatGraphResponse>>(
            `/api/v2/asset-groups/${assetGroupId}/combo-node`,
            Object.assign(
                {
                    params: { domainsid: domainsid },
                },
                options
            )
        );
    };

    /* qa */

    getADQualityStats = (
        domainId: string,
        start?: Date,
        end?: Date,
        limit?: number,
        sort_by?: string,
        options?: RequestOptions
    ) => {
        return this.baseClient.get<ActiveDirectoryDataQualityResponse>(
            `/api/v2/ad-domains/${domainId}/data-quality-stats`,
            Object.assign(
                {
                    params: {
                        start: start?.toISOString(),
                        end: end?.toISOString(),
                        limit: limit,
                        sort_by: sort_by,
                    },
                },
                options
            )
        );
    };

    getAzureQualityStats = (
        tenantId: string,
        start?: Date,
        end?: Date,
        limit?: number,
        sort_by?: string,
        options?: RequestOptions
    ) => {
        return this.baseClient.get<AzureDataQualityResponse>(
            `/api/v2/azure-tenants/${tenantId}/data-quality-stats`,
            Object.assign(
                {
                    params: {
                        start: start?.toISOString(),
                        end: end?.toISOString(),
                        limit: limit,
                        sort_by: sort_by,
                    },
                },
                options
            )
        );
    };

    getPlatformQualityStats = (
        platformtype: string,
        start?: Date,
        end?: Date,
        limit?: number,
        sort_by?: string,
        options?: RequestOptions
    ) => {
        return this.baseClient.get(
            `/api/v2/platform/${platformtype}/data-quality-stats`,
            Object.assign(
                {
                    params: {
                        start: start?.toISOString(),
                        end: end?.toISOString(),
                        limit: limit,
                        sort_by: sort_by,
                    },
                },
                options
            )
        );
    };

    /* posture */

    getPostureStats = (from: Date, to: Date, environmentId?: string, sortBy?: string, options?: RequestOptions) => {
        const params: PostureRequest = {
            from: from.toISOString(),
            to: to.toISOString(),
        };

        if (environmentId) params.domain_sid = `eq:${environmentId}`;
        if (sortBy) params.sort_by = sortBy;

        return this.baseClient.get<PostureResponse>(
            '/api/v2/posture-stats',
            Object.assign(
                {
                    params: params,
                },
                options
            )
        );
    };

    getPostureFindingTrends = (options?: RequestOptions) =>
        this.baseClient.get<PostureFindingTrendsResponse>(`/api/v2/attack-paths/finding-trends`, {
            ...options,
            paramsSerializer: { indexes: null },
        });

    getPostureHistory = (dataType: string, options?: RequestOptions) =>
        this.baseClient.get<PostureHistoryResponse>(`/api/v2/posture-history/${dataType}`, {
            ...options,
            paramsSerializer: { indexes: null },
        });

    /* explore search */

    getPathfindingResult = (startNode: string, endNode: string, options?: RequestOptions) =>
        this.baseClient.get(
            '/api/v2/pathfinding',
            Object.assign(
                {
                    params: {
                        start_node: startNode,
                        end_node: endNode,
                    },
                },
                options
            )
        );

    getSearchResult = (query: string, searchType: string, options?: RequestOptions) =>
        this.baseClient.get<BasicResponse<types.FlatGraphResponse>>(
            '/api/v2/graph-search',
            Object.assign(
                {
                    params: {
                        query: query,
                        type: searchType,
                    },
                },
                options
            )
        );

    /* ingest */

    ingestData = (options?: RequestOptions) => this.baseClient.post('/api/v2/ingest', options);

    /* clients */

    getClients = (
        skip: number = 0,
        limit: number = 10,
        hydrateDomains?: boolean,
        hydrateOUs?: boolean,
        options?: RequestOptions
    ) =>
        this.baseClient.get<GetClientResponse>(
            '/api/v2/clients',
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        hydrate_domains: hydrateDomains,
                        hydrate_ous: hydrateOUs,
                    },
                },
                options
            )
        );

    createClient = (client: CreateSharpHoundClientRequest | CreateAzureHoundClientRequest, options?: RequestOptions) =>
        this.baseClient.post('/api/v2/clients', client, options);

    getClient = (clientId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/clients/${clientId}`, options);

    updateClient = (
        clientId: string,
        client: UpdateSharpHoundClientRequest | UpdateAzureHoundClientRequest,
        options?: RequestOptions
    ) => this.baseClient.put(`/api/v2/clients/${clientId}`, client, options);

    regenerateClientToken = (clientId: string, options?: RequestOptions) =>
        this.baseClient.put(`/api/v2/clients/${clientId}/token`, options);

    deleteClient = (clientId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/clients/${clientId}`, options);

    getClientCompletedJobs = (clientId: string, skip: number, limit: number, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/clients/${clientId}/completed-jobs`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                    },
                },
                options
            )
        );

    createScheduledJob = (clientId: string, scheduledJob: CreateScheduledJobRequest, options?: RequestOptions) =>
        this.baseClient.post(`/api/v2/clients/${clientId}/jobs`, scheduledJob, options);

    getFinishedJobs = (
        {
            limit,
            skip,
            status,
            start_time,
            end_time,
            client_id,
            ad_structure_collection,
            ca_registry_collection,
            cert_services_collection,
            dc_registry_collection,
            hydrate_domains = false,
            hydrate_ous = false,
            local_group_collection,
            session_collection,
        }: {
            limit: number;
            skip: number;
            status?: number;
            start_time?: string;
            end_time?: string;
            client_id?: string;
            ad_structure_collection?: boolean;
            ca_registry_collection?: boolean;
            cert_services_collection?: boolean;
            dc_registry_collection?: boolean;
            hydrate_domains?: boolean;
            hydrate_ous?: boolean;
            local_group_collection?: boolean;
            session_collection?: boolean;
        },
        options?: RequestOptions
    ) =>
        this.baseClient.get<GetScheduledJobDisplayResponse>(
            '/api/v2/jobs/finished',
            Object.assign(
                {
                    params: omitUndefined({
                        skip,
                        limit,
                        hydrate_domains,
                        hydrate_ous,
                        status: prefixValue('eq', status),
                        start_time: prefixValue('gte', start_time),
                        end_time: prefixValue('lte', end_time),
                        client_id: prefixValue('eq', client_id),
                        ad_structure_collection: prefixValue('eq', ad_structure_collection),
                        ca_registry_collection: prefixValue('eq', ca_registry_collection),
                        cert_services_collection: prefixValue('eq', cert_services_collection),
                        dc_registry_collection: prefixValue('eq', dc_registry_collection),
                        local_group_collection: prefixValue('eq', local_group_collection),
                        session_collection: prefixValue('eq', session_collection),
                    }),
                },
                options
            )
        );

    /* events */
    getEvents = (hydrateDomains?: boolean, hydrateOUs?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            '/api/v2/events',
            Object.assign(
                {
                    params: {
                        hydrate_domains: hydrateDomains,
                        hydrate_ous: hydrateOUs,
                    },
                },
                options
            )
        );

    getEvent = (eventId: string, options?: RequestOptions) => this.baseClient.get(`/api/v2/events/${eventId}`, options);

    createEvent = (event: CreateSharpHoundEventRequest | CreateAzureHoundEventRequest, options?: RequestOptions) =>
        this.baseClient.post('/api/v2/events', event, {
            params: {
                hydrate_ous: false,
                hydrate_domains: false,
            },
            ...options,
        });

    updateEvent = (
        eventId: string,
        event: UpdateSharpHoundEventRequest | UpdateAzureHoundEventRequest,
        options?: RequestOptions
    ) => this.baseClient.put(`/api/v2/events/${eventId}`, event, options);

    deleteEvent = (eventId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/events/${eventId}`, options);

    /* file ingest */
    listFileIngestJobs = (
        {
            skip,
            limit,
            sortBy,
            status,
            user_id,
            start_time,
            end_time,
        }: {
            skip?: number;
            limit?: number;
            sortBy?: string;
            status?: number;
            user_id?: string;
            start_time?: string;
            end_time?: string;
        },
        options?: RequestOptions
    ) =>
        this.baseClient.get<ListFileIngestJobsResponse>(
            'api/v2/file-upload',
            Object.assign(
                {
                    params: omitUndefined({
                        skip,
                        limit,
                        sort_by: sortBy,
                        status: prefixValue('eq', status),
                        user_id: prefixValue('eq', user_id),
                        start_time: prefixValue('gte', start_time),
                        end_time: prefixValue('lte', end_time),
                    }),
                },
                options
            )
        );

    listFileTypesForIngest = () =>
        this.baseClient.get<ListFileTypesForIngestResponse>('/api/v2/file-upload/accepted-types');

    startFileIngest = () => {
        return this.baseClient.post<StartFileIngestResponse>('/api/v2/file-upload/start');
    };

    uploadFileToIngestJob = (
        ingestId: string,
        json: any,
        contentType: string,
        options: AxiosRequestConfig<any> = {}
    ) => {
        const mergedOptions: AxiosRequestConfig<any> = {
            ...options,
            headers: {
                ...(options?.headers ?? {}),
                'Content-Type': contentType,
                'X-File-Upload-Name': json.name,
            },
        };

        return this.baseClient.post<UploadFileToIngestResponse>(`/api/v2/file-upload/${ingestId}`, json, mergedOptions);
    };

    getFileUpload = (uploadId: string, options?: RequestOptions) =>
        this.baseClient.get<FileIngestCompletedTasksResponse>(
            `/api/v2/file-upload/${uploadId}/completed-tasks`,
            options
        );

    endFileIngest = (ingestId: string) =>
        this.baseClient.post<EndFileIngestResponse>(`/api/v2/file-upload/${ingestId}/end`);

    uploadSchemaFile = (json: any, options?: RequestOptions) =>
        this.baseClient.put('/api/v2/extensions', json, options);

    /* custom node kinds */
    getCustomNodeKinds = (options?: RequestOptions) =>
        this.baseClient.get<GetCustomNodeKindsResponse>('/api/v2/custom-nodes', options);

    /* jobs */
    getJobs = (hydrateDomains?: boolean, hydrateOUs?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            '/api/v2/jobs',
            Object.assign(
                {
                    params: {
                        hydrate_domains: hydrateDomains,
                        hydrate_ous: hydrateOUs,
                    },
                },
                options
            )
        );

    getJob = (jobId: string, options?: RequestOptions) => this.baseClient.get(`/api/v2/jobs/${jobId}`, options);

    cancelScheduledJob = (jobId: string, options?: RequestOptions) =>
        this.baseClient.put(`/api/v2/jobs/${jobId}/cancel`, undefined, options);

    getJobLogFile = (jobId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/jobs/${jobId}/log`, options);

    /* auth */
    login = (credentials: LoginRequest, options?: RequestOptions) =>
        this.baseClient.post<types.LoginResponse>('/api/v2/login', credentials, options);

    getSelf = (options?: RequestOptions) => this.baseClient.get<GetSelfResponse>('/api/v2/self', options);

    logout = (options?: RequestOptions) => this.baseClient.post('/api/v2/logout', options);

    createSAMLProviderFromFile = (
        data: { name: string; metadata: File } & types.SSOProviderConfiguration,
        options?: RequestOptions
    ) => {
        // form data is limited to strings or blobs so we have to deconstruct the config payload
        const formData = new FormData();
        formData.append('name', data.name);
        formData.append('metadata', data.metadata);
        formData.append('config.auto_provision.enabled', data.config.auto_provision.enabled.toString());
        formData.append('config.auto_provision.default_role_id', data.config.auto_provision.default_role_id.toString());
        formData.append('config.auto_provision.role_provision', data.config.auto_provision.role_provision.toString());
        return this.baseClient.post(`/api/v2/sso-providers/saml`, formData, options);
    };

    updateSAMLProviderFromFile = (
        ssoProviderId: types.SSOProvider['id'],
        data: { name?: string; metadata?: File; config?: types.SSOProviderConfiguration['config'] },
        options?: RequestOptions
    ) => {
        const formData = new FormData();
        if (data.name) {
            formData.append('name', data.name);
        }
        if (data.metadata) {
            formData.append('metadata', data.metadata);
        }
        if (data.config) {
            formData.append('config.auto_provision.enabled', data.config.auto_provision.enabled.toString());
            formData.append(
                'config.auto_provision.default_role_id',
                data.config.auto_provision.default_role_id.toString()
            );
            formData.append(
                'config.auto_provision.role_provision',
                data.config.auto_provision.role_provision.toString()
            );
        }
        return this.baseClient.patch(`/api/v2/sso-providers/${ssoProviderId}`, formData, options);
    };

    getSAMLProviderSigningCertificate = (ssoProviderId: types.SSOProvider['id'], options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/sso-providers/${ssoProviderId}/signing-certificate`, options);

    deleteSSOProvider = (ssoProviderId: types.SSOProvider['id'], options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/sso-providers/${ssoProviderId}`, options);

    createOIDCProvider = (oidcProvider: CreateOIDCProviderRequest) =>
        this.baseClient.post(`/api/v2/sso-providers/oidc`, oidcProvider);

    updateOIDCProvider = (ssoProviderId: types.SSOProvider['id'], oidcProvider: UpdateOIDCProviderRequest) =>
        this.baseClient.patch(`/api/v2/sso-providers/${ssoProviderId}`, oidcProvider);

    listSSOProviders = (options?: RequestOptions) =>
        this.baseClient.get<types.ListSSOProvidersResponse>(`/api/v2/sso-providers`, options);

    permissionList = (options?: RequestOptions) => this.baseClient.get('/api/v2/permissions', options);

    permissionGet = (permissionId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/permissions/${permissionId}`, options);

    getRoles = (options?: RequestOptions) => this.baseClient.get<types.ListRolesResponse>(`/api/v2/roles`, options);

    getRole = (roleId: string, options?: RequestOptions) => this.baseClient.get(`/api/v2/roles/${roleId}`, options);

    getUserTokens = (userId: string, options?: RequestOptions) =>
        this.baseClient.get<ListAuthTokensResponse>(
            `/api/v2/tokens`,
            Object.assign(
                {
                    params: {
                        user_id: `eq:${userId}`,
                    },
                },
                options
            )
        );

    createUserToken = (userId: string, tokenName: string, options?: RequestOptions) =>
        this.baseClient.post<CreateAuthTokenResponse>(
            `/api/v2/tokens`,
            {
                user_id: userId,
                token_name: tokenName,
            },
            options
        );

    deleteUserToken = (tokenId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/tokens/${tokenId}`, options);

    listUsers = (options?: RequestOptions) =>
        this.baseClient.get<types.ListUsersResponse>('/api/v2/bloodhound-users', options);

    listUsersMinimal = (options?: RequestOptions) =>
        this.baseClient.get<types.ListUsersResponse>('/api/v2/bloodhound-users-minimal', options);

    getUser = (userId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/bloodhound-users/${userId}`, options);

    createUser = (user: CreateUserRequest, options?: RequestOptions) =>
        this.baseClient.post('/api/v2/bloodhound-users', user, options);

    updateUser = (userId: string, user: UpdateUserRequest, options?: RequestOptions) =>
        this.baseClient.patch(`/api/v2/bloodhound-users/${userId}`, user, options);

    deleteUser = (userId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/bloodhound-users/${userId}`, options);

    expireUserAuthSecret = (userId: string, options?: RequestOptions) =>
        this.baseClient.delete(`/api/v2/bloodhound-users/${userId}/secret`, options);

    putUserAuthSecret = (userId: string, payload: PutUserAuthSecretRequest, options?: RequestOptions) =>
        this.baseClient.put(
            `/api/v2/bloodhound-users/${userId}/secret`,
            {
                current_secret: payload.currentSecret,
                needs_password_reset: payload.needsPasswordReset,
                secret: payload.secret,
            },
            options
        );

    enrollMFA = (userId: string, data: { secret: string }, options?: RequestOptions) =>
        this.baseClient.post(`/api/v2/bloodhound-users/${userId}/mfa`, data, options);

    disenrollMFA = (userId: string, data: { secret?: string }, options?: RequestOptions) =>
        this.baseClient.delete(
            `/api/v2/bloodhound-users/${userId}/mfa`,
            Object.assign(
                {
                    headers: { 'Content-Type': 'application/json' },
                    data,
                },
                options
            )
        );

    getMFAActivationStatus = (userId: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/bloodhound-users/${userId}/mfa-activation`, options);

    activateMFA = (userId: string, data: { otp: string }, options?: RequestOptions) =>
        this.baseClient.post(`/api/v2/bloodhound-users/${userId}/mfa-activation`, data, options);

    acceptEULA = (options?: RequestOptions) => this.baseClient.put('/api/v2/accept-eula', options);

    acceptFedRAMPEULA = (options?: RequestOptions) => this.baseClient.put('/api/v2/fed-eula/accept', options);

    getFedRAMPEULAStatus = (options?: RequestOptions) =>
        this.baseClient.get<{ data: { accepted: boolean } }>('/api/v2/fed-eula/status', options);

    getFedRAMPEULAText = (options?: RequestOptions) =>
        this.baseClient.get<{ data: string }>('/api/v2/fed-eula/text', options);

    getFeatureFlags = (options?: RequestOptions) => this.baseClient.get('/api/v2/features', options);

    toggleFeatureFlag = (flagId: string | number, options?: RequestOptions) =>
        this.baseClient.put(`/api/v2/features/${flagId}/toggle`, options);

    getCollectors = (collectorType: types.CommunityCollectorType, options?: RequestOptions) =>
        this.baseClient.get<GetCollectorsResponse>(`/api/v2/collectors/${collectorType}`, options);

    getCommunityCollectors = (options?: RequestOptions): Promise<AxiosResponse<GetCommunityCollectorsResponse>> =>
        this.baseClient.get<GetCommunityCollectorsResponse>('/api/v2/kennel/manifest', options);

    getEnterpriseCollectors = (options?: RequestOptions): Promise<AxiosResponse<GetEnterpriseCollectorsResponse>> =>
        this.baseClient.get<GetEnterpriseCollectorsResponse>('/api/v2/kennel/enterprise-manifest', options);

    downloadCollector = (collectorType: types.CommunityCollectorType, version: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/collectors/${collectorType}/${version}`,
            Object.assign(
                {
                    responseType: 'blob',
                },
                options
            )
        );

    downloadCollectorChecksum = (
        collectorType: types.CommunityCollectorType,
        version: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/collectors/${collectorType}/${version}/checksum`,
            Object.assign(
                {
                    responseType: 'blob',
                },
                options
            )
        );

    downloadCollectorManifestAsset = (fileName: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/kennel/download/${fileName}`,
            Object.assign(
                {
                    responseType: 'blob',
                },
                options
            )
        );

    //Entity Endpoints
    getAZEntityInfoV2 = (
        entityType: string,
        id: string,
        relatedEntityType?: string,
        counts?: boolean,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/azure/${entityType}`,
            Object.assign(
                {
                    params: {
                        object_id: id,
                        related_entity_type: relatedEntityType,
                        counts,
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getBaseV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/base/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getBaseControllablesV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/base/${id}/controllables`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getBaseControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/base/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getComputerSessionsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/sessions`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerAdminUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/admin-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerRDPUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/rdp-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerDCOMUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/dcom-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerPSRemoteUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/ps-remote-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerSQLAdminsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/sql-admins`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerGroupMembershipV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/group-membership`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerAdminRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/admin-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerRDPRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/rdp-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerDCOMRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/dcom-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerPSRemoteRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/ps-remote-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerConstrainedUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/constrained-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerConstrainedDelegationRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/constrained-delegation-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getComputerControllablesV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/computers/${id}/controllables`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getDomainUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainGroupsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/groups`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/computers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainOUsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/ous`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainGPOsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/gpos`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainForeignUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/foreign-users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainForeignGroupsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/foreign-groups`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainForeignAdminsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/foreign-admins`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainForeignGPOControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/foreign-gpo-controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainInboundTrustsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/inbound-trusts`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainOutboundTrustsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/outbound-trusts`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainDCSyncersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/dc-syncers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainLinkedGPOsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/linked-gpos`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getDomainADCSEscalationsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${id}/adcs-escalations`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGPOV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getGPOOUsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}/ous`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGPOComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}/computers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGPOUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}/users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGPOControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGPOTierZeroV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/gpos/${id}/tier-zero`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getOUV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ous/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getOUGPOsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ous/${id}/gpos`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getOUUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ous/${id}/users`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getOUGroupsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ous/${id}/groups`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getOUComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ous/${id}/computers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getUserSessionsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/sessions`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserMembershipsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/memberships`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserAdminRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/admin-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserRDPRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/rdp-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserDCOMRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/dcom-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserPSRemoteRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/ps-remote-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserSQLAdminRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/sql-admin-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserConstrainedDelegationRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/users/${id}/constrained-delegation-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getUserControllablesV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/users/${id}/controllables`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getGroupSessionsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/sessions`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupMembersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/members`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupMembershipsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/memberships`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupAdminRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/admin-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupRDPRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/rdp-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupDCOMRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/dcom-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupPSRemoteRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/ps-remote-rights`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupControllablesV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/controllables`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getGroupControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/groups/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getContainerV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/containers/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getContainerControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/containers/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getAIACAV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/aiacas/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getAIACAControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/aiacas/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getAIACAPKIHierarchyV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/aiacas/${id}/pki-hierarchy`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getRootCAV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/rootcas/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getRootCAControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/rootcas/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getRootCAPKIHierarchyV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/rootcas/${id}/pki-hierarchy`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getEnterpriseCAV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/enterprisecas/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getEnterpriseCAControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/enterprisecas/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getEnterpriseCAPKIHierarchyV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/enterprisecas/${id}/pki-hierarchy`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getEnterpriseCAPublishedTemplatesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/enterprisecas/${id}/published-templates`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getNTAuthStoreV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ntauthstores/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getNTAuthStoreControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/ntauthstores/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getNTAuthStoreTrustedCAsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/ntauthstores/${id}/trusted-cas`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getCertTemplateV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/certtemplates/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getCertTemplateControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/certtemplates/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getCertTemplatePublishedToCAsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/certtemplates/${id}/published-to-cas`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getIssuancePolicyV2 = (id: string, counts?: boolean, options?: RequestOptions) =>
        this.baseClient.get(
            `/api/v2/issuancepolicies/${id}`,
            Object.assign(
                {
                    params: {
                        counts,
                    },
                },
                options
            )
        );

    getIssuancePolicyControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/issuancepolicies/${id}/controllers`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getIssuancePolicyLinkedTemplatesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/issuancepolicies/${id}/linkedtemplates`,
            Object.assign(
                {
                    params: {
                        skip,
                        limit,
                        type,
                    },
                },
                options
            )
        );

    getMetaV2 = (id: string, options?: RequestOptions) => this.baseClient.get(`/api/v2/meta/${id}`, options);

    getShortestPathV2 = (startNode: string, endNode: string, relationshipKinds?: string, options?: RequestOptions) =>
        this.baseClient.get<GraphResponse>(
            '/api/v2/graphs/shortest-path',
            Object.assign(
                {
                    params: {
                        start_node: startNode,
                        end_node: endNode,
                        relationship_kinds: relationshipKinds,
                        only_traversable: true,
                    },
                },
                options
            )
        );

    getEdgeComposition = (sourceNode: number, targetNode: number, edgeType: string, options?: RequestOptions) =>
        this.baseClient.get<GraphResponse>(
            '/api/v2/graphs/edge-composition',
            Object.assign(
                {
                    params: {
                        source_node: sourceNode,
                        target_node: targetNode,
                        edge_type: edgeType,
                    },
                },
                options
            )
        );

    getRelayTargets = (sourceNode: number, targetNode: number, edgeType: string, options?: RequestOptions) =>
        this.baseClient.get<GraphResponse>(
            '/api/v2/graphs/relay-targets',
            Object.assign(
                {
                    params: {
                        source_node: sourceNode,
                        target_node: targetNode,
                        edge_type: edgeType,
                    },
                },
                options
            )
        );

    getACLInheritance = (sourceNode: number, targetNode: number, edgeType: string, options?: RequestOptions) =>
        this.baseClient.get<GraphResponse>(
            '/api/v2/graphs/acl-inheritance',
            Object.assign(
                {
                    params: {
                        source_node: sourceNode,
                        target_node: targetNode,
                        edge_type: edgeType,
                    },
                },
                options
            )
        );

    /* remote assets */
    getRemoteAsset = (assetPath: string, options?: RequestOptions) =>
        this.baseClient.get(`/api/v2/assets/${assetPath}`, options);

    /* configuration */
    getConfiguration = (options?: RequestOptions) =>
        this.baseClient.get<GetConfigurationResponse>('/api/v2/config', options);

    updateConfiguration = (payload: UpdateConfigurationRequest, options?: RequestOptions) =>
        this.baseClient.put<UpdateConfigurationResponse>('/api/v2/config', payload, options);

    getEdgeTypes = (options?: RequestOptions) =>
        this.baseClient.get<GetEdgeTypesResponse>('/api/v2/graph-schema/edges', options);

    getDogTags = (options?: RequestOptions) => this.baseClient.get('/api/v2/dog-tags', options);
}

export default BHEAPIClient;
