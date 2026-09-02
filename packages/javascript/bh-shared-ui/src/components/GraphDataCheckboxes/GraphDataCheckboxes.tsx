// Copyright 2024 Specter Ops, Inc.
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

import { Checkbox, FormControlLabel } from '@mui/material';
import { Skeleton } from 'doodle-ui';
import { type SourceKind } from 'js-client-library';
import { type FC } from 'react';
import { useSourceKindsQuery } from '../../hooks/useSourceKinds';

type GraphDataOption = {
    key: string;
    label: string;
};

type SelectionAmount = 'all' | 'none' | 'some';

type GraphDataSelection = {
    selected?: true;
    options?: Record<string, true>;
};

type GraphDataChecked = Record<number, GraphDataSelection>;

export type GraphDataSelections = {
    sourceKinds: number[];
    relationships: string[];
    allGraphData: boolean;
};

// Displayed while source kinds are loading
const LOADING_CHECKBOXES = (
    <>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
    </>
);

// The default source kind names are replaced with friendlier ones
const KIND_LABEL_MAP: Record<string, string> = {
    Base: 'Active Directory',
    AZBase: 'Azure',
};

const SOURCE_KIND_GRAPH_DATA_OPTIONS: Record<string, GraphDataOption[]> = {
    Base: [{ key: 'HasSession', label: 'HasSession edges' }],
};

const getSelectionAmount = (checkedCount: number, totalCount: number): SelectionAmount => {
    if (checkedCount === 0 || totalCount === 0) {
        return 'none';
    }

    return checkedCount === totalCount ? 'all' : 'some';
};

const getSourceKindGraphDataOptions = (sourceKindName: string): GraphDataOption[] => {
    return SOURCE_KIND_GRAPH_DATA_OPTIONS[sourceKindName] ?? [];
};

const getSourceKindGraphDataOptionKeys = (sourceKindName: string): string[] => {
    return getSourceKindGraphDataOptions(sourceKindName).map((graphDataOption) => graphDataOption.key);
};

const hasGraphDataSelection = (selection: GraphDataSelection): boolean => {
    return Boolean(selection.selected) || Object.keys(selection.options ?? {}).length > 0;
};

const getGraphDataSelections = (checked: GraphDataChecked, sourceKinds: SourceKind[] = []): GraphDataSelections => {
    return {
        sourceKinds: Object.entries(checked).flatMap(([sourceKindId, selection]) => {
            return selection.selected ? [Number(sourceKindId)] : [];
        }),
        relationships: [
            ...new Set(Object.values(checked).flatMap((selection) => Object.keys(selection.options ?? {}))),
        ],
        allGraphData: getAllGraphDataSelectionAmount(checked, sourceKinds) === 'all',
    };
};

const getGraphDataChecked = (
    value: Pick<GraphDataSelections, 'sourceKinds' | 'relationships'>,
    sourceKinds: SourceKind[] = []
): GraphDataChecked => {
    const checked = value.sourceKinds.reduce<GraphDataChecked>((allChecked, sourceKindId) => {
        allChecked[sourceKindId] = { selected: true };
        return allChecked;
    }, {});

    return sourceKinds.reduce((allChecked, sourceKind) => {
        const sourceKindSelected = value.sourceKinds.includes(sourceKind.id);
        const graphDataOptionKeys = getSourceKindGraphDataOptionKeys(sourceKind.name);

        return graphDataOptionKeys.reduce((nextChecked, key) => {
            return sourceKindSelected || value.relationships.includes(key)
                ? setGraphDataOptionChecked(nextChecked, sourceKind.id, key, true)
                : nextChecked;
        }, allChecked);
    }, checked);
};

const getGraphDataSelection = (checked: GraphDataChecked, sourceKindId: number): GraphDataSelection => {
    return checked[sourceKindId] ?? {};
};

const isSourceKindChecked = (checked: GraphDataChecked, sourceKindId: number): boolean => {
    return Boolean(getGraphDataSelection(checked, sourceKindId).selected);
};

const isGraphDataOptionChecked = (checked: GraphDataChecked, sourceKindId: number, key: string): boolean => {
    return Boolean(getGraphDataSelection(checked, sourceKindId).options?.[key]);
};

const isEffectiveGraphDataOptionChecked = (checked: GraphDataChecked, sourceKindId: number, key: string): boolean => {
    return isSourceKindChecked(checked, sourceKindId) || isGraphDataOptionChecked(checked, sourceKindId, key);
};

const updateGraphDataSelection = (
    checked: GraphDataChecked,
    sourceKindId: number,
    updater: (selection: GraphDataSelection) => void
): GraphDataChecked => {
    const nextChecked = { ...checked };
    const nextSelection = { ...getGraphDataSelection(checked, sourceKindId) };

    updater(nextSelection);

    if (hasGraphDataSelection(nextSelection)) {
        nextChecked[sourceKindId] = nextSelection;
    } else {
        delete nextChecked[sourceKindId];
    }

    return nextChecked;
};

const setGraphDataOptionsChecked = (
    checked: GraphDataChecked,
    sourceKindId: number,
    keys: string[],
    isChecked: boolean
): GraphDataChecked => {
    return keys.reduce(
        (nextChecked, key) => setGraphDataOptionChecked(nextChecked, sourceKindId, key, isChecked),
        checked
    );
};

const setSourceKindChecked = (
    checked: GraphDataChecked,
    sourceKindId: number,
    isChecked: boolean
): GraphDataChecked => {
    return updateGraphDataSelection(checked, sourceKindId, (selection) => {
        if (isChecked) {
            selection.selected = true;
        } else {
            delete selection.selected;
        }
    });
};

const setGraphDataOptionChecked = (
    checked: GraphDataChecked,
    sourceKindId: number,
    key: string,
    isChecked: boolean
): GraphDataChecked => {
    return updateGraphDataSelection(checked, sourceKindId, (selection) => {
        const nextOptions = { ...(selection.options ?? {}) };

        if (isChecked) {
            nextOptions[key] = true;
        } else {
            delete nextOptions[key];
        }

        if (Object.keys(nextOptions).length > 0) {
            selection.options = nextOptions;
        } else {
            delete selection.options;
        }
    });
};

const getSourceKindSelectionDetails = (
    checked: GraphDataChecked,
    sourceKind: SourceKind
): { graphDataOptionKeys: string[]; checkedCount: number; totalCount: number } => {
    const graphDataOptionKeys = getSourceKindGraphDataOptionKeys(sourceKind.name);
    const checkedCount =
        Number(isSourceKindChecked(checked, sourceKind.id)) +
        graphDataOptionKeys.filter((key) => isEffectiveGraphDataOptionChecked(checked, sourceKind.id, key)).length;

    return {
        graphDataOptionKeys,
        checkedCount,
        totalCount: graphDataOptionKeys.length + 1,
    };
};

const getAllGraphDataSelectionAmount = (checked: GraphDataChecked, sourceKinds: SourceKind[]): SelectionAmount => {
    const counts = sourceKinds.reduce(
        (allCounts, sourceKind) => {
            const { checkedCount, totalCount } = getSourceKindSelectionDetails(checked, sourceKind);

            allCounts.checkedCount += checkedCount;
            allCounts.totalCount += totalCount;

            return allCounts;
        },
        { checkedCount: 0, totalCount: 0 }
    );

    return getSelectionAmount(counts.checkedCount, counts.totalCount);
};

const getAllGraphDataChecked = (sourceKinds: SourceKind[]): GraphDataChecked => {
    return sourceKinds.reduce<GraphDataChecked>((allChecked, sourceKind) => {
        const graphDataOptions = getSourceKindGraphDataOptionKeys(sourceKind.name).reduce<Record<string, true>>(
            (allOptions, key) => {
                allOptions[key] = true;
                return allOptions;
            },
            {}
        );

        allChecked[sourceKind.id] = {
            selected: true,
            ...(Object.keys(graphDataOptions).length > 0 ? { options: graphDataOptions } : {}),
        };

        return allChecked;
    }, {});
};

export const GraphDataCheckboxes: FC<{
    checkedSourceKinds: number[];
    checkedRelationships: string[];
    disabled?: boolean;
    onChange: (checked: GraphDataSelections) => void;
}> = ({ checkedSourceKinds, checkedRelationships, disabled = true, onChange }) => {
    const { data: sourceKinds, isLoading, isSuccess } = useSourceKindsQuery();
    const checked = getGraphDataChecked(
        {
            sourceKinds: checkedSourceKinds,
            relationships: checkedRelationships,
        },
        sourceKinds
    );

    // Feature disabled is passed in prop or if query fails
    const isDisabled = disabled || !isSuccess;
    const amountChecked = isSuccess ? getAllGraphDataSelectionAmount(checked, sourceKinds) : 'none';
    const notifyChange = (nextChecked: GraphDataChecked) => onChange(getGraphDataSelections(nextChecked, sourceKinds));

    // If all boxes are checked, they are all unchecked; other wise all boxes are checked
    const toggleAllChecked = () => {
        if (sourceKinds) {
            const nextChecked = amountChecked === 'all' ? {} : getAllGraphDataChecked(sourceKinds);

            notifyChange(nextChecked);
        }
    };

    // Toggle a source kind on or off, then update the set of checked boxes
    const toggleSourceKind = (sourceKind: SourceKind) => () => {
        const { graphDataOptionKeys, checkedCount, totalCount } = getSourceKindSelectionDetails(checked, sourceKind);
        const selectionAmount = getSelectionAmount(checkedCount, totalCount);
        const isChecked = selectionAmount !== 'all';
        const newChecked = setGraphDataOptionsChecked(
            setSourceKindChecked(checked, sourceKind.id, isChecked),
            sourceKind.id,
            graphDataOptionKeys,
            isChecked
        );

        notifyChange(newChecked);
    };

    const toggleGraphDataOption = (sourceKindId: number, key: string) => () => {
        if (isSourceKindChecked(checked, sourceKindId)) {
            return;
        }

        const newChecked = setGraphDataOptionChecked(
            checked,
            sourceKindId,
            key,
            !isGraphDataOptionChecked(checked, sourceKindId, key)
        );

        notifyChange(newChecked);
    };

    return (
        <div className='flex flex-col' data-testid='source-kinds-checkboxes'>
            <FormControlLabel
                label='All graph data'
                control={
                    <Checkbox
                        checked={amountChecked === 'all'}
                        disabled={isDisabled}
                        indeterminate={amountChecked === 'some'}
                        name='All GraphData'
                        onChange={toggleAllChecked}
                    />
                }
            />

            {isLoading && LOADING_CHECKBOXES}

            {isSuccess &&
                sourceKinds.map((sourceKind) => {
                    const { checkedCount, totalCount } = getSourceKindSelectionDetails(checked, sourceKind);
                    const selectionAmount = getSelectionAmount(checkedCount, totalCount);
                    const graphDataOptions = getSourceKindGraphDataOptions(sourceKind.name);

                    return (
                        <div className='flex flex-col' key={sourceKind.id}>
                            <FormControlLabel
                                className='pl-8'
                                control={
                                    <Checkbox
                                        checked={selectionAmount === 'all'}
                                        indeterminate={selectionAmount === 'some'}
                                        onChange={toggleSourceKind(sourceKind)}
                                        name={sourceKind.name}
                                        disabled={isDisabled}
                                    />
                                }
                                label={(KIND_LABEL_MAP[sourceKind.name] ?? sourceKind.name) + ' data'}
                            />

                            {graphDataOptions.map((graphDataOption) => (
                                <FormControlLabel
                                    className='pl-16'
                                    control={
                                        <Checkbox
                                            checked={isEffectiveGraphDataOptionChecked(
                                                checked,
                                                sourceKind.id,
                                                graphDataOption.key
                                            )}
                                            onChange={toggleGraphDataOption(sourceKind.id, graphDataOption.key)}
                                            name={graphDataOption.key}
                                            disabled={isDisabled || isSourceKindChecked(checked, sourceKind.id)}
                                        />
                                    }
                                    label={graphDataOption.label}
                                    key={`${sourceKind.id}-${graphDataOption.key}`}
                                />
                            ))}
                        </div>
                    );
                })}
        </div>
    );
};
