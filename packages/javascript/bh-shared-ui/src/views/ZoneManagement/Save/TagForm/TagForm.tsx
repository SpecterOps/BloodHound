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
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    Input,
    Label,
    Skeleton,
    Switch,
    Textarea,
} from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeTier,
    CreateAssetGroupTagRequest,
    UpdateAssetGroupTagRequest,
} from 'js-client-library';
import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { SubmitHandler, useForm } from 'react-hook-form';
import DeleteConfirmationDialog from '../../../../components/DeleteConfirmationDialog';
import {
    useAssetGroupTagInfo,
    useAssetGroupTags,
    useCreateAssetGroupTag,
    useDeleteAssetGroupTag,
    usePatchAssetGroupTag,
} from '../../../../hooks/useAssetGroupTags';
import { useZonePathParams } from '../../../../hooks/useZoneParams';
import { useNotifications } from '../../../../providers';
import { useAppNavigate } from '../../../../utils';
import { ZoneManagementContext } from '../../ZoneManagementContext';
import { handleError } from '../utils';
import { useTagFormUtils } from './utils';

const MAX_NAME_LENGTH = 250;

export const TagForm: FC = () => {
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    const navigate = useAppNavigate();
    const { addNotification } = useNotifications();

    const { tagId } = useZonePathParams();
    const {
        privilegeZoneAnalysisEnabled,
        disableNameInput,
        isLabelLocation,
        isUpdateTierLocation,
        showAnalysisToggle,
        showDeleteButton,
        formTitle,
        tagKind,
        tagKindDisplay,
        handleCreateNavigate,
        handleUpdateNavigate,
        handleDeleteNavigate,
    } = useTagFormUtils();

    const tagsQuery = useAssetGroupTags();
    const tagQuery = useAssetGroupTagInfo(tagId);

    const { TierList, SalesMessage } = useContext(ZoneManagementContext);
    const showSalesMessage = isUpdateTierLocation && SalesMessage;
    const showTierList = isUpdateTierLocation && TierList;

    const form = useForm<UpdateAssetGroupTagRequest>({
        defaultValues: {
            name: '',
            description: '',
            analysis_enabled: false,
            position: -1,
        },
    });

    const { isDirty } = form.formState;

    const createTagMutation = useCreateAssetGroupTag();
    const updateTagMutation = usePatchAssetGroupTag(tagId);
    const deleteTagMutation = useDeleteAssetGroupTag();

    const handleCreateTag = useCallback(
        async (formData: CreateAssetGroupTagRequest) => {
            try {
                const requestValues = {
                    name: formData.name,
                    description: formData.description,
                    position: null,
                    type: isLabelLocation ? AssetGroupTagTypeLabel : AssetGroupTagTypeTier,
                };

                const response = await createTagMutation.mutateAsync({
                    values: requestValues,
                });

                addNotification(`${tagKindDisplay} was created successfully!`, undefined, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });

                handleCreateNavigate(response.id);
            } catch (error) {
                handleError(error, 'creating', tagKind, addNotification);
            }
        },
        [createTagMutation, addNotification, handleCreateNavigate, tagKind, tagKindDisplay, isLabelLocation]
    );

    const handleUpdateTag = useCallback(
        async (formData: UpdateAssetGroupTagRequest) => {
            try {
                if (!isDirty) {
                    addNotification('No changes detected', `zone-management_update-tag_no-changes-warn_${tagId}`, {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    });
                    return;
                }

                const updatedValues = { ...formData };

                if (!privilegeZoneAnalysisEnabled) delete updatedValues.analysis_enabled;

                await updateTagMutation.mutateAsync({
                    updatedValues,
                    tagId,
                });

                addNotification(
                    `${tagKindDisplay} was updated successfully!`,
                    `zone-management_update-${tagKind}_success_${tagId}`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );

                handleUpdateNavigate();
            } catch (error) {
                handleError(error, 'updating', tagKind, addNotification);
            }
        },
        [tagId, handleUpdateNavigate, addNotification, updateTagMutation, tagKind, tagKindDisplay, isDirty]
    );

    const handleDeleteTag = useCallback(async () => {
        try {
            await deleteTagMutation.mutateAsync(tagId);

            addNotification(
                `${tagKindDisplay} was deleted successfully!`,
                `zone-management_delete-${tagKind}_success_${tagId}`,
                {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );

            setDeleteDialogOpen(false);
            handleDeleteNavigate();
        } catch (error) {
            handleError(error, 'deleting', tagKind, addNotification);
        }
    }, [tagId, deleteTagMutation, addNotification, handleDeleteNavigate, tagKind, tagKindDisplay]);

    const onSubmit: SubmitHandler<UpdateAssetGroupTagRequest | CreateAssetGroupTagRequest> = useCallback(
        (formData) => {
            if (tagId === '') {
                handleCreateTag(formData as CreateAssetGroupTagRequest);
            } else {
                handleUpdateTag(formData);
            }
        },
        [tagId, handleCreateTag, handleUpdateTag]
    );

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    useEffect(() => {
        if (tagQuery.data) {
            form.reset({
                name: tagQuery.data.name,
                description: tagQuery.data.description,
                position: tagQuery.data.position,
                analysis_enabled: tagQuery.data.analysis_enabled || false,
            });
        }
    }, [tagQuery.data, form]);

    if (tagQuery.isLoading)
        return (
            <form className='flex gap-x-6 mt-6'>
                <div className='flex flex-col justify-between min-w-96 w-[672px]'>
                    <Card className='p-3 mb-4'>
                        <CardHeader>
                            <CardTitle>{formTitle}</CardTitle>
                        </CardHeader>
                        <Skeleton className='' />
                        <CardContent>
                            <div className='flex justify-between'>
                                <span>{`${tagKindDisplay} Information`}</span>
                            </div>
                            <div className='flex flex-col gap-6 mt-6'>
                                <div className='grid gap-2'>
                                    <Label>Name</Label>
                                    <Skeleton className='h-10 w-full' />
                                </div>
                                <div className='grid gap-2'>
                                    <Label>Description</Label>
                                    <Skeleton className='h-16 w-full' />
                                </div>
                                {showAnalysisToggle && (
                                    <div className='grid gap-2'>
                                        <Label>Enable Analysis</Label>
                                        <Skeleton className='h-3 w-6' />
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                    {showSalesMessage && <SalesMessage />}
                    <div className='flex justify-end gap-6 mt-4 min-w-96 max-w-[672px]'>
                        {showDeleteButton() && (
                            <Button
                                data-testid='zone-management_save_tag-form_delete-button'
                                variant={'text'}
                                onClick={() => {
                                    setDeleteDialogOpen(true);
                                }}>
                                <span>
                                    <FontAwesomeIcon icon={faTrashCan} className='mr-2' />
                                    {`Delete ${tagKindDisplay}`}
                                </span>
                            </Button>
                        )}
                        <Button
                            data-testid='zone-management_save_tag-form_cancel-button'
                            variant={'secondary'}
                            onClick={() => {
                                navigate(-1);
                            }}>
                            Cancel
                        </Button>
                        <Button data-testid='zone-management_save_tag-form_save-button' variant={'primary'}>
                            {tagId === '' ? 'Define Selector' : 'Save Edits'}
                        </Button>
                    </div>
                </div>

                <Skeleton className='w-[28rem] p-3' />
            </form>
        );

    if (tagQuery.isError) return <div>There was an error fetching the tag information.</div>;

    return (
        <Form {...form}>
            <form className='flex gap-x-6 mt-6'>
                <div className='flex flex-col justify-between min-w-96 w-[672px]'>
                    <Card className='p-3 mb-4'>
                        <CardHeader>
                            <CardTitle>{formTitle}</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className='flex justify-between'>
                                <span>{`${tagKindDisplay} Information`}</span>
                            </div>
                            <div className='flex flex-col gap-6 mt-6'>
                                <FormField
                                    control={form.control}
                                    name='name'
                                    rules={{
                                        required: `Please provide a name for the ${tagKindDisplay}`,
                                        maxLength: {
                                            value: MAX_NAME_LENGTH,
                                            message: `Name cannot exceed ${MAX_NAME_LENGTH} characters. Please provide a shorter name`,
                                        },
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel aria-labelledby='name'>Name</FormLabel>
                                            <FormControl>
                                                <Input
                                                    {...field}
                                                    type='text'
                                                    autoComplete='off'
                                                    disabled={disableNameInput}
                                                    data-testid='zone-management_save_tag-form_name-input'
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name='description'
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Description</FormLabel>
                                            <FormControl>
                                                <Textarea
                                                    onChange={field.onChange}
                                                    value={field.value}
                                                    data-testid='zone-management_save_tag-form_description-input'
                                                    placeholder='Description Input'
                                                    rows={3}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                {showAnalysisToggle && (
                                    <FormField
                                        control={form.control}
                                        name='analysis_enabled'
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Enable Analysis</FormLabel>
                                                <FormControl>
                                                    <Switch
                                                        {...field}
                                                        value={''}
                                                        data-testid='zone-management_save_tag-form_enable-analysis-toggle'
                                                        checked={field.value || false}
                                                        onCheckedChange={field.onChange}
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                )}
                                <div className='hidden'>
                                    <FormField
                                        control={form.control}
                                        name='position'
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Position</FormLabel>
                                                <FormControl>
                                                    <Input
                                                        data-testid='zone-management_save_tag-form_position-input'
                                                        type='number'
                                                        {...field}
                                                        value={field.value || -1}
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    {showSalesMessage && <SalesMessage />}
                    <div className='flex justify-end gap-6 mt-4 min-w-96 max-w-[672px]'>
                        {showDeleteButton() && (
                            <Button
                                data-testid='zone-management_save_tag-form_delete-button'
                                variant={'text'}
                                onClick={() => {
                                    setDeleteDialogOpen(true);
                                }}>
                                <span>
                                    <FontAwesomeIcon icon={faTrashCan} className='mr-2' />
                                    {`Delete ${tagKindDisplay}`}
                                </span>
                            </Button>
                        )}
                        <Button
                            data-testid='zone-management_save_tag-form_cancel-button'
                            variant={'secondary'}
                            onClick={() => {
                                navigate(-1);
                            }}>
                            Cancel
                        </Button>
                        <Button
                            data-testid='zone-management_save_tag-form_save-button'
                            variant={'primary'}
                            onClick={form.handleSubmit(onSubmit)}>
                            {tagId === '' ? 'Define Selector' : 'Save Edits'}
                        </Button>
                    </div>
                </div>

                {showTierList && (
                    <TierList
                        tiers={tagsQuery.data?.filter((tag) => tag.type === AssetGroupTagTypeTier) || []}
                        setPosition={(position: number | undefined) => {
                            form.setValue('position', position);
                        }}
                        name={tagQuery.data?.name || 'New Tier'}
                    />
                )}
            </form>
            <DeleteConfirmationDialog
                isLoading={tagQuery.isLoading}
                itemName={tagQuery.data?.name || tagKind}
                itemType={tagKind}
                onCancel={handleCancel}
                onConfirm={handleDeleteTag}
                open={deleteDialogOpen}
            />
        </Form>
    );
};
