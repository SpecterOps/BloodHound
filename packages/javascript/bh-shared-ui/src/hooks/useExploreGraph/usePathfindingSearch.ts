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

export const usePathfindingSearch = () => {
    const [sourceSearchTerm, setSourceSearchTerm] = useState<string>('');
    const [sourceSelectedItem, setSourceSelectedItem] = useState<SearchValue | undefined>(undefined);
    const [destinationSearchTerm, setDestinationSearchTerm] = useState<string>('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

    // Wire up search queries. we should only recompute keywords when the param values change
    const { keyword: sourceKeyword, type: sourceType } = useMemo(
        () => getKeywordAndTypeValues(primarySearch ?? undefined),
        [primarySearch]
    );
    const { keyword: destinationKeyword, type: destinationType } = useMemo(
        () => getKeywordAndTypeValues(secondarySearch ?? undefined),
        [secondarySearch]
    );
    const { data: sourceSearchData } = useSearch(sourceKeyword, sourceType);
    const { data: destinationSearchData } = useSearch(destinationKeyword, destinationType);

    // Watch query params and seperately sync them to each search field
    useEffect(() => {
        if (primarySearch && sourceSearchData) {
            const matchedNode = Object.values(sourceSearchData).find((node) => node.objectid === primarySearch);

            if (matchedNode) {
                setSourceSearchTerm(matchedNode.name);
                setSourceSelectedItem(matchedNode);
            }
        }
    }, [primarySearch, sourceSearchData]);

    useEffect(() => {
        if (secondarySearch && destinationSearchData) {
            const matchedNode = Object.values(destinationSearchData).find((node) => node.objectid === secondarySearch);

            if (matchedNode) {
                setDestinationSearchTerm(matchedNode.name);
                setDestinationSelectedItem(matchedNode);
            }
        }
    }, [secondarySearch, destinationSearchData]);

    // Handle syncing each search field up to query params to trigger a graph query. Should trigger pathfinding if both have been selected and node
    // if only one has been selected
    const handleSourceNodeSelected = (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        setSourceSelectedItem(selected);
        setSourceSearchTerm(term);

        // if i have the other node, set type to 'pathfinding' and search term to the objectid
        // if not, set type to 'node' and clear out the opposing search term
        if (secondarySearch && destinationSelectedItem) {
            setExploreParams({
                searchType: 'pathfinding',
                primarySearch: objectId,
            });
        } else {
            setExploreParams({
                searchType: 'node',
                primarySearch: objectId,
                secondarySearch: null,
            });
        }
    };

    const handleDestinationNodeSelected = (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        setDestinationSelectedItem(selected);
        setDestinationSearchTerm(term);

        if (primarySearch && sourceSelectedItem) {
            setExploreParams({
                searchType: 'pathfinding',
                secondarySearch: objectId,
            });
        } else {
            setExploreParams({
                searchType: 'node',
                secondarySearch: objectId,
                primarySearch: null,
            });
        }
    };

    const handleSwapPathfindingInputs = () => {
        if (sourceSelectedItem && destinationSelectedItem) {
            console.log('setting params: ', destinationSelectedItem, sourceSelectedItem);
            setExploreParams({
                searchType: 'pathfinding',
                primarySearch: destinationSelectedItem.objectid,
                secondarySearch: sourceSelectedItem.objectid,
            });
        }
    };

    // Handle changes internal to the search form that should not trigger a graph query. Each param should sync independently
    const handleSourceNodeEdited = (edit: string) => {
        setSourceSelectedItem(undefined);
        setSourceSearchTerm(edit);
    };

    const handleDestinationNodeEdited = (edit: string) => {
        setDestinationSelectedItem(undefined);
        setDestinationSearchTerm(edit);
    };

    return {
        sourceSearchTerm,
        sourceSelectedItem,
        destinationSearchTerm,
        destinationSelectedItem,
        handleSourceNodeEdited,
        handleSourceNodeSelected,
        handleDestinationNodeEdited,
        handleDestinationNodeSelected,
        handleSwapPathfindingInputs,
    };
};
