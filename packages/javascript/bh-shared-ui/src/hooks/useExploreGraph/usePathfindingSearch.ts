import { useEffect, useState } from 'react';
import { SearchValue } from '../../store';
import { useExploreParams } from '../useExploreParams';
import { useExploreGraph } from './useExploreGraph';

export const usePathfindingSearch = () => {
    const [sourceSearchTerm, setSourceSearchTerm] = useState<string>('');
    const [sourceSelectedItem, setSourceSelectedItem] = useState<SearchValue | undefined>(undefined);

    const [destinationSearchTerm, setDestinationSearchTerm] = useState<string>('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>(undefined);

    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();
    const { data: graphData } = useExploreGraph();

    // Watch query params and seperately sync them to each search field
    useEffect(() => {
        if (primarySearch && graphData) {
            const matchedNode = Object.values(graphData).find((node) => node.data['objectid'] === primarySearch);

            if (matchedNode) {
                setSourceSearchTerm(matchedNode.data['name']);

                setSourceSelectedItem({
                    name: matchedNode.data['name'],
                    objectid: matchedNode.data['objectid'],
                    type: matchedNode.data['nodetype'],
                });
            }
        }
    }, [primarySearch, graphData]);

    useEffect(() => {
        if (secondarySearch && graphData) {
            const matchedNode = Object.values(graphData).find((node) => node.data['objectid'] === secondarySearch);

            if (matchedNode) {
                setDestinationSearchTerm(matchedNode.data['name']);

                setDestinationSelectedItem({
                    name: matchedNode.data['name'],
                    objectid: matchedNode.data['objectid'],
                    type: matchedNode.data['nodetype'],
                });
            }
        }
    }, [secondarySearch, graphData]);

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
