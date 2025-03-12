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
import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { EdgeCheckboxType, MappedStringLiteral } from '../..';
import { EntityInfoDataTableProps } from '../../utils/content';
import { setParamsFactory } from '../../utils/searchParams/searchParams';

export type ExploreSearchTab = 'node' | 'pathfinding' | 'cypher';
type SearchType = ExploreSearchTab | 'relationship' | 'composition';

export type ExploreQueryParams = {
    exploreSearchTab: ExploreSearchTab | null;
    primarySearch: string | null;
    secondarySearch: string | null;
    cypherSearch: string | null;
    searchType: SearchType | null;
    graphSelection: string | null;
    panelSelection: string | null;
    expandedRelationships: EntityInfoDataTableProps['label'][] | null;
    pathFilters: EdgeCheckboxType['edgeType'][] | null;
};

export const acceptedExploreSearchTabs = {
    node: 'node',
    pathfinding: 'pathfinding',
    cypher: 'cypher',
} satisfies MappedStringLiteral<ExploreSearchTab, ExploreSearchTab>;

export const parseSearchTab = (paramValue: string | null): ExploreSearchTab | null => {
    if (paramValue && paramValue in acceptedExploreSearchTabs) {
        return paramValue as ExploreSearchTab;
    }
    return null;
};

export const acceptedSearchTypes = {
    ...acceptedExploreSearchTabs,
    relationship: 'relationship',
    composition: 'composition',
} satisfies MappedStringLiteral<SearchType, SearchType>;

export const parseSearchType = (paramValue: string | null): SearchType | null => {
    if (paramValue && paramValue in acceptedSearchTypes) {
        return paramValue as SearchType;
    }
    return null;
};

interface UseExploreParamsReturn extends ExploreQueryParams {
    setExploreParams: (params: Partial<ExploreQueryParams>) => void;
}

export const useExploreParams = (): UseExploreParamsReturn => {
    const [searchParams, setSearchParams] = useSearchParams();

    return {
        exploreSearchTab: parseSearchTab(searchParams.get('exploreSearchTab')),
        primarySearch: searchParams.get('primarySearch'),
        secondarySearch: searchParams.get('secondarySearch'),
        cypherSearch: searchParams.get('cypherSearch'),
        searchType: parseSearchType(searchParams.get('searchType')),
        graphSelection: searchParams.get('graphSelection'),
        panelSelection: searchParams.get('panelSelection'),
        expandedRelationships: searchParams.getAll('expandedRelationship'),
        pathFilters: searchParams.getAll('pathFilters'),
        // react doesnt like this because it doesnt know the params needed for the function factory return function.
        // but the params needed are not needed in the deps array
        // eslint-disable-next-line react-hooks/exhaustive-deps
        setExploreParams: useCallback(
            setParamsFactory(setSearchParams, [
                'exploreSearchTab',
                'primarySearch',
                'secondarySearch',
                'cypherSearch',
                'searchType',
                'graphSelection',
                'panelSelection',
                'expandedRelationships',
                'pathFilters',
            ]),
            [setSearchParams]
        ),
    };
};
