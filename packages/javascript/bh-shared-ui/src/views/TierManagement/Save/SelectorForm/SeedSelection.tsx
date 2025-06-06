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
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { SubmitHandler, useFormContext } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { DeleteConfirmationDialog } from '../../../../components';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { useDebouncedValue } from '../../../../hooks';
import { useNotifications } from '../../../../providers';
import { apiClient, cn, useAppNavigate } from '../../../../utils';
import { Cypher } from '../../Cypher/Cypher';
import { getTagUrlValue } from '../../utils';
import { handleError } from '../utils';
import DeleteSelectorButton from './DeleteSelectorButton';
import ObjectSelect from './ObjectSelect';
import SelectorFormContext from './SelectorFormContext';
import { useDeleteSelector } from './hooks';
import { SelectorFormInputs } from './types';

const getListScalar = (windowHeight: number) => {
    if (windowHeight > 1080) return 18;
    if (1080 >= windowHeight && windowHeight > 900) return 14;
    if (900 >= windowHeight) return 10;
    return 8;
};

const SeedSelection: FC<{
    onSubmit: SubmitHandler<SelectorFormInputs>;
}> = ({ onSubmit }) => {
    const navigate = useAppNavigate();
    const { tierId = '', labelId, selectorId = '' } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const { addNotification } = useNotifications();

    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    const { seeds, selectorType, selectorQuery } = useContext(SelectorFormContext);
    const { handleSubmit, register } = useFormContext<SelectorFormInputs>();

    const previewQuery = useQuery({
        queryKey: ['tier-management', 'preview-selectors', selectorType, seeds],
        queryFn: async ({ signal }) => {
            return apiClient
                .assetGroupTagsPreviewSelectors({ seeds: seeds }, { signal })
                .then((res) => res.data.data['members']);
        },
        retry: false,
        refetchOnWindowFocus: false,
        enabled: seeds.length > 0,
    });

    const deleteSelectorMutation = useDeleteSelector();

    const handleDeleteSelector = useCallback(async () => {
        try {
            if (!tagId || !selectorId)
                throw new Error(`Missing required entity IDs; tagId: ${tagId} , selectorId: ${selectorId}`);

            await deleteSelectorMutation.mutateAsync({ tagId, selectorId });

            addNotification('Selector was deleted successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            setDeleteDialogOpen(false);

            navigate(`/tier-management/details/${getTagUrlValue(labelId)}/${tagId}`);
        } catch (error) {
            handleError(error, 'deleting', 'selector', addNotification);
        }
    }, [tagId, labelId, selectorId, navigate, deleteSelectorMutation, addNotification]);

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    const [heightScalar, setHeightScalar] = useState(getListScalar(window.innerHeight));

    const updateHeightScalar = useDebouncedValue(() => setHeightScalar(getListScalar(window.innerHeight)), 100);

    useEffect(() => {
        window.addEventListener('resize', updateHeightScalar);
        return () => window.removeEventListener('resize', updateHeightScalar);
    }, [updateHeightScalar]);

    if (selectorQuery.isLoading) return <Skeleton />;
    if (selectorQuery.isError) return <div>There was an error fetching the selector data</div>;

    const firstSeed = seeds.values().next().value;

    return (
        <>
            <div className='grow'>
                <div className='flex justify-center'>
                    <div
                        className={cn('w-full max-w-[60rem]', {
                            'max-w-[42rem] max-md:w-96 max-lg:w-[28rem] max-xl:w-[36rem]':
                                selectorType === SeedTypeObjectId,
                        })}>
                        <Input {...register('seeds', { value: Array.from(seeds) })} className='hidden w-0' />
                        {selectorType === SeedTypeObjectId ? (
                            <ObjectSelect />
                        ) : (
                            <Cypher preview={false} initialInput={firstSeed?.value} />
                        )}
                        <div className={cn('flex justify-end gap-6 mt-6 w-full')}>
                            <DeleteSelectorButton
                                selectorId={selectorId}
                                selectorData={selectorQuery.data}
                                onClick={() => {
                                    setDeleteDialogOpen(true);
                                }}
                            />
                            <Button variant={'secondary'} onClick={() => navigate(-1)}>
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
                    <VirtualizedNodeList nodes={previewQuery.data ?? []} itemSize={46} heightScalar={heightScalar} />
                </CardContent>
            </Card>
            <DeleteConfirmationDialog
                open={deleteDialogOpen}
                itemName={selectorQuery.data?.name || 'Selector'}
                itemType='selector'
                onConfirm={handleDeleteSelector}
                onCancel={handleCancel}
            />
        </>
    );
};

export default SeedSelection;
