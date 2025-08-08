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

import { faPlay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button } from '@bloodhoundenterprise/doodleui';
import { 
    DropdownSelector, 
    DropdownOption, 
    PathfindingSearch,
    usePathfindingFilters,
    usePathfindingSearch
} from 'bh-shared-ui';
import { useState } from 'react';

const sniffDeepOptions: DropdownOption[] = [
    { key: 0, value: 'All' },
    { key: 1, value: 'DCSync' },
];

const SniffDeepSearch = ({
    pathfindingSearchState,
    pathfindingFilterState,
}: {
    pathfindingSearchState: ReturnType<typeof usePathfindingSearch>;
    pathfindingFilterState: ReturnType<typeof usePathfindingFilters>;
}) => {
    const [selectedOption, setSelectedOption] = useState<DropdownOption>(sniffDeepOptions[0]);

    const handleDropdownChange = (option: DropdownOption) => {
        setSelectedOption(option);
        console.log('Selected Sniff Deep option:', option.value);
    };

    const handlePlayButtonClick = () => {
        // TODO: Implement manual search trigger logic based on selected option
        console.log('Play button clicked - starting search with option:', selectedOption.value);
        // This is where you would trigger the actual search based on the selected dropdown option
    };

    return (
        <div className='relative'>
            {/* Wrapper div with increased height for the gray search container */}
            <div className='bg-gray-100 dark:bg-gray-800 p-2 origin-top-center rounded-lg min-h-[100px]'>
                {/* Original PathfindingSearch component */}
                <PathfindingSearch 
                    pathfindingSearchState={pathfindingSearchState}
                    pathfindingFilterState={pathfindingFilterState}
                />
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
                    title="Start Sniff Deep search">
                    <FontAwesomeIcon icon={faPlay} size='xs' />
                </Button>
            </div>
        </div>
    );
};

export default SniffDeepSearch;
