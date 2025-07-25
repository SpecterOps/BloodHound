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

import { CreateUserQueryRequest, RequestOptions, SavedQuery, UpdateUserQueryRequest } from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils/api';

export const savedQueryKeys = {
    all: ['savedQueries'] as const,
    permissions: ['savedQueries'] as const,
};

export const getSavedQueries = (options?: RequestOptions): Promise<SavedQuery[]> => {
    return apiClient.getUserSavedQueries(options).then((response) => response.data.data);
};

export const getExportQueries = (): Promise<any> => {
    return apiClient.getExportCypherQueries().then((response: any) => response);
};

export const getExportQuery = (id: number): Promise<void> => {
    return apiClient.getExportCypherQuery(id).then((response: any) => response);
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

export const getQueryPermissions = (id: number, options?: RequestOptions): Promise<any> => {
    return apiClient.getUserQueryPermissions(id, options).then((response: any) => response.data);
};

export const useSavedQueries = () => useQuery(savedQueryKeys.all, ({ signal }) => getSavedQueries({ signal }));

export const useCreateSavedQuery = () => {
    const queryClient = useQueryClient();

    return useMutation(createSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};

export const useQueryPermissions = (id: number) => {
    const queryClient = useQueryClient();

    return useMutation(() => getQueryPermissions(id), {
        onSuccess: (data) => {
            queryClient.invalidateQueries(savedQueryKeys.permissions);
            return data;
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
