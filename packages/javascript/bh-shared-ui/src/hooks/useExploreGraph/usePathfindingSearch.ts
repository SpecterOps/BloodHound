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

import { useEffect, useState } from 'react';
import { SearchValue } from '../../views/Explore/ExploreSearch/types';
import { ExploreQueryParams, useExploreParams } from '../useExploreParams';
import { useKeywordAndTypeValues, useSearch } from '../useSearch';

const MAX_NODES = 4;

const SEARCH_PARAM_KEYS = ['primarySearch', 'secondarySearch', 'tertiarySearch', 'quaternarySearch'] as const;

type PathfindingNode = {
    searchTerm: string;
    selectedItem: SearchValue | undefined;
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

    // Sync URL params to node state
    useEffect(() => {
        syncNodeFromParam(0, primarySearch, data0);
    }, [primarySearch, data0]);

    useEffect(() => {
        syncNodeFromParam(1, secondarySearch, data1);
    }, [secondarySearch, data1]);

    useEffect(() => {
        syncNodeFromParam(2, tertiarySearch, data2);
        if (tertiarySearch && data2) {
            setExtraNodeCount((prev) => Math.max(prev, 1));
        }
    }, [tertiarySearch, data2]);

    useEffect(() => {
        syncNodeFromParam(3, quaternarySearch, data3);
        if (quaternarySearch && data3) {
            setExtraNodeCount((prev) => Math.max(prev, 2));
        }
    }, [quaternarySearch, data3]);

    const syncNodeFromParam = (index: number, param: string | null, data: any) => {
        if (param && data) {
            const matchedNode = Object.values(data).find((node: any) => node.objectid === param) as
                | SearchValue
                | undefined;
            if (matchedNode) {
                updateNode(index, { searchTerm: matchedNode.name, selectedItem: matchedNode });
            }
        } else if (!param) {
            updateNode(index, emptyNode());
        }
    };

    const updateNode = (index: number, update: Partial<PathfindingNode>) => {
        setNodes((prev) => {
            const next = [...prev];
            while (next.length <= index) next.push(emptyNode());
            next[index] = { ...next[index], ...update };
            return next;
        });
    };

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
        const [moved] = currentNodes.splice(fromIndex, 1);
        currentNodes.splice(toIndex, 0, moved);

        setNodes(currentNodes);

        const params = getParamsFromNodes(currentNodes);
        triggerPathfinding(params);
    };

    const handleAddNode = () => {
        if (totalNodeCount < MAX_NODES) {
            setExtraNodeCount((prev) => prev + 1);
            setNodes((prev) => {
                const next = [...prev];
                while (next.length < 2 + extraNodeCount + 1) next.push(emptyNode());
                return next;
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
