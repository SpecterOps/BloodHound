// Copyright 2024 Specter Ops, Inc.
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

import { ConfigurationPayload, RequestOptions } from 'js-client-library';
import { apiClient } from '../utils';
import { useMutation, useQuery, useQueryClient } from 'react-query';

const configurationKeys = {
    all: ['configuration'] as const,
};

const getConfiguration = (options?: RequestOptions) => {
    return apiClient.getConfiguration(options).then((res) => res.data);
};

export const useGetConfiguration = () => {
    return useQuery(configurationKeys.all, ({ signal }) => getConfiguration({ signal }), {
        refetchOnWindowFocus: false,
    });
};

const updateConfiguration = (payload: ConfigurationPayload, options?: RequestOptions) => {
    return apiClient.updateConfiguration(payload, options).then((res) => res.data);
};

export const useUpdateConfiguration = () => {
    const queryClient = useQueryClient();

    return useMutation(updateConfiguration, {
        onSuccess: () => {
            queryClient.invalidateQueries(configurationKeys.all);
        },
    });
};
