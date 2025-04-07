// Copyright 2023 Specter Ops, Inc.
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

import { faBullseye, faCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import ExploreSearchCombobox from '../../../components/ExploreSearchCombobox';
import { SearchValue } from '../../../store';
import { EdgeFilter, PathfindingFilterState } from './EdgeFilter';
import PathfindingSwapButton from './PathfindingSwapButton';

type PathfindingSearchState = {
    sourceSearchTerm: string;
    destinationSearchTerm: string;
    sourceSelectedItem: SearchValue | undefined;
    destinationSelectedItem: SearchValue | undefined;
    handleSourceNodeEdited: (edit: string) => void;
    handleDestinationNodeEdited: (edit: string) => void;
    handleSourceNodeSelected: (selected: SearchValue) => void;
    handleDestinationNodeSelected: (selected: SearchValue) => void;
    handleSwapPathfindingInputs: () => void;
};

const PathfindingSearch = ({
    pathfindingSearchState,
    pathfindingFilterState,
}: {
    pathfindingSearchState: PathfindingSearchState;
    pathfindingFilterState: PathfindingFilterState;
}) => {
    const {
        sourceSearchTerm,
        destinationSearchTerm,
        sourceSelectedItem,
        destinationSelectedItem,
        handleSourceNodeEdited,
        handleDestinationNodeEdited,
        handleSourceNodeSelected,
        handleDestinationNodeSelected,
        handleSwapPathfindingInputs,
    } = pathfindingSearchState;

    return (
        <div className='flex items-center gap-2'>
            <SourceToBullseyeIcon />

            <div className='flex flex-col flex-grow gap-2'>
                <ExploreSearchCombobox
                    handleNodeEdited={handleSourceNodeEdited}
                    handleNodeSelected={handleSourceNodeSelected}
                    inputValue={sourceSearchTerm}
                    selectedItem={sourceSelectedItem || null}
                    labelText='Start Node'
                />
                <ExploreSearchCombobox
                    handleNodeEdited={handleDestinationNodeEdited}
                    handleNodeSelected={handleDestinationNodeSelected}
                    inputValue={destinationSearchTerm}
                    selectedItem={destinationSelectedItem || null}
                    labelText='Destination Node'
                />
            </div>

            <PathfindingSwapButton
                disabled={!sourceSelectedItem || !destinationSelectedItem}
                onSwapPathfindingInputs={handleSwapPathfindingInputs}
            />
            <EdgeFilter pathfindingFilterState={pathfindingFilterState} />
        </div>
    );
};

const SourceToBullseyeIcon = () => {
    return (
        <div className='flex flex-col items-center'>
            <FontAwesomeIcon icon={faCircle} size='xs' />
            <div className='border-l border-dotted border-primary dark:border-white my-2 h-4'></div>
            <FontAwesomeIcon icon={faBullseye} size='xs' />
        </div>
    );
};

export default PathfindingSearch;
