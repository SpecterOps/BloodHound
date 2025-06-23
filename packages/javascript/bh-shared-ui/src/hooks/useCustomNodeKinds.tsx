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

import { findIconDefinition, IconName } from '@fortawesome/fontawesome-svg-core';
import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { apiClient, DEFAULT_ICON_BACKGROUND, GenericQueryOptions, IconDictionary } from '../utils';

export const getCustomNodeKinds = async (options: RequestOptions): Promise<IconDictionary> =>
    apiClient.getCustomNodeKinds(options).then((res) => {
        const customIcons: IconDictionary = {};

        if (Array.isArray(res?.data?.data)) {
            res.data.data.forEach((node) => {
                const iconName = node.config.icon.name as IconName;

                const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: iconName });
                if (iconDefinition == undefined) {
                    return;
                }

                customIcons[node.kindName] = {
                    icon: iconDefinition,
                    color: node.config.icon.color ? node.config.icon.color : DEFAULT_ICON_BACKGROUND,
                };
            });
        }

        return customIcons;
    });

export const useCustomNodeKinds = (
    queryOptions?: GenericQueryOptions<IconDictionary>
): UseQueryResult<IconDictionary> => {
    return useQuery({
        queryKey: ['getCustomNodeKinds'],
        queryFn: ({ signal }) => getCustomNodeKinds({ signal }),
        staleTime: 2 * (60 * 1000),
        cacheTime: 5 * (60 * 1000),
        ...queryOptions,
        // TODO: remove this! Custom-nodes call kept failing and causing rerenders
        enabled: false,
    });
};
