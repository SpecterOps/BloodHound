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

import { RequestOptions } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from 'bh-shared-ui';

export type CollectorType = 'sharphound' | 'azurehound';

export const COLLECTOR_TYPE: Record<CollectorType, string> = {
    sharphound: 'SharpHound',
    azurehound: 'AzureHound',
};

export const collectorKeys = {
    all: ['collectors'] as const,
    listByType: (type: CollectorType) => [...collectorKeys.all, type] as const,
    detail: (userId: number) => [...collectorKeys.all, userId] as const,
};

export const getCollectorsByType = (type: CollectorType, options?: RequestOptions) =>
    apiClient.getCollectors(type, options).then((res) => res.data);

export const useGetCollectorsByType = (type: CollectorType) =>
    useQuery(collectorKeys.listByType(type), ({ signal }) => getCollectorsByType(type, { signal }));
