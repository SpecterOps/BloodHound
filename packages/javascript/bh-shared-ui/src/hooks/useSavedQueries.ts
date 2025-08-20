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
    CreateUserQueryRequest,
    DeleteUserQueryPermissionsRequest,
    QueryScope,
    RequestOptions,
    SavedQuery,
    UpdateUserQueryPermissionsRequest,
    UpdateUserQueryRequest,
} from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils/api';
export const savedQueryKeys = {
    all: ['savedQueries'] as const,
    permissions: ['permissions'] as const,
};

export const getSavedQueries = (scope: QueryScope, options?: RequestOptions): Promise<SavedQuery[]> => {
    return apiClient.getUserSavedQueries(scope, options).then((response) => response.data.data);
};

export const getExportQueries = (): Promise<any> => {
    return apiClient.getExportCypherQueries().then((response: any) => response);
};

export const getExportQuery = (id: number) => {
    return apiClient.getExportCypherQuery(id).then((response) => response);
};

export const createSavedQuery = (savedQuery: CreateUserQueryRequest, options?: RequestOptions): Promise<SavedQuery> => {
    return apiClient.createUserQuery(savedQuery, options).then((response) => response.data.data);
};

export const updateSavedQuery = (savedQuery: UpdateUserQueryRequest, options?: RequestOptions): Promise<SavedQuery> => {
    return apiClient.updateUserQuery(savedQuery, options).then((response) => response.data.data);
};

export const importSavedQuery = (savedQuery: any, options?: RequestOptions): Promise<any> => {
    return apiClient.importUserQuery(savedQuery, options).then((response) => response.data);
};

export const deleteSavedQuery = (id: number): Promise<void> => {
    return apiClient.deleteUserQuery(id).then((response) => response.data);
};

export const getQueryPermissions = async (id: number, options?: RequestOptions): Promise<any> => {
    try {
        return await apiClient.getUserQueryPermissions(id, options).then((response) => response.data.data);
    } catch (error: any) {
        if (error.status === 404 || error.status === 400) {
            return { query_id: undefined, public: false, shared_to_user_ids: [] };
        }
        return error;
    }
};

export const useQueryPermissions = (id: number) =>
    useQuery(savedQueryKeys.permissions, ({ signal }) => getQueryPermissions(id, { signal }), {
        retry: false,
    });

export const updateQueryPermissions = (
    { id, payload }: { id: number; payload: UpdateUserQueryPermissionsRequest },
    options?: RequestOptions
) => apiClient.updateUserQueryPermissions(id, payload, options).then((res) => res.data);

export const useUpdateQueryPermissions = () => {
    const queryClient = useQueryClient();
    return useMutation(updateQueryPermissions, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.permissions);
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
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.permissions);
        },
    });
};

export const useSavedQueries = (scope: QueryScope = QueryScope.ALL) => {
    return useQuery(savedQueryKeys.all, ({ signal }) => getSavedQueries(scope, { signal }));
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
        onSuccess: (data) => {
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
