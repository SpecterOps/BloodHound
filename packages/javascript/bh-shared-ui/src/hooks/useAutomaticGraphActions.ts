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

import { useExploreParams } from './useExploreParams';
import { useExploreSelectedItem } from './useExploreSelectedItem';

export const useAutomaticGraphActions = (graphData: Record<string, any> | undefined) => {
    const { searchType, primarySearch } = useExploreParams();
    const { setSelectedItem, clearSelectedItem } = useExploreSelectedItem();

    if (searchType === 'pathfinding') {
        clearSelectedItem();
        return;
    }

    if (searchType !== 'node' || !primarySearch || !graphData) return;

    // Handle both data structures: BHCE uses { nodes: {...}, edges: [...] }, BE uses flat object
    const nodeData = graphData.nodes || graphData;
    const nodeId = Object.keys(nodeData).length === 1 ? Object.keys(nodeData)[0] : undefined;

    if (nodeId) {
        setSelectedItem(nodeId);
    }
};
