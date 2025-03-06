import { useCallback, useMemo } from 'react';
import { parseItemId } from '../utils';
import { useExploreParams } from './useExploreParams';
import { useGraphItem } from './useGraphItem';

export const useExploreSelectedItem = () => {
    const { selectedItem, setExploreParams } = useExploreParams();
    const selectedItemQuery = useGraphItem(selectedItem!);

    const setSelectedItem = useCallback(
        (itemId: string) => {
            if (itemId !== selectedItem)
                setExploreParams({
                    selectedItem: itemId,
                });
        },
        [selectedItem, setExploreParams]
    );

    const selectedItemType = useMemo(
        () => (selectedItem ? parseItemId(selectedItem).itemType : undefined),
        [selectedItem]
    );

    return {
        selectedItem,
        selectedItemQuery,
        setSelectedItem,
        selectedItemType,
    };
};
