import { useState } from 'react';
import { SearchValue } from '../../../store';
import { useExploreParams } from '../../useExploreParams';

export const useNodeSearch = () => {
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [selectedItem, setSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { setExploreParams } = useExploreParams();

    const selectSourceNode = (selected?: SearchValue) => {
        setSelectedItem(selected);
        setSearchTerm(selected?.name || '');
        setExploreParams({
            searchType: 'node',
            primarySearch: selected?.objectid,
        });
    };

    const editSourceNode = (edit: string) => setSearchTerm(edit);

    return {
        searchTerm,
        selectedItem,
        editSourceNode,
        selectSourceNode,
    };
};
