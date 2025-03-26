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

import { useEffect, useMemo, useState } from 'react';
import { SearchValue } from '../../store';
import { useExploreParams } from '../useExploreParams';
import { getKeywordAndTypeValues, useSearch } from '../useSearch';

/* Reusable logic for syncing up a single node search field with browser query params on the Explore page. The value of the search field is tracked
internally, and is only pushed to query params once the event handler is called by the consumer component. Direct changes to the associated query
params will be synced back to the search field. */
export const useNodeSearch = () => {
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [selectedItem, setSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, searchType, setExploreParams } = useExploreParams();

    // Wire up search query. we should only recompute the keyword/type when the param value changes
    const { keyword, type } = useMemo(() => getKeywordAndTypeValues(primarySearch ?? undefined), [primarySearch]);
    const { data: searchData } = useSearch(keyword, type);

    // Watch query params for a new incoming node search and sync to internal state
    useEffect(() => {
        if (primarySearch && searchData) {
            const matchedNode = searchData.find((node) => node.objectid === primarySearch);

            if (matchedNode) {
                setSearchTerm(matchedNode.name);
                setSelectedItem(matchedNode);
            }
        } else {
            setSearchTerm('');
            setSelectedItem(undefined);
        }
    }, [primarySearch, searchType, searchData]);

    // Handles syncing the local search state up to query params to trigger a graph query
    const selectSourceNode = (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        // These will get set again by our param sync, but this lets the user see the selection instantly in the search field
        setSelectedItem(selected);
        setSearchTerm(term);

        setExploreParams({
            searchType: 'node',
            primarySearch: objectId,
        });
    };

    // Handle changes internal to the search form that should not trigger a graph query
    const editSourceNode = (edit: string) => {
        setSelectedItem(undefined);
        setSearchTerm(edit);
    };

    return {
        searchTerm,
        selectedItem,
        editSourceNode,
        selectSourceNode,
    };
};
