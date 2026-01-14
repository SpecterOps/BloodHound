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
import { FC, useContext, useEffect } from 'react';
import { Control } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useLocation } from 'react-router-dom';
import { usePZPathParams } from '../../../../hooks';
import { apiClient, queriesAreLoadingOrErrored } from '../../../../utils';
import { PrivilegeZonesContext } from '../../PrivilegeZonesContext';
import RuleFormContext from './RuleFormContext';
import { RuleFormInputs } from './types';

const BasicInfo: FC<{ control: Control<RuleFormInputs, any, RuleFormInputs> }> = ({ control }) => {
    const location = useLocation();
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
                                                    <strong>All Objects</strong> - Means every object, including those
                                                    tied to direct objects, is certified automatically.
                                                </p>
                                                <p>
                                                    <strong>Off</strong> - Means all certification is manual.
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
        </div>
    );
};

export default BasicInfo;
