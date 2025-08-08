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
        // TODO: Add logic to handle the filter change based on the selected option
        console.log('Selected Sniff Deep option:', option.value);
    };

    return (
        <div className='relative'>
            {/* Original PathfindingSearch component */}
            <PathfindingSearch 
                pathfindingSearchState={pathfindingSearchState}
                pathfindingFilterState={pathfindingFilterState}
            />
            
            {/* Compact dropdown positioned absolutely under the buttons */}
            <div className='absolute top-[52px] right-0 scale-75 origin-top-right'>
                <DropdownSelector
                    options={sniffDeepOptions}
                    selectedText={selectedOption.value}
                    onChange={handleDropdownChange}
                />
            </div>
        </div>
    );
};

export default SniffDeepSearch;
