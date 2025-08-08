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

import { faBullseye, faCircle, faPlay, faFilter } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from '@bloodhoundenterprise/doodleui';
import { 
    DropdownSelector, 
    DropdownOption,
    ExploreSearchCombobox,
    usePathfindingFilters,
    usePathfindingSearch,
    useExploreParams,
    encodeCypherQuery
} from 'bh-shared-ui';
import { useState, useCallback } from 'react';
import { SearchValue } from 'bh-shared-ui/src/views/Explore/ExploreSearch/types';

const sniffDeepOptions: DropdownOption[] = [
    { key: 0, value: 'All' },
    { key: 1, value: 'DCSync' },
];

// Custom search state interface for sniff deep
interface SniffDeepSearchState {
    sourceSearchTerm: string;
    sourceSelectedItem: SearchValue | undefined;
    destinationSearchTerm: string;
    destinationSelectedItem: SearchValue | undefined;
    handleSourceNodeEdited: (edit: string) => void;
    handleSourceNodeSelected: (selected: SearchValue) => void;
    handleDestinationNodeEdited: (edit: string) => void;
    handleDestinationNodeSelected: (selected: SearchValue) => void;
}

// Custom hook for sniff deep search functionality
const useSniffDeepSearch = (): SniffDeepSearchState => {
    const [sourceSearchTerm, setSourceSearchTerm] = useState('');
    const [sourceSelectedItem, setSourceSelectedItem] = useState<SearchValue | undefined>();
    const [destinationSearchTerm, setDestinationSearchTerm] = useState('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>();

    const handleSourceNodeEdited = useCallback((edit: string) => {
        setSourceSearchTerm(edit);
        // Clear selected item when editing
        if (sourceSelectedItem) {
            setSourceSelectedItem(undefined);
        }
    }, [sourceSelectedItem]);

    const handleSourceNodeSelected = useCallback((selected: SearchValue) => {
        setSourceSelectedItem(selected);
        setSourceSearchTerm(selected.name || '');
    }, []);

    const handleDestinationNodeEdited = useCallback((edit: string) => {
        setDestinationSearchTerm(edit);
        // Clear selected item when editing
        if (destinationSelectedItem) {
            setDestinationSelectedItem(undefined);
        }
    }, [destinationSelectedItem]);

    const handleDestinationNodeSelected = useCallback((selected: SearchValue) => {
        setDestinationSelectedItem(selected);
        setDestinationSearchTerm(selected.name || '');
    }, []);

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

const SniffDeepSearch = ({
    pathfindingSearchState,
    pathfindingFilterState,
}: {
    pathfindingSearchState: ReturnType<typeof usePathfindingSearch>;
    pathfindingFilterState: ReturnType<typeof usePathfindingFilters>;
}) => {
    const [selectedOption, setSelectedOption] = useState<DropdownOption>(sniffDeepOptions[0]);
    const sniffDeepSearchState = useSniffDeepSearch();
    const { setExploreParams } = useExploreParams();

    const handleDropdownChange = (option: DropdownOption) => {
        setSelectedOption(option);
        console.log('Selected Sniff Deep option:', option.value);
    };

    const executeSniffDeepSearch = useCallback(async () => {
        if (!sniffDeepSearchState.sourceSelectedItem || !sniffDeepSearchState.destinationSelectedItem) {
            console.warn('Both source and destination nodes must be selected for sniff deep search');
            return;
        }

        const sourceNodeId = sniffDeepSearchState.sourceSelectedItem.objectid;
        const destinationNodeId = sniffDeepSearchState.destinationSelectedItem.objectid;
        const selectedSearchType = selectedOption.value;

        try {
            let cypherQuery = '';

            if (selectedSearchType === 'All' || selectedSearchType === 'DCSync') {
                cypherQuery = `MATCH p_changes = (x1:Base)-[:GetChanges]->(d:Domain)
                WHERE d.objectid = "${destinationNodeId}"
                MATCH p_changesall = (x2:Base)-[:GetChangesAll]->(d)
                MATCH p_tochanges = shortestpath((n:Base)-[:GenericAll|AddMember|MemberOf*0..]->(x1))
                WHERE NOT (n)-[:GenericAll|AddMember|MemberOf*0..10]->()-[:DCSync]->(d)
                    AND n.objectid = "${sourceNodeId}"
                MATCH p_tochangesall = shortestpath((n)-[:GenericAll|AddMember|MemberOf*0..]->(x2)) 
                WHERE n.objectid = "${sourceNodeId}"
                RETURN p_changes,p_tochanges, p_changesall,p_tochangesall`;
            }

            if (cypherQuery) {
                console.log('Executing sniff deep search query:', cypherQuery);
                console.log('Source node:', sourceNodeId, 'Destination node:', destinationNodeId);
                
                // Execute the cypher query via the explore params system
                // This will trigger the same query execution and graph visualization as cypher search
                setExploreParams({
                    searchType: 'cypher',
                    cypherSearch: encodeCypherQuery(cypherQuery),
                });
            }

        } catch (error) {
            console.error('Error executing sniff deep search:', error);
        }
    }, [sniffDeepSearchState.sourceSelectedItem, sniffDeepSearchState.destinationSelectedItem, selectedOption.value, setExploreParams]);

    const handlePlayButtonClick = () => {
        console.log('Play button clicked - starting sniff deep search with option:', selectedOption.value);
        executeSniffDeepSearch();
    };

    return (
        <div className='relative'>
            {/* Wrapper div with increased height for the gray search container */}
            <div className='bg-gray-100 dark:bg-gray-800 p-2 origin-top-center rounded-lg min-h-[100px]'>
                {/* Custom sniff deep search UI that looks like PathfindingSearch */}
                <div className='flex items-center gap-2'>
                    {/* Source to bullseye icon */}
                    <div className='flex flex-col items-center'>
                        <FontAwesomeIcon icon={faCircle} size='xs' />
                        <div className='border-l border-dotted border-primary dark:border-white my-2 h-4'></div>
                        <FontAwesomeIcon icon={faBullseye} size='xs' />
                    </div>

                    {/* Search inputs */}
                    <div className='flex flex-col flex-grow gap-2'>
                        {/* Source node - functional search */}
                        <ExploreSearchCombobox
                            handleNodeEdited={sniffDeepSearchState.handleSourceNodeEdited}
                            handleNodeSelected={sniffDeepSearchState.handleSourceNodeSelected}
                            inputValue={sniffDeepSearchState.sourceSearchTerm}
                            selectedItem={sniffDeepSearchState.sourceSelectedItem || null}
                            labelText='Source Node'
                        />
                        
                        {/* Destination node - functional search */}
                        <ExploreSearchCombobox
                            handleNodeEdited={sniffDeepSearchState.handleDestinationNodeEdited}
                            handleNodeSelected={sniffDeepSearchState.handleDestinationNodeSelected}
                            inputValue={sniffDeepSearchState.destinationSearchTerm}
                            selectedItem={sniffDeepSearchState.destinationSelectedItem || null}
                            labelText='Destination Node'
                        />
                    </div>

                    {/* Filter button (placeholder, non-functional for now) */}
                    <Button
                        className='h-7 w-7 min-w-7 p-0 rounded-[4px] border-black/25 text-white'
                        disabled={true}
                        title="Edge filters (not available for sniff deep)">
                        <FontAwesomeIcon icon={faFilter} size='xs' />
                    </Button>
                </div>
            </div>
            
            {/* Compact dropdown positioned above the buttons area */}
            <div className='absolute top-[-5px] right-[8px] scale-75 origin-top-right z-10'>
                <DropdownSelector
                    options={sniffDeepOptions}
                    selectedText={selectedOption.value}
                    onChange={handleDropdownChange}
                />
            </div>

            {/* Play button positioned as third button in the row */}
            <div className='absolute top-[68px] right-[25px] z-10'>
                <Button
                    className='h-7 w-7 min-w-7 p-0 rounded-[4px] border-black/25 text-white'
                    onClick={handlePlayButtonClick}
                    disabled={!sniffDeepSearchState.sourceSelectedItem || !sniffDeepSearchState.destinationSelectedItem}
                    title="Start Sniff Deep search">
                    <FontAwesomeIcon icon={faPlay} size='xs' />
                </Button>
            </div>
        </div>
    );
};

export default SniffDeepSearch;
