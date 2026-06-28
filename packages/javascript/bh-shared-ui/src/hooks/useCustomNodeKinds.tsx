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
import { CustomNodeKindType, RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import {
    apiClient,
    DEFAULT_ICON_BACKGROUND_COLOR,
    GenericQueryOptions,
    IconDictionary,
    RACF_NODE_ICONS,
} from '../utils';

const isStubbedCustomNodeKind = (node: CustomNodeKindType) =>
    node.config.icon.name === 'question' && node.config.icon.color.toUpperCase() === '#FFFFFF';

export const createCustomIconDictionary = (nodes: CustomNodeKindType[] | undefined): IconDictionary => {
    const customIcons: IconDictionary = { ...RACF_NODE_ICONS };

    nodes?.forEach((node) => {
        if (node.kindName in RACF_NODE_ICONS && isStubbedCustomNodeKind(node)) {
            return;
        }

        const iconName = node.config.icon.name as IconName;
        const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: iconName });
        if (iconDefinition == undefined) {
            return;
        }

        customIcons[node.kindName] = {
            icon: iconDefinition,
            color: node.config.icon.color ? node.config.icon.color : DEFAULT_ICON_BACKGROUND_COLOR,
        };
    });

    return customIcons;
};

export const getCustomNodeKinds = async (options: RequestOptions): Promise<IconDictionary> =>
    apiClient.getCustomNodeKinds(options).then((res) => {
        return createCustomIconDictionary(Array.isArray(res?.data?.data) ? res.data.data : undefined);
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
    });
};
