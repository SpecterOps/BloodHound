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

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { SearchValue } from '../../views/Explore/ExploreSearch/types';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';
import { SearchResults, useKeywordAndTypeValues, useSearch } from '../useSearch';

const MAX_NODES = 4;

const SEARCH_PARAM_KEYS = ['primarySearch', 'secondarySearch', 'tertiarySearch', 'quaternarySearch'] as const;

export type PathfindingNode = {
    searchTerm: string;
    selectedItem: SearchValue | undefined;
};

export type PathfindingSearchState = {
    sourceSearchTerm: string;
    destinationSearchTerm: string;
    sourceSelectedItem: SearchValue | undefined;
    destinationSelectedItem: SearchValue | undefined;
    nodes: PathfindingNode[];
    totalNodeCount: number;
    maxNodes: number;
    handleSourceNodeEdited: (edit: string) => void;
    handleDestinationNodeEdited: (edit: string) => void;
    handleSourceNodeSelected: (selected: SearchValue) => void;
    handleDestinationNodeSelected: (selected: SearchValue) => void;
    handleNodeEdited: (index: number) => (edit: string) => void;
    handleNodeSelected: (index: number) => (selected: SearchValue) => void;
    handleSwapPathfindingInputs: () => void;
    handleReorderNodes: (fromIndex: number, toIndex: number) => void;
    handleRemoveNode: (index: number) => void;
    handleAddNode: () => void;
};

const emptyNode = (): PathfindingNode => ({ searchTerm: '', selectedItem: undefined });

export const usePathfindingSearch = () => {
    const [nodes, setNodes] = useState<PathfindingNode[]>([emptyNode(), emptyNode()]);
    const [extraNodeCount, setExtraNodeCount] = useState(0);

    const { primarySearch, secondarySearch, tertiarySearch, quaternarySearch, setExploreParams } = useExploreParams();

    // Wire up search queries for each param
    const { keyword: kw0, type: t0 } = useKeywordAndTypeValues(primarySearch);
    const { keyword: kw1, type: t1 } = useKeywordAndTypeValues(secondarySearch);
    const { keyword: kw2, type: t2 } = useKeywordAndTypeValues(tertiarySearch);
    const { keyword: kw3, type: t3 } = useKeywordAndTypeValues(quaternarySearch);

    const { data: data0 } = useSearch(kw0, t0);
    const { data: data1 } = useSearch(kw1, t1);
    const { data: data2 } = useSearch(kw2, t2);
    const { data: data3 } = useSearch(kw3, t3);

    const updateNode = useCallback((index: number, updatedNode: Partial<PathfindingNode>) => {
        setNodes((previousNodes) => {
            const newNodes = [...previousNodes];
            // Pad the array with empty nodes so the target index exists
            while (newNodes.length <= index) newNodes.push(emptyNode());
            // Merge the update into the node at the target index
            newNodes[index] = { ...newNodes[index], ...updatedNode };
            return newNodes;
        });
    }, []);

    const syncNodeFromParam = useCallback(
        (index: number, param: string | null, data: SearchResults | undefined) => {
            if (param && data) {
                const matchedNode = data.find((node) => node.objectid === param);
                if (matchedNode) {
                    updateNode(index, { searchTerm: matchedNode.name, selectedItem: matchedNode });
                }
            } else if (!param) {
                updateNode(index, emptyNode());
            }
        },
        [updateNode]
    );

    const params = useMemo(
        () => [primarySearch, secondarySearch, tertiarySearch, quaternarySearch],
        [primarySearch, secondarySearch, tertiarySearch, quaternarySearch]
    );
    const searchData = useMemo(() => [data0, data1, data2, data3], [data0, data1, data2, data3]);

    // Sync URL params to node state. Each node is synced only when its own param or search
    // data changes — re-syncing every node whenever any one of them changes would clear a
    // sibling's typed-but-not-yet-selected search term.
    const lastSynced = useRef<{ param: string | null; data: SearchResults | undefined }[]>([]);

    useEffect(() => {
        params.forEach((param, index) => {
            const previous = lastSynced.current[index];
            if (previous && previous.param === param && previous.data === searchData[index]) return;

            lastSynced.current[index] = { param, data: searchData[index] };
            syncNodeFromParam(index, param, searchData[index]);
        });
    }, [syncNodeFromParam, params, searchData]);

    // Keep extraNodeCount in sync with URL params: if quaternary then 2, if not and tertiary
    // then 1, if neither then 0. Counting the set params instead would collapse a node when a
    // hand-edited URL sets quaternary without tertiary, hiding an input the query still uses.
    useEffect(() => {
        const extraCount = quaternarySearch ? 2 : tertiarySearch ? 1 : 0;
        setExtraNodeCount(extraCount);
    }, [tertiarySearch, quaternarySearch]);

    const totalNodeCount = 2 + extraNodeCount;

    const getParamsFromNodes = (nodeList: PathfindingNode[]): Partial<ExploreQueryParams> => {
        const params: Partial<ExploreQueryParams> = {};
        SEARCH_PARAM_KEYS.forEach((key, i) => {
            params[key] = nodeList[i]?.selectedItem?.objectid ?? null;
        });
        return params;
    };

    const triggerPathfinding = (params: Partial<ExploreQueryParams>) => {
        const merged = { ...getParamsFromNodes(nodes), ...params };
        const source = merged.primarySearch;
        const dest = merged.secondarySearch;

        if (source && dest) {
            setExploreParams({ searchType: 'pathfinding', ...params });
        } else if (source || dest) {
            setExploreParams({ searchType: 'node', ...params });
        }
    };

    // Handle node selection — triggers query
    const handleNodeSelected = (index: number) => (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        updateNode(index, { searchTerm: term, selectedItem: selected });

        const paramKey = SEARCH_PARAM_KEYS[index];

        if (index === 0) {
            // Source node
            if (secondarySearch && nodes[1]?.selectedItem) {
                setExploreParams({ searchType: 'pathfinding', [paramKey]: objectId });
            } else {
                setExploreParams({ searchType: 'node', [paramKey]: objectId, secondarySearch: null });
            }
        } else if (index === 1) {
            // First destination
            if (primarySearch && nodes[0]?.selectedItem) {
                setExploreParams({ searchType: 'pathfinding', [paramKey]: objectId });
            } else {
                setExploreParams({ searchType: 'node', [paramKey]: objectId, primarySearch: null });
            }
        } else {
            // Extra destinations
            triggerPathfinding({ [paramKey]: objectId });
        }
    };

    // Handle text edit — does not trigger query
    const handleNodeEdited = (index: number) => (edit: string) => {
        updateNode(index, { searchTerm: edit, selectedItem: undefined });
    };

    const handleSwapPathfindingInputs = () => {
        if (nodes[0]?.selectedItem && nodes[1]?.selectedItem) {
            setExploreParams({
                searchType: 'pathfinding',
                primarySearch: nodes[1].selectedItem.objectid,
                secondarySearch: nodes[0].selectedItem.objectid,
            });
        }
    };

    const handleReorderNodes = (fromIndex: number, toIndex: number) => {
        const currentNodes = nodes.slice(0, totalNodeCount);
        const [movedNode] = currentNodes.splice(fromIndex, 1);
        currentNodes.splice(toIndex, 0, movedNode);

        setNodes(currentNodes);

        const params = getParamsFromNodes(currentNodes);
        triggerPathfinding(params);
    };

    const handleAddNode = () => {
        if (totalNodeCount < MAX_NODES) {
            setExtraNodeCount((previousCount) => previousCount + 1);
            setNodes((previousNodes) => {
                const newNodes = [...previousNodes];
                const targetNodeCount = totalNodeCount + 1;
                // Pad the array so the newly added destination has a node to render
                while (newNodes.length < targetNodeCount) newNodes.push(emptyNode());
                return newNodes;
            });
        }
    };

    const handleRemoveNode = (index: number) => {
        if (index === 0 || totalNodeCount <= 2) return;

        const currentNodes = nodes.slice(0, totalNodeCount);
        currentNodes.splice(index, 1);
        setExtraNodeCount((prev) => prev - 1);

        // Pad back to at least 2
        while (currentNodes.length < 2) currentNodes.push(emptyNode());
        setNodes(currentNodes);

        // Update URL params
        const params = getParamsFromNodes(currentNodes);
        // Clear any now-unused param
        for (let i = currentNodes.length; i < MAX_NODES; i++) {
            params[SEARCH_PARAM_KEYS[i]] = null;
        }
        triggerPathfinding(params);
    };

    // Build the return shape that PathfindingSearch.tsx expects
    const sourceSearchTerm = nodes[0]?.searchTerm ?? '';
    const sourceSelectedItem = nodes[0]?.selectedItem;
    const destinationSearchTerm = nodes[1]?.searchTerm ?? '';
    const destinationSelectedItem = nodes[1]?.selectedItem;

    return {
        sourceSearchTerm,
        sourceSelectedItem,
        destinationSearchTerm,
        destinationSelectedItem,
        nodes,
        totalNodeCount,
        maxNodes: MAX_NODES,
        handleSourceNodeEdited: handleNodeEdited(0),
        handleSourceNodeSelected: handleNodeSelected(0),
        handleDestinationNodeEdited: handleNodeEdited(1),
        handleDestinationNodeSelected: handleNodeSelected(1),
        handleNodeEdited,
        handleNodeSelected,
        handleSwapPathfindingInputs,
        handleReorderNodes,
        handleRemoveNode,
        handleAddNode,
    };
};
