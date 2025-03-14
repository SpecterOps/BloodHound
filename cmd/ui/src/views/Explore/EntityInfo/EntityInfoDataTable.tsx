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
    getOpenExpandedPanelSections,
    searchbarActions,
    transformFlatGraphResponse,
    useExploreParams,
    useFeatureFlag,
} from 'bh-shared-ui';
import { useQuery } from 'react-query';
import { useDispatch } from 'react-redux';
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { putGraphData, putGraphError, saveResponseForExport, setGraphLoading } from 'src/ducks/explore/actions';
import { addSnackbar } from 'src/ducks/global/actions';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';

const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({
    id,
    label,
    queryKey,
    endpoint,
    countLabel,
    sections,
    parentLabels = [],
}) => {
    const dispatch = useDispatch();
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const { setExploreParams, expandedPanelSections } = useExploreParams();
    const { expandedSections, setExpandedSections, toggleSection } = useEntityInfoPanelContext();

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

    const setExpandedPanelSectionsParams = () => {
        const labelList = [...(parentLabels as string[]), label];

        setExploreParams({
            expandedPanelSections: labelList,
            searchType: 'relationship',
            relationshipQueryType: queryKey,
            relationshipQueryItemId: id,
        });
    };

    const handleOnChange = (isOpen: boolean) => {
        handleCurrentSectionToggle();
        handleSetGraph(isOpen);
    };

    const handleSetGraph = async (isOpen: boolean) => {
        if (!endpoint) {
            if (backButtonFlag?.enabled && isOpen) {
                setExpandedPanelSectionsParams();
            }
            return;
        }

        if (isOpen && countQuery.data?.count < NODE_GRAPH_RENDER_LIMIT) {
            abortEntitySectionRequest();
            if (backButtonFlag?.enabled) {
                setExpandedPanelSectionsParams();
                return;
            }
            dispatch(setGraphLoading(true));
            await endpoint({ type: 'graph' })
                .then((result) => {
                    const formattedData = transformFlatGraphResponse(result);

                    dispatch(saveResponseForExport(formattedData));
                    dispatch(putGraphData(result));
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

    const handleCurrentSectionToggle = () => {
        if (backButtonFlag?.enabled) {
            setExpandedSections(
                getOpenExpandedPanelSections(expandedPanelSections as string[], parentLabels as string[], label)
            );
        } else {
            toggleSection(label);
        }
    };

    const setNodeSearchParams = (item: SelectedNode) => {
        setExploreParams({
            primarySearch: item.id,
            searchType: 'node',
            exploreSearchTab: 'node',
        });
    };

    const setSourceNodeSelected = (item: SelectedNode) => {
        dispatch(
            searchbarActions.sourceNodeSelected({
                objectid: item.id,
                type: item.type,
                name: item.name,
            })
        );
    };

    const handleOnClick = (item: SelectedNode) => {
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
                        parentLabels={[...(parentLabels as string[]), label]}
                        {...nestedSection}
                    />
                ))}
        </EntityInfoCollapsibleSection>
    );
};

export default EntityInfoDataTable;
