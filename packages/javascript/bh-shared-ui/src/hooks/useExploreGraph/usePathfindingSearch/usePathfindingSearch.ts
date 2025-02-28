import { useEffect, useState } from 'react';
import { SearchValue } from '../../../store';
import { useExploreParams } from '../../useExploreParams';

export const usePathfindingSearch = () => {
    const [sourceSearchTerm, setSourceSearchTerm] = useState<string>('');
    const [sourceSelectedItem, setSourceSelectedItem] = useState<SearchValue | undefined>(undefined);

    const [destinationSearchTerm, setDestinationSearchTerm] = useState<string>('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

    // Watch query params and seperately sync them to each search field
    useEffect(() => {
        if (primarySearch) {
            setSourceSearchTerm(primarySearch);
        }
    }, [primarySearch]);

    useEffect(() => {
        if (secondarySearch) {
            setSourceSearchTerm(secondarySearch);
        }
    }, [secondarySearch]);

    // Handle syncing each search field up to query params to trigger a graph query. Should trigger pathfinding if both have been selected and node
    // if only one has been selected
    const handleSourceNodeSelected = (selected?: SearchValue) => {
        const currentSearch = selected?.name ?? selected?.objectid ?? '';

        setSourceSelectedItem(selected);
        setSourceSearchTerm(currentSearch);

        setExploreParams({
            searchType: secondarySearch ? 'pathfinding' : 'node',
            primarySearch: currentSearch,
        });
    };

    const handleDestinationNodeSelected = (selected?: SearchValue) => {
        const currentSearch = selected?.name ?? selected?.objectid ?? '';

        setDestinationSelectedItem(selected);
        setDestinationSearchTerm(currentSearch);

        setExploreParams({
            searchType: primarySearch ? 'pathfinding' : 'node',
            secondarySearch: currentSearch,
        });
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
    };
};
