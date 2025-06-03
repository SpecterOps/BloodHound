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

import {
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Input,
    Label,
    Skeleton,
    Switch,
} from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagTypeLabel, AssetGroupTagTypeTier } from 'js-client-library';
import { FC, useCallback } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import { useParams } from 'react-router-dom';
import { useNotifications } from '../../../../providers';
import { cn, useAppNavigate } from '../../../../utils';
import { OWNED_ID, TIER_ZERO_ID, getTagUrlValue } from '../../utils';
import { handleError } from '../utils';
import { useAssetGroupTagInfo, useCreateAssetGroupTag, useDeleteAssetGroupTag, usePatchAssetGroupTag } from './hooks';

type TagFormInputs = {
    name: string;
    description: string;
    position: number | null;
    certificationRequired: boolean;
};

export const TagForm: FC = () => {
    const { tierId = '', labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const navigate = useAppNavigate();

    const { addNotification } = useNotifications();

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<TagFormInputs>();

    const createTagMutation = useCreateAssetGroupTag();
    const updateTagMutation = usePatchAssetGroupTag(tagId);
    const deleteTagMutation = useDeleteAssetGroupTag();

    const handleCreateTag = useCallback(
        async (formData: TagFormInputs) => {
            try {
                await createTagMutation.mutateAsync({
                    values: { ...formData, type: labelId ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier },
                });

                addNotification(`${labelId ? 'Label' : 'Tier'} was created successfully!`, undefined, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });

                navigate(`/tier-management/details/${getTagUrlValue(labelId)}/${tagId}`);
            } catch (error) {
                handleError(error, 'creating', getTagUrlValue(labelId), addNotification);
            }
        },
        [labelId, tagId, navigate, createTagMutation, addNotification]
    );

    const handleUpdateTag = useCallback(
        async (formData: TagFormInputs) => {
            try {
                await updateTagMutation.mutateAsync({
                    updatedValues: {
                        ...formData,
                        type: labelId ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
                    },
                    tagId,
                });

                addNotification(
                    `${labelId ? 'Label' : 'Tier'} was updated successfully!`,
                    `tier-management_update-${getTagUrlValue(labelId)}_success_${tagId}`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );

                navigate(`/tier-management/details/${getTagUrlValue(labelId)}/${tagId}`);
            } catch (error) {
                handleError(error, 'updating', getTagUrlValue(labelId), addNotification);
            }
        },
        [labelId, tagId, navigate, addNotification, updateTagMutation]
    );

    const handleDeleteTag = useCallback(async () => {
        try {
            await deleteTagMutation.mutateAsync(tagId);

            addNotification(
                `${labelId ? 'Label' : 'Tier'} was deleted successfully!`,
                `tier-management_delete-${getTagUrlValue(labelId)}_success_${tagId}`,
                {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );

            navigate(`/tier-management/details/${getTagUrlValue(labelId)}/${tagId}`);
        } catch (error) {
            handleError(error, 'deleting', getTagUrlValue(labelId), addNotification);
        }
    }, [labelId, tagId, deleteTagMutation, addNotification, navigate]);

    const onSubmit: SubmitHandler<TagFormInputs> = useCallback(
        (formData) => {
            if (tagId === '') {
                handleCreateTag(formData);
            } else {
                handleUpdateTag(formData);
            }
        },
        [tagId, handleCreateTag, handleUpdateTag]
    );

    const tagQuery = useAssetGroupTagInfo(tagId);

    if (tagQuery.isLoading) return <Skeleton />;
    if (tagQuery.isError) return <div>There was an error fetching the tag information.</div>;

    const showDeleteButton = tagId && labelId ? tagId !== OWNED_ID : tagId !== TIER_ZERO_ID;

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <Card className='min-w-96 w-[672px] p-3 mt-6'>
                <CardHeader>
                    <CardTitle>Tier Details</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className='flex justify-between'>
                        <span>Tier Information</span>
                    </div>
                    <div onSubmit={handleSubmit(onSubmit)} className='flex flex-col gap-6 mt-6'>
                        <div>
                            <Label htmlFor='name'>Name</Label>
                            <Input
                                id='name'
                                type='text'
                                {...register('name', { required: true, value: tagQuery.data?.name, maxLength: 250 })}
                                className={
                                    'rounded-none text-base bg-transparent dark:bg-transparent border-t-0 border-x-0 border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b-[1px] focus-visible:outline-none focus:border-t-0 focus:border-x-0 focus-visible:ring-offset-0 focus-visible:ring-transparent focus-visible:border-secondary focus-visible:border-b-2 focus:border-secondary focus:border-b-2 dark:focus-visible:outline-none dark:focus:border-t-0 dark:focus:border-x-0 dark:focus-visible:ring-offset-0 dark:focus-visible:ring-transparent dark:focus-visible:border-secondary-variant-2 dark:focus-visible:border-b-2 dark:focus:border-secondary-variant-2 dark:focus:border-b-2 hover:border-b-2'
                                }
                            />
                            {errors.name && (
                                <p className='text-sm text-rose-700'>Please provide a name value for your tag</p>
                            )}
                        </div>
                        <div>
                            <Label htmlFor='description'>Description</Label>
                            <textarea
                                id='description'
                                {...register('description', { value: tagQuery.data?.description })}
                                placeholder='Description Input'
                                rows={3}
                                className={cn(
                                    'resize-none rounded-md dark:bg-neutral-dark-5 pl-2 w-full mt-2 focus-visible:outline-none focus:ring-secondary focus-visible:ring-secondary focus:outline-secondary focus-visible:outline-secondary dark:focus:ring-secondary-variant-2 dark:focus-visible:ring-secondary-variant-2 dark:focus:outline-secondary-variant-2 dark:focus-visible:outline-secondary-variant-2'
                                )}
                            />
                        </div>
                        <div className='hidden'>
                            <Label htmlFor='certificationRequired'>Required Certificaiton</Label>
                            <Switch
                                id='certificationRequired'
                                {...register('certificationRequired', {
                                    value: tagQuery.data?.requireCertify || true,
                                })}
                                label='Enable this to mandate certification for all objects within this tier'
                                defaultChecked></Switch>
                        </div>
                        <div className='hidden'>
                            <Label htmlFor='position'>Required Certificaiton</Label>
                            <Input
                                id='position'
                                type='number'
                                {...register('position', { value: tagQuery.data?.position })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>
            <div className='flex justify-end gap-6 mt-6 w-[672px]'>
                {showDeleteButton && (
                    <Button variant={'text'} onClick={handleDeleteTag}>
                        <span>
                            <FontAwesomeIcon icon={faTrashCan} className='mr-2' />
                            {`Delete ${labelId ? 'Label' : 'Tier'}`}
                        </span>
                    </Button>
                )}
                <Button
                    variant={'secondary'}
                    onClick={() => {
                        navigate(-1);
                    }}>
                    Cancel
                </Button>
                <Button variant={'primary'} onClick={handleSubmit(onSubmit)}>
                    Save
                </Button>
            </div>
        </form>
    );
};
