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

import {
    Button,
    Card,
    Checkbox,
    DialogActions,
    DialogClose,
    DialogTitle,
    Form,
    FormControl,
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
    Tooltip,
} from '@bloodhoundenterprise/doodleui';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert } from '@mui/material';
import { Environment, EnvironmentRequest, Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import { Minus } from 'lucide-react';
import React, { useCallback, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments/useAvailableEnvironments';
import { apiClient } from '../../utils';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const UpdateUserForm: React.FC<{
    //onSubmit: (user: UpdateUserRequestForm) => void;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
    //open?: boolean;
    showEnvironmentAccessControls?: boolean;
}> = ({
    //onSubmit,
    userId,
    hasSelectedSelf,
    isLoading,
    error,
    showEnvironmentAccessControls,
}) => {
    const getUserQuery = useQuery(
        ['getUser', userId],
        ({ signal }) => apiClient.getUser(userId, { signal }).then((res) => res.data.data),
        {
            cacheTime: 0,
        }
    );

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data.data.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data.data)
    );

    if (getUserQuery.isLoading || getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) {
        return (
            <Card>
                <Skeleton className='rounded-md w-10' />
                <DialogActions>
                    <DialogClose asChild>
                        <Button
                            data-testid='update-user-dialog_button-cancel'
                            disabled={isLoading}
                            role='button'
                            type='button'
                            variant='tertiary'>
                            Cancel
                        </Button>
                    </DialogClose>
                </DialogActions>
            </Card>
        );
    }

    if (getUserQuery.isError || getRolesQuery.isError || listSSOProvidersQuery.isError) {
        return (
            <Card>
                <div>User not found.</div>

                <DialogActions>
                    <DialogClose asChild>
                        <Button
                            data-testid='update-user-dialog_button-cancel'
                            disabled={isLoading}
                            role='button'
                            type='button'
                            variant='tertiary'>
                            Close
                        </Button>
                    </DialogClose>
                </DialogActions>
            </Card>
        );
    }

    return (
        <UpdateUserFormInner
            initialData={{
                emailAddress: getUserQuery.data.email_address || '',
                principal: getUserQuery.data.principal_name || '',
                firstName: getUserQuery.data.first_name || '',
                lastName: getUserQuery.data.last_name || '',
                SSOProviderId: getUserQuery.data.sso_provider_id?.toString() || undefined,
                roles: getUserQuery.data.roles ? getUserQuery.data.roles?.map((role: any) => role.id) : [],
                all_environments: getUserQuery.data.all_environments,
                environment_targeted_access_control: {
                    environments:
                        getUserQuery.data.all_environments === false
                            ? getUserQuery.data.environment_targeted_access_control?.map(
                                  (environment: EnvironmentRequest) => environment
                              )
                            : null,
                },
            }}
            error={error}
            hasSelectedSelf={hasSelectedSelf}
            isLoading={isLoading}
            //onSubmit={onSubmit}
            roles={getRolesQuery.data}
            SSOProviders={listSSOProvidersQuery.data}
            showEnvironmentAccessControls={showEnvironmentAccessControls}
        />
    );
};

const UpdateUserFormInner: React.FC<{
    error: any;
    hasSelectedSelf: boolean;
    initialData: UpdateUserRequestForm;
    isLoading: boolean;
    //open?: boolean;
    //onSubmit: (user: UpdateUserRequestForm) => void;
    roles?: Role[];
    showEnvironmentAccessControls?: boolean;
    SSOProviders?: SSOProvider[];
}> = ({
    error,
    hasSelectedSelf,
    initialData,
    isLoading,
    //onSubmit,
    roles,
    showEnvironmentAccessControls,
    SSOProviders,
}) => {
    const form = useForm<UpdateUserRequestForm & { authenticationMethod: 'sso' | 'password' }>({
        defaultValues: {
            ...initialData,
            authenticationMethod: initialData.SSOProviderId ? 'sso' : 'password',
        },
    });

    const [selectedRoleValue, setSelectedRoleValue] = useState<number[]>(initialData.roles);
    const roleInputValue = form.watch('roles');
    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';
    const authenticationMethod = form.watch('authenticationMethod');
    const [searchInput, setSearchInput] = useState<string>('');

    const selectedSSOProviderHasRoleProvisionEnabled = !!SSOProviders?.find(
        ({ id }) => id === Number(form.watch('SSOProviderId'))
    )?.config?.auto_provision?.role_provision;

    const { data: availableEnvironments } = useAvailableEnvironments();

    const initialEnvironmentsSelected = initialData.environment_targeted_access_control?.environments?.map(
        (item) => item.environment_id
    );

    const filteredEnvironments = availableEnvironments?.filter((environment: Environment) =>
        environment.name.toLowerCase().includes(searchInput.toLowerCase())
    );

    const returnMappedEnvironments: any = availableEnvironments?.map((environment) => environment.id);

    // TODO: REMOVE?
    /*
    const formatReturnedEnvironments: EnvironmentRequest[] | null = returnMappedEnvironments?.map((item: string) => ({
        environment_id: item,
    }));
    */

    const matchingEnvironmentValues = initialEnvironmentsSelected?.filter(
        (value) => returnMappedEnvironments && returnMappedEnvironments.includes(value)
    );

    const checkedEnvironments =
        initialData.all_environments === true ? returnMappedEnvironments : matchingEnvironmentValues;

    const [selectedEnvironments, setSelectedEnvironments] = useState<any>(checkedEnvironments);

    const handleSelectAllEnvironmentsChange = (allEnvironmentsChecked: any) => {
        if (allEnvironmentsChecked) {
            setSelectedEnvironments(returnMappedEnvironments);
            form.setValue('all_environments', true);
            form.setValue('environment_targeted_access_control.environments', null);
        } else {
            setSelectedEnvironments([]);
        }
    };

    const formatSelectedEnvironments: EnvironmentRequest[] | null = selectedEnvironments?.map((item: string) => ({
        environment_id: item,
    }));

    const handleEnvironmentSelectChange = (itemId: string, checked: string | boolean) => {
        if (checked) {
            setSelectedEnvironments((prevSelected: any) => [...prevSelected, itemId]);
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', formatSelectedEnvironments);
        } else {
            setSelectedEnvironments((prevSelected: any) => prevSelected?.filter((id: string) => id !== itemId));
            form.setValue('all_environments', false);
            form.setValue('environment_targeted_access_control.environments', formatSelectedEnvironments);
        }
    };

    const allEnvironmentsSelected =
        selectedEnvironments &&
        selectedEnvironments.length === availableEnvironments?.length &&
        availableEnvironments!.length > 0;

    const allEnvironmentsCheckboxRef = React.useRef<HTMLButtonElement>(null);
    const allEnvironmentsIndeterminate =
        selectedEnvironments &&
        selectedEnvironments.length > 0 &&
        selectedEnvironments.length < availableEnvironments!.length;

    const onError = () => {
        if (error) {
            const errMsg = error.response?.data?.errors[0]?.message.toLowerCase();
            if (error.response?.status === 400) {
                if (errMsg.includes('role provision enabled')) {
                    form.setError('root.generic', {
                        type: 'custom',
                        message: 'Cannot modify user roles for role provision enabled SSO providers.',
                    });
                }
            } else if (error.response?.status === 409) {
                if (errMsg.includes('principal name')) {
                    form.setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (errMsg.includes('email')) {
                    form.setError('emailAddress', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    form.setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            } else {
                form.setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    };

    const onSubmit = useCallback(
        (user: UpdateUserRequestForm) => {
            user;

            if (authenticationMethod === 'password') {
                form.setValue('SSOProviderId', undefined);
            }

            // user selects all environment checkboxes, all_environments should be true and environment_targeted_access_control.environments should be null
            if (allEnvironmentsCheckboxRef.current) {
                if (allEnvironmentsCheckboxRef.current.dataset.state === 'checked') {
                    form.setValue('all_environments', true);
                    form.setValue('environment_targeted_access_control.environments', null);
                }
            }

            // user unselects all environment checkboxes, all_environments should be false and environment_targeted_access_control.environments should be null
            if (allEnvironmentsCheckboxRef.current) {
                if (
                    allEnvironmentsCheckboxRef.current.dataset.state === 'unchecked' &&
                    selectedEnvironments.length === 0
                ) {
                    form.setValue('all_environments', false);
                    form.setValue('environment_targeted_access_control.environments', null);
                }
            }

            // user selects between 0 and all environment checkboxes, all_environments should be false and environment_targeted_access_control.environments should be null
            if (allEnvironmentsCheckboxRef.current) {
                if (
                    allEnvironmentsCheckboxRef.current.dataset.state === 'indeterminate' &&
                    selectedEnvironments.length > 0
                ) {
                    form.setValue('all_environments', false);
                    form.setValue('environment_targeted_access_control.environments', formatSelectedEnvironments);
                }
            }

            if (allEnvironmentsCheckboxRef.current) {
                allEnvironmentsCheckboxRef.current.dataset.state = allEnvironmentsIndeterminate
                    ? 'indeterminate'
                    : allEnvironmentsSelected
                      ? 'checked'
                      : 'unchecked';
            }
        },
        [
            authenticationMethod,
            form,
            form.setValue,
            allEnvironmentsIndeterminate,
            allEnvironmentsSelected,
            formatSelectedEnvironments,
            selectedEnvironments,
        ]
    );

    useEffect(() => {
        // on submit for password and setvalues on form and error states
        /*
        if (authenticationMethod === 'password') {
            form.setValue('SSOProviderId', undefined);
        }
        */
        /*
        if (error) {
            const errMsg = error.response?.data?.errors[0]?.message.toLowerCase();
            if (error.response?.status === 400) {
                if (errMsg.includes('role provision enabled')) {
                    form.setError('root.generic', {
                        type: 'custom',
                        message: 'Cannot modify user roles for role provision enabled SSO providers.',
                    });
                }
            } else if (error.response?.status === 409) {
                if (errMsg.includes('principal name')) {
                    form.setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (errMsg.includes('email')) {
                    form.setError('emailAddress', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    form.setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            } else {
                form.setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
        */
    }, [
        /*
        authenticationMethod,
        error,
        form,
        form.setError,
        form.setValue,
        allEnvironmentsIndeterminate,
        allEnvironmentsSelected,
        formatSelectedEnvironments,
                selectedEnvironments,

        */
        form,
        form.setValue,
        formatSelectedEnvironments,
        selectedEnvironments,
        checkedEnvironments, // this is making tests weird
    ]);

    console.log(form.watch('all_environments'));
    console.log(form.watch('environment_targeted_access_control.environments'));

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(onSubmit, onError)}>
                <div className='flex gap-x-4 justify-center'>
                    <Card className='p-6 rounded shadow max-w-[600px] w-full'>
                        <DialogTitle>{'Edit User'}</DialogTitle>

                        <div className='flex flex-col mt-4 w-full' data-testid='update-user-dialog_dialog-content'>
                            {!hasSelectedSelf && (
                                <div className='mb-4'>
                                    <FormField
                                        name='roles.0'
                                        control={form.control}
                                        defaultValue={1}
                                        rules={{
                                            required: 'Role is required',
                                        }}
                                        render={({ field }) => (
                                            <>
                                                <FormItem>
                                                    <div className='flex row'>
                                                        <FormLabel className='mr-2 font-medium !text-sm' htmlFor='role'>
                                                            Role
                                                        </FormLabel>
                                                        <Tooltip
                                                            tooltip='Only User, Read-Only, Upload-Only roles contain the limited access functionality.'
                                                            contentProps={{
                                                                className: 'max-w-80 dark:bg-neutral-dark-5 border-0',
                                                            }}
                                                        />
                                                    </div>
                                                    <FormControl>
                                                        <Select
                                                            onValueChange={(field) => {
                                                                form.setValue('roles', [Number(field)]);
                                                                setSelectedRoleValue([Number(field)]);
                                                            }}
                                                            value={String(selectedRoleValue)}>
                                                            <FormControl className='pointer-events-auto'>
                                                                <SelectTrigger
                                                                    variant='underlined'
                                                                    className='bg-transparent'
                                                                    id='role'
                                                                    disabled={
                                                                        selectedSSOProviderHasRoleProvisionEnabled
                                                                    }>
                                                                    <SelectValue placeholder={field.value} />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectPortal>
                                                                <SelectContent>
                                                                    {roles?.map((role: Role) => (
                                                                        <SelectItem
                                                                            className='hover:cursor-pointer'
                                                                            key={role.id}
                                                                            role='option'
                                                                            value={role.id.toString()}>
                                                                            {role.name}
                                                                        </SelectItem>
                                                                    ))}
                                                                </SelectContent>
                                                            </SelectPortal>
                                                        </Select>
                                                    </FormControl>
                                                    {selectedSSOProviderHasRoleProvisionEnabled && (
                                                        <FormMessage id='role-helper-text'>
                                                            SSO Provider has enabled role provision.
                                                        </FormMessage>
                                                    )}
                                                </FormItem>
                                            </>
                                        )}
                                    />
                                </div>
                            )}

                            <div className='mb-4'>
                                <FormField
                                    control={form.control}
                                    name='emailAddress'
                                    rules={{
                                        required: 'Email Address is required',
                                        maxLength: {
                                            value: MAX_EMAIL_LENGTH,
                                            message: `Email address must be less than ${MAX_EMAIL_LENGTH} characters`,
                                        },
                                        pattern: {
                                            value: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
                                            message: 'Please follow the example@domain.com format',
                                        },
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className='font-medium !text-sm' htmlFor='emailAddress'>
                                                Email Address
                                            </FormLabel>
                                            <FormControl>
                                                <Input {...field} id='emailAddress' type='email' />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <div className='mb-4'>
                                <FormField
                                    name='principal'
                                    control={form.control}
                                    rules={{
                                        required: 'Principal Name is required',
                                        maxLength: {
                                            value: MAX_NAME_LENGTH,
                                            message: `Principal Name must be less than ${MAX_NAME_LENGTH} characters`,
                                        },
                                        minLength: {
                                            value: MIN_NAME_LENGTH,
                                            message: `Principal Name must be ${MIN_NAME_LENGTH} characters or more`,
                                        },
                                        validate: (value) => {
                                            const trimmed = value.trim();
                                            if (value !== trimmed) {
                                                return 'Principal Name does not allow leading or trailing spaces';
                                            }
                                            return true;
                                        },
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className='font-medium !text-sm' htmlFor='principal'>
                                                Principal Name
                                            </FormLabel>
                                            <FormControl>
                                                <Input {...field} id='principal' />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <div className='mb-4'>
                                <FormField
                                    name='firstName'
                                    control={form.control}
                                    rules={{
                                        required: 'First Name is required',
                                        maxLength: {
                                            value: MAX_NAME_LENGTH,
                                            message: `First Name must be less than ${MAX_NAME_LENGTH} characters`,
                                        },
                                        minLength: {
                                            value: MIN_NAME_LENGTH,
                                            message: `First Name must be ${MIN_NAME_LENGTH} characters or more`,
                                        },
                                        validate: (value) => {
                                            const trimmed = value.trim();
                                            if (value !== trimmed) {
                                                return 'First Name does not allow leading or trailing spaces';
                                            }
                                            return true;
                                        },
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className='font-medium !text-sm' htmlFor='firstName'>
                                                First Name
                                            </FormLabel>
                                            <FormControl>
                                                <Input {...field} id='firstName' />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <div className='mb-4'>
                                <FormField
                                    name='lastName'
                                    control={form.control}
                                    rules={{
                                        required: 'Last Name is required',
                                        maxLength: {
                                            value: MAX_NAME_LENGTH,
                                            message: `Last Name must be less than ${MAX_NAME_LENGTH} characters`,
                                        },
                                        minLength: {
                                            value: MIN_NAME_LENGTH,
                                            message: `Last Name must be ${MIN_NAME_LENGTH} characters or more`,
                                        },
                                        validate: (value) => {
                                            const trimmed = value.trim();
                                            if (value !== trimmed) {
                                                return 'Last Name does not allow leading or trailing spaces';
                                            }
                                            return true;
                                        },
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className='font-medium !text-sm' htmlFor='lastName'>
                                                Last Name
                                            </FormLabel>
                                            <FormControl>
                                                <Input {...field} id='lastName' />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <>
                                {!hasSelectedSelf && (
                                    <div className=''>
                                        <FormField
                                            name='authenticationMethod'
                                            control={form.control}
                                            rules={{
                                                required: 'Authentication Method is required',
                                            }}
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel
                                                        className='font-medium !text-sm'
                                                        htmlFor='authenticationMethod'>
                                                        Authentication Method
                                                    </FormLabel>

                                                    <Select
                                                        defaultValue={field.value}
                                                        onValueChange={(field: any) => {
                                                            form.setValue('authenticationMethod', field);
                                                            //setAuthenticationMethod(field);
                                                        }}
                                                        value={field.value}>
                                                        <FormControl className='pointer-events-auto'>
                                                            <SelectTrigger
                                                                variant='underlined'
                                                                className='bg-transparent'
                                                                id='authenticationMethod'>
                                                                <SelectValue
                                                                    placeholder={
                                                                        authenticationMethod === 'password'
                                                                            ? 'Username / Password'
                                                                            : 'Single Sign-On (SSO)'
                                                                    }
                                                                />
                                                            </SelectTrigger>
                                                        </FormControl>
                                                        <SelectPortal>
                                                            <SelectContent>
                                                                <SelectItem value='password'>
                                                                    Username / Password
                                                                </SelectItem>
                                                                {SSOProviders && SSOProviders.length > 0 && (
                                                                    <SelectItem value='sso'>
                                                                        Single Sign-On (SSO)
                                                                    </SelectItem>
                                                                )}
                                                            </SelectContent>
                                                        </SelectPortal>
                                                    </Select>
                                                </FormItem>
                                            )}
                                        />
                                    </div>
                                )}

                                {authenticationMethod === 'sso' && !hasSelectedSelf && (
                                    <div className=''>
                                        <FormField
                                            name='SSOProviderId'
                                            control={form.control}
                                            rules={{
                                                required: 'SSO Provider is required',
                                            }}
                                            render={({ field }) => (
                                                <FormItem>
                                                    <FormLabel
                                                        className='font-medium !text-sm'
                                                        htmlFor='sso'
                                                        id='SSOProviderId-label'>
                                                        SSO Provider
                                                    </FormLabel>

                                                    <Select
                                                        onValueChange={(field: any) => {
                                                            form.setValue('authenticationMethod', field.value);
                                                        }}
                                                        value={field.value}>
                                                        <FormControl>
                                                            <SelectTrigger
                                                                variant='underlined'
                                                                className='bg-transparent'
                                                                id='sso'>
                                                                <SelectValue placeholder='SSO Provider' />
                                                            </SelectTrigger>
                                                        </FormControl>
                                                        <SelectPortal>
                                                            <SelectContent>
                                                                {SSOProviders?.map((SSOProvider: SSOProvider) => (
                                                                    <SelectItem
                                                                        role='option'
                                                                        value={SSOProvider.id.toString()}
                                                                        key={SSOProvider.id}>
                                                                        {SSOProvider.name}
                                                                    </SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </SelectPortal>
                                                    </Select>
                                                </FormItem>
                                            )}
                                        />
                                    </div>
                                )}
                            </>

                            {!!form.formState.errors.root?.generic && (
                                <div>
                                    <Alert severity='error'>{form.formState.errors.root.generic.message}</Alert>
                                </div>
                            )}
                        </div>
                        <DialogActions className='mt-8 flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button
                                    data-testid='update-user-dialog_button-cancel'
                                    disabled={isLoading}
                                    role='button'
                                    type='button'
                                    variant='tertiary'>
                                    Cancel
                                </Button>
                            </DialogClose>
                            <Button
                                data-testid='update-user-dialog_button-save'
                                disabled={isLoading}
                                role='button'
                                type='submit'>
                                Save
                            </Button>
                        </DialogActions>
                    </Card>
                    {showEnvironmentAccessControls && selectedRole && (
                        <Card className='flex-1 p-4 rounded shadow max-w-[400px]'>
                            <DialogTitle>Environmental Targeted Access Control</DialogTitle>
                            <div
                                className='flex flex-col h-full pb-6'
                                data-testid='update-user-dialog_environments-checkboxes-dialog'>
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
                                                        ref={allEnvironmentsCheckboxRef}
                                                        checked={
                                                            allEnvironmentsSelected || allEnvironmentsIndeterminate
                                                        }
                                                        id='allEnvironments'
                                                        onCheckedChange={handleSelectAllEnvironmentsChange}
                                                        className={
                                                            allEnvironmentsIndeterminate
                                                                ? 'data-[state=indeterminate]'
                                                                : 'data-[state=checked]:bg-primary data-[state=checked]:border-[#2C2677]'
                                                        }
                                                        icon={
                                                            allEnvironmentsIndeterminate && (
                                                                <Minus
                                                                    className='h-full w-full'
                                                                    absoluteStrokeWidth={true}
                                                                    strokeWidth={3}
                                                                />
                                                            )
                                                        }
                                                        data-testid='update-user-dialog_select-all-environments-checkbox'
                                                    />
                                                    <FormLabel
                                                        className='ml-3 w-full cursor-pointer font-medium !text-sm'
                                                        htmlFor='allEnvironments'>
                                                        Select All Environments
                                                    </FormLabel>
                                                </FormItem>
                                            )}
                                        />
                                    </div>
                                    <div
                                        className='flex flex-col'
                                        data-testid='update-user-dialog_environments-checkboxes'>
                                        {filteredEnvironments &&
                                            filteredEnvironments?.map((item) => {
                                                return (
                                                    <div
                                                        key={item.id}
                                                        className='flex justify-start items-center ml-5'
                                                        data-testid='update-user-dialog_environments-checkbox'>
                                                        <FormField
                                                            name='environment_targeted_access_control.environments'
                                                            control={form.control}
                                                            render={() => (
                                                                <FormItem className='flex flex-row items-center'>
                                                                    <Checkbox
                                                                        checked={selectedEnvironments?.includes(
                                                                            item.id
                                                                        )}
                                                                        className='m-2 data-[state=checked]:bg-primary data-[state=checked]:border-[#2C2677]'
                                                                        id='environments'
                                                                        onCheckedChange={(checked) =>
                                                                            handleEnvironmentSelectChange(
                                                                                item.id,
                                                                                checked
                                                                            )
                                                                        }
                                                                        value={item.name}
                                                                    />
                                                                    <FormLabel
                                                                        className=' w-full cursor-pointer ml-3 w-full cursor-pointer font-medium !text-sm'
                                                                        htmlFor='environments'>
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
                    )}
                </div>
            </form>
        </Form>
    );
};

export default UpdateUserForm;
