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
    AssetGroupTag,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeZone,
    CreateAssetGroupTagRequest,
    UpdateAssetGroupTagRequest,
} from 'js-client-library';
import isEmpty from 'lodash/isEmpty';
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
import { usePZPathParams } from '../../../../hooks/usePZParams';
import { useNotifications } from '../../../../providers';
import { useAppNavigate } from '../../../../utils';
import { PrivilegeZonesContext } from '../../PrivilegeZonesContext';
import { handleError } from '../utils';
import { useTagFormUtils } from './utils';

const MAX_NAME_LENGTH = 250;

export const TagForm: FC = () => {
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    const navigate = useAppNavigate();
    const { addNotification } = useNotifications();

    const { tagId } = usePZPathParams();
    const {
        privilegeZoneAnalysisEnabled,
        disableNameInput,
        isLabelPage,
        isUpdateZoneLocation,
        isZonePage,
        showAnalysisToggle,
        showDeleteButton,
        formTitle,
        tagType: tagTypePlural,
        tagTypeDisplay,
        handleCreateNavigate,
        handleUpdateNavigate,
        handleDeleteNavigate,
    } = useTagFormUtils();
    const tagType = tagTypePlural.slice(0, -1) as 'label' | 'zone';

    const tagsQuery = useAssetGroupTags();
    const tagQuery = useAssetGroupTagInfo(tagId);

    const { ZoneList, SalesMessage } = useContext(PrivilegeZonesContext);
    const showSalesMessage = isUpdateZoneLocation && SalesMessage;
    const showZoneList = isUpdateZoneLocation && ZoneList;

    const diffValues = (
        data: AssetGroupTag | undefined,
        formValues: UpdateAssetGroupTagRequest
    ): Partial<UpdateAssetGroupTagRequest> => {
        if (data === undefined) return formValues;
        const workingCopy = { ...formValues };
        const diffed: Partial<UpdateAssetGroupTagRequest> = {};

        if (data.name !== workingCopy.name) diffed.name = workingCopy.name;
        if (data.description !== workingCopy.description) diffed.description = workingCopy.description;
        if (data.position !== workingCopy.position) diffed.position = workingCopy.position;
        if (data.require_certify != workingCopy.require_certify) diffed.require_certify = workingCopy.require_certify;
        if (data.analysis_enabled !== workingCopy.analysis_enabled)
            diffed.analysis_enabled = workingCopy.analysis_enabled;

        return diffed;
    };

    const form = useForm<UpdateAssetGroupTagRequest>({
        defaultValues: {
            name: '',
            description: '',
            require_certify: false,
            analysis_enabled: false,
            position: -1,
        },
    });

    const { control, getValues, handleSubmit, reset, setValue } = form;

    const createTagMutation = useCreateAssetGroupTag();
    const updateTagMutation = usePatchAssetGroupTag(tagId);
    const deleteTagMutation = useDeleteAssetGroupTag();

    const handleCreateTag = useCallback(
        async (formData: CreateAssetGroupTagRequest) => {
            try {
                const requestValues = {
                    name: formData.name,
                    description: formData.description,
                    require_certify: isZonePage ? formData.require_certify : null,
                    position: null,
                    type: isLabelPage ? AssetGroupTagTypeLabel : AssetGroupTagTypeZone,
                };

                const response = await createTagMutation.mutateAsync({
                    values: requestValues,
                });

                addNotification(`${tagTypeDisplay} was created successfully!`, undefined, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });

                handleCreateNavigate(response.id);
            } catch (error) {
                handleError(error, 'creating', tagType, addNotification);
            }
        },
        [isZonePage, isLabelPage, createTagMutation, addNotification, tagTypeDisplay, handleCreateNavigate, tagType]
    );

    const handleUpdateTag = useCallback(async () => {
        try {
            const diffedValues = diffValues(tagQuery.data, { ...getValues() });
            if (isEmpty(diffedValues)) {
                addNotification('No changes detected', `privilege-zones_update-tag_no-changes-warn_${tagId}`, {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                });
                return;
            }

            const updatedValues = { ...diffedValues };

            if (!privilegeZoneAnalysisEnabled) delete updatedValues.analysis_enabled;
            if (isLabelPage) delete updatedValues.require_certify;

            await updateTagMutation.mutateAsync({
                updatedValues,
                tagId,
            });

            addNotification(
                `${tagTypeDisplay} was updated successfully!`,
                `privilege-zones_update-${tagType}_success_${tagId}`,
                {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );

            handleUpdateNavigate();
        } catch (error) {
            handleError(error, 'updating', tagType, addNotification);
        }
    }, [
        tagQuery.data,
        getValues,
        privilegeZoneAnalysisEnabled,
        isLabelPage,
        updateTagMutation,
        tagId,
        addNotification,
        tagTypeDisplay,
        tagType,
        handleUpdateNavigate,
    ]);

    const handleDeleteTag = useCallback(async () => {
        try {
            await deleteTagMutation.mutateAsync(tagId);

            addNotification(
                `${tagTypeDisplay} was deleted successfully!`,
                `privilege-zones_delete-${tagType}_success_${tagId}`,
                {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );

            setDeleteDialogOpen(false);
            handleDeleteNavigate();
        } catch (error) {
            handleError(error, 'deleting', tagType, addNotification);
        }
    }, [tagId, deleteTagMutation, addNotification, handleDeleteNavigate, tagType, tagTypeDisplay]);

    const onSubmit: SubmitHandler<UpdateAssetGroupTagRequest | CreateAssetGroupTagRequest> = useCallback(
        (formData) => {
            if (tagId === '') {
                handleCreateTag(formData as CreateAssetGroupTagRequest);
            } else {
                handleUpdateTag();
            }
        },
        [tagId, handleCreateTag, handleUpdateTag]
    );

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    useEffect(() => {
        if (tagQuery.data) {
            reset({
                name: tagQuery.data.name,
                description: tagQuery.data.description,
                position: tagQuery.data.position,
                require_certify: tagQuery.data.require_certify || false,
                analysis_enabled: tagQuery.data.analysis_enabled || false,
            });
        }
    }, [tagQuery.data, reset]);

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
                                <span>{`${tagTypeDisplay} Information`}</span>
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
                                {isZonePage && (
                                    <div className='grid gap-2'>
                                        <Label>Require Certification</Label>
                                        <Skeleton className='h-3 w-6' />
                                    </div>
                                )}
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
                                data-testid='privilege-zones_save_tag-form_delete-button'
                                variant={'text'}
                                onClick={() => {
                                    setDeleteDialogOpen(true);
                                }}>
                                <span>
                                    <FontAwesomeIcon icon={faTrashCan} className='mr-2' />
                                    {`Delete ${tagTypeDisplay}`}
                                </span>
                            </Button>
                        )}
                        <Button
                            data-testid='privilege-zones_save_tag-form_cancel-button'
                            variant={'secondary'}
                            onClick={() => {
                                navigate(-1);
                            }}>
                            Cancel
                        </Button>
                        <Button data-testid='privilege-zones_save_tag-form_save-button' variant={'primary'}>
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
                                <span>{`${tagTypeDisplay} Information`}</span>
                            </div>
                            <div className='flex flex-col gap-6 mt-6'>
                                <FormField
                                    control={control}
                                    name='name'
                                    rules={{
                                        required: `Please provide a name for the ${tagTypeDisplay}`,
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
                                                    data-testid='privilege-zones_save_tag-form_name-input'
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={control}
                                    name='description'
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Description</FormLabel>
                                            <FormControl>
                                                <Textarea
                                                    onChange={field.onChange}
                                                    value={field.value}
                                                    data-testid='privilege-zones_save_tag-form_description-input'
                                                    placeholder='Description Input'
                                                    rows={3}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                {isZonePage && (
                                    <FormField
                                        control={control}
                                        name='require_certify'
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Require Certification</FormLabel>
                                                <div className='flex gap-2'>
                                                    <FormControl>
                                                        <Switch
                                                            {...field}
                                                            value={field.value?.toString()}
                                                            data-testid='privilege-zones_save_tag-form_require-certify-toggle'
                                                            checked={field.value || false}
                                                            onCheckedChange={field.onChange}></Switch>
                                                    </FormControl>
                                                    <p className='text-sm'>
                                                        Enable this to mandate certification for all members within this
                                                        zone
                                                    </p>
                                                </div>

                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                )}

                                {showAnalysisToggle && (
                                    <FormField
                                        control={control}
                                        name='analysis_enabled'
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Enable Analysis</FormLabel>
                                                <FormControl>
                                                    <Switch
                                                        {...field}
                                                        value={''}
                                                        data-testid='privilege-zones_save_tag-form_enable-analysis-toggle'
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
                                        control={control}
                                        name='position'
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Position</FormLabel>
                                                <FormControl>
                                                    <Input
                                                        data-testid='privilege-zones_save_tag-form_position-input'
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
                                data-testid='privilege-zones_save_tag-form_delete-button'
                                variant={'text'}
                                onClick={() => {
                                    setDeleteDialogOpen(true);
                                }}>
                                <span>
                                    <FontAwesomeIcon icon={faTrashCan} className='mr-2' />
                                    {`Delete ${tagTypeDisplay}`}
                                </span>
                            </Button>
                        )}
                        <Button
                            data-testid='privilege-zones_save_tag-form_cancel-button'
                            variant={'secondary'}
                            onClick={() => {
                                navigate(-1);
                            }}>
                            Cancel
                        </Button>
                        <Button
                            data-testid='privilege-zones_save_tag-form_save-button'
                            variant={'primary'}
                            onClick={handleSubmit(onSubmit)}>
                            {tagId === '' ? 'Define Selector' : 'Save Edits'}
                        </Button>
                    </div>
                </div>

                {showZoneList && (
                    <ZoneList
                        zones={tagsQuery.data?.filter((tag) => tag.type === AssetGroupTagTypeZone) || []}
                        setPosition={(position: number | undefined) => {
                            setValue('position', position, { shouldDirty: true });
                        }}
                        name={tagQuery.data?.name || 'New Zone'}
                    />
                )}
            </form>
            <DeleteConfirmationDialog
                isLoading={tagQuery.isLoading}
                itemName={tagQuery.data?.name || tagType}
                itemType={tagType}
                onCancel={handleCancel}
                onConfirm={handleDeleteTag}
                open={deleteDialogOpen}
            />
        </Form>
    );
};
