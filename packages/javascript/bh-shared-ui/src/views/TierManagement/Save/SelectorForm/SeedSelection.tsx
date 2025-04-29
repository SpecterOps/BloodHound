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

import { Button, Card, CardContent, CardHeader, Input, Skeleton } from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { createBrowserHistory } from 'history';
import { GraphNodes, SeedTypeObjectId, SeedTypes } from 'js-client-library';
import { RequestOptions, SelectorSeedRequest } from 'js-client-library/dist/requests';
import { FC, useCallback, useEffect, useRef, useState } from 'react';
import { SubmitHandler, useFormContext } from 'react-hook-form';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { AssetGroupSelectorObjectSelect, DeleteConfirmationDialog } from '../../../../components';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { useNotifications } from '../../../../providers';
import { apiClient, cn } from '../../../../utils';
import { Cypher } from '../../Cypher';
import { DeleteSelectorParams, SelectorFormInputs } from './types';
import { handleError } from './utils';

const deleteSelector = async (ids: DeleteSelectorParams, options?: RequestOptions) =>
    await apiClient.deleteAssetGroupTagSelector(ids.tagId, ids.selectorId, options).then((res) => res.data.data);

const useDeleteSelector = (tagId: string | number | undefined) => {
    const queryClient = useQueryClient();
    return useMutation(deleteSelector, {
        onSettled: () => {
            queryClient.invalidateQueries(['tier-management', 'tags', tagId, 'selectors']);
        },
    });
};

const getListScalar = (windoHeight: number) => {
    if (windoHeight > 1080) return 18;
    if (1080 >= windoHeight && windoHeight > 900) return 14;
    if (900 >= windoHeight) return 10;
    return 8;
};

const SeedSelection: FC<{ selectorType: SeedTypes; onSubmit: SubmitHandler<SelectorFormInputs> }> = ({
    selectorType,
    onSubmit,
}) => {
    const { tagId = '', selectorId = '' } = useParams();

    const [results, setResults] = useState<GraphNodes | null>(null);
    const [seeds, setSeeds] = useState<SelectorSeedRequest[]>([]);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    const { handleSubmit, register } = useFormContext<SelectorFormInputs>();

    const history = createBrowserHistory();
    const navigate = useNavigate();

    const { addNotification } = useNotifications();

    const deleteSelectorMutation = useDeleteSelector(tagId);

    const heightScalar = useRef(getListScalar(window.innerHeight));

    useEffect(() => {
        const updateHeightScalar = () => {
            heightScalar.current = getListScalar(window.innerHeight);
        };

        window.addEventListener('resize', updateHeightScalar);
        return () => window.removeEventListener('resize', updateHeightScalar);
    }, []);

    const selectorQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors', selectorId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTagSelector(tagId, selectorId);
            return response.data.data;
        },
        enabled: selectorId !== '',
    });

    const handleDeleteSelector = useCallback(
        async (response: boolean) => {
            if (response === false) {
                setDeleteDialogOpen(false);
            } else {
                try {
                    if (!tagId || !selectorId)
                        throw new Error(`Missing required entity IDs; tagId: ${tagId} , selectorId: ${selectorId}`);

                    await deleteSelectorMutation.mutateAsync({ tagId, selectorId });

                    navigate(`/tier-management/details/tags/${tagId}`);
                } catch (error) {
                    handleError(error, 'deleting', addNotification);
                }
            }
        },
        [tagId, selectorId, navigate, deleteSelectorMutation, addNotification]
    );

    if (selectorQuery.isLoading) return <Skeleton />;
    if (selectorQuery.isError) throw new Error();

    return (
        <>
            <div className='grow'>
                <div className='flex justify-center'>
                    <div
                        className={cn('w-full max-w-[60rem]', {
                            'max-w-[42rem] max-md:w-96 max-lg:w-[28rem] max-xl:w-[36rem]':
                                selectorType === SeedTypeObjectId,
                        })}>
                        <Input {...register('seeds', { value: seeds })} className='hidden w-0' />
                        {selectorType === SeedTypeObjectId ? (
                            <AssetGroupSelectorObjectSelect setSeeds={setSeeds} />
                        ) : (
                            <Cypher
                                preview={false}
                                setCypherSearchResults={setResults}
                                setSeeds={setSeeds}
                                initialInput={selectorQuery.data?.selector.seeds[0].value}
                            />
                        )}
                        <div className={cn('flex justify-end gap-6 mt-6 w-full')}>
                            {selectorId !== '' && (
                                <Button
                                    variant={'text'}
                                    onClick={() => {
                                        setDeleteDialogOpen(true);
                                    }}>
                                    <span>
                                        <FontAwesomeIcon icon={faTrashCan} /> Delete Selector
                                    </span>
                                </Button>
                            )}
                            <Button variant={'secondary'} onClick={history.back}>
                                Cancel
                            </Button>
                            <Button variant={'primary'} onClick={handleSubmit(onSubmit)}>
                                Save
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
            <Card className='max-h-full min-w-[27rem]'>
                <CardHeader className='pl-6 first:py-6 text-xl font-bold'>Sample Results</CardHeader>
                <CardContent className='pl-4'>
                    <div className='font-bold pl-2 mb-2'>
                        <span>Type</span>
                        <span className='ml-8'>Object Name</span>
                    </div>
                    <VirtualizedNodeList
                        nodes={results ? Object.values(results) : []}
                        itemSize={46}
                        heightScalar={heightScalar.current}
                    />
                </CardContent>
            </Card>
            <DeleteConfirmationDialog
                open={deleteDialogOpen}
                itemName={selectorQuery.data?.selector.name || 'Selector'}
                itemType='selector'
                onClose={handleDeleteSelector}
            />
        </>
    );
};

export default SeedSelection;
