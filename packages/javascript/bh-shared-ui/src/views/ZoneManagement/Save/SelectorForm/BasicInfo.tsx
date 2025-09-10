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
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Skeleton,
    Switch,
    Textarea,
} from '@bloodhoundenterprise/doodleui';
import { SeedTypeCypher, SeedTypeObjectId, SeedTypesMap } from 'js-client-library';
import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { Control } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useLocation, useParams } from 'react-router-dom';
import { DeleteConfirmationDialog } from '../../../../components';
import { useDeleteSelector } from '../../../../hooks/useAssetGroupTags';
import { useNotifications } from '../../../../providers';
import { ROUTE_PRIVILEGE_ZONES_ROOT } from '../../../../routes';
import { apiClient, queriesAreLoadingOrErrored, useAppNavigate } from '../../../../utils';
import { getTagUrlValue } from '../../utils';
import { handleError } from '../utils';
import DeleteSelectorButton from './DeleteSelectorButton';
import SelectorFormContext from './SelectorFormContext';
import { SelectorFormInputs } from './types';

const BasicInfo: FC<{ control: Control<SelectorFormInputs, any, SelectorFormInputs> }> = ({ control }) => {
    const location = useLocation();
    const navigate = useAppNavigate();

    const { zoneId = '', labelId, selectorId = '' } = useParams();
    const tagId = labelId === undefined ? zoneId : labelId;

    const { dispatch, selectorType, selectorQuery } = useContext(SelectorFormContext);
    const receivedData = location.state;

    useEffect(() => {
        if (receivedData) {
            dispatch({ type: 'set-selector-type', selectorType: SeedTypeCypher });
        }
    }, []);

    const tagQuery = useQuery({
        queryKey: ['privilege-zones', 'tags', tagId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTag(tagId);
            return response.data.data['tag'];
        },
        enabled: tagId !== '',
    });

    const { isLoading, isError } = queriesAreLoadingOrErrored(tagQuery, selectorQuery);

    const { addNotification } = useNotifications();

    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

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

            navigate(`${ROUTE_PRIVILEGE_ZONES_ROOT}/${getTagUrlValue(labelId)}/${tagId}/details`);
        } catch (error) {
            handleError(error, 'deleting', 'selector', addNotification);
        }
    }, [tagId, labelId, selectorId, navigate, deleteSelectorMutation, addNotification]);

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    if (isLoading) return <Skeleton />;
    if (isError) return <div>There was an error fetching the selector information.</div>;

    return (
        <div className={'max-lg:w-full w-96 h-[36rem] '}>
            <Card className={'p-3'}>
                <CardHeader className='text-xl font-bold'>Defining Selector</CardHeader>
                <CardContent>
                    {selectorId !== '' && (
                        <div className='mb-4'>
                            <FormField
                                control={control}
                                name='disabled'
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Selector Status</FormLabel>
                                        <FormControl>
                                            <Switch
                                                {...field}
                                                value={''}
                                                data-testid='zone-management_save_selector-form_disable-switch'
                                                disabled={
                                                    selectorQuery.data === undefined
                                                        ? false
                                                        : !selectorQuery.data.allow_disable
                                                }
                                                checked={!field.value}
                                                onCheckedChange={(checked: boolean) => {
                                                    field.onChange(!checked);
                                                }}
                                            />
                                        </FormControl>
                                        <FormDescription>{!field.value ? 'Enabled' : 'Disabled'}</FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    )}
                    <p className='font-bold'>
                        {location.pathname.includes('label') ? 'Label' : 'Zone'}:{' '}
                        <span className='font-normal'>{tagQuery.data?.name}</span>
                    </p>
                    <div className='flex flex-col gap-6 mt-6'>
                        <div className='flex flex-col gap-6'>
                            <FormField
                                control={control}
                                name='name'
                                rules={{
                                    required: `Please provide a name for the Selector`,
                                }}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Name</FormLabel>
                                        <FormControl>
                                            <Input
                                                {...field}
                                                type='text'
                                                autoComplete='off'
                                                data-testid='zone-management_save_selector-form_name-input'
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
                                                data-testid='zone-management_save_selector-form_description-input'
                                                placeholder='Description Input'
                                                rows={3}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <div>
                                <Label className='text-base font-bold'>Selector Type</Label>
                                <Select
                                    data-testid='zone-management_save_selector-form_type-select'
                                    value={selectorType.toString()}
                                    onValueChange={(value: string) => {
                                        if (value === SeedTypeObjectId.toString()) {
                                            dispatch({ type: 'set-selector-type', selectorType: SeedTypeObjectId });
                                        } else if (value === SeedTypeCypher.toString()) {
                                            dispatch({ type: 'set-selector-type', selectorType: SeedTypeCypher });
                                        }
                                    }}>
                                    <SelectTrigger aria-label='select selector seed type'>
                                        <SelectValue placeholder='Choose a Selector Type' />
                                    </SelectTrigger>
                                    <SelectPortal>
                                        <SelectContent>
                                            {Object.entries(SeedTypesMap).map(([seedType, displayValue]) => (
                                                <SelectItem key={seedType} value={seedType}>
                                                    {displayValue}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </SelectPortal>
                                </Select>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>
            <div className='flex justify-end gap-2 mt-6'>
                <DeleteSelectorButton
                    selectorId={selectorId}
                    selectorData={selectorQuery.data}
                    onClick={() => {
                        setDeleteDialogOpen(true);
                    }}
                />
                <Button
                    data-testid='zone-management_save_selector-form_cancel-button'
                    variant={'secondary'}
                    onClick={() => navigate(-1)}>
                    Cancel
                </Button>
                <Button data-testid='zone-management_save_selector-form_save-button' variant={'primary'} type='submit'>
                    {selectorId === '' ? 'Save' : 'Save Edits'}
                </Button>
            </div>
            <DeleteConfirmationDialog
                open={deleteDialogOpen}
                itemName={selectorQuery.data?.name || 'Selector'}
                itemType='selector'
                onConfirm={handleDeleteSelector}
                onCancel={handleCancel}
            />
        </div>
    );
};

export default BasicInfo;
