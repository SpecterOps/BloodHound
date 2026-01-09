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
import { AssetGroupTagSelectorAutoCertifyMap, SeedTypeCypher } from 'js-client-library';
import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { Control } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useLocation } from 'react-router-dom';
import { DeleteConfirmationDialog } from '../../../../components';
import { usePZPathParams } from '../../../../hooks';
import { useDeleteRule } from '../../../../hooks/useAssetGroupTags';
import { useNotifications } from '../../../../providers';
import { detailsPath, privilegeZonesPath } from '../../../../routes';
import { apiClient, queriesAreLoadingOrErrored, useAppNavigate } from '../../../../utils';
import { PrivilegeZonesContext } from '../../PrivilegeZonesContext';
import { handleError } from '../utils';
import DeleteRuleButton from './DeleteRuleButton';
import RuleFormContext from './RuleFormContext';
import { RuleFormInputs } from './types';

const BasicInfo: FC<{ control: Control<RuleFormInputs, any, RuleFormInputs> }> = ({ control }) => {
    const location = useLocation();
    const navigate = useAppNavigate();
    const { ruleId = '', tagId, tagType, tagTypeDisplay } = usePZPathParams();
    const { dispatch, ruleQuery } = useContext(RuleFormContext);
    const { Certification } = useContext(PrivilegeZonesContext);
    const receivedData = location.state;

    useEffect(() => {
        if (receivedData) {
            dispatch({ type: 'set-rule-type', ruleType: SeedTypeCypher });
        }
    }, [dispatch, receivedData]);

    const tagQuery = useQuery({
        queryKey: ['privilege-zones', 'tags', tagId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTag(tagId);
            return response.data.data['tag'];
        },
        enabled: tagId !== '',
    });

    const { isLoading, isError } = queriesAreLoadingOrErrored(tagQuery, ruleQuery);
    const { addNotification } = useNotifications();
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const deleteRuleMutation = useDeleteRule();

    const handleDeleteRule = useCallback(async () => {
        try {
            if (!tagId || !ruleId) throw new Error(`Missing required entity IDs; tagId: ${tagId} , ruleId: ${ruleId}`);

            await deleteRuleMutation.mutateAsync({ tagId, ruleId });

            addNotification('Rule was deleted successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            setDeleteDialogOpen(false);

            navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${detailsPath}`);
        } catch (error) {
            handleError(error, 'deleting', 'rule', addNotification);
        }
    }, [tagId, ruleId, navigate, deleteRuleMutation, addNotification, tagType]);

    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

    if (isLoading) return <Skeleton />;
    if (isError) return <div>There was an error fetching the rule information.</div>;

    return (
        <div className={'max-lg:w-full w-96 h-[36rem] '}>
            <Card className={'p-3'}>
                <CardHeader className='text-xl font-bold'>Defining Rule</CardHeader>
                <CardContent>
                    {ruleId !== '' && (
                        <div className='mb-4'>
                            <FormField
                                control={control}
                                name='disabled'
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Enable Rule</FormLabel>
                                        <div className='flex gap-3'>
                                            <FormControl>
                                                <Switch
                                                    {...field}
                                                    value={''}
                                                    data-testid='privilege-zones_save_rule-form_disable-switch'
                                                    disabled={
                                                        ruleQuery.data === undefined
                                                            ? false
                                                            : !ruleQuery.data.allow_disable
                                                    }
                                                    checked={!field.value}
                                                    onCheckedChange={(checked: boolean) => {
                                                        field.onChange(!checked);
                                                    }}
                                                />
                                            </FormControl>
                                            <FormDescription>{!field.value ? 'Enabled' : 'Disabled'}</FormDescription>
                                        </div>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    )}
                    <p className='font-bold'>
                        {tagTypeDisplay}: <span className='font-normal'>{tagQuery.data?.name}</span>
                    </p>
                    <div className='flex flex-col gap-6 mt-6'>
                        <div className='flex flex-col gap-6'>
                            <FormField
                                control={control}
                                name='name'
                                rules={{
                                    required: `Please provide a name for the Rule`,
                                }}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Name</FormLabel>
                                        <FormControl>
                                            <Input
                                                {...field}
                                                type='text'
                                                autoComplete='off'
                                                data-testid='privilege-zones_save_rule-form_name-input'
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
                                                data-testid='privilege-zones_save_rule-form_description-input'
                                                placeholder='Description Input'
                                                rows={3}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            {tagType === 'zones' && Certification && (
                                <FormField
                                    control={control}
                                    name='auto_certify'
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel aria-labelledby='auto_certify'>
                                                Automatic Certification
                                            </FormLabel>
                                            <div className='text-sm [&>p]:mt-2'>
                                                Choose how new objects are certified.
                                                <p>
                                                    <strong>Direct Objects</strong> - Only the object explicitly
                                                    selected either by object ID or cypher query are certified
                                                    automatically.
                                                </p>
                                                <p>
                                                    <strong>All Objects</strong> - means every object, including those
                                                    tied to direct objects, is certified automatically.
                                                </p>
                                                <p>
                                                    <strong>Off</strong> - means all certification is manual.
                                                </p>
                                            </div>
                                            <Select
                                                value={field.value}
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}>
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue
                                                            data-testid='privilege-zones_save_rule-form_default-certify'
                                                            placeholder='Off'
                                                            {...field}
                                                        />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        {Object.entries(AssetGroupTagSelectorAutoCertifyMap).map(
                                                            ([autoCertifyOption, displayValue]) => (
                                                                <SelectItem
                                                                    key={autoCertifyOption}
                                                                    value={autoCertifyOption}>
                                                                    {displayValue}
                                                                </SelectItem>
                                                            )
                                                        )}
                                                    </SelectContent>
                                                </SelectPortal>
                                            </Select>
                                        </FormItem>
                                    )}
                                />
                            )}
                        </div>
                    </div>
                </CardContent>
            </Card>
            <div className='flex justify-end gap-2 mt-6'>
                <DeleteRuleButton
                    ruleId={ruleId}
                    ruleData={ruleQuery.data}
                    onClick={() => {
                        setDeleteDialogOpen(true);
                    }}
                />
                <Button
                    data-testid='privilege-zones_save_rule-form_cancel-button'
                    variant={'secondary'}
                    onClick={() => navigate(-1)}>
                    Back
                </Button>
                <Button data-testid='privilege-zones_save_rule-form_save-button' variant={'primary'} type='submit'>
                    {ruleId === '' ? 'Create Selector' : 'Save Edits'}
                </Button>
            </div>
            <DeleteConfirmationDialog
                open={deleteDialogOpen}
                itemName={ruleQuery.data?.name || 'Rule'}
                itemType='rule'
                onConfirm={handleDeleteRule}
                onCancel={handleCancel}
            />
        </div>
    );
};

export default BasicInfo;
