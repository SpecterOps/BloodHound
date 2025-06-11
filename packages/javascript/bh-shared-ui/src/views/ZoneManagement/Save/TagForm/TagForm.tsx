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
} from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    AssetGroupTag,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeTier,
    AssetGroupTagTypes,
    UpdateAssetGroupTagRequest,
} from 'js-client-library';
import { FC, useCallback, useState } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import { Location, useLocation, useParams } from 'react-router-dom';
import DeleteConfirmationDialog from '../../../../components/DeleteConfirmationDialog';
import { useNotifications } from '../../../../providers';
import { cn, useAppNavigate } from '../../../../utils';
import { OWNED_ID, TIER_ZERO_ID, getTagUrlValue } from '../../utils';
import { handleError } from '../utils';
import { useAssetGroupTagInfo, useCreateAssetGroupTag, useDeleteAssetGroupTag, usePatchAssetGroupTag } from './hooks';

type TagFormInputs = {
    name: string;
    description: string;
    position: number | null;
    type: AssetGroupTagTypes;
};

const MAX_NAME_LENGTH = 250;

const formTitleFromPath = (labelId: string | undefined, tierId: string, location: Location): string => {
    if (location.pathname.includes('save/label') && !labelId) return 'Create new Label';
    if (location.pathname.includes('save/tier') && tierId === '') return 'Create new Tier';
    if (location.pathname.includes('save/label') && labelId) return 'Edit Label Details';
    if (location.pathname.includes('save/tier') && tierId !== '') return 'Edit Tier Details';

    // We should never reach this default return
    return 'Tag Details';
};

const showDeleteButton = (labelId: string | undefined, tierId: string) => {
    if (tierId === '' && !labelId) return false;
    if (labelId === OWNED_ID) return false;
    if (tierId === TIER_ZERO_ID) return false;
    return true;
};

const diffValues = (data: AssetGroupTag | undefined, formValues: TagFormInputs): UpdateAssetGroupTagRequest => {
    if (data === undefined) return formValues;

    const workingCopy = { ...formValues };

    const diffed: UpdateAssetGroupTagRequest = {};

    if (data.name !== workingCopy.name) diffed.name = workingCopy.name;
    if (data.description !== workingCopy.description) diffed.description = workingCopy.description;
    if (data.position !== workingCopy.position) diffed.position = workingCopy.position;

    return diffed;
};

export const TagForm: FC = () => {
    const { tierId = '', labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const navigate = useAppNavigate();
    const location = useLocation();

    const { addNotification } = useNotifications();
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<TagFormInputs>();

    const tagQuery = useAssetGroupTagInfo(tagId);

    const createTagMutation = useCreateAssetGroupTag();
    const updateTagMutation = usePatchAssetGroupTag(tagId);
    const deleteTagMutation = useDeleteAssetGroupTag();

    const handleCreateTag = useCallback(
        async (formData: TagFormInputs) => {
            try {
                const response = await createTagMutation.mutateAsync({
                    values: {
                        ...formData,
                        type: location.pathname.includes('label') ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
                    },
                });

                addNotification(`${labelId ? 'Label' : 'Tier'} was created successfully!`, undefined, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });

                // Upon creation of this tag the user should be moved to creating a selector for the newly created tag, e.g., /save/tier/<NEW_TIER_ID>/selector
                // This means that we have to await for the ID of the new tag in order to go to the URL for creating a new selector associated with this tag
                // In addition, once at the create selector form, the cancel button needs go back to the form for the newly created tag
                // but the URL for creating a new tag does not have the recently created tag ID in the path, i.e., /save/tier vs /save/tier/<NEW_TIER_ID>
                // that means the location history needs to be manipulated (replaced) in order to have that available once at the selector form

                navigate(`${location.pathname}/${response.id}`, { replace: true });
                navigate(`${location.pathname}/${response.id}/selector`);
            } catch (error) {
                handleError(error, 'creating', getTagUrlValue(labelId), addNotification);
            }
        },
        [labelId, navigate, createTagMutation, addNotification, location]
    );

    const handleUpdateTag = useCallback(
        async (formData: TagFormInputs) => {
            try {
                const diffedValues = diffValues(tagQuery.data, formData);

                await updateTagMutation.mutateAsync({
                    updatedValues: {
                        ...diffedValues,
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

                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}`);
            } catch (error) {
                handleError(error, 'updating', getTagUrlValue(labelId), addNotification);
            }
        },
        [labelId, tagId, navigate, addNotification, updateTagMutation, tagQuery.data]
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

            const tagValue = getTagUrlValue(labelId);

            setDeleteDialogOpen(false);
            navigate(`/zone-management/details/${tagValue}/${tagValue === 'tier' ? TIER_ZERO_ID : OWNED_ID}`);
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

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    if (tagQuery.isLoading) return <Skeleton />;
    if (tagQuery.isError) return <div>There was an error fetching the tag information.</div>;

    return (
        <form>
            <Card className='min-w-96 w-[672px] p-3 mt-6'>
                <CardHeader>
                    <CardTitle>{formTitleFromPath(labelId, tierId, location)}</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className='flex justify-between'>
                        <span>{`${location.pathname.includes('label') ? 'Label' : 'Tier'} Information`}</span>
                    </div>
                    <div className='flex flex-col gap-6 mt-6'>
                        <div>
                            <Label htmlFor='name'>Name</Label>
                            <Input
                                id='name'
                                type='text'
                                disabled={tagId === TIER_ZERO_ID || tagId === OWNED_ID}
                                {...register('name', {
                                    required: `Please provide a name for the ${labelId ? 'Label' : 'Tier'}`,
                                    value: tagQuery.data?.name,
                                    maxLength: {
                                        value: MAX_NAME_LENGTH,
                                        message: `Name cannot exceed ${MAX_NAME_LENGTH} characters. Please provide a shorter name`,
                                    },
                                })}
                                className={
                                    'rounded-none text-base bg-transparent dark:bg-transparent border-t-0 border-x-0 border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b-[1px] focus-visible:outline-none focus:border-t-0 focus:border-x-0 focus-visible:ring-offset-0 focus-visible:ring-transparent focus-visible:border-secondary focus-visible:border-b-2 focus:border-secondary focus:border-b-2 dark:focus-visible:outline-none dark:focus:border-t-0 dark:focus:border-x-0 dark:focus-visible:ring-offset-0 dark:focus-visible:ring-transparent dark:focus-visible:border-secondary-variant-2 dark:focus-visible:border-b-2 dark:focus:border-secondary-variant-2 dark:focus:border-b-2 hover:border-b-2'
                                }
                            />
                            {errors.name && (
                                <p className='text-sm text-[#B44641] dark:text-[#E9827C]'>{errors.name.message}</p>
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
                            <Label htmlFor='position'>Position</Label>
                            <Input
                                id='position'
                                type='number'
                                {...register('position', { value: tagQuery.data?.position || null })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>
            <div className='flex justify-end gap-6 mt-6 w-[672px]'>
                {showDeleteButton(labelId, tierId) && (
                    <Button
                        variant={'text'}
                        onClick={() => {
                            setDeleteDialogOpen(true);
                        }}>
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
                    {tagId === '' ? 'Define Selector' : 'Save Edits'}
                </Button>
            </div>
            <DeleteConfirmationDialog
                isLoading={tagQuery.isLoading}
                itemName={tagQuery.data?.name || getTagUrlValue(labelId)}
                itemType={getTagUrlValue(labelId)}
                onCancel={handleCancel}
                onConfirm={handleDeleteTag}
                open={deleteDialogOpen}
            />
        </form>
    );
};
