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

import { Card, Checkbox, DialogTitle, FormField, FormItem, FormLabel, Input } from '@bloodhoundenterprise/doodleui';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Environment, EnvironmentRequest } from 'js-client-library';
import { Minus } from 'lucide-react';
import React, { useEffect } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments/useAvailableEnvironments';
import { UpdateUserRequestForm } from '../UpdateUserForm';

const EnvironmentSelectPanel: React.FC<{
    createUser?: boolean; // TODO: make required
    updateUser?: boolean; // TODO: make required
    initialData?: UpdateUserRequestForm;
    form: UseFormReturn;
}> = ({
    createUser,
    //updateUser,
    initialData,
    form,
}) => {
    const { data: availableEnvironments } = useAvailableEnvironments();

    const initialEnvironmentsSelected = initialData?.environment_targeted_access_control?.environments?.map(
        (item) => item.environment_id
    );
    const returnMappedEnvironments: any = availableEnvironments?.map((environment) => environment.id);

    const matchingEnvironmentValues = initialEnvironmentsSelected?.filter(
        (value) => returnMappedEnvironments && returnMappedEnvironments.includes(value)
    );

    const checkedEnvironments =
        !createUser && initialData?.all_environments === true
            ? returnMappedEnvironments
            : matchingEnvironmentValues || (createUser && []);

    const [searchInput, setSearchInput] = React.useState<string>('');
    const [selectedEnvironments, setSelectedEnvironments] = React.useState<string[]>(checkedEnvironments);

    const filteredEnvironments = availableEnvironments?.filter((environment: Environment) =>
        environment.name.toLowerCase().includes(searchInput.toLowerCase())
    );

    const allEnvironmentsSelected: any =
        selectedEnvironments &&
        selectedEnvironments.length === availableEnvironments?.length &&
        availableEnvironments!.length > 0;

    const allEnvironmentsIndeterminate =
        selectedEnvironments &&
        selectedEnvironments.length > 0 &&
        selectedEnvironments.length < availableEnvironments!.length;

    const handleSelectAllEnvironmentsChange = (allEnvironmentsChecked: string | boolean) => {
        if (allEnvironmentsChecked || allEnvironmentsIndeterminate) {
            const returnMappedEnvironments: string[] | undefined = availableEnvironments?.map(
                (environment) => environment.id
            );
            returnMappedEnvironments && setSelectedEnvironments(returnMappedEnvironments);
        } else {
            setSelectedEnvironments([]);
        }
    };

    const formatReturnedEnvironments: EnvironmentRequest[] = selectedEnvironments?.map((itemId: string) => ({
        environment_id: itemId,
    }));

    const handleEnvironmentSelectChange = (itemId: string, checked: string | boolean) => {
        if (checked) {
            setSelectedEnvironments((prevSelected) => [...prevSelected, itemId]);
        } else {
            setSelectedEnvironments((prevSelected) => prevSelected.filter((id) => id !== itemId));
        }
    };

    useEffect(() => {
        if (allEnvironmentsSelected) {
            form.setValue('all_environments', true);
            form.setValue('environment_targeted_access_control.environments', null);
        }

        if (allEnvironmentsIndeterminate) {
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', formatReturnedEnvironments);
        }

        if (!allEnvironmentsIndeterminate && !allEnvironmentsSelected) {
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', null);
        }
    }, [selectedEnvironments]);

    return (
        <Card className='flex-1 p-4 rounded shadow max-w-[400px]'>
            <DialogTitle>Environmental Targeted Access Control </DialogTitle>
            <div className='flex flex-col h-full pb-6' data-testid='create-user-dialog_environments-checkboxes-dialog'>
                <div className='border border-color-[#CACFD3] mt-3 box-border h-full overflow-y-auto'>
                    <div className='border border-solid border-color-[#CACFD3] flex'>
                        <FontAwesomeIcon className='ml-4 mt-3' icon={faSearch} />
                        <Input
                            variant='underlined'
                            className='w-full ml-3'
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
                                        checked={allEnvironmentsSelected || allEnvironmentsIndeterminate}
                                        id='allEnvironments'
                                        onCheckedChange={handleSelectAllEnvironmentsChange}
                                        className={
                                            allEnvironmentsSelected && '!bg-primary border-[#2C2677] dark:!bg-[#f4f4f4]'
                                        }
                                        icon={
                                            allEnvironmentsIndeterminate && (
                                                <Minus
                                                    className='h-full w-full bg-[#f4f4f4] text-neutral-dark-1 dark:bg-[#222222] dark:text-[#f4f4f4]'
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
                    <div className='flex flex-col' data-testid='create-user-dialog_environments-checkboxes-div'>
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
                                                        className='m-3 data-[state=checked]:bg-primary data-[state=checked]:border-[#2C2677]'
                                                        id='environments'
                                                        onCheckedChange={(checked) =>
                                                            handleEnvironmentSelectChange(item.id, checked)
                                                        }
                                                        value={item.name}
                                                        data-testid='create-user-dialog_environments-checkboxes'
                                                    />
                                                    <FormLabel
                                                        htmlFor='environments'
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
            </div>
        </Card>
    );
};

export default EnvironmentSelectPanel;
