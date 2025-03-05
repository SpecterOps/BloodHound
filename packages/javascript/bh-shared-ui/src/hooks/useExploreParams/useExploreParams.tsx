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
import { MappedStringLiteral } from '../..';
import { EntityInfoDataTableProps } from '../../utils/content';
import { setParamsFactory } from '../../utils/searchParams/searchParams';

type SearchTab = 'node' | 'pathfinding' | 'cypher';
type SearchType = SearchTab | 'relationship' | 'composition';

export type ExploreQueryParams = {
    searchTab: SearchTab | null;
    primarySearch: string | null;
    secondarySearch: string | null;
    cypherSearch: string | null;
    searchType: SearchType | null;
    graphSelection: string | null;
    panelSelection: string | null;
    expandedRelationships: EntityInfoDataTableProps['label'][] | null;
};

const acceptedSearchTabs = {
    node: 'node',
    pathfinding: 'pathfinding',
    cypher: 'cypher',
} satisfies MappedStringLiteral<SearchTab, SearchTab>;

export const parseSearchTab = (paramValue: string | null): SearchTab | null => {
    if (paramValue && paramValue in acceptedSearchTabs) {
        return paramValue as SearchTab;
    }
    return null;
};

export const acceptedSearchTypes = {
    ...acceptedSearchTabs,
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
        searchTab: parseSearchTab(searchParams.get('searchTab')),
        primarySearch: searchParams.get('primarySearch'),
        secondarySearch: searchParams.get('secondarySearch'),
        cypherSearch: searchParams.get('cypherSearch'),
        searchType: parseSearchType(searchParams.get('searchType')),
        graphSelection: searchParams.get('graphSelection'),
        panelSelection: searchParams.get('panelSelection'),
        expandedRelationships: searchParams.getAll('expandedRelationship'),
        // react doesnt like this because it doesnt know the params needed for the function factory return function.
        // but the params needed are not needed in the deps array
        // eslint-disable-next-line react-hooks/exhaustive-deps
        setExploreParams: useCallback(
            setParamsFactory(setSearchParams, [
                'searchTab',
                'primarySearch',
                'secondarySearch',
                'cypherSearch',
                'searchType',
                'graphSelection',
                'panelSelection',
                'expandedRelationships',
            ]),
            [setSearchParams]
        ),
    };
};
