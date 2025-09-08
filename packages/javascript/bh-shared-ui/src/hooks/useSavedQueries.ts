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

import {
    BasicResponse,
    CreateUserQueryRequest,
    DeleteUserQueryPermissionsRequest,
    GetExportQueryResponse,
    QueryScope,
    RequestOptions,
    SavedQuery,
    SavedQueryPermissionsResponse,
    UpdateUserQueryPermissionsRequest,
    UpdateUserQueryRequest,
} from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { GenericQueryOptions, getQueryKey } from '../utils';
import { apiClient } from '../utils/api';
export const savedQueryKeys = {
    all: ['savedQueries'] as const,
    permissions: ['permissions'] as const,
    queryId: (id?: number) => [`query-id-${id}`],
};

export const getSavedQueries = (scope: QueryScope, options?: RequestOptions): Promise<SavedQuery[]> => {
    return apiClient.getUserSavedQueries(scope, options).then((response) => response.data.data);
};

export const getExportQuery = (id: number, options?: RequestOptions): Promise<GetExportQueryResponse> => {
    return apiClient.getExportCypherQuery(id, options).then((response) => response);
};

export const createSavedQuery = (savedQuery: CreateUserQueryRequest, options?: RequestOptions): Promise<SavedQuery> => {
    return apiClient.createUserQuery(savedQuery, options).then((response) => response.data.data);
};

export const updateSavedQuery = (savedQuery: UpdateUserQueryRequest): Promise<SavedQuery> => {
    return apiClient.updateUserQuery(savedQuery).then((response) => response.data.data);
};

export const importSavedQuery = (
    savedQuery: FormData | Blob | object,
    options?: RequestOptions
): Promise<BasicResponse<any>> => {
    return apiClient.importUserQuery(savedQuery, options).then((response) => {
        return response.data;
    });
};

export const deleteSavedQuery = (id: number): Promise<void> => {
    return apiClient.deleteUserQuery(id).then((response) => response.data);
};

export const getQueryPermissions = async (
    id: number,
    options?: RequestOptions
): Promise<SavedQueryPermissionsResponse> => {
    const emptyPermissions = { query_id: undefined, public: false, shared_to_user_ids: [] };
    if (!id) {
        return emptyPermissions;
    }
    try {
        return await apiClient.getUserQueryPermissions(id, options).then((response) => response.data.data);
    } catch (error: any) {
        const status = error?.response?.status ?? error?.status;
        if (status === 404 || status === 400) {
            return emptyPermissions;
        }
        throw error;
    }
};

export const useQueryPermissions = (id?: number) =>
    useQuery({
        queryKey: getQueryKey([...savedQueryKeys.permissions], savedQueryKeys.queryId(id)),
        queryFn: ({ signal }) => getQueryPermissions(id as number, { signal }),
        retry: false,
        enabled: typeof id !== 'undefined',
    });

export const updateQueryPermissions = (
    { id, payload }: { id: number; payload: UpdateUserQueryPermissionsRequest },
    options?: RequestOptions
) => apiClient.updateUserQueryPermissions(id, payload, options).then((res) => res.data);

export const useUpdateQueryPermissions = () => {
    const queryClient = useQueryClient();
    return useMutation(updateQueryPermissions, {
        onSuccess: (data) => {
            queryClient.invalidateQueries(
                getQueryKey([...savedQueryKeys.permissions], savedQueryKeys.queryId(data.query_id))
            );
        },
    });
};

export const deleteQueryPermissions = (
    { id, payload }: { id: number; payload: DeleteUserQueryPermissionsRequest },
    options?: RequestOptions
) => apiClient.deleteUserQueryPermissions(id, payload, options).then((res) => res.data);

export const useDeleteQueryPermissions = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteQueryPermissions, {
        onSuccess: (data) => {
            queryClient.invalidateQueries(
                getQueryKey([...savedQueryKeys.permissions], savedQueryKeys.queryId(data.query_id))
            );
        },
    });
};

export const useSavedQueries = (
    scope: QueryScope = QueryScope.ALL,
    queryOptions?: GenericQueryOptions<SavedQuery[]>
) => {
    return useQuery({
        queryKey: ['savedQueries'],
        queryFn: ({ signal }) => getSavedQueries(scope, { signal }),
        ...queryOptions,
    });
};
export const useCreateSavedQuery = () => {
    const queryClient = useQueryClient();
    return useMutation(createSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};

export const useUpdateSavedQuery = () => {
    const queryClient = useQueryClient();
    return useMutation(updateSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};

export const useDeleteSavedQuery = () => {
    const queryClient = useQueryClient();
    return useMutation(deleteSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};

export const useImportSavedQuery = () => {
    const queryClient = useQueryClient();
    return useMutation(importSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};
