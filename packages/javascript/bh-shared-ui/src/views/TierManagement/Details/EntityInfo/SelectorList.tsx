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
import { useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../../components/AppIcon';
import { apiClient, cn } from '../../../../utils';
import { itemSkeletons } from '../utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

type SelectorListProps = {
    selectedTag: number;
    selectedObject: number;
};

const SelectorList: React.FC<SelectorListProps> = ({ selectedTag, selectedObject }) => {
    const [menuOpen, setMenuOpen] = useState<{ [key: number]: boolean }>({});

    const selectorsQuery = useQuery(['asset-group-member-info'], () => {
        return apiClient.getAssetGroupLabelMemberInfo(selectedTag, selectedObject).then((res) => {
            return res.data.data['member'];
        });
    });

    const handleMenuClick = (index: number) => {
        setMenuOpen((prev) => ({
            ...prev,
            [index]: !prev[index], //Toggle only the clicked popover
        }));
    };

    const handleViewClick = () => {};
    const handleEditClick = () => {};

    if (selectorsQuery.isLoading) {
        return itemSkeletons.map((skeleton, index) => {
            return skeleton('object-selector', index);
        });
    }
    if (selectorsQuery.isError) {
        return (
            <li className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                <span className='text-base'>There was an error fetching this data</span>
            </li>
        );
    }

    if (selectorsQuery.isSuccess) {
        return (
            <EntityInfoCollapsibleSection label='Selectors' count={selectorsQuery.data.selectors?.length}>
                {selectorsQuery.data.selectors?.map((selector, index) => {
                    return (
                        <div
                            className={cn('flex items-center gap-2 p-2', {
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
                                        onClick={handleViewClick}>
                                        View
                                    </div>
                                    <div
                                        className='cursor-pointer p-2 hover:bg-neutral-light-4 hover:dark:bg-neutral-dark-4'
                                        onClick={handleEditClick}>
                                        Edit
                                    </div>
                                </PopoverContent>
                            </Popover>
                            {selector.name}
                        </div>
                    );
                })}
            </EntityInfoCollapsibleSection>
        );
    }
};

export default SelectorList;
