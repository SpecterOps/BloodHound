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
import { QueryScope, SavedQuery } from 'js-client-library';
import { useMemo } from 'react';
import { CommonSearches as prebuiltSearchListAGI } from '../../commonSearchesAGI';
import { CommonSearches as prebuiltSearchListAGT } from '../../commonSearchesAGT';
import { useFeatureFlag } from '../../hooks/useFeatureFlags';
import { useSavedQueries } from '../../hooks/useSavedQueries';
import { QueryLineItem } from '../../types';
import { useSelf } from '../useSelf';

export const usePrebuiltQueries = () => {
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const { getSelfId } = useSelf();
    const { data: selfId } = getSelfId;

    const queryDataMapper = (data: SavedQuery[]) => {
        return (
            data?.map((query: SavedQuery) => ({
                name: query.name,
                description: query.description,
                query: query.query,
                canEdit: query.user_id === selfId,
                id: query.id,
                user_id: query.user_id,
            })) || []
        );
    };

    const userQueries = useSavedQueries(QueryScope.ALL, {
        select: queryDataMapper,
    });

    const savedQueries = {
        category: 'Saved Queries',
        subheader: '',
        queries: userQueries.data || [],
    };

    const queryList = tierFlag?.enabled
        ? [...prebuiltSearchListAGT, savedQueries]
        : [...prebuiltSearchListAGI, savedQueries];

    return queryList;
};

export const useGetSelectedQuery = (cypherQuery: string, id?: number) => {
    const groups = usePrebuiltQueries();

    const selected = useMemo<QueryLineItem | undefined>(() => {
        const queryList: QueryLineItem[] = groups.flatMap((g) => g.queries ?? []);

        // Prefer direct id match if provided
        if (id != undefined) {
            const byId = queryList.find((q) => q.id === id);
            if (byId) return byId;
        }

        // Fallback: match by cypher string (could be multiple “Save As” copies)
        const matches = queryList.filter((q) => q.query === cypherQuery);
        if (matches.length === 0) return undefined;
        if (matches.length === 1) return matches[0];

        // If multiples, prefer the user-saved (has an id) over hardcoded
        return matches.find((q) => q.id != undefined) ?? matches[0];
    }, [groups, id, cypherQuery]);

    return selected;
};
