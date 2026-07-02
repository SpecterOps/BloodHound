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

import { faBullseye, faCircle, faGripVertical, faPlus, faTimes } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useRef, useState } from 'react';
import ExploreSearchCombobox from '../../../components/ExploreSearchCombobox';
import { EdgeFilter, PathfindingFilterState } from './EdgeFilter/EdgeFilter';
import PathfindingSwapButton from './PathfindingSwapButton';
import { SearchValue } from './types';

type PathfindingNode = {
    searchTerm: string;
    selectedItem: SearchValue | undefined;
};

type PathfindingSearchState = {
    sourceSearchTerm: string;
    destinationSearchTerm: string;
    sourceSelectedItem: SearchValue | undefined;
    destinationSelectedItem: SearchValue | undefined;
    nodes: PathfindingNode[];
    totalNodeCount: number;
    maxNodes: number;
    handleSourceNodeEdited: (edit: string) => void;
    handleDestinationNodeEdited: (edit: string) => void;
    handleSourceNodeSelected: (selected: SearchValue) => void;
    handleDestinationNodeSelected: (selected: SearchValue) => void;
    handleNodeEdited: (index: number) => (edit: string) => void;
    handleNodeSelected: (index: number) => (selected: SearchValue) => void;
    handleSwapPathfindingInputs: () => void;
    handleReorderNodes: (fromIndex: number, toIndex: number) => void;
    handleRemoveNode: (index: number) => void;
    handleAddNode: () => void;
};

const PathfindingSearch = ({
    pathfindingSearchState,
    pathfindingFilterState,
}: {
    pathfindingSearchState: PathfindingSearchState;
    pathfindingFilterState: PathfindingFilterState;
}) => {
    const {
        sourceSelectedItem,
        destinationSelectedItem,
        nodes,
        totalNodeCount,
        maxNodes,
        handleNodeEdited,
        handleNodeSelected,
        handleSwapPathfindingInputs,
        handleReorderNodes,
        handleRemoveNode,
        handleAddNode,
    } = pathfindingSearchState;

    const [dragIndex, setDragIndex] = useState<number | null>(null);
    const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
    const dragCounter = useRef<Record<number, number>>({});

    const handleDragStart = (index: number) => (e: React.DragEvent) => {
        setDragIndex(index);
        e.dataTransfer.effectAllowed = 'move';
    };

    const handleDragEnter = (index: number) => (e: React.DragEvent) => {
        e.preventDefault();
        dragCounter.current[index] = (dragCounter.current[index] || 0) + 1;
        if (dragIndex !== null && index !== dragIndex) {
            setDragOverIndex(index);
        }
    };

    const handleDragLeave = (index: number) => () => {
        dragCounter.current[index] = (dragCounter.current[index] || 0) - 1;
        if (dragCounter.current[index] <= 0) {
            dragCounter.current[index] = 0;
            if (dragOverIndex === index) {
                setDragOverIndex(null);
            }
        }
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
    };

    const handleDrop = (toIndex: number) => (e: React.DragEvent) => {
        e.preventDefault();
        if (dragIndex !== null && dragIndex !== toIndex) {
            handleReorderNodes(dragIndex, toIndex);
        }
        setDragIndex(null);
        setDragOverIndex(null);
        dragCounter.current = {};
    };

    const handleDragEnd = () => {
        setDragIndex(null);
        setDragOverIndex(null);
        dragCounter.current = {};
    };

    const visibleNodes = nodes.slice(0, totalNodeCount).map((node, index) => ({
        label: index === 0 ? 'Start Node' : 'Destination Node',
        searchTerm: node.searchTerm,
        selectedItem: node.selectedItem,
        removable: index > 0 && totalNodeCount > 2,
        autoFocus:
            index === 0
                ? !node.searchTerm
                : index === 1
                  ? !!(nodes[0]?.searchTerm && !node.searchTerm)
                  : !node.searchTerm,
    }));

    return (
        <div className='flex items-center gap-2' data-testid='pathfinding-search'>
            <SourceToBullseyeIcon destinationCount={totalNodeCount - 1} />

            <div className='flex flex-col flex-grow gap-2'>
                {visibleNodes.map((node, index) => (
                    <div
                        key={index}
                        draggable
                        onDragStart={handleDragStart(index)}
                        onDragEnter={handleDragEnter(index)}
                        onDragLeave={handleDragLeave(index)}
                        onDragOver={handleDragOver}
                        onDrop={handleDrop(index)}
                        onDragEnd={handleDragEnd}
                        className={`relative flex items-center gap-1 rounded transition-all group ${
                            dragIndex === index ? 'opacity-40' : ''
                        } ${dragOverIndex === index ? 'ring-2 ring-primary ring-offset-1' : ''}`}>
                        <div className='cursor-grab text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300 opacity-0 group-hover:opacity-100 transition-opacity'>
                            <FontAwesomeIcon icon={faGripVertical} size='sm' />
                        </div>
                        <div className='flex-grow'>
                            <ExploreSearchCombobox
                                autoFocus={node.autoFocus}
                                handleNodeEdited={handleNodeEdited(index)}
                                handleNodeSelected={handleNodeSelected(index)}
                                inputValue={node.searchTerm}
                                selectedItem={node.selectedItem || null}
                                labelText={node.label}
                            />
                        </div>
                        {node.removable && (
                            <button
                                onClick={() => handleRemoveNode(index)}
                                className='absolute right-1 top-1/2 -translate-y-1/2 p-1 text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300 z-10'
                                aria-label='Remove destination'
                                title='Remove destination'>
                                <FontAwesomeIcon icon={faTimes} size='sm' />
                            </button>
                        )}
                    </div>
                ))}
                {totalNodeCount < maxNodes && (
                    <button
                        onClick={handleAddNode}
                        className='flex items-center gap-1.5 text-xs text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300 py-0.5 cursor-pointer'
                        aria-label='Add destination'>
                        <FontAwesomeIcon icon={faPlus} size='xs' />
                        <span>Add Destination</span>
                    </button>
                )}
            </div>

            {totalNodeCount === 2 && (
                <PathfindingSwapButton
                    disabled={!sourceSelectedItem || !destinationSelectedItem}
                    onSwapPathfindingInputs={handleSwapPathfindingInputs}
                />
            )}
            <EdgeFilter pathfindingFilterState={pathfindingFilterState} />
        </div>
    );
};

const SourceToBullseyeIcon = ({ destinationCount }: { destinationCount: number }) => {
    return (
        <div className='flex flex-col items-center'>
            <FontAwesomeIcon icon={faCircle} size='xs' />
            {Array.from({ length: destinationCount }).map((_, i) => (
                <span key={i} className='flex flex-col items-center'>
                    <div className='border-l border-dotted border-primary dark:border-white my-2 h-4'></div>
                    <FontAwesomeIcon icon={faBullseye} size='xs' />
                </span>
            ))}
        </div>
    );
};

export default PathfindingSearch;
