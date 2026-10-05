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
import { useExtensionsQuery } from '../../hooks/useExtensions';
import { useFeatureFlag } from '../../hooks/useFeatureFlags';
import { useSavedQueries } from '../../hooks/useSavedQueries';
import { QueryLineItem, QueryListSection } from '../../types';
import { useSelf } from '../useSelf';

export const usePrebuiltQueries = () => {
    const { data: tierFlag } = useFeatureFlag('tier_management_engine');
    const { getSelfId } = useSelf();
    const { data: selfId } = getSelfId;
    const { data: extensions } = useExtensionsQuery();
    const userQueries = useSavedQueries(QueryScope.ALL);

    const savedQuerySections = useMemo<QueryListSection[]>(() => {
        const queries = (userQueries.data || []).map((query: SavedQuery) => ({
            name: query.name,
            description: query.description,
            query: query.query,
            canEdit: query.extension_id == null && query.user_id === selfId,
            id: query.id,
            user_id: query.user_id,
            category: query.category,
            schema_extension_id: query.extension_id,
        }));
        const extensionsById = new Map(
            extensions?.map((extension) => [Number(extension.id), extension.display_name || extension.name])
        );
        const sections = new Map<string, QueryListSection>();

        for (const query of queries) {
            const platform =
                query.schema_extension_id == null
                    ? 'Saved Queries'
                    : extensionsById.get(query.schema_extension_id) || 'Extension';
            const category = query.category?.trim() || 'Uncategorized';
            const key = JSON.stringify([platform, category]);
            if (!sections.has(key)) {
                sections.set(key, { category: platform, subheader: category, queries: [] });
            }
            sections.get(key)?.queries.push(query);
        }

        return [...sections.values()];
    }, [userQueries.data, extensions, selfId]);

    const queryList = tierFlag?.enabled
        ? [...prebuiltSearchListAGT, ...savedQuerySections]
        : [...prebuiltSearchListAGI, ...savedQuerySections];

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
