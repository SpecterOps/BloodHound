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

import { SagaIterator } from 'redux-saga';
import { all, call, fork, put, takeLatest } from 'redux-saga/effects';
import { apiClient } from 'bh-shared-ui';
import { putGraphData, putGraphError, putGraphVars, saveResponseForExport } from 'src/ducks/explore/actions';
import {
    AssetGroupRequest,
    GraphEndpoints,
    GraphRequestType,
    GRAPH_START,
    NodeInfoRequest,
    PathfindingRequest,
    SearchRequest,
    ShortestPathRequest,
    CypherQueryRequest,
} from 'src/ducks/explore/types';
import { addSnackbar } from 'src/ducks/global/actions';
import { getLinksIndex, getNodesIndex } from 'src/ducks/graph/graphutils';
import { ActiveDirectoryRelationshipKind, AzureRelationshipKind } from 'bh-shared-ui';
import { transformFlatGraphResponse, transformToFlatGraphResponse } from 'src/utils';

function* graphQueryWatcher(): SagaIterator {
    yield takeLatest(GRAPH_START, graphQueryWorker);
}

function isNodeInfoRequest(request: GraphRequestType): request is NodeInfoRequest {
    return (request as NodeInfoRequest).url !== undefined;
}

function isSearchRequest(request: GraphRequestType): request is SearchRequest {
    return (request as SearchRequest).objectid !== undefined;
}

function isPathfindingRequest(request: GraphRequestType): request is PathfindingRequest {
    return (request as PathfindingRequest).edges !== undefined;
}

function isShortestPathRequest(request: GraphRequestType): request is ShortestPathRequest {
    return (request as ShortestPathRequest).edges !== undefined;
}

function isAssetGroupRequest(request: GraphRequestType): request is AssetGroupRequest {
    return (request as AssetGroupRequest).assetGroupId !== undefined;
}

function isCypherQueryRequest(request: GraphRequestType): request is CypherQueryRequest {
    return (request as CypherQueryRequest).cypherQuery !== undefined;
}

function* graphQueryWorker(payload: GraphRequestType) {
    if (isPathfindingRequest(payload)) {
        yield call(runPathfindingQuery, payload);
    } else if (isShortestPathRequest(payload)) {
        /* empty */
    } else if (isNodeInfoRequest(payload)) {
        yield call(runNodeInfoQuery, payload);
    } else if (isSearchRequest(payload)) {
        yield call(runSearchQuery, payload);
    } else if (isAssetGroupRequest(payload)) {
        yield call(runAssetGroupQuery, payload);
    } else if (isCypherQueryRequest(payload)) {
        yield call(runCypherSearchQuery, payload);
    } else {
        yield call(runApiQuery, payload.endpoint);
    }
}

function* runNodeInfoQuery(payload: NodeInfoRequest): SagaIterator {
    let response;
    try {
        response = yield call(apiClient.baseClient.get, payload.url);
        const data = response;
        yield put(putGraphData(data));
    } catch (e) {
        yield put(putGraphError(e));
        yield put(addSnackbar('Query failed. Please try again.', 'nodeInfoQueryFailure', {}));
        return;
    }
}

function* runSearchQuery(payload: SearchRequest): SagaIterator {
    let response;
    try {
        response = yield call(apiClient.getSearchResult, payload.objectid, payload.searchType);
        const data = response.data.data;

        const formattedData = transformFlatGraphResponse(data);
        yield put(saveResponseForExport(formattedData));

        yield put(putGraphData(data));
    } catch (e) {
        yield put(putGraphError(e));
        return;
    }
}

function* runPathfindingQuery(payload: PathfindingRequest): SagaIterator {
    const standardExclusions = [
        ActiveDirectoryRelationshipKind.LocalToComputer,
        ActiveDirectoryRelationshipKind.RemoteInteractiveLogonPrivilege,
        ActiveDirectoryRelationshipKind.MemberOfLocalGroup,
        ActiveDirectoryRelationshipKind.GetChanges,
        ActiveDirectoryRelationshipKind.GetChangesAll,
        ActiveDirectoryRelationshipKind.RootCAFor,
        ActiveDirectoryRelationshipKind.PublishedTo,
        ActiveDirectoryRelationshipKind.ManageCertificates,
        ActiveDirectoryRelationshipKind.ManageCA,
        ActiveDirectoryRelationshipKind.DelegatedEnrollmentAgent,
        ActiveDirectoryRelationshipKind.Enroll,
        ActiveDirectoryRelationshipKind.WritePKIEnrollmentFlag,
        ActiveDirectoryRelationshipKind.WritePKINameFlag,
        ActiveDirectoryRelationshipKind.NTAuthStoreFor,
        ActiveDirectoryRelationshipKind.TrustedForNTAuth,
        ActiveDirectoryRelationshipKind.EnterpriseCAFor,
        ActiveDirectoryRelationshipKind.IssuedSignedBy,
        ActiveDirectoryRelationshipKind.EnrollOnBehalfOf,
        ActiveDirectoryRelationshipKind.HostsCAService,
        AzureRelationshipKind.ApplicationReadWriteAll,
        AzureRelationshipKind.AppRoleAssignmentReadWriteAll,
        AzureRelationshipKind.DirectoryReadWriteAll,
        AzureRelationshipKind.GroupReadWriteAll,
        AzureRelationshipKind.GroupMemberReadWriteAll,
        AzureRelationshipKind.RoleManagementReadWriteDirectory,
        AzureRelationshipKind.ServicePrincipalEndpointReadWriteAll,
    ];

    // treat no edges selected the same as all edges selected
    const defaultRelationshipKindsFilter = `nin:${standardExclusions.join(',')}`;
    let relationshipKindsFilter = defaultRelationshipKindsFilter;

    // if edges have been selected, use them to filter
    if (payload.edges.length !== 0) {
        relationshipKindsFilter = `in:${payload.edges.join(',')}`;
    }

    try {
        const { data } = yield call(
            apiClient.getShortestPathV2,
            payload.start,
            payload.end,
            `${relationshipKindsFilter}`
        );
        yield put(saveResponseForExport(data));

        const flatGraph = transformToFlatGraphResponse(data);
        yield put(putGraphData(flatGraph));
    } catch (error: any) {
        if (error?.response?.status === 404) {
            yield put(putGraphData({}));
            yield put(addSnackbar('Path not found.', 'shortestPathNotFound', {}));
        } else if (error?.response?.status === 503) {
            yield put(
                addSnackbar(
                    'Calculating the requested Attack Path exceeded memory limitations due to the complexity of paths involved.',
                    'shortestPathOutOfMemory',
                    {}
                )
            );
        } else if (error?.response?.status === 504) {
            yield put(
                addSnackbar(
                    'The results took too long to compute, possibly due to the complexity of paths involved.',
                    'shortestPathTimeout',
                    {}
                )
            );
        } else {
            yield put(addSnackbar('An unknown error occurred. Please try again.', 'shortestPathUnknown', {}));
        }
        yield put(putGraphError(error));
    }
}

function* runCypherSearchQuery(payload: CypherQueryRequest): SagaIterator {
    try {
        const { data } = yield call(apiClient.cypherSearch, payload.cypherQuery);

        const resultNodesAreEmpty = Object.keys(data?.data?.nodes).length === 0;
        const resultEdgesAreEmpty = Object.keys(data?.data?.edges).length === 0;

        if (resultNodesAreEmpty && !resultEdgesAreEmpty) {
            yield put(putGraphData({}));
            yield put(
                addSnackbar(
                    'The results are not rendered since only edges were returned',
                    'cypherSearchOnlyContainsEdges'
                )
            );
        } else if (resultNodesAreEmpty && resultEdgesAreEmpty) {
            yield put(putGraphData({}));
            yield put(addSnackbar('No results match your criteria', 'cypherSearchEmptyResponse'));
        } else {
            yield put(saveResponseForExport(data));

            const flatGraph = transformToFlatGraphResponse(data);
            yield put(putGraphData(flatGraph));
        }
    } catch (error: any) {
        const apiErrorMessage = error?.response?.data?.errors?.[0]?.message;

        if (error?.response?.status === 400) {
            yield put(addSnackbar(apiErrorMessage, 'cypherSearchBadRequest'));
        } else {
            if (apiErrorMessage) {
                yield put(addSnackbar(`${apiErrorMessage}`, 'cypherSearch'));
            } else {
                yield put(addSnackbar('An error occured. Please try again', 'cypherSearch'));
            }
        }
        yield put(putGraphError(error));
    }
}

function* runAssetGroupQuery(payload: AssetGroupRequest): SagaIterator {
    try {
        if (payload.domainId) {
            const response = yield call(apiClient.getAssetGroupComboNode, payload.assetGroupId, payload.domainId);
            yield put(putGraphData(response.data.data));
        } else {
            const response = yield call(apiClient.getAssetGroupComboNode, payload.assetGroupId);
            const rels = getLinksIndex(response.data.data);
            const nodes = getNodesIndex(response.data.data);

            const filteredNodes = Object.fromEntries(
                Object.entries(nodes).filter((node: any) => {
                    const nodeData = node[1];
                    if (payload.domainType === 'active-directory-platform') {
                        return !!nodeData.data.domainsid;
                    } else if (payload.domainType === 'azure-platform') {
                        return !!nodeData.data.tenantid;
                    } else return false;
                })
            );

            yield put(putGraphData({ ...filteredNodes, ...rels }));
        }
        yield put(
            putGraphVars({
                combine: {
                    properties: ['nodetype', 'category'],
                    level: 2,
                },
            })
        );
    } catch (e) {
        yield put(putGraphError(e));
    }
}

function* runApiQuery(endpoint: GraphEndpoints, ext?: string): SagaIterator {
    const fullUrl = window.location.origin + endpoint + (ext !== undefined ? '/' + ext : '');

    try {
        const response = yield call(apiClient.baseClient.get, fullUrl);

        yield put(putGraphData(response));
    } catch (e) {
        yield put(putGraphError(e));
        return;
    }
}

export default function* startGraphSagas() {
    yield all([fork(graphQueryWatcher)]);
}
