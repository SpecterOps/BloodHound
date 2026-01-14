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

import { useCallback, useMemo, useState } from 'react';
import { parseItemId } from '../utils';
import { useExploreParams } from './useExploreParams';
import { useGraphItem } from './useGraphItem';

export const useExploreSelectedItem = () => {
    const [previousSelectedItem, setPreviousSelectedItem] = useState<string | null>();
    const { selectedItem, setExploreParams } = useExploreParams();

    const selectedItemQuery = useGraphItem(selectedItem!);

    /** Set the selected node or edge. The most recently selected item will stay highlighted */
    const setSelectedItem = useCallback(
        (itemId: string) => {
            if (itemId !== selectedItem) {
                setPreviousSelectedItem(selectedItem);
                setExploreParams({
                    expandedPanelSections: null,
                    selectedItem: itemId,
                });
            }
        },
        [selectedItem, setExploreParams]
    );

    const clearSelectedItem = useCallback(() => {
        if (selectedItem) {
            setExploreParams({
                expandedPanelSections: null,
                selectedItem: null,
            });
        }
    }, [setExploreParams, selectedItem]);

    const selectedItemType = useMemo(
        () => (selectedItem ? parseItemId(selectedItem).itemType : undefined),
        [selectedItem]
    );

    const selectedHiddenEdge = {
        id: selectedItem ? selectedItem : 'Hidden',
        name: '** Hidden Edge **',
        data: {},
        sourceNode: {
            id: 'HIDDEN',
            objectId: 'HIDDEN',
            name: 'HIDDEN',
            type: 'HIDDEN',
        },
        targetNode: {
            id: 'HIDDEN',
            objectId: 'HIDDEN',
            name: 'HIDDEN',
            type: 'HIDDEN',
        },
    };

    return {
        selectedItem,
        selectedItemQuery,
        setSelectedItem,
        selectedItemType,
        clearSelectedItem,
        selectedHiddenEdge,
        previousSelectedItem,
        setPreviousSelectedItem,
    };
};
