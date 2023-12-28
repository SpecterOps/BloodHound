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

import axios, { AxiosInstance } from 'axios';
import * as types from './types';
import {
    BasicResponse,
    CreateAuthTokenResponse,
    ListAuthTokensResponse,
    AssetGroupMembersResponse,
    AssetGroupResponse,
    PaginatedResponse,
    PostureResponse,
    SavedQuery,
} from './responses';

class BHEAPIClient {
    baseClient: AxiosInstance;

    constructor(config: types.RequestOptions) {
        this.baseClient = axios.create(config);
    }

    /* health */
    health = (options?: types.RequestOptions) => this.baseClient.get('/health', options);

    /* version */
    version = (options?: types.RequestOptions) => this.baseClient.get('/api/version', options);

    /* datapipe status */
    getDatapipeStatus = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/datapipe/status', options);

    /* search */
    searchHandler = (keyword: string, type?: string, options?: types.RequestOptions) => {
        return this.baseClient.get(
            '/api/v2/search',
            Object.assign(
                {
                    params: {
                        q: keyword,
                        type: type,
                    },
                    headers: {
                        Prefer: 'wait=60',
                    },
                },
                options
            )
        );
    };

    cypherSearch = (query: string, includeProperties?: boolean, options?: types.RequestOptions) => {
        return this.baseClient.post('/api/v2/graphs/cypher', { query, include_properties: includeProperties }, options);
    };

    getUserSavedQueries = (options?: types.RequestOptions) => {
        return this.baseClient.get<PaginatedResponse<SavedQuery[]>>(
            '/api/v2/saved-queries',
            Object.assign(
                {
                    params: {
                        sort_by: 'name',
                    },
                },
                options
            )
        );
    };

    createUserQuery = (payload: types.CreateUserQueryRequest, options?: types.RequestOptions) => {
        return this.baseClient.post<BasicResponse<SavedQuery>>('/api/v2/saved-queries', payload, options);
    };

    deleteUserQuery = (queryId: number, options?: types.RequestOptions) => {
        return this.baseClient.delete(`/api/v2/saved-queries/${queryId}`, options);
    };

    getAvailableDomains = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/available-domains', options);

    /* audit */
    getAuditLogs = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/audit', options);

    /* asset groups */
    createAssetGroup = (assetGroup: types.CreateAssetGroupRequest, options?: types.RequestOptions) =>
        this.baseClient.post('/api/v2/asset-groups', assetGroup, options);

    getAssetGroup = (assetGroupId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/asset-groups/${assetGroupId}`, options);

    deleteAssetGroup = (assetGroupId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-groups/${assetGroupId}`, options);

    updateAssetGroup = (
        assetGroupId: string,
        assetGroup: types.UpdateAssetGroupRequest,
        options?: types.RequestOptions
    ) => this.baseClient.put(`/api/v2/asset-groups/${assetGroupId}`, assetGroup, options);

    updateAssetGroupSelector = (
        assetGroupId: string,
        selectorChangeset: types.UpdateAssetGroupSelectorRequest[],
        options?: types.RequestOptions
    ) => this.baseClient.put(`/api/v2/asset-groups/${assetGroupId}/selectors`, selectorChangeset, options);

    deleteAssetGroupSelector = (assetGroupId: string, selectorId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/asset-groups/${assetGroupId}/selectors/${selectorId}`, options);

    listAssetGroupCollections = (assetGroupId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/asset-groups/${assetGroupId}/collections`, options);

    listAssetGroupMembers = (
        assetGroupId: string,
        params?: types.AssetGroupMemberParams,
        options?: types.RequestOptions
    ) =>
        this.baseClient.get<AssetGroupMembersResponse>(
            `/api/v2/asset-groups/${assetGroupId}/members`,
            Object.assign({ params }, options)
        );

    listAssetGroups = (options?: types.RequestOptions) =>
        this.baseClient.get<AssetGroupResponse>('/api/v2/asset-groups', options);

    /* analysis */
    getComboTreeGraph = (domainId: string, nodeId: string | null = null, options?: types.RequestOptions) =>
        this.baseClient.get(
            `/api/v2/meta-trees/${domainId}`,
            Object.assign({ params: nodeId && { node_id: nodeId } }, options)
        );

    getLatestTierZeroComboNode = (domainId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/meta-nodes/${domainId}`, options);

    getAssetGroupComboNode = (assetGroupId: string, domainsid?: string, options?: types.RequestOptions) => {
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
        options?: types.RequestOptions
    ) => {
        return this.baseClient.get(
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
        options?: types.RequestOptions
    ) => {
        return this.baseClient.get(
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
        options?: types.RequestOptions
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

    getPostureStats = (from: Date, to: Date, domainSID?: string, sortBy?: string, options?: types.RequestOptions) => {
        const params: types.PostureRequest = {
            from: from.toISOString(),
            to: to.toISOString(),
        };

        if (domainSID) params.domain_sid = `eq:${domainSID}`;
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

    /* ingest */
    ingestData = (options?: types.RequestOptions) => this.baseClient.post('/api/v2/ingest', options);

    getPathfindingResult = (startNode: string, endNode: string, options?: types.RequestOptions) =>
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

    getSearchResult = (query: string, searchType: string, options?: types.RequestOptions) =>
        this.baseClient.get(
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

    /* clients */
    getClients = (
        skip: number = 0,
        limit: number = 10,
        hydrateDomains?: boolean,
        hydrateOUs?: boolean,
        options?: types.RequestOptions
    ) =>
        this.baseClient.get(
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

    createClient = (
        client: types.CreateSharpHoundClientRequest | types.CreateAzureHoundClientRequest,
        options?: types.RequestOptions
    ) => this.baseClient.post('/api/v2/clients', client, options);

    getClient = (clientId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/clients/${clientId}`, options);

    updateClient = (
        clientId: string,
        client: types.UpdateSharpHoundClientRequest | types.UpdateAzureHoundClientRequest,
        options?: types.RequestOptions
    ) => this.baseClient.put(`/api/v2/clients/${clientId}`, client, options);

    regenerateClientToken = (clientId: string, options?: types.RequestOptions) =>
        this.baseClient.put(`/api/v2/clients/${clientId}/token`, options);

    deleteClient = (clientId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/clients/${clientId}`, options);

    getClientCompletedJobs = (clientId: string, skip: number, limit: number, options?: types.RequestOptions) =>
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

    createScheduledJob = (
        clientId: string,
        scheduledJob: types.CreateScheduledJobRequest,
        options?: types.RequestOptions
    ) => this.baseClient.post(`/api/v2/clients/${clientId}/jobs`, scheduledJob, options);

    getFinishedJobs = (
        skip: number,
        limit: number,
        hydrateDomains?: boolean,
        hydrateOUs?: boolean,
        options?: types.RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/jobs/finished`,
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

    /* events */
    getEvents = (hydrateDomains?: boolean, hydrateOUs?: boolean, options?: types.RequestOptions) =>
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

    getEvent = (eventId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/events/${eventId}`, options);

    createEvent = (
        event: types.CreateSharpHoundEventRequest | types.CreateAzureHoundEventRequest,
        options?: types.RequestOptions
    ) => this.baseClient.post('/api/v2/events', event, options);

    updateEvent = (
        eventId: string,
        event: types.UpdateSharpHoundEventRequest | types.UpdateAzureHoundEventRequest,
        options?: types.RequestOptions
    ) => this.baseClient.put(`/api/v2/events/${eventId}`, event, options);

    deleteEvent = (eventId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/events/${eventId}`, options);

    /* file ingest */
    listFileIngestJobs = (skip?: number, limit?: number, sortBy?: string) =>
        this.baseClient.get(
            'api/v2/file-upload',
            Object.assign({
                params: {
                    skip,
                    limit,
                    sort_by: sortBy,
                },
            })
        );

    startFileIngest = () => this.baseClient.post('/api/v2/file-upload/start');

    uploadFileToIngestJob = (ingestId: string, json: any) => {
        return this.baseClient.post(`/api/v2/file-upload/${ingestId}`, json);
    };

    endFileIngest = (ingestId: string) => this.baseClient.post(`/api/v2/file-upload/${ingestId}/end`);

    /* jobs */
    getJobs = (hydrateDomains?: boolean, hydrateOUs?: boolean, options?: types.RequestOptions) =>
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

    getJob = (jobId: string, options?: types.RequestOptions) => this.baseClient.get(`/api/v2/jobs/${jobId}`, options);

    cancelScheduledJob = (jobId: string, options?: types.RequestOptions) =>
        this.baseClient.put(`/api/v2/jobs/${jobId}/cancel`, undefined, options);

    getJobLogFile = (jobId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/jobs/${jobId}/log`, options);

    getRiskDetails = (
        domainId: string,
        finding: string,
        skip: number,
        limit: number,
        filterAccepted?: boolean,
        options?: types.RequestOptions
    ) => {
        const params: types.RiskDetailsRequest = {
            finding: finding,
            skip: skip,
            limit: limit,
        };

        if (typeof filterAccepted === 'boolean') params.Accepted = `eq:${filterAccepted}`;

        return this.baseClient.get(
            `/api/v2/domains/${domainId}/details`,
            Object.assign(
                {
                    params: params,
                },
                options
            )
        );
    };

    getRiskSparklineValues = (
        domainId: string,
        finding: string,
        from?: Date,
        to?: Date,
        sortBy?: string,
        options?: types.RequestOptions
    ) =>
        this.baseClient.get(
            `/api/v2/domains/${domainId}/sparkline`,
            Object.assign(
                {
                    params: {
                        finding,
                        from: from?.toISOString(),
                        to: to?.toISOString(),
                        sort_by: sortBy,
                    },
                },
                options
            )
        );

    changeRiskAcceptance = (
        attackPathId: string,
        riskType: string,
        accepted: boolean,
        acceptUntil?: Date,
        options?: types.RequestOptions
    ) =>
        this.baseClient.put(
            `/api/v2/attack-paths/${attackPathId}/acceptance`,
            {
                risk_type: riskType,
                accepted: accepted,
                accept_until: acceptUntil && acceptUntil.toISOString(),
            },
            options
        );

    genRisks = (options?: types.RequestOptions) => this.baseClient.put('/api/v2/attack-paths', options);

    getAvailableRiskTypes = (domainId: string, sortBy?: string, options?: types.RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${domainId}/available-types`,
            Object.assign(
                {
                    params: {
                        sort_by: sortBy,
                    },
                },
                options
            )
        );

    exportRiskFindings = (domainId: string, findingType: string, accepted?: boolean, options?: types.RequestOptions) =>
        this.baseClient.get(
            `/api/v2/domains/${domainId}/attack-path-findings`,
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

    /* auth */
    login = (credentials: types.LoginRequest, options?: types.RequestOptions) =>
        this.baseClient.post<types.LoginResponse>('/api/v2/login', credentials, options);

    getSelf = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/self', options);

    logout = (options?: types.RequestOptions) => this.baseClient.post('/api/v2/logout', options);

    listSAMLSignOnEndpoints = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/saml/sso', options);

    listSAMLProviders = (options?: types.RequestOptions) => this.baseClient.get(`/api/v2/saml`, options);

    getSAMLProvider = (samlProviderId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/saml/providers/${samlProviderId}`, options);

    createSAMLProvider = (
        data: {
            name: string;
            displayName: string;
            signingCertificate: string;
            issuerUri: string;
            singleSignOnUri: string;
            principalAttributeMappings: string[];
        },
        options?: types.RequestOptions
    ) =>
        this.baseClient.post(
            `/api/v2/saml`,
            {
                name: data.name,
                display_name: data.displayName,
                signing_certificate: data.signingCertificate,
                issuer_uri: data.issuerUri,
                single_signon_uri: data.singleSignOnUri,
                principal_attribute_mappings: data.principalAttributeMappings,
            },
            options
        );

    createSAMLProviderFromFile = (data: { name: string; metadata: File }, options?: types.RequestOptions) => {
        const formData = new FormData();
        formData.append('name', data.name);
        formData.append('metadata', data.metadata);
        return this.baseClient.post(`/api/v2/saml/providers`, formData, options);
    };

    validateSAMLProvider = (
        data: {
            name: string;
            displayName: string;
            signingCertificate: string;
            issuerUri: string;
            singleSignOnUri: string;
            principalAttributeMappings: string[];
        },
        options?: types.RequestOptions
    ) =>
        this.baseClient.post(
            `/api/v2/saml/validate-idp`,
            {
                name: data.name,
                display_name: data.displayName,
                signing_certificate: data.signingCertificate,
                issuer_uri: data.issuerUri,
                single_signon_uri: data.singleSignOnUri,
                principal_attribute_mappings: data.principalAttributeMappings,
            },
            options
        );

    deleteSAMLProvider = (SAMLProviderId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/saml/providers/${SAMLProviderId}`, options);

    permissionList = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/permissions', options);

    permissionGet = (permissionId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/permissions/${permissionId}`, options);

    getRoles = (options?: types.RequestOptions) => this.baseClient.get(`/api/v2/roles`, options);

    getRole = (roleId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/roles/${roleId}`, options);

    getUserTokens = (userId: string, options?: types.RequestOptions) =>
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

    createUserToken = (userId: string, tokenName: string, options?: types.RequestOptions) =>
        this.baseClient.post<CreateAuthTokenResponse>(
            `/api/v2/tokens`,
            {
                user_id: userId,
                token_name: tokenName,
            },
            options
        );

    deleteUserToken = (tokenId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/tokens/${tokenId}`, options);

    listUsers = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/bloodhound-users', options);

    getUser = (userId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/bloodhound-users/${userId}`, options);

    createUser = (
        user: {
            firstName: string;
            lastName: string;
            emailAddress: string;
            principal: string;
            roles: number[];
            SAMLProviderId?: string;
            password?: string;
            needsPasswordReset?: boolean;
        },
        options?: types.RequestOptions
    ) =>
        this.baseClient.post(
            '/api/v2/bloodhound-users',
            {
                first_name: user.firstName,
                last_name: user.lastName,
                email_address: user.emailAddress,
                principal: user.principal,
                roles: user.roles,
                saml_provider_id: user.SAMLProviderId,
                secret: user.password,
                needs_password_reset: user.needsPasswordReset,
            },
            options
        );

    updateUser = (
        userId: string,
        user: {
            firstName: string;
            lastName: string;
            emailAddress: string;
            principal: string;
            roles: number[];
            SAMLProviderId?: string;
            is_disabled?: boolean;
        },
        options?: types.RequestOptions
    ) =>
        this.baseClient.patch(
            `/api/v2/bloodhound-users/${userId}`,
            {
                first_name: user.firstName,
                last_name: user.lastName,
                email_address: user.emailAddress,
                principal: user.principal,
                roles: user.roles,
                saml_provider_id: user.SAMLProviderId,
                is_disabled: user.is_disabled,
            },
            options
        );

    deleteUser = (userId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/bloodhound-users/${userId}`, options);

    expireUserAuthSecret = (userId: string, options?: types.RequestOptions) =>
        this.baseClient.delete(`/api/v2/bloodhound-users/${userId}/secret`, options);

    putUserAuthSecret = (userId: string, userSecret: types.PutUserAuthSecretRequest, options?: types.RequestOptions) =>
        this.baseClient.put(`/api/v2/bloodhound-users/${userId}/secret`, userSecret, options);

    enrollMFA = (userId: string, data: { secret: string }, options?: types.RequestOptions) =>
        this.baseClient.post(`/api/v2/bloodhound-users/${userId}/mfa`, data, options);

    disenrollMFA = (userId: string, data: { secret: string }, options?: types.RequestOptions) =>
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

    getMFAActivationStatus = (userId: string, options?: types.RequestOptions) =>
        this.baseClient.get(`/api/v2/bloodhound-users/${userId}/mfa-activation`, options);

    activateMFA = (userId: string, data: { otp: string }, options?: types.RequestOptions) =>
        this.baseClient.post(`/api/v2/bloodhound-users/${userId}/mfa-activation`, data, options);

    acceptEULA = (options?: types.RequestOptions) => this.baseClient.put('/api/v2/accept-eula', options);

    getFeatureFlags = (options?: types.RequestOptions) => this.baseClient.get('/api/v2/features', options);

    toggleFeatureFlag = (flagId: string | number, options?: types.RequestOptions) =>
        this.baseClient.put(`/api/v2/features/${flagId}/toggle`, options);

    getCollectors = (collectorType: 'sharphound' | 'azurehound', options?: types.RequestOptions) =>
        this.baseClient.get<types.GetCollectorsResponse>(`/api/v2/collectors/${collectorType}`, options);

    downloadCollector = (collectorType: 'sharphound' | 'azurehound', version: string, options?: types.RequestOptions) =>
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
        collectorType: 'sharphound' | 'azurehound',
        version: string,
        options?: types.RequestOptions
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

    //Entity Endpoints
    getAZEntityInfoV2 = (
        entityType: string,
        id: string,
        relatedEntityType?: string,
        counts?: boolean,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
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

    getBaseV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getBaseControllablesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getBaseControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getComputerV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getComputerSessionsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerAdminUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerRDPUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerDCOMUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerPSRemoteUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerSQLAdminsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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
        options?: types.RequestOptions
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

    getComputerAdminRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerRDPRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerDCOMRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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
        options?: types.RequestOptions
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
        options?: types.RequestOptions
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
        options?: types.RequestOptions
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

    getComputerControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getComputerControllablesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getDomainUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainGroupsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainOUsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainGPOsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainForeignUsersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainForeignGroupsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainForeignAdminsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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
        options?: types.RequestOptions
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

    getDomainInboundTrustsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainOutboundTrustsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getDomainDCSyncersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getDomainLinkedGPOsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGPOV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getGPOOUsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGPOComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGPOUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGPOControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGPOTierZeroV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getOUV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getOUGPOsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getOUUsersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getOUGroupsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getOUComputersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getUserSessionsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserMembershipsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserAdminRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserRDPRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserDCOMRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserPSRemoteRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getUserSQLAdminRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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
        options?: types.RequestOptions
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

    getUserControllersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getUserControllablesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGroupV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getGroupSessionsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGroupMembersV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGroupMembershipsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGroupAdminRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGroupRDPRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGroupDCOMRightsV2 = (id: string, skip?: number, limit?: number, type?: string, options?: types.RequestOptions) =>
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

    getGroupPSRemoteRightsV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGroupControllablesV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getGroupControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getContainerV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getContainerControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getAIACAV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getAIACAControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getRootCAV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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

    getRootCAControllersV2 = (
        id: string,
        skip?: number,
        limit?: number,
        type?: string,
        options?: types.RequestOptions
    ) =>
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

    getEnterpriseCAV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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
        options?: types.RequestOptions
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

    getNTAuthStoreV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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
        options?: types.RequestOptions
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

    getCertTemplateV2 = (id: string, counts?: boolean, options?: types.RequestOptions) =>
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
        options?: types.RequestOptions
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

    getMetaV2 = (id: string, options?: types.RequestOptions) => this.baseClient.get(`/api/v2/meta/${id}`, options);

    getShortestPathV2 = (
        startNode: string,
        endNode: string,
        relationshipKinds?: string,
        options?: types.RequestOptions
    ) =>
        this.baseClient.get<types.GraphResponse>(
            '/api/v2/graphs/shortest-path',
            Object.assign(
                {
                    params: {
                        start_node: startNode,
                        end_node: endNode,
                        relationship_kinds: relationshipKinds,
                    },
                },
                options
            )
        );

    getEdgeComposition = (sourceNode: number, targetNode: number, edgeType: string, options?: types.RequestOptions) =>
        this.baseClient.get<types.GraphResponse>(
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
}

export default BHEAPIClient;
