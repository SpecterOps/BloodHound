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

import { RequestOptions, SavedQuery, CreateUserQueryRequest } from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../utils/api';

export const savedQueryKeys = {
    all: ['savedQueries'] as const,
};

export const getSavedQueries = (options?: RequestOptions): Promise<SavedQuery[]> => {
    return apiClient.getUserSavedQueries(options).then((response) => response.data.data);
};

export const createSavedQuery = (savedQuery: CreateUserQueryRequest, options?: RequestOptions): Promise<SavedQuery> => {
    return apiClient.createUserQuery(savedQuery, options).then((response) => response.data.data);
};

export const deleteSavedQuery = (id: number): Promise<void> => {
    return apiClient.deleteUserQuery(id).then((response) => response.data);
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

export const useDeleteSavedQuery = () => {
    const queryClient = useQueryClient();

    return useMutation(deleteSavedQuery, {
        onSuccess: () => {
            queryClient.invalidateQueries(savedQueryKeys.all);
        },
    });
};
