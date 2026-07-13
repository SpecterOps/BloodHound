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

import { RequestOptions } from 'js-client-library';
import { useQuery, UseQueryResult } from 'react-query';
import { apiClient } from '../../utils/api';
import { entityInformationEndpoints } from '../../utils/content';
import { validateNodeType } from './utils';

export type FetchEntityInfoParams = {
    objectId?: string;
    nodeType?: string;
    databaseId?: string;
};

export type EntityInfo = UseQueryResult<{
    kinds: string[];
    properties: {
        [k: string]: any;
        objectid: string;
    };
}>;

export type FetchEntityInfoResult = EntityInfo & {
    informationAvailable: boolean;
};

export const FetchEntityCacheId = 'entity-properties' as const;

export const useFetchEntityInfo: (param: FetchEntityInfoParams) => FetchEntityInfoResult = ({
    objectId,
    nodeType,
    databaseId,
}) => {
    const validatedKind = nodeType ? validateNodeType(nodeType) : undefined;

    const informationAvailable = !!databaseId || !!validatedKind;

    return {
        ...useQuery(
            [FetchEntityCacheId, nodeType, objectId, databaseId],
            ({ signal }) => {
                if (validatedKind && objectId) {
                    const endpoint = entityInformationEndpoints[validatedKind] as (
                        params: string,
                        options?: RequestOptions,
                        includeProperties?: boolean
                    ) => Promise<Record<string, any>>;
                    return endpoint(objectId, { signal }, true).then((res) => {
                        const kinds = res.data.data.kinds;
                        const properties = res.data.data.props;
                        return { kinds, properties };
                    });
                } else if (databaseId) {
                    return apiClient.getNodeByID(Number(databaseId), { signal }).then((res) => {
                        const kinds = res.data.data.kinds.map((kind) => kind.name);
                        const properties = res.data.data.properties;
                        return { kinds, properties };
                    });
                }
                return Promise.resolve({ kinds: [], properties: {} });
            },
            {
                refetchOnWindowFocus: false,
                retry: false,
                enabled: informationAvailable,
            }
        ),
        informationAvailable,
    };
};
