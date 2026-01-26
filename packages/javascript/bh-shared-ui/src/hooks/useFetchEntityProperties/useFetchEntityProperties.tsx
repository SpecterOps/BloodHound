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
import { useQuery } from 'react-query';
import { GenericQueryOptions } from '../../utils';
import { apiClient } from '../../utils/api';
import { entityInformationEndpoints } from '../../utils/content';
import { getNodeByDatabaseIdCypher } from '../../utils/entityInfoDisplay';
import { useTagsQuery } from '../useAssetGroupTags';
import { validateNodeType } from '../useSearch/useSearch';

export type FetchEntityPropertiesParams = {
    objectId: string;
    nodeType: string;
    databaseId?: string;
};

export type EntityProperties = {
    [k: string]: any;
    objectid: string;
};

type FetchEntityPropertiesExport = {
    entityProperties: EntityProperties;
    informationAvailable: boolean;
    isLoading: boolean;
    isError: boolean;
    isSuccess: boolean;
};

type FetchEntityZone = {
    zoneName?: string;
    informationAvailable: boolean;
    isLoading: boolean;
    isError: boolean;
    isSuccess: boolean;
};

export const FetchEntityCacheId = 'entity-properties' as const;

const useSharedQuery = (
    requestDetails: {
        endpoint: (
            params: string,
            options?: RequestOptions,
            includeProperties?: boolean
        ) => Promise<Record<string, any>>;
        param: string;
    },
    params: FetchEntityPropertiesParams,
    informationAvailable: boolean,
    options?: GenericQueryOptions<any>
) => {
    return useQuery(
        [FetchEntityCacheId, params.nodeType, params.objectId],
        ({ signal }) =>
            requestDetails.endpoint(requestDetails.param, { signal }, true).then((res) => {
                return res.data;
            }),
        {
            refetchOnWindowFocus: false,
            retry: false,
            enabled: informationAvailable,
            ...(options ?? {}),
        }
    );
};

export const useFetchEntityProperties: (param: FetchEntityPropertiesParams) => FetchEntityPropertiesExport = ({
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
    const {
        data: entityProperties,
        isLoading,
        isError,
        isSuccess,
    } = useSharedQuery(requestDetails, { objectId, nodeType }, informationAvailable, {
        select: (data) => {
            if (validatedKind) return data.data.props;
            else if (databaseId) return Object.values(data.data.nodes as Record<string, any>)[0].properties;
            else return {};
        },
    });
    return {
        entityProperties,
        informationAvailable,
        isLoading,
        isError,
        isSuccess,
    };
};
export const useFetchEntityKind: (param: FetchEntityPropertiesParams) => FetchEntityZone = ({
    objectId,
    nodeType,
    databaseId,
}) => {
    //normalize value for string matching
    const normalize = (value: string) => value.replace(/^Tag_/, '').replace(/_/g, ' ').trim().toLowerCase();
    const tagsQuery = useTagsQuery();
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
    const { data, isLoading, isError, isSuccess } = useSharedQuery(
        requestDetails,
        { objectId, nodeType },
        informationAvailable
    );

    //build a Map to match tag name to kind
    const tagMap = new Map(
        tagsQuery.data?.map((tag) => {
            const normalized = normalize(tag.name);
            return [normalized, tag];
        })
    );

    //string matching for zone name
    const match = data?.data?.kinds?.find((kind: string) => tagMap.has(normalize(kind)));
    const zoneName = match ? tagMap.get(normalize(match))?.name : undefined;

    return {
        zoneName,
        informationAvailable,
        isLoading,
        isError,
        isSuccess,
    };
};
