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

import { Button } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';
import { useQuery } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { AppIcon, CreateMenu } from '../../components';
import { ROUTE_TIER_MANAGEMENT_CREATE, ROUTE_TIER_MANAGEMENT_EDIT } from '../../routes';
import { apiClient } from '../../utils';
import { DetailsList } from './DetailsList';

const innerDetail = (
    selectedObject: number | null,
    selectedSelector: number | null,
    selectedTier: number
): SelectedDetailsProps => {
    if (selectedObject !== null) return { id: selectedObject, type: 'object' };

    if (selectedSelector !== null) return { id: selectedSelector, type: 'selector' };

    return { id: selectedTier, type: 'tier' };
};

type SelectedDetailsProps = {
    type: 'tier' | 'label' | 'selector' | 'object';
    id: number;
    cypher?: boolean;
};

const SelectedDetails: FC<SelectedDetailsProps> = ({ type, id, cypher }) => {
    if (type === 'object')
        return (
            <>
                <div>{`Object Info Panel - ${type}-${id}`}</div>
            </>
        );

    if (type === 'selector') {
        if (cypher)
            return (
                <>
                    <div>{`Dynamic Details - ${type}-${id}`}</div>
                    <div>Cypher Input</div>
                </>
            );
        else
            return (
                <>
                    <div>{`Dynamic Details - ${type}-${id}`}</div>
                    <div>Object Count Panel</div>
                </>
            );
    }

    return (
        <>
            <div>{`Dynamic Details - ${type}-${id}`}</div>
            <div>Object Count Panel</div>
        </>
    );
};

const Details: FC = () => {
    const [selectedTier, setSelectedTier] = useState(1);
    const [selectedSelector, setSelectedSelector] = useState<number | null>(null);
    const [selectedObject, setSelectedObject] = useState<number | null>(null);
    const [showCypher, setShowCypher] = useState(false);
    const navigate = useNavigate();

    const labelsQuery = useQuery(
        ['asset-group-labels'],
        () => {
            return apiClient.getAssetGroupLabels().then((res) => {
                return res.data.data['asset_group_labels'];
            });
        },
        { cacheTime: Infinity, keepPreviousData: true, staleTime: Infinity }
    );

    const selectorsQuery = useQuery(
        ['asset-group-selectors', selectedTier],
        () => {
            return apiClient.getAssetGroupSelectors(selectedTier).then((res) => {
                return res.data.data['selectors'];
            });
        },
        { cacheTime: Infinity, keepPreviousData: true, staleTime: Infinity }
    );

    const objectsQuery = useQuery(
        ['asset-group-members', selectedTier, selectedSelector],
        async () => {
            if (selectedSelector === null)
                return apiClient.getAssetGroupLabelMembers(selectedTier).then((res) => {
                    return res.data.data['members'];
                });

            return apiClient.getAssetGroupSelectorMembers(selectedTier, selectedSelector).then((res) => {
                return res.data.data['members'];
            });
        },
        { cacheTime: Infinity, keepPreviousData: true, staleTime: Infinity }
    );

    const disableEditButton =
        selectedObject !== null ||
        (selectorsQuery.isLoading && labelsQuery.isLoading) ||
        (selectorsQuery.isError && labelsQuery.isError);

    const { type, id } = innerDetail(selectedObject, selectedSelector, selectedTier);

    return (
        <div>
            <div className='flex mt-6'>
                <div className='flex justify-around basis-2/3'>
                    <div className='flex justify-start gap-4 items-center basis-2/3'>
                        <CreateMenu
                            createMenuTitle='Create'
                            menuItems={[
                                {
                                    title: 'Tier',
                                    onClick: () => {
                                        navigate(ROUTE_TIER_MANAGEMENT_CREATE, { state: { type: 'Tier' } });
                                    },
                                },
                                {
                                    title: 'Label',
                                    onClick: () => {
                                        navigate(ROUTE_TIER_MANAGEMENT_CREATE, { state: { type: 'Label' } });
                                    },
                                },
                                {
                                    title: 'Selector',
                                    onClick: () => {
                                        navigate(ROUTE_TIER_MANAGEMENT_CREATE, {
                                            state: { type: 'Selector', within: selectedTier },
                                        });
                                    },
                                },
                            ]}
                        />
                        <div className='flex items-center align-middle'>
                            <div>
                                <AppIcon.Info className='mr-4 ml-2' size={24} />
                            </div>
                            <span>
                                To create additional tiers{' '}
                                <Button
                                    variant='text'
                                    asChild
                                    className='p-0 text-base text-secondary dark:text-secondary-variant-2'>
                                    <a href='#'>contact sales</a>
                                </Button>{' '}
                                in order to upgrade for multi-tier analysis.
                            </span>
                        </div>
                    </div>
                    <div className='flex justify-start basis-1/3'>
                        <input type='text' placeholder='search' className='hidden' />
                    </div>
                </div>

                <div className='basis-1/3'>
                    <Button
                        onClick={() => {
                            navigate(ROUTE_TIER_MANAGEMENT_EDIT, { state: { type: type, id: id } });
                        }}
                        variant={'secondary'}
                        disabled={disableEditButton}>
                        Edit
                    </Button>
                </div>
            </div>
            <div className='flex gap-8 mt-4'>
                <div className='flex basis-2/3 bg-neutral-light-2 dark:bg-neutral-dark-2 rounded-lg'>
                    <div className='grow min-h-96'>
                        <DetailsList
                            title='Tiers'
                            listQuery={labelsQuery}
                            selected={selectedTier}
                            onSelect={(id: number) => {
                                setSelectedTier(id);
                                setSelectedSelector(null);
                                setSelectedObject(null);
                                setShowCypher(false);
                            }}
                        />
                    </div>
                    <div className='border-neutral-light-3 dark:border-neutral-dark-3 grow min-h-96'>
                        <DetailsList
                            title='Selectors'
                            listQuery={selectorsQuery}
                            selected={selectedSelector}
                            onSelect={(id) => {
                                setSelectedSelector(id);

                                const selected = selectorsQuery.data?.find((item) => {
                                    return item.id === id;
                                });
                                if (selected?.seeds?.[0].type === 1) setShowCypher(true);
                                else setShowCypher(false);

                                setSelectedObject(null);
                            }}
                            sortable
                        />
                    </div>
                    <div className='grow min-h-96'>
                        <DetailsList
                            title='Objects'
                            listQuery={objectsQuery}
                            selected={selectedObject}
                            onSelect={(id) => {
                                setSelectedObject(id);
                                setShowCypher(false);
                            }}
                            sortable
                            nodeIcon
                        />
                    </div>
                </div>
                <div className='flex flex-col basis-1/3'>
                    <SelectedDetails type={type} id={id} cypher={showCypher} />
                </div>
            </div>
        </div>
    );
};

export default Details;
