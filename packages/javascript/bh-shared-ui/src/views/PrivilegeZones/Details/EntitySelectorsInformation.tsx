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

import { Popover, PopoverContent, PopoverTrigger, Skeleton } from '@bloodhoundenterprise/doodleui';
import { useCallback, useState } from 'react';
import { AppIcon } from '../../../components';
import EntityInfoCollapsibleSection from '../../../components/EntityInfo/EntityInfoCollapsibleSection';
import { useExploreParams, useMemberInfo, usePZPathParams, usePZQueryParams } from '../../../hooks';
import { cn, useAppNavigate } from '../../../utils';

const EntitySelectorsInformation: React.FC = () => {
    const navigate = useAppNavigate();
    const [menuOpen, setMenuOpen] = useState<{ [key: number]: boolean }>({});

    const { setExploreParams, expandedPanelSections, selectedItem: selected } = useExploreParams();
    const { tagId: pathTagId, memberId: pathMemberId, ruleDetailsLink, ruleEditLink } = usePZPathParams();
    const { assetGroupTagId: queryTagId } = usePZQueryParams();

    const assetGroupTagId = pathTagId ? pathTagId : queryTagId;
    const selectedItem = pathMemberId ? pathMemberId : selected;

    const isExpandedPanelSection = expandedPanelSections?.includes('Rules');

    const handleOnChange = () => {
        isExpandedPanelSection
            ? setExploreParams({
                  expandedPanelSections: [],
              })
            : setExploreParams({
                  expandedPanelSections: ['Rules'],
              });
    };

    const memberInfoQuery = useMemberInfo(assetGroupTagId?.toString() ?? '', selectedItem ?? '');

    const handleMenuClick = (index: number) => {
        setMenuOpen((prev) => ({
            ...prev,
            [index]: !prev[index], //Toggle only the clicked popover
        }));
    };

    const handleViewClick = useCallback(
        (id: number) => {
            if (!assetGroupTagId) return;
            navigate(ruleDetailsLink(assetGroupTagId, id));
        },
        [assetGroupTagId, navigate, ruleDetailsLink]
    );

    const handleEditClick = useCallback(
        (id: number) => {
            if (!assetGroupTagId) return;
            navigate(ruleEditLink(assetGroupTagId, id));
        },
        [assetGroupTagId, navigate, ruleEditLink]
    );

    if (memberInfoQuery.isLoading) {
        return <Skeleton className='h-10 w-full' />;
    }

    if (memberInfoQuery.isError) {
        return <span className='text-error'>There was an error fetching this data</span>;
    }

    if (memberInfoQuery.isSuccess) {
        return (
            <>
                <EntityInfoCollapsibleSection
                    label='Rules'
                    count={memberInfoQuery.data.selectors?.length}
                    isExpanded={!!isExpandedPanelSection}
                    onChange={handleOnChange}>
                    {memberInfoQuery.data.selectors?.map((selector, index) => {
                        return (
                            <div
                                className={cn('flex items-center gap-2 p-2 overflow-hidden', {
                                    'bg-neutral-4': index % 2 === 0,
                                })}
                                key={index}>
                                <Popover open={!!menuOpen[index]}>
                                    <PopoverTrigger asChild>
                                        <button onClick={() => handleMenuClick(index)}>
                                            <AppIcon.VerticalEllipsis />
                                        </button>
                                    </PopoverTrigger>
                                    <PopoverContent
                                        className='w-80 px-4 py-2 flex flex-col gap-2'
                                        onInteractOutside={() => setMenuOpen({})}
                                        onEscapeKeyDown={() => setMenuOpen({})}>
                                        <div
                                            className='cursor-pointer p-2 hover:bg-neutral-4'
                                            onClick={() => {
                                                handleViewClick(selector.id);
                                            }}>
                                            View
                                        </div>
                                        <div
                                            className='cursor-pointer p-2 hover:bg-neutral-4'
                                            onClick={() => {
                                                handleEditClick(selector.id);
                                            }}>
                                            Edit
                                        </div>
                                    </PopoverContent>
                                </Popover>
                                <div className='truncate' title={selector.name}>
                                    {selector.name}
                                </div>
                            </div>
                        );
                    })}
                </EntityInfoCollapsibleSection>
            </>
        );
    }
};

export default EntitySelectorsInformation;
