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
import { getNodeByDatabaseIdCypher } from '../../utils/entityInfoDisplay';
import { validateNodeType } from '../useSearch/useSearch';

export type FetchEntityInfoParams = {
    objectId: string;
    nodeType: string;
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
    const requestDetails: {
        endpoint: (
            params: string,
            options?: RequestOptions,
            includeProperties?: boolean
        ) => Promise<Record<string, any>>;
        param: string;
    } = {
        endpoint: async function () {
            return {};
        },
        param: '',
    };

    const validatedKind = validateNodeType(nodeType);

    if (validatedKind) {
        requestDetails.endpoint = entityInformationEndpoints[validatedKind];
        requestDetails.param = objectId;
    } else if (databaseId) {
        requestDetails.endpoint = apiClient.cypherSearch;
        requestDetails.param = getNodeByDatabaseIdCypher(databaseId);
    }

    const informationAvailable = !!validatedKind || !!databaseId;

    return {
        ...useQuery(
            [FetchEntityCacheId, nodeType, objectId, databaseId],
            ({ signal }) =>
                requestDetails.endpoint(requestDetails.param, { signal }, true).then((res) => {
                    if (validatedKind) {
                        const kinds = res.data.data.kinds;
                        const properties = res.data.data.props;
                        return { kinds, properties };
                    } else if (databaseId) {
                        const data = Object.values(res.data.data.nodes as Record<string, any>)[0];
                        const kinds = data.kinds;
                        const properties = data.properties;
                        return { kinds, properties };
                    } else return { kinds: [], properties: {} };
                }),
            {
                refetchOnWindowFocus: false,
                retry: false,
                enabled: informationAvailable,
            }
        ),
        informationAvailable,
    };
};
