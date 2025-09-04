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
import { RequestOptions } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils/api';

const getSelf = (options?: RequestOptions) => apiClient.getSelf(options).then((res) => res.data.data);

export const useSelf = () => {
    const getSelfId = useQuery(['getSelf'], ({ signal }) => getSelf({ signal }), {
        cacheTime: Number.POSITIVE_INFINITY,
        select: (data) => {
            return data.id;
        },
    });

    const getSelfRoles = useQuery(['getSelf'], ({ signal }) => getSelf({ signal }), {
        cacheTime: Number.POSITIVE_INFINITY,
        select: (data) => {
            const userRoles = data?.roles.map((role: any) => role.name) || [];
            return userRoles;
        },
    });

    return { getSelfId, getSelfRoles };
};
