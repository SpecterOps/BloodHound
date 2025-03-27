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
import { ActiveDirectoryNodeKind } from '../../../graphSchema';
import { ROUTE_TIER_MANAGEMENT_CREATE, ROUTE_TIER_MANAGEMENT_EDIT } from '../../../routes';
import { apiClient } from '../../../utils';
import { DetailsList } from './DetailsList';
import DynamicDetails from './DynamicDetails';
import EntityInfoPanel from './EntityInfo/EntityInfoPanel';
import { MembersList } from './MembersList';
import ObjectCountPanel from './ObjectCountPanel';

const innerDetail = (
    selectedObjectId: number | null,
    selectedSelectorId: number | null,
    selectedLabelId: number,
    selectedObjectData: AssetGroupTagSelectorNode | null,
    labelsList: AssetGroupTag[],
    selectorsList: AssetGroupTagSelector[]
): SelectedDetailsProps => {
    if (selectedObjectId !== null && selectedObjectData !== null) {
        return { id: selectedObjectId, type: 'object', data: selectedObjectData };
    }

    if (selectedSelectorId !== null) {
        const selectedSelector = selectorsList.find((object) => object.id === selectedSelectorId);
        return { id: selectedSelectorId, type: 'selector', data: selectedSelector };
    }

    const selectedLabel = labelsList.find((object) => object.id === selectedLabelId);
    return { id: selectedLabelId, type: 'tier', data: selectedLabel };
};

type SelectedDetailsProps = {
    type: 'tier' | 'label' | 'selector' | 'object';
    id: number;
    data?: AssetGroupTagSelectorNode | AssetGroupTag | AssetGroupTagSelector;
    cypher?: boolean;
};

const isObject = (data: any): data is AssetGroupTagSelectorNode => {
    const objectData = data || {};
    return 'node_id' in objectData;
};

const SelectedDetails: FC<SelectedDetailsProps> = ({ type, id, cypher, data }) => {
    if (isObject(data)) {
        const selectedNode = {
            id: '3',
            type: ActiveDirectoryNodeKind.User,
            //properties: data.properties
            properties: {},
            name: data.name,
        };
        return (
            <>
                <EntityInfoPanel selectedNode={selectedNode} />
            </>
        );
    }

    if (type === 'selector') {
        if (cypher)
            return (
                <>
                    <DynamicDetails data={data} isCypher={cypher} />
                    <div>Cypher Input</div>
                </>
            );
        else
            return (
                <>
                    <DynamicDetails data={data} />
                    {/* <ObjectCountPanel data={objectsCount} /> */}
                </>
            );
    }

    return (
        <>
            <DynamicDetails data={data} />
            <ObjectCountPanel selectedTier={id} />
        </>
    );
};

const Details: FC = () => {
    const [selectedTier, setSelectedTier] = useState(1);
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

    const selectorsQuery = useQuery(['asset-group-selectors', selectedTier], () => {
        return apiClient.getAssetGroupSelectors(selectedTier).then((res) => {
            return res.data.data['selectors'];
        });
    });

    const objectsQuery = useQuery(['asset-group-members', selectedTier, selectedSelector], async () => {
        if (selectedSelector === null)
            return apiClient.getAssetGroupLabelMembers(selectedTier, 0, 1).then((res) => {
                return res.data.count;
            });

        return apiClient.getAssetGroupSelectorMembers(selectedTier, selectedSelector, 0, 1).then((res) => {
            return res.data.count;
        });
    });

    const disableEditButton =
        selectedObject !== null ||
        (selectorsQuery.isLoading && labelsQuery.isLoading) ||
        (selectorsQuery.isError && labelsQuery.isError);

    const { type, id, data } = innerDetail(
        selectedObject,
        selectedSelector,
        selectedTier,
        selectedObjectData,
        labelsQuery.data || [],
        selectorsQuery.data || []
    );

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
                    <div className='min-h-96 grow-0 basis-1/3'>
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
                    <div className='border-neutral-light-3 dark:border-neutral-dark-3 min-h-96 grow-0 basis-1/3'>
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
                    <div className='min-h-96 grow-0 basis-1/3'>
                        <MembersList
                            itemCount={objectsQuery.data || 1000}
                            onClick={(id, data) => {
                                setSelectedObject(id);
                                setShowCypher(false);
                                setSelectedObjectData(data);
                            }}
                            selected={selectedObject}
                            selectedSelector={selectedSelector}
                            selectedTier={selectedTier}
                        />
                    </div>
                </div>
                <div className='flex flex-col basis-1/3'>
                    <SelectedDetails type={type} id={id} cypher={showCypher} data={data} />
                </div>
            </div>
        </div>
    );
};

export default Details;
