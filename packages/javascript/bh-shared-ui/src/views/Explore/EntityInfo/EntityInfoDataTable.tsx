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
import { useQuery } from 'react-query';
import InfiniteScrollingTable from '../../../components/InfiniteScrollingTable';
import { NODE_GRAPH_RENDER_LIMIT } from '../../../constants';
import { useExploreParams } from '../../../hooks';
import { EntityInfoDataTableProps, SelectedNode, entityRelationshipEndpoints } from '../../../utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({
    id,
    label,
    queryType,
    countLabel,
    sections,
    parentLabels = [],
}) => {
    const { setExploreParams, expandedPanelSections } = useExploreParams();

    const endpoint = queryType ? entityRelationshipEndpoints[queryType] : undefined;
    const isExpandedPanelSection = (expandedPanelSections as string[]).includes(label);

    const countQuery = useQuery(
        ['relatedCount', label, id],
        () => {
            if (endpoint) {
                return endpoint({ id, skip: 0, limit: 128 });
            }
            if (sections)
                return Promise.all(
                    sections.map((section: EntityInfoDataTableProps) => {
                        const endpoint = section.queryType ? entityRelationshipEndpoints[section.queryType] : undefined;
                        return endpoint ? endpoint({ id, skip: 0, limit: 128 }) : Promise.resolve();
                    })
                );
            return Promise.reject('Invalid call data provided for relationship list query');
        },
        { refetchOnWindowFocus: false, retry: false }
    );
    const isUnderRenderLimit = countQuery.data?.count < NODE_GRAPH_RENDER_LIMIT;

    const removeExpandedPanelSectionParams = () => {
        setExploreParams({
            expandedPanelSections: parentLabels,
        });
    };

    const setParentExpandedSectionParam = () => {
        const labelList = [...(parentLabels as string[]), label];

        setExploreParams({
            expandedPanelSections: labelList,
        });
    };

    const setExpandedPanelSectionsParams = () => {
        const labelList = [...(parentLabels as string[]), label];

        setExploreParams({
            expandedPanelSections: labelList,
            ...(isUnderRenderLimit && {
                searchType: 'relationship',
                relationshipQueryType: queryType,
                relationshipQueryItemId: id,
            }),
        });
    };

    const handleOnChange = (isOpen: boolean) => {
        if (isOpen) handleSetGraph();
        else removeExpandedPanelSectionParams();
    };
    const handleSetGraph = async () => {
        if (!endpoint) {
            setParentExpandedSectionParam();
        } else {
            setExpandedPanelSectionsParams();
        }
    };

    const setNodeSearchParams = (item: SelectedNode) => {
        setExploreParams({
            primarySearch: item.id,
            searchType: 'node',
            exploreSearchTab: 'node',
        });
    };

    const handleOnClick = (item: SelectedNode) => {
        setNodeSearchParams(item);
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
            isExpanded={isExpandedPanelSection}
            isLoading={countQuery.isLoading}
            isError={countQuery.isError}
            error={countQuery.error}
            onChange={handleOnChange}>
            {endpoint && (
                <InfiniteScrollingTable
                    itemCount={count}
                    fetchDataCallback={(params: { skip: number; limit: number }) => endpoint({ id, ...params })}
                    onClick={handleOnClick}
                />
            )}
            {sections &&
                sections.map((nestedSection: EntityInfoDataTableProps, nestedIndex: number) => (
                    <EntityInfoDataTable
                        key={nestedIndex}
                        parentLabels={[...(parentLabels as string[]), label]}
                        {...nestedSection}
                    />
                ))}
        </EntityInfoCollapsibleSection>
    );
};

export default EntityInfoDataTable;
