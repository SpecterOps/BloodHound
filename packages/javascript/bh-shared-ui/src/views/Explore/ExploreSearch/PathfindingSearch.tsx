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
import { PathfindingNode, PathfindingSearchState } from '../../../hooks/useExploreGraph/usePathfindingSearch';
import { cn } from '../../../utils';
import { EdgeFilter, PathfindingFilterState } from './EdgeFilter/EdgeFilter';
import PathfindingSwapButton from './PathfindingSwapButton';

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
    const dragImageRefs = useRef<Record<number, HTMLDivElement | null>>({});

    const handleDragStart = (index: number) => (e: React.DragEvent) => {
        setDragIndex(index);
        e.dataTransfer.effectAllowed = 'move';

        // The row owns the drag, but a drag image is rasterized from the whole painted subtree of
        // whatever element it is given. Snapshot the input box alone: the row would drag along the
        // node icon and connector line, and the combobox would drag along its results dropdown,
        // which is open whenever the input has focus.
        const dragImageElement = dragImageRefs.current[index];
        if (dragImageElement) {
            const { left, top, width, height } = dragImageElement.getBoundingClientRect();

            // The grab point can sit outside the input box (on the grip or the node icon), which
            // would leave the drag image floating away from the cursor.
            const cursorOffsetX = Math.min(Math.max(e.clientX - left, 0), width);
            const cursorOffsetY = Math.min(Math.max(e.clientY - top, 0), height);

            e.dataTransfer.setDragImage(dragImageElement, cursorOffsetX, cursorOffsetY);
        }
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

    // The first destination only steals focus once the start node has a term, so tabbing into
    // an empty form lands on the start node rather than jumping ahead.
    const shouldAutoFocus = (index: number, node: PathfindingNode): boolean => {
        if (index === 1) return !!(nodes[0]?.searchTerm && !node.searchTerm);
        return !node.searchTerm;
    };

    const visibleNodes = nodes.slice(0, totalNodeCount).map((node, index) => {
        // Every destination shows the same label on screen, so the accessible name appends the
        // destination number to keep the rows distinguishable to assistive tech. It keeps the
        // visible text as a prefix so the spoken name still matches the screen (WCAG 2.5.3).
        const label = index === 0 ? 'Start Node' : 'Destination Node';

        return {
            label,
            ariaLabel: index === 0 ? label : `${label} ${index}`,
            searchTerm: node.searchTerm,
            selectedItem: node.selectedItem,
            removable: index > 0 && totalNodeCount > 2,
            autoFocus: shouldAutoFocus(index, node),
        };
    });

    return (
        <div className='flex items-center gap-2' data-testid='pathfinding-search'>
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
                        className={cn('relative flex items-center gap-1 rounded transition-all group', {
                            'opacity-40': dragIndex === index,
                        })}>
                        <PathfindingNodeIcon isStartNode={index === 0} showConnector={index > 0} />
                        <div className='relative flex flex-grow items-center gap-1'>
                            <div
                                role='button'
                                tabIndex={0}
                                aria-label={`Reorder ${node.ariaLabel}, position ${index + 1} of ${visibleNodes.length}`}
                                aria-roledescription='sortable'
                                onKeyDown={(e) => {
                                    if (e.key === 'ArrowUp' && index > 0) {
                                        e.preventDefault();
                                        handleReorderNodes(index, index - 1);
                                    } else if (e.key === 'ArrowDown' && index < visibleNodes.length - 1) {
                                        e.preventDefault();
                                        handleReorderNodes(index, index + 1);
                                    }
                                }}
                                className='cursor-grab text-light hover:text-main dark:text-common-white dark:hover:text-neutral-light-5 opacity-0 group-hover:opacity-100 focus-visible:opacity-100 transition-opacity'>
                                <FontAwesomeIcon icon={faGripVertical} size='sm' />
                            </div>
                            <div
                                className={cn('flex-grow rounded', {
                                    'ring-2 ring-primary ring-offset-1': dragOverIndex === index,
                                })}>
                                <ExploreSearchCombobox
                                    ariaLabel={node.ariaLabel}
                                    inputContainerRef={(element) => {
                                        dragImageRefs.current[index] = element;
                                    }}
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
                                    className='absolute right-1 top-1/2 -translate-y-1/2 p-1 text-light hover:text-main dark:text-common-white dark:hover:text-neutral-light-5 z-10'
                                    aria-label={`Remove ${node.ariaLabel}`}
                                    title={`Remove ${node.ariaLabel}`}>
                                    <FontAwesomeIcon icon={faTimes} size='sm' />
                                </button>
                            )}
                        </div>
                    </div>
                ))}
                {totalNodeCount < maxNodes && (
                    <button
                        onClick={handleAddNode}
                        className='flex items-center gap-1.5 ml-4 text-xs text-light hover:text-main dark:text-common-white dark:hover:text-neutral-light-5 py-0.5 cursor-pointer'
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

// The icon sits inside its own input row and stretches to the row's full height, so `items-center`
// on the row keeps it centered on the input no matter how tall that input renders. The connector
// is absolutely positioned so it spans the gap up to the previous row's icon, stopping 0.5rem short
// at each end to leave the same breathing room around both icons.
const PathfindingNodeIcon = ({ isStartNode, showConnector }: { isStartNode: boolean; showConnector: boolean }) => {
    return (
        <div className='relative flex w-3 shrink-0 items-center justify-center self-stretch'>
            {showConnector && (
                <div className='absolute bottom-[calc(50%+0.5rem)] left-1/2 h-[calc(100%-0.5rem)] -translate-x-1/2 border-l border-dotted border-primary dark:border-white'></div>
            )}
            <FontAwesomeIcon icon={isStartNode ? faCircle : faBullseye} size='xs' />
        </div>
    );
};

export default PathfindingSearch;
