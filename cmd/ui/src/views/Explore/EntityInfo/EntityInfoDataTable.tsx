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
import { SelectedNode } from 'src/ducks/entityinfo/types';
import { putGraphData, putGraphError, saveResponseForExport, setGraphLoading } from 'src/ducks/explore/actions';
import { addSnackbar } from 'src/ducks/global/actions';
import { transformFlatGraphResponse } from 'src/utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';

const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({
    id,
    label,
    queryKey,
    endpoint,
    countLabel,
    sections,
    sectionsMapper,
}) => {
    const dispatch = useDispatch();
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const { setExploreParams } = useExploreParams();
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

    const findLabelLocation = () => {
        const filteredArray: string[] = sectionsMapper.find((nestedArray: string[]) =>
            nestedArray.includes(label)
        ) as string[];
        const index = filteredArray.indexOf(label);
        return { filteredArray, index };
    };

    const setExpandedRelationshipsParams = () => {
        const expandedRelationshipHelperArray: string[] = [];

        const { filteredArray, index } = findLabelLocation();
        expandedRelationshipHelperArray.push(label);
        if (index > 0) expandedRelationshipHelperArray.unshift(filteredArray[0]);

        setExploreParams({
            expandedRelationships: expandedRelationshipHelperArray,
            searchType: 'relationship',
            relationshipQueryType: queryKey,
            relationshipQueryObjectId: id,
        });
    };

    const handleOnChange = (label: string, isOpen: boolean) => {
        handleCurrentSectionToggle();
        handleSetGraph(label, isOpen);
    };

    const handleSetGraph = async (label: string, isOpen: boolean) => {
        if (!endpoint) {
            if (backButtonFlag?.enabled && isOpen) {
                setExpandedRelationshipsParams();
            }
            return;
        }

        if (isOpen && countQuery.data?.count < NODE_GRAPH_RENDER_LIMIT) {
            abortEntitySectionRequest();
            if (backButtonFlag?.enabled) {
                setExpandedRelationshipsParams();
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

    const isParentOfLabel = (key: string) => {
        const { filteredArray, index } = findLabelLocation();
        if (index > 0) {
            return key === filteredArray[0];
        }
        return false;
    };

    const collapseOtherSections = () => {
        for (const [key] of Object.entries(expandedSections)) {
            const isNotObjectInformation = key !== 'Object Information';
            const isNotParentOfSection = !isParentOfLabel(key);
            const isNotClickedSection = key !== label; // to not interfere with normal toggle flow
            if (isNotObjectInformation && isNotParentOfSection && isNotClickedSection) {
                expandedSections[key] = false;
            }
        }
    };

    const handleCurrentSectionToggle = () => {
        if (backButtonFlag?.enabled) {
            collapseOtherSections(); // We want to keep only one open at a time
        }
        toggleSection(label);
    };

    const setNodeSearchParams = (item: SelectedNode) => {
        setExploreParams({
            primarySearch: item.id ?? item.name,
            searchType: 'node',
            searchTab: 'node',
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
                    <EntityInfoDataTable sectionsMapper={sectionsMapper} key={nestedSectionIndex} {...nestedSection} />
                ))}
        </EntityInfoCollapsibleSection>
    );
};

export default EntityInfoDataTable;
