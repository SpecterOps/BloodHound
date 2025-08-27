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
import { useQuery } from 'react-query';
import { CommonSearches as prebuiltSearchListAGI } from '../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../commonSearchesAGT';
import { useFeatureFlag, useSavedQueries } from '../../hooks';
import { QueryLineItem } from '../../types';
import { apiClient } from '../../utils';

export const usePrebuiltQueries = () => {
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const userQueries = useSavedQueries();
    const getSelf = useQuery(['getSelf'], ({ signal }) => apiClient.getSelf({ signal }).then((res) => res.data.data));

    //Get master list of queries to validate against
    const savedLineItems: QueryLineItem[] =
        userQueries.data?.map((query) => ({
            name: query.name,
            description: query.description,
            query: query.query,
            canEdit: query.user_id === getSelf.data?.id,
            id: query.id,
            user_id: query.user_id,
        })) || [];

    const savedQueries = {
        category: 'Saved Queries',
        subheader: '',
        queries: savedLineItems,
    };
    const queryList = tierFlag?.enabled
        ? [...prebuiltSearchListAGT, savedQueries]
        : [...prebuiltSearchListAGI, savedQueries];
    return queryList;
};

export const useGetSelectedQuery = (cypherQuery: string, id?: number) => {
    const queryList = usePrebuiltQueries();
    for (const item of queryList) {
        let result = null;
        result = item.queries.find((query) => {
            if (id && query.id === id) {
                return query;
            } else if (query.query === cypherQuery) {
                return query;
            }
        });
        if (result) {
            return result;
        }
    }
};
