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
    usePathfindingSearch
} from 'bh-shared-ui';
import { useState, useCallback } from 'react';
import { SearchValue } from 'bh-shared-ui/src/views/Explore/ExploreSearch/types';

const sniffDeepOptions: DropdownOption[] = [
    { key: 0, value: 'All' },
    { key: 1, value: 'DCSync' },
];

// Custom search state interface for sniff deep
interface SniffDeepSearchState {
    destinationSearchTerm: string;
    destinationSelectedItem: SearchValue | undefined;
    handleDestinationNodeEdited: (edit: string) => void;
    handleDestinationNodeSelected: (selected: SearchValue) => void;
}

// Custom hook for sniff deep search functionality
const useSniffDeepSearch = (): SniffDeepSearchState => {
    const [destinationSearchTerm, setDestinationSearchTerm] = useState('');
    const [destinationSelectedItem, setDestinationSelectedItem] = useState<SearchValue | undefined>();

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
        destinationSearchTerm,
        destinationSelectedItem,
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

    const handleDropdownChange = (option: DropdownOption) => {
        setSelectedOption(option);
        console.log('Selected Sniff Deep option:', option.value);
    };

    const executeSniffDeepSearch = useCallback(async () => {
        if (!sniffDeepSearchState.destinationSelectedItem) {
            console.warn('No destination node selected for sniff deep search');
            return;
        }

        const destinationNodeId = sniffDeepSearchState.destinationSelectedItem.objectid;
        const selectedSearchType = selectedOption.value;

        try {
            let cypherQueries: string[] = [];

            if (selectedSearchType === 'All' || selectedSearchType === 'DCSync') {
                // Path 1: GetChanges edge from Group nodes to the destination node
                const getChangesQuery = `
                    MATCH p=(g:Group)-[:GetChanges]->(d)
                    WHERE d.objectid = "${destinationNodeId}"
                    RETURN p
                    LIMIT 1000
                `;
                cypherQueries.push(getChangesQuery);

                // Path 2: GetChangesAll edge from Group nodes to the destination node  
                const getChangesAllQuery = `
                    MATCH p=(g:Group)-[:GetChangesAll]->(d)
                    WHERE d.objectid = "${destinationNodeId}"
                    RETURN p
                    LIMIT 1000
                `;
                cypherQueries.push(getChangesAllQuery);
            }

            // Execute the cypher queries
            for (const query of cypherQueries) {
                console.log('Executing sniff deep search query:', query);
                console.log('Target destination node:', destinationNodeId);
                
                // TODO: Execute the actual cypher query via apiClient
                // This will need to be integrated with the graph visualization
                // For now, just log the query that would be executed
                
                // Example of how this might look:
                // const result = await apiClient.cypherSearch(query, {}, true);
                // console.log('Sniff deep search result:', result);
            }

        } catch (error) {
            console.error('Error executing sniff deep search:', error);
        }
    }, [sniffDeepSearchState.destinationSelectedItem, selectedOption.value]);

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
                        {/* Source node - shows "Group nodes" as placeholder/fixed */}
                        <div className='relative'>
                            <div className='min-h-[40px] px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700 flex items-center text-sm text-gray-500 dark:text-gray-400'>
                                Group nodes (source)
                            </div>
                        </div>
                        
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
                    disabled={!sniffDeepSearchState.destinationSelectedItem}
                    title="Start Sniff Deep search">
                    <FontAwesomeIcon icon={faPlay} size='xs' />
                </Button>
            </div>
        </div>
    );
};

export default SniffDeepSearch;
