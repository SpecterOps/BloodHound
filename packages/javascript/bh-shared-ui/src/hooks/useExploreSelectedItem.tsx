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

import { useCallback, useEffect, useMemo, useState } from 'react';
import { parseItemId } from '../utils';
import { useExploreParams } from './useExploreParams';
import { useGraphItem } from './useGraphItem';

let initTimeout: NodeJS.Timeout;

// Scoped outside of module so it only applies on first app load
let isInitialized = false;

export const useExploreSelectedItem = () => {
    const { cypherSearch, exploreSearchTab, selectedItem, setExploreParams } = useExploreParams();
    const [infoPanelItem, setInfoPanelItem] = useState<string | null>(selectedItem);

    const selectedItemQuery = useGraphItem(selectedItem!);

    const setSelectedItem = useCallback(
        (itemId?: string) => {
            if (itemId !== selectedItem)
                setExploreParams({
                    selectedItem: itemId,
                    expandedPanelSections: null,
                });
        },
        [selectedItem, setExploreParams]
    );

    const selectedItemType = useMemo(
        () => (selectedItem ? parseItemId(selectedItem).itemType : undefined),
        [selectedItem]
    );

    // Indicates when initial load completes so selected item isn't cleared while gathering state
    useEffect(() => {
        clearTimeout(initTimeout);
        initTimeout = setTimeout(() => (isInitialized = true), 50);
    }, [cypherSearch, exploreSearchTab, selectedItem]);

    // EntityInfoPanel should dislay last selected item instead of currently
    // selected so that it remains open when an item is unselected
    useEffect(() => {
        if (!selectedItem) return;
        setInfoPanelItem(selectedItem);
    }, [selectedItem]);

    // Close info panel if displayed item no longer on graph
    useEffect(() => {
        if (isInitialized) {
            setExploreParams({
                selectedItem: null,
            });
            setInfoPanelItem(null);
        }
    }, [cypherSearch, exploreSearchTab, setExploreParams]);

    return {
        infoPanelItem,
        selectedItem,
        selectedItemQuery,
        setSelectedItem,
        selectedItemType,
    };
};
