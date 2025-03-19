import { Button } from '@bloodhoundenterprise/doodleui';
import { AssetLabel, AssetSelector, SelectorNode } from 'js-client-library';
import { FC, useState } from 'react';
import { useQuery } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { AppIcon, CreateMenu } from '../../components';
import { ActiveDirectoryNodeKind } from '../../graphSchema';
import { ROUTE_TIER_MANAGEMENT_CREATE, ROUTE_TIER_MANAGEMENT_EDIT } from '../../routes';
import { apiClient } from '../../utils';
import DynamicDetails from './Details/DynamicDetails';
import EntityInfoPanel from './Details/EntityInfo/EntityInfoPanel';
import ObjectCountPanel from './Details/ObjectCountPanel';
import { DetailsList } from './DetailsList';

const innerDetail = (
    selectedObjectId: number | null,
    selectedSelectorId: number | null,
    selectedLabelId: number,
    objectsList: SelectorNode[],
    labelsList: AssetLabel[],
    selectorsList: AssetSelector[]
): SelectedDetailsProps => {
    if (selectedObjectId !== null) {
        const selectedObject = objectsList.find((object) => object.id === selectedObjectId);
        return { id: selectedObjectId, type: 'object', data: selectedObject };
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
    data?: SelectorNode | AssetLabel | AssetSelector;
    cypher?: boolean;
};

// const isSelector = (data: any): data is AssetSelector => {
//     return 'seeds' in data;
// };

// const isLabel = (data: any): data is AssetLabel => {
//     return 'asset_group_tier_id' in data;
// };
const isObject = (data: any): data is SelectorNode => {
    const objectData = data || {};
    return 'node_id' in objectData;
};

const SelectedDetails: FC<SelectedDetailsProps> = ({ type, cypher, data }) => {
    if (isObject(data)) {
        const selectedNode = {
            id: data.node_id.toString(),
            type: ActiveDirectoryNodeKind.User,
            name: data.name,
        };
        return (
            <>
                <EntityInfoPanel selectedNode={selectedNode} />
                {/* <div>{`Object Info Panel - ${type}-${id}`}</div> */}
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
                    <ObjectCountPanel data={data} />
                </>
            );
    }

    return (
        <>
            <DynamicDetails data={data} />
            {/* <div>{`Dynamic Details - ${type}-${id}`}</div> */}
            <ObjectCountPanel data={data} />
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

    const { type, id, data } = innerDetail(
        selectedObject,
        selectedSelector,
        selectedTier,
        objectsQuery.data || [],
        labelsQuery.data || [],
        selectorsQuery.data || []
    );

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
                    <SelectedDetails type={type} id={id} cypher={showCypher} data={data} />
                </div>
            </div>
        </div>
    );
};

export default Details;
