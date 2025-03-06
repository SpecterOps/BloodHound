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

import {
    EntityInfoDataTableProps,
    InfiniteScrollingTable,
    NODE_GRAPH_RENDER_LIMIT,
    abortEntitySectionRequest,
    searchbarActions,
    useExploreParams,
    useFeatureFlag,
} from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { useDispatch } from 'react-redux';
import { putGraphData, putGraphError, saveResponseForExport, setGraphLoading } from 'src/ducks/explore/actions';
import { addSnackbar } from 'src/ducks/global/actions';
import { transformFlatGraphResponse } from 'src/utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';

const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({
    id,
    label,
    endpoint,
    countLabel,
    sections,
    parentSectionIndex,
}) => {
    const dispatch = useDispatch();
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const { setExploreParams, expandedRelationships } = useExploreParams();
    const { expandedSections, toggleSection } = useEntityInfoPanelContext();

    const countQuery = useQuery(
        ['relatedCount', label, id],
        () => {
            if (endpoint) {
                return endpoint({ skip: 0, limit: 128 });
            }
            if (sections) return Promise.all(sections.map((section) => section.endpoint?.({ skip: 0, limit: 128 })));
            return Promise.reject('Invalid call data provided for relationship list query');
        },
        { refetchOnWindowFocus: false, retry: false }
    );

    const setExpandedRelationshipsParams = () => {
        let expandedRelationshipHelperArray: string[] = [];
        const expandedRelationshipsLength = (expandedRelationships as string[]).length;

        if (!expandedRelationshipsLength) {
            expandedRelationshipHelperArray = [`${parentSectionIndex}-${label}`];
        } else {
            const parentIndexOfNested = parseInt((expandedRelationships as string[])?.at(-1)?.split('-')[0] as string);
            const isNestedSameSection = parentIndexOfNested === parentSectionIndex;
            if (expandedRelationshipsLength === 1) {
                if (isNestedSameSection) {
                    expandedRelationshipHelperArray = [
                        ...(expandedRelationships as string[]),
                        `${parentSectionIndex}-${label}`,
                    ];
                } else {
                    expandedRelationshipHelperArray = [`${parentSectionIndex}-${label}`];
                }
            }
            if (expandedRelationshipsLength >= 2) {
                if (isNestedSameSection) {
                    expandedRelationships?.pop();
                    expandedRelationshipHelperArray = [
                        ...(expandedRelationships as string[]),
                        `${parentSectionIndex}-${label}`,
                    ];
                } else {
                    expandedRelationshipHelperArray = [`${parentSectionIndex}-${label}`];
                }
            }
        }

        setExploreParams({
            expandedRelationships: expandedRelationshipHelperArray,
            searchType: 'relationship',
        });
    };

    const handleOnChange = (label: string, isOpen: boolean) => {
        if (backButtonFlag?.enabled && isOpen && !endpoint) {
            setExpandedRelationshipsParams();
        }
        handleCurrentSectionToggle();
        handleSetGraph(label, isOpen);
    };

    const handleSetGraph = async (label: string, isOpen: boolean) => {
        if (!endpoint) {
            return;
        }

        if (isOpen && countQuery.data?.count < NODE_GRAPH_RENDER_LIMIT) {
            abortEntitySectionRequest();

            dispatch(setGraphLoading(true));

            await endpoint({ type: 'graph' })
                .then((result) => {
                    const formattedData = transformFlatGraphResponse(result);

                    dispatch(saveResponseForExport(formattedData));
                    backButtonFlag?.enabled ? setExpandedRelationshipsParams() : dispatch(putGraphData(result));
                })
                .catch((err) => {
                    if (err?.code === 'ERR_CANCELED') {
                        return;
                    }
                    dispatch(putGraphError(err));
                    dispatch(addSnackbar('Query failed. Please try again.', 'nodeRelationshipGraphQuery', {}));
                })
                .finally(() => {
                    dispatch(setGraphLoading(false));
                });
        }
    };

    const isParentSection = (key: string) => {
        // for it to be a parent key/label that is being checked needs to be label of 0 index of expanded relationships nad needs to have the index same as the current section index we are evaluating
        const checkKey = expandedRelationships?.at(0)?.split('-')[1] === key;
        const checkIndex = expandedRelationships?.at(0)?.split('-')[0] == parentSectionIndex;
        return checkKey && checkIndex;
    };

    const handleCurrentSectionToggle = () => {
        if (backButtonFlag?.enabled) {
            if (expandedSections && (expandedRelationships as string[]).length > 0) {
                for (const [key] of Object.entries(expandedSections)) {
                    // Closes if the key that is being evaluated is not a direct parent, is not object information and its not the same label that we are trying to open
                    if (key !== 'Object Information' && !isParentSection(key) && key !== label) {
                        expandedSections[key] = false;
                    }
                }
            }
        }
        toggleSection(label);
    };

    const setNodeSearchParams = (item: any) => {
        setExploreParams({
            primarySearch: item.id ?? item.name,
            searchType: 'node',
            searchTab: 'node',
        });
    };

    const setSourceNodeSelected = (item: any) => {
        dispatch(
            searchbarActions.sourceNodeSelected({
                objectid: item.id,
                type: item.type,
                name: item.name,
            })
        );
    };

    const handleOnClick = (item: any) => {
        if (backButtonFlag?.enabled) {
            setNodeSearchParams(item);
        } else {
            setSourceNodeSelected(item);
        }
    };

    let count: number | undefined;
    if (Array.isArray(countQuery.data)) {
        if (countLabel !== undefined) {
            countQuery.data.forEach((sectionData: any) => {
                if (sectionData.countLabel === countLabel) count = sectionData.count;
            });
        } else {
            count = countQuery.data.reduce((acc, val) => {
                const count = val?.count ?? 0;
                return acc + count;
            }, 0);
        }
    } else if (countQuery.data) {
        count = countQuery.data?.count ?? 0;
    }

    return (
        <EntityInfoCollapsibleSection
            label={label}
            count={count}
            isExpanded={!!expandedSections[label]}
            isLoading={countQuery.isLoading}
            isError={countQuery.isError}
            error={countQuery.error}
            onChange={handleOnChange}>
            {endpoint && (
                <InfiniteScrollingTable itemCount={count} fetchDataCallback={endpoint} onClick={handleOnClick} />
            )}
            {sections &&
                sections.map((nestedSection, nestedSectionIndex) => (
                    <EntityInfoDataTable
                        key={nestedSectionIndex}
                        parentSectionIndex={parentSectionIndex}
                        {...nestedSection}
                    />
                ))}
        </EntityInfoCollapsibleSection>
    );
};

export default EntityInfoDataTable;
