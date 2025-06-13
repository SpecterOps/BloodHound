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

import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useCallback, useState } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { AppIcon } from '../../../../components/AppIcon';
import { apiClient, cn, useAppNavigate } from '../../../../utils';
import { getTagUrlValue, itemSkeletons } from '../../utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

type SelectorListProps = {
    tagId: string;
    memberId: string;
};

const SelectorList: React.FC<SelectorListProps> = ({ tagId, memberId }) => {
    const navigate = useAppNavigate();
    const { labelId } = useParams();
    const [menuOpen, setMenuOpen] = useState<{ [key: number]: boolean }>({});

    const memberInfoQuery = useQuery(['asset-group-member-info'], () => {
        return apiClient.getAssetGroupTagMemberInfo(tagId, memberId).then((res) => {
            return res.data.data['member'];
        });
    });

    const handleMenuClick = (index: number) => {
        setMenuOpen((prev) => ({
            ...prev,
            [index]: !prev[index], //Toggle only the clicked popover
        }));
    };

    const handleViewClick = useCallback(
        (id: number) => {
            navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`);
        },
        [tagId, navigate, labelId]
    );

    const handleEditClick = useCallback(
        (id: number) => {
            navigate(`/zone-management/save/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`);
        },
        [tagId, navigate, labelId]
    );

    if (memberInfoQuery.isLoading) {
        return itemSkeletons.map((skeleton, index) => {
            return skeleton('object-selector', index);
        });
    }
    if (memberInfoQuery.isError) {
        return (
            <li className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                <span className='text-base'>There was an error fetching this data</span>
            </li>
        );
    }

    if (memberInfoQuery.isSuccess) {
        return (
            <EntityInfoCollapsibleSection label='Selectors' count={memberInfoQuery.data.selectors?.length}>
                {memberInfoQuery.data.selectors?.map((selector, index) => {
                    return (
                        <div
                            className={cn('flex items-center gap-2 p-2 overflow-hidden', {
                                'bg-neutral-light-4 dark:bg-neutral-dark-4': index % 2 === 0,
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
                                        className='cursor-pointer p-2 hover:bg-neutral-light-4 hover:dark:bg-neutral-dark-4'
                                        onClick={() => {
                                            handleViewClick(selector.id);
                                        }}>
                                        View
                                    </div>
                                    <div
                                        className='cursor-pointer p-2 hover:bg-neutral-light-4 hover:dark:bg-neutral-dark-4'
                                        onClick={() => {
                                            handleEditClick(selector.id);
                                        }}>
                                        Edit
                                    </div>
                                </PopoverContent>
                            </Popover>
                            <div className='truncate max-w-[350px]' title={selector.name}>
                                {selector.name}
                            </div>
                        </div>
                    );
                })}
            </EntityInfoCollapsibleSection>
        );
    }
};

export default SelectorList;
