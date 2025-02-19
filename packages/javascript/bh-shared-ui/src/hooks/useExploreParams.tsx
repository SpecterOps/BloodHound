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

import { useSearchParams } from 'react-router-dom';
import { setParamsFactory } from '../utils/searchParams';

interface ExploreQuery {
    primarySearch: string | null;
    secondarySearch: string | null;
    cypherSearch: string | null;
    searchType: 'node' | 'pathfinding' | 'cypher' | 'relationship' | 'composition' | null;
    graphSelection: string | null;
    panelSelection: string | null;
    expandedRelationship: string[] | null;
}

export type ExploreQueryParams = Partial<ExploreQuery>;

const parseSearchQuery = (param: string | null) => {
    if (
        param === 'node' ||
        param === 'pathfinding' ||
        param === 'cypher' ||
        param === 'relationship' ||
        param === 'composition'
    ) {
        return param;
    }
    return null;
};

interface useExploreParamsReturn extends ExploreQueryParams {
    setExploreParams: (params: ExploreQueryParams) => void;
}

export const useExploreParams = (): useExploreParamsReturn => {
    const [searchParams, setSearchParams] = useSearchParams();

    const primarySearch = searchParams.get('primarySearch');
    const secondarySearch = searchParams.get('secondarySearch');
    const cypherSearch = searchParams.get('cypherSearch');
    const searchType = parseSearchQuery(searchParams.get('searchType'));
    const graphSelection = searchParams.get('graphSelection');
    const panelSelection = searchParams.get('panelSelection');
    const expandedRelationship = searchParams.getAll('expandedRelationship');

    const setExploreParams = setParamsFactory<ExploreQueryParams>(setSearchParams, [
        'primarySearch',
        'secondarySearch',
        'cypherSearch',
        'searchType',
        'graphSelection',
        'panelSelection',
        'expandedRelationship',
    ]);

    return {
        primarySearch,
        secondarySearch,
        cypherSearch,
        searchType,
        graphSelection,
        panelSelection,
        expandedRelationship,
        setExploreParams,
    };
};
