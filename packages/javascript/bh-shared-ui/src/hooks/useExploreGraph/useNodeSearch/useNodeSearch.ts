import { useEffect, useState } from 'react';
import { SearchValue } from '../../../store';
import { useExploreParams } from '../../useExploreParams';

/* Reusable logic for syncing up a single node search field with browser query params on the Explore page. The value of the search field is tracked internally, and is only pushed to query params once the event handler is called by the consumer component. Direct changes to the associated query params will be synced back to the search field. */
export const useNodeSearch = () => {
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [selectedItem, setSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, searchType, setExploreParams } = useExploreParams();

    // Watch query params for a new incoming node search and sync to internal state
    useEffect(() => {
        if (primarySearch) {
            setSearchTerm(primarySearch);
        }
    }, [primarySearch, searchType]);

    // Handles syncing the local search state up to query params to trigger a graph query
    const selectSourceNode = (selected?: SearchValue) => {
        const searchTerm = selected?.name ?? selected?.objectid ?? '';

        setSelectedItem(selected);
        setSearchTerm(searchTerm);

        setExploreParams({
            searchType: 'node',
            primarySearch: searchTerm,
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
