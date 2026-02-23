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
import { Extension } from 'js-client-library';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient, GenericQueryOptions } from '../utils';

export const extensionsKeys = {
    all: ['extensions'],
};

export function useExtensionsQuery(queryOptions?: GenericQueryOptions<Extension[]>) {
    return useQuery({
        queryKey: extensionsKeys.all,
        queryFn: ({ signal }) => apiClient.getExtensions({ signal }).then((res) => res.data.data.extensions),
        ...queryOptions,
    });
}

export function useDeleteExtension() {
    const queryClient = useQueryClient();
    return useMutation((extensionId: string) => apiClient.deleteExtension(extensionId), {
        onSuccess: () => {
            queryClient.invalidateQueries(extensionsKeys.all);
        },
    });
}
