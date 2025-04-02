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
import { AssetGroupTag, AssetGroupTagSelector, AssetGroupTagSelectorNode } from 'js-client-library';
import { FC, useState } from 'react';
import { useQuery } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { AppIcon, CreateMenu } from '../../../components';
import { ROUTE_TIER_MANAGEMENT_CREATE, ROUTE_TIER_MANAGEMENT_EDIT } from '../../../routes';
import { apiClient } from '../../../utils';
import { Cypher } from '../Cypher';
import { DetailsList } from './DetailsList';
import DynamicDetails from './DynamicDetails';
import EntityInfoPanel from './EntityInfo/EntityInfoPanel';
import { MembersList } from './MembersList';
import ObjectCountPanel from './ObjectCountPanel';

type SelectedDetailsProps = {
    cypher?: boolean;
    selectedTierId: number;
    selectedSelectorId: number | null;
    selectedObjectId: number | null;
    selectedTier: AssetGroupTag | undefined;
    selectedSelector: AssetGroupTagSelector | undefined;
    selectedObject: AssetGroupTagSelectorNode | null;
};

const isObject = (data: any): data is AssetGroupTagSelectorNode => {
    const objectData = data || {};
    return 'node_id' in objectData;
};

const SelectedDetails: FC<SelectedDetailsProps> = ({
    cypher,
    selectedTierId,
    selectedSelectorId,
    selectedObjectId,
    selectedTier,
    selectedSelector,
    selectedObject,
}) => {
    if (selectedObjectId !== null) {
        if (isObject(selectedObject)) {
            return (
                <EntityInfoPanel
                    selectedNode={selectedObject}
                    selectedTag={selectedTierId}
                    selectedObject={selectedObjectId}
                />
            );
        }
    }

    if (selectedSelectorId !== null) {
        if (cypher)
            return (
                <>
                    <DynamicDetails data={selectedSelector} isCypher={cypher} />
                    <Cypher />
                </>
            );
        else
            return (
                <>
                    <DynamicDetails data={selectedSelector} />
                    <ObjectCountPanel selectedTier={selectedTierId} />
                </>
            );
    }

    return (
        <>
            <DynamicDetails data={selectedTier} />
            <ObjectCountPanel selectedTier={selectedTierId} />
        </>
    );
};

const Details: FC = () => {
    const [selectedTag, setSelectedTag] = useState(1);
    const [selectedSelector, setSelectedSelector] = useState<number | null>(null);
    const [selectedObject, setSelectedObject] = useState<number | null>(null);
    const [selectedObjectData, setSelectedObjectData] = useState<AssetGroupTagSelectorNode | null>(null);
    const [showCypher, setShowCypher] = useState(false);
    const navigate = useNavigate();

    const labelsQuery = useQuery(['asset-group-labels'], () => {
        return apiClient.getAssetGroupLabels().then((res) => {
            return res.data.data['asset_group_labels'];
        });
    });

    const selectorsQuery = useQuery(['asset-group-selectors', selectedTag], () => {
        return apiClient.getAssetGroupSelectors(selectedTag).then((res) => {
            return res.data.data['selectors'];
        });
    });

    const objectsQuery = useQuery(['asset-group-members', selectedTag, selectedSelector], async () => {
        if (selectedSelector === null)
            return apiClient.getAssetGroupLabelMembers(selectedTag, 0, 1).then((res) => {
                return res.data.count;
            });

        return apiClient.getAssetGroupSelectorMembers(selectedTag, selectedSelector, 0, 1).then((res) => {
            return res.data.count;
        });
    });

    const disableEditButton =
        selectedObject !== null ||
        (selectorsQuery.isLoading && labelsQuery.isLoading) ||
        (selectorsQuery.isError && labelsQuery.isError);

    return (
        <div>
            <div className='flex mt-6'>
                <div className='flex justify-around basis-2/3'>
                    <div className='flex justify-start gap-4 items-center basis-2/3 invisible'>
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
                                            state: { type: 'Selector', within: selectedTag },
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
                            navigate(ROUTE_TIER_MANAGEMENT_EDIT);
                        }}
                        variant={'secondary'}
                        disabled={disableEditButton}>
                        Edit
                    </Button>
                </div>
            </div>
            <div className='flex gap-8 mt-4 grow-1'>
                <div className='flex basis-2/3 bg-neutral-light-2 dark:bg-neutral-dark-2 rounded-lg max-h-[560px] *:grow-0 *:basis-1/3 *:min-h-96'>
                    <div className='max-h-[518px]'>
                        <DetailsList
                            title='Tiers'
                            listQuery={labelsQuery}
                            selected={selectedTag}
                            onSelect={(id: number) => {
                                setSelectedTag(id);
                                setSelectedSelector(null);
                                setSelectedObject(null);
                                setShowCypher(false);
                            }}
                        />
                    </div>
                    <div className='border-neutral-light-3 dark:border-neutral-dark-3 max-h-[518px]'>
                        <DetailsList
                            title='Selectors'
                            listQuery={selectorsQuery}
                            selected={selectedSelector}
                            onSelect={(id: number | null) => {
                                setSelectedSelector(id);

                                const selected = selectorsQuery.data?.find((item) => {
                                    return item.id === id;
                                });
                                if (selected?.seeds?.[0].type === 1) setShowCypher(true);
                                else setShowCypher(false);

                                setSelectedObject(null);
                            }}
                        />
                    </div>
                    <div>
                        <MembersList
                            itemCount={objectsQuery.data || 1000}
                            onClick={(id, data) => {
                                setSelectedObject(id);
                                setShowCypher(false);
                                setSelectedObjectData(data);
                            }}
                            selected={selectedObject}
                            selectedSelector={selectedSelector}
                            selectedTag={selectedTag}
                        />
                    </div>
                </div>
                <div className='flex flex-col basis-1/3'>
                    <SelectedDetails
                        cypher={showCypher}
                        selectedTierId={selectedTag}
                        selectedSelectorId={selectedSelector}
                        selectedObjectId={selectedObject}
                        selectedTier={labelsQuery.data?.find((tag) => tag.id === selectedTag)}
                        selectedSelector={selectorsQuery.data?.find((selector) => selector.id === selectedSelector)}
                        selectedObject={selectedObjectData}
                    />
                </div>
            </div>
        </div>
    );
};

export default Details;
