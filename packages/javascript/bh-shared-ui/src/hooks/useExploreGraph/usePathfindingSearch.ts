import { useEffect, useMemo, useState } from 'react';
import { SearchValue } from '../../store';
import { useExploreParams } from '../useExploreParams';
import { getKeywordAndTypeValues, useSearch } from '../useSearch';

export const usePathfindingSearch = () => {
    const [sourceSearchTerm, setSourceSearchTerm] = useState<string>('');
    const [sourceSelectedItem, setSourceSelectedItem] = useState<SearchValue | undefined>(undefined);
    const [destinationSearchTerm, setDestinationSearchTerm] = useState<string>('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

    // Wire up search queries. we should only recompute keywords when the param values change
    const { keyword: sourceKeyword, type: sourceType } = useMemo(
        () => getKeywordAndTypeValues(primarySearch ?? undefined),
        [primarySearch]
    );
    const { keyword: destinationKeyword, type: destinationType } = useMemo(
        () => getKeywordAndTypeValues(secondarySearch ?? undefined),
        [secondarySearch]
    );
    const { data: sourceSearchData } = useSearch(sourceKeyword, sourceType);
    const { data: destinationSearchData } = useSearch(destinationKeyword, destinationType);

    // Watch query params and seperately sync them to each search field
    useEffect(() => {
        if (primarySearch && sourceSearchData) {
            const matchedNode = Object.values(sourceSearchData).find((node) => node.objectid === primarySearch);

            if (matchedNode) {
                setSourceSearchTerm(matchedNode.name);
                setSourceSelectedItem(matchedNode);
            }
        }
    }, [primarySearch, sourceSearchData]);

    useEffect(() => {
        if (secondarySearch && destinationSearchData) {
            const matchedNode = Object.values(destinationSearchData).find((node) => node.objectid === secondarySearch);

            if (matchedNode) {
                setDestinationSearchTerm(matchedNode.name);
                setDestinationSelectedItem(matchedNode);
            }
        }
    }, [secondarySearch, destinationSearchData]);

    // Handle syncing each search field up to query params to trigger a graph query. Should trigger pathfinding if both have been selected and node
    // if only one has been selected
    const handleSourceNodeSelected = (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        setSourceSelectedItem(selected);
        setSourceSearchTerm(term);

        setExploreParams({
            searchType: secondarySearch ? 'pathfinding' : 'node',
            primarySearch: objectId,
        });
    };

    const handleDestinationNodeSelected = (selected?: SearchValue) => {
        const objectId = selected?.objectid ?? '';
        const term = selected?.name ?? objectId;

        setDestinationSelectedItem(selected);
        setDestinationSearchTerm(term);

        setExploreParams({
            searchType: primarySearch ? 'pathfinding' : 'node',
            secondarySearch: objectId,
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
