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
import { useSearchParams } from 'react-router-dom';
import { EntityInfoDataTableProps, entityRelationshipEndpoints } from '../../utils';
import EntityInfoCollapsibleSection from '../EntityInfo/EntityInfoCollapsibleSection';
import InfiniteScrollingTable from '../InfiniteScrollingTable';

export const EntityInfoDataTable: React.FC<EntityInfoDataTableProps> = ({
    id,
    label,
    queryType,
    countLabel,
    sections,
    parentLabels = [],
}) => {
    const [searchParams, setSearchParams] = useSearchParams();

    const endpoint = queryType ? entityRelationshipEndpoints[queryType] : undefined;
    const isExpandedPanelSection = (searchParams.getAll('expandedPanelSections') as string[]).includes(label);

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

    const removeExpandedPanelSectionParams = () => {
        const params = new URLSearchParams(searchParams);
        params.delete('expandedPanelSections');
        setSearchParams(params);
    };

    const setParentExpandedSectionParam = () => {
        const labelList = [...(parentLabels as string[]), label];

        setSearchParams({ expandedPanelSections: labelList });
    };

    const setExpandedPanelSectionsParams = () => {
        const params = new URLSearchParams(searchParams);
        const labelList = [...(parentLabels as string[]), label];
        labelList.forEach((section) => params.set('expandedPanelSections', section));

        setSearchParams(params);
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
            count={count}
            error={countQuery.error}
            isError={countQuery.isError}
            isExpanded={isExpandedPanelSection}
            isLoading={countQuery.isLoading}
            label={label}
            onChange={handleOnChange}>
            {endpoint && (
                <InfiniteScrollingTable
                    itemCount={count}
                    fetchDataCallback={(params: { skip: number; limit: number }) => endpoint({ id, ...params })}
                />
            )}
            {sections &&
                sections.map((nestedSection: EntityInfoDataTableProps, nestedIndex: number) => (
                    <EntityInfoDataTable
                        {...nestedSection}
                        data-testid='entity-info-data-table'
                        key={nestedIndex}
                        parentLabels={[...(parentLabels as string[]), label]}
                    />
                ))}
        </EntityInfoCollapsibleSection>
    );
};
