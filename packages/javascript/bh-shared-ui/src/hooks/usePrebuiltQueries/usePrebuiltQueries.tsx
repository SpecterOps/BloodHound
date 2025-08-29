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
    const matchedResults: any[] = [];
    //first we check for a matched id
    for (const item of queryList) {
        let result = null;
        result = item.queries.find((query) => {
            if (id && query.id === id) {
                return query;
            }
        });
        if (result) {
            return result;
        }
    }

    //next we check for matched query string
    //setting an array in case of duplicated queries via save as
    for (const item of queryList) {
        item.queries.find((query) => {
            if (query.query === cypherQuery) {
                matchedResults.push(query);
            }
        });
    }

    //prefer user saved version over hardcoded version
    //this is useful on initial load or refresh where we dont have the specific query id to compare against
    if (!matchedResults) return null;
    if (matchedResults.length === 1) {
        return matchedResults[0];
    } else if (matchedResults.length > 1) {
        const resultWithId = matchedResults.find((query) => {
            if (query.id !== undefined) {
                return query;
            }
        });
        if (resultWithId) {
            return resultWithId;
        }
    }
};
