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

import { EdgeType } from 'js-client-library';
import { useQuery } from 'react-query';
import { apiClient } from '../utils/api';
import { useFeatureFlag } from './useFeatureFlags';

export const edgeTypesKeys = {
    all: ['getEdgeTypes'] as const,
};

// Edge types are only served when OpenGraph extension management is enabled.
export const useEdgeTypesQuery = <T = EdgeType[]>(select?: (data: EdgeType[]) => T) => {
    const { data: openGraphFeatureFlag } = useFeatureFlag('opengraph_extension_management');
    return useQuery({
        queryKey: edgeTypesKeys.all,
        queryFn: ({ signal }) => apiClient.getEdgeTypes({ signal }).then((response) => response.data.data),
        select,
        enabled: !!openGraphFeatureFlag?.enabled,
        retry: false,
    });
};
