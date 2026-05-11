// Copyright 2026 Specter Ops, Inc.
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

import { FlatGraphResponse, GraphResponse } from 'js-client-library';
import { useCallback } from 'react';
import { useExploreParams } from './useExploreParams';
import { useExploreSelectedItem } from './useExploreSelectedItem';

type NodeRecord = Record<string, { objectId: string }>;

export const useNodeAutoSelect = (
    extractNodes: (data: GraphResponse | FlatGraphResponse) => NodeRecord | undefined
) => {
    const { searchType, primarySearch } = useExploreParams();
    const { selectedItem, setSelectedItem, clearSelectedItem } = useExploreSelectedItem();

    return useCallback(
        (data: GraphResponse | FlatGraphResponse) => {
            if (searchType === 'pathfinding') {
                clearSelectedItem();
                return;
            }

            if (searchType !== 'node' || !primarySearch) return;

            const nodes = extractNodes(data);
            if (!nodes) return;

            if (selectedItem && nodes[selectedItem]?.objectId === primarySearch) return;

            const matchedEntry = Object.entries(nodes).find(([, node]) => node.objectId === primarySearch);
            if (matchedEntry) setSelectedItem(matchedEntry[0]);
        },
        [searchType, primarySearch, selectedItem, setSelectedItem, clearSelectedItem, extractNodes]
    );
};
