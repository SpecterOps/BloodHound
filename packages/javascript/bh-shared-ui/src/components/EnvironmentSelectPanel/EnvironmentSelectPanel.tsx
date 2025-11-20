// Copyright 2023 Specter Ops, Inc.
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

import {Card, Checkbox, DialogTitle, FormField, FormItem, FormLabel, Input} from '@bloodhoundenterprise/doodleui';
import {faSearch} from '@fortawesome/free-solid-svg-icons';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {Environment, EnvironmentRequest} from 'js-client-library';
import {Minus} from 'lucide-react';
import React, {useEffect, useMemo, useState} from 'react';
import {UseFormReturn} from 'react-hook-form';
import {CreateUserRequestForm, useExploreToggleable} from '../..';
import {useAvailableEnvironments} from '../../hooks';
import {cn} from '../../utils';
import {UpdateUserRequestForm} from '../UpdateUserForm';
import {Typography} from "@mui/material";

const EnvironmentSelectPanel: React.FC<{
    initialData?: UpdateUserRequestForm;
    form: UseFormReturn<CreateUserRequestForm | UpdateUserRequestForm>;
}> = ({initialData, form}) => {
    const {data, isLoading} = useAvailableEnvironments();

    const allEnvironmentIds = data?.map((environment) => environment.id);

    // Get an array of all the environment IDs from the initial data and match them against the available environments from the API
    const mapInitialEnvironments = () => {
        return allEnvironmentIds?.filter((id) => {
            return initialData?.environment_targeted_access_control?.environments?.some((e) => e.environment_id === id);
        });
    };

    // If the all_environments flag is checked in the initial data, we can skip the above calculations
    const initialEnvironments = initialData?.all_environments ? allEnvironmentIds : mapInitialEnvironments();

    if (!isLoading) {
        return (
            <EnvironmentSelectPanelInner
                availableEnvironments={data || []}
                initialEnvironments={initialEnvironments}
                form={form}
            />
        );
    } else {
        return null;
    }
};

const EnvironmentSelectPanelInner: React.FC<{
    initialEnvironments?: string[];
    availableEnvironments: Environment[];
    form: UseFormReturn<CreateUserRequestForm | UpdateUserRequestForm>;
}> = ({initialEnvironments = [], availableEnvironments, form}) => {
    const [searchInput, setSearchInput] = useState<string>('');
    const [selectedEnvironments, setSelectedEnvironments] = useState<string[]>(initialEnvironments);
    const [exploreEnabled, setExploreEnabled] = useState<boolean>(false);
    const exploreToggleable = useExploreToggleable();


    const filteredEnvironments = useMemo(() => {
        const searchInputLowered = searchInput.toLowerCase();
        return availableEnvironments?.filter((environment: Environment) =>
            environment.name.toLowerCase().includes(searchInputLowered)
        );
    }, [searchInput, availableEnvironments]);

    const areAllEnvironmentsSelected =
        selectedEnvironments &&
        selectedEnvironments.length === (availableEnvironments?.length ?? 0) &&
        (availableEnvironments?.length ?? 0) > 0;

    const areAllEnvironmentsIndeterminate =
        selectedEnvironments &&
        selectedEnvironments.length > 0 &&
        selectedEnvironments.length < (availableEnvironments?.length ?? 0);

    const handleSelectAllEnvironmentsChange = (allEnvironmentsChecked: string | boolean) => {
        if (allEnvironmentsChecked || areAllEnvironmentsIndeterminate) {
            const returnMappedEnvironments: string[] | undefined = availableEnvironments?.map(
                (environment) => environment.id
            );
            returnMappedEnvironments && setSelectedEnvironments(returnMappedEnvironments);
        } else {
            setSelectedEnvironments([]);
        }
    };

    const handleEnvironmentSelectChange = (itemId: string, checked: string | boolean) => {
        if (checked) {
            setSelectedEnvironments((prevSelected) => (prevSelected ? [...prevSelected, itemId] : [itemId]));
        } else {
            setSelectedEnvironments((prevSelected) => prevSelected.filter((id) => id !== itemId));
        }
    };

    const handleExploreEnabledCheckChanged = (checked: boolean) => {
        form.setValue('explore_enabled', checked);
        setExploreEnabled(checked);
    }

    useEffect(() => {
        const formatReturnedEnvironments: EnvironmentRequest[] =
            selectedEnvironments &&
            selectedEnvironments.map((itemId: string) => ({
                environment_id: itemId,
            }));

        if (areAllEnvironmentsSelected) {
            form.setValue('all_environments', true);
            form.setValue('environment_targeted_access_control.environments', null);
        }

        if (areAllEnvironmentsIndeterminate) {
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', formatReturnedEnvironments);
        }

        if (!areAllEnvironmentsIndeterminate && !areAllEnvironmentsSelected) {
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', null);
        }
    }, [selectedEnvironments, areAllEnvironmentsSelected, areAllEnvironmentsIndeterminate, form]);

    return (
        <Card className='flex-1 p-4 rounded shadow max-w-[400px] overflow-y-hidden'>
            <DialogTitle>Environmental Targeted Access Control </DialogTitle>
            <div
                className='flex flex-col relative pb-2 h-full'
                data-testid='create-user-dialog_environments-checkboxes-dialog'>
                <div className='border border-neutral-5 mt-3 flex-1 max-h-[720px]'>
                    <div className='flex border-b border-neutral-dark-1 dark:border-b-neutral-light-5'>
                        <FontAwesomeIcon className='ml-4 mt-3' icon={faSearch}/>
                        <Input
                            variant='underlined'
                            className='w-full ml-3 border-b-0 focus:!border-b-0 hover:!border-b-0 dark:focus:!border-b-0 dark:hover:!border-b-0'
                            id='search'
                            type='text'
                            placeholder='Search'
                            onChange={(e) => {
                                setSearchInput(e.target.value);
                            }}
                        />
                    </div>
                    <div
                        className='flex flex-row ml-4 mt-6 mb-2 items-center'
                        data-testid='create-user-dialog_select-all-environments-checkbox-div'>
                        <FormField
                            name='all_environments'
                            control={form.control}
                            render={() => (
                                <FormItem className='flex flex-row items-center'>
                                    <Checkbox
                                        checked={areAllEnvironmentsSelected || areAllEnvironmentsIndeterminate}
                                        id='allEnvironments'
                                        onCheckedChange={handleSelectAllEnvironmentsChange}
                                        className={cn(
                                            areAllEnvironmentsSelected &&
                                            '!bg-primary border-neutral-dark-1 dark:!bg-neutral-light-2'
                                        )}
                                        icon={
                                            areAllEnvironmentsIndeterminate && (
                                                <Minus
                                                    className='h-full w-full bg-neutral-light-2 text-neutral-dark-1 dark:bg-neutral-dark-2 dark:text-neutral-light-2'
                                                    absoluteStrokeWidth={true}
                                                    strokeWidth={3}
                                                />
                                            )
                                        }
                                        data-testid='create-user-dialog_select-all-environments-checkbox'
                                    />
                                    <FormLabel
                                        htmlFor='allEnvironments'
                                        className='ml-3 w-full cursor-pointer font-normal'>
                                        Select All Environments
                                    </FormLabel>
                                </FormItem>
                            )}
                        />
                    </div>
                    <div
                        className='flex flex-col max-h-[640px] overflow-y-auto'
                        data-testid='create-user-dialog_environments-checkboxes-div'>
                        {filteredEnvironments &&
                            filteredEnvironments?.map((item) => {
                                return (
                                    <div
                                        key={item.id}
                                        className='flex justify-start items-center ml-5'
                                        data-testid='create-user-dialog_environments-checkbox'>
                                        <FormField
                                            name='environment_targeted_access_control.environments'
                                            control={form.control}
                                            render={() => (
                                                <FormItem className='flex flex-row items-center'>
                                                    <Checkbox
                                                        checked={
                                                            selectedEnvironments &&
                                                            selectedEnvironments.includes(item.id)
                                                        }
                                                        className='m-3 data-[state=checked]:bg-primary data-[state=checked]:border-neutral-dark-2'
                                                        id={item.id}
                                                        onCheckedChange={(checked) =>
                                                            handleEnvironmentSelectChange(item.id, checked)
                                                        }
                                                        value={item.name}
                                                        data-testid='create-user-dialog_environments-checkboxes'
                                                    />
                                                    <FormLabel
                                                        htmlFor={item.id}
                                                        className='mr-3 w-full cursor-pointer font-normal'>
                                                        {item.name}
                                                    </FormLabel>
                                                </FormItem>
                                            )}
                                        />
                                    </div>
                                );
                            })}
                    </div>
                </div>
                {exploreToggleable &&
                    (
                        <div
                            className='flex flex-row ml-4 mt-6 mb-2 items-center'
                            data-testid='create-user-dialog_select-explore-enabled-checkbox-div'>
                            <FormField
                                name='explore_enabled'
                                control={form.control}
                                render={() => (
                                    <FormItem>
                                        <Checkbox data-testid='create-user-dialog_select-explore-enabled-checkbox'
                                                  checked={exploreEnabled}
                                                  onCheckedChange={handleExploreEnabledCheckChanged}
                                                  defaultChecked={false}
                                                  name='exploreEnabled'
                                                  value="Enable Explore Access"/>
                                        <FormLabel
                                            htmlFor='exploreEnabled'
                                            className='ml-3 w-full cursor-pointer font-normal'>
                                            Enable Explore Access
                                        </FormLabel>
                                        <Typography className='italic'>
                                            Explore access can not be filtered by
                                            environment. Granting a user access to this page may include results from
                                            other environments.
                                        </Typography>
                                    </FormItem>
                                )}/>
                        </div>
                    )
                }
            </div>
        </Card>
    );
};

export default EnvironmentSelectPanel;
