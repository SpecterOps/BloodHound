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
} from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagSelector, SeedTypeCypher, SeedTypeObjectId, SeedTypesMap } from 'js-client-library';
import { FC, useContext, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useQuery } from 'react-query';
import { useLocation, useParams } from 'react-router-dom';
import { apiClient, queriesAreLoadingOrErrored } from '../../../../utils';
import SelectorFormContext from './SelectorFormContext';
import { SelectorFormInputs } from './types';

/**
 * selectorStatus takes in the selectorId from the path param in the url and the selector's data.
 * It returns a boolean value associated with whether the selector is enabled or not.
 */
const selectorStatus = (id: string, data: AssetGroupTagSelector | undefined) => {
    if (id === '') return true;
    if (data === undefined) return true;
    if (data.disabled_at !== null) return false;
    return true;
};

const BasicInfo: FC = () => {
    const location = useLocation();
    const { tierId = '', labelId, selectorId = '' } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const { dispatch, selectorType, selectorQuery } = useContext(SelectorFormContext);

    const {
        formState: { errors },
        register,
        setValue,
    } = useFormContext<SelectorFormInputs>();

    const tagQuery = useQuery({
        queryKey: ['zone-management', 'tags', tagId],
        queryFn: async () => {
            const response = await apiClient.getAssetGroupTag(tagId);
            return response.data.data['tag'];
        },
        enabled: tagId !== '',
    });

    const [enabled, setEnabled] = useState(selectorStatus(selectorId, selectorQuery.data));

    const { isLoading, isError } = queriesAreLoadingOrErrored(tagQuery, selectorQuery);

    if (isLoading) return <Skeleton />;
    if (isError) return <div>There was an error fetching the selector information.</div>;

    return (
        <Card className={'w-full max-w-[40rem] min-w-80 sm:w-80 md:w-96 lg:w-lg p-3 h-[30rem]'}>
            <CardHeader className='text-xl font-bold'>Defining Selector</CardHeader>
            <CardContent>
                {selectorId === '' ? null : (
                    <div className='mb-4'>
                        <Label htmlFor='disabled' className='text-base font-bold'>
                            Selector Status
                        </Label>
                        <div className='flex gap-2 items-center mt-2'>
                            <Switch
                                data-testid='zone-management_save_selector-form_disable-switch'
                                id='disabled'
                                checked={enabled}
                                disabled={selectorQuery.data === undefined ? false : !selectorQuery.data.allow_disable}
                                {...register('disabled')}
                                onCheckedChange={(checked: boolean) => {
                                    setEnabled((prev) => !prev);
                                    if (checked) {
                                        setValue('disabled', false);
                                    } else {
                                        setValue('disabled', true);
                                    }
                                }}
                            />
                            <p className='flex items-center ml-2'>{enabled ? 'Enabled' : 'Disabled'}</p>
                        </div>
                    </div>
                )}
                <p className='font-bold'>
                    {location.pathname.includes('label') ? 'Label' : 'Tier'}:{' '}
                    <span className='font-normal'>{tagQuery.data?.name}</span>
                </p>
                <div className='flex flex-col gap-6 mt-6'>
                    <div className='flex flex-col gap-6'>
                        <div>
                            <Label className='text-base font-bold' htmlFor='name'>
                                Name
                            </Label>
                            <Input
                                id='name'
                                data-testid='zone-management_save_selector-form_name-input'
                                {...register('name', { required: true, value: selectorQuery.data?.name })}
                                autoComplete='off'
                                className={
                                    'pointer-events-auto rounded-none text-base bg-transparent dark:bg-transparent border-t-0 border-x-0 border-b-neutral-dark-5 dark:border-b-neutral-light-5 border-b-[1px] focus-visible:outline-none focus:border-t-0 focus:border-x-0 focus-visible:ring-offset-0 focus-visible:ring-transparent focus-visible:border-secondary focus-visible:border-b-2 focus:border-secondary focus:border-b-2 dark:focus-visible:outline-none dark:focus:border-t-0 dark:focus:border-x-0 dark:focus-visible:ring-offset-0 dark:focus-visible:ring-transparent dark:focus-visible:border-secondary-variant-2 dark:focus-visible:border-b-2 dark:focus:border-secondary-variant-2 dark:focus:border-b-2 hover:border-b-2'
                                }
                            />
                            {errors.name && (
                                <p className='text-sm text-[#B44641] dark:text-[#E9827C]'>
                                    Please provide a name for the selector
                                </p>
                            )}
                        </div>
                        <div>
                            <Label htmlFor='description' className='text-base font-bold block'>
                                Description
                            </Label>
                            <textarea
                                data-testid='zone-management_save_selector-form_description-input'
                                id='description'
                                {...register('description', { value: selectorQuery.data?.description })}
                                rows={3}
                                className={
                                    'resize-none rounded-md dark:bg-neutral-dark-5 pl-2 w-full mt-2 focus-visible:outline-none focus:ring-secondary focus-visible:ring-secondary focus:outline-secondary focus-visible:outline-secondary dark:focus:ring-secondary-variant-2 dark:focus-visible:ring-secondary-variant-2 dark:focus:outline-secondary-variant-2 dark:focus-visible:outline-secondary-variant-2'
                                }
                                placeholder='Description Input'
                            />
                        </div>
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
                                <SelectTrigger
                                    className='focus-visible:outline-secondary focus:outline-secondary focus:outline-1 focus:ring-secondary dark:focus-visible:outline-secondary-variant-2 dark:focus:outline-secondary-variant-2 dark:focus:outline-1 dark:focus:ring-secondary-variant-2 hover:border-b-2'
                                    aria-label='select selector seed type'>
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
    );
};

export default BasicInfo;
