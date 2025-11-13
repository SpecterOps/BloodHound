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
import { Alert } from '@mui/material';
import { EnvironmentRequest, Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import React from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { useListDisplayRoles } from '../../hooks/useListDisplayRoles/useListDisplayRoles';
import { apiClient } from '../../utils';
import EnvironmentSelectPanel from '../EnvironmentSelectPanel';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'sso_provider_id'> & {
    sso_provider_id: string | undefined;
};

const UpdateUserForm: React.FC<{
    onSubmit: (user: UpdateUserRequestForm) => void;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
    showEnvironmentAccessControls?: boolean;
}> = ({ onSubmit, userId, hasSelectedSelf, isLoading, error, showEnvironmentAccessControls }) => {
    const getUserQuery = useQuery(
        ['getUser', userId],
        ({ signal }) => apiClient.getUser(userId, { signal }).then((res) => res.data.data),
        {
            cacheTime: 0,
        }
    );

    const getRolesQuery = useListDisplayRoles();

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
                email_address: getUserQuery.data.email_address || '',
                principal: getUserQuery.data.principal_name || '',
                first_name: getUserQuery.data.first_name || '',
                last_name: getUserQuery.data.last_name || '',
                sso_provider_id: getUserQuery.data.sso_provider_id?.toString() || undefined,
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
            onSubmit={onSubmit}
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
    onSubmit: (user: UpdateUserRequestForm) => void;
    roles?: Role[];
    showEnvironmentAccessControls?: boolean;
    SSOProviders?: SSOProvider[];
}> = ({
    error,
    hasSelectedSelf,
    initialData,
    isLoading,
    onSubmit,
    roles,
    showEnvironmentAccessControls,
    SSOProviders,
}) => {
    const form = useForm<UpdateUserRequestForm & { authenticationMethod: 'sso' | 'password' }>({
        defaultValues: {
            ...initialData,
            authenticationMethod: initialData.sso_provider_id ? 'sso' : 'password',
        },
    });

    const authenticationMethod = form.watch('authenticationMethod');
    const selectedRoleValue = form.watch('roles.0');

    const matchingRole = roles?.find((item) => selectedRoleValue === item.id)?.name;

    const selectedETACEnabledRole = matchingRole && ['Read-Only', 'User'].includes(matchingRole);
    const selectedAdminOrPowerUserRole = matchingRole && ['Administrator', 'Power User'].includes(matchingRole);

    const selectedSSOProviderHasRoleProvisionEnabled = !!SSOProviders?.find(
        ({ id }) => id === Number(form.watch('sso_provider_id'))
    )?.config?.auto_provision?.role_provision;

    const onError = () => {
        // onSubmit error
        if (error) {
            const errMsg = error.response?.data?.errors[0]?.message.toLowerCase();
            if (error.response?.status === 400) {
                if (errMsg.includes('role provision enabled')) {
                    form.setError('root.generic', {
                        type: 'custom',
                        message: 'Cannot modify user roles for role provision enabled SSO providers.',
                    });
                }
            }
            if (error.response?.status === 409) {
                if (errMsg.includes('principal name')) {
                    form.setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (errMsg.includes('email')) {
                    form.setError('email_address', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    form.setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            }
        }
    };

    const handleOnSave = () => {
        const values = form.getValues();

        // Filter out uneeded fields before form submission
        const { authenticationMethod, environment_targeted_access_control, ...filteredValues } = values;

        const formData = {
            ...filteredValues,
            sso_provider_id: values.authenticationMethod === 'password' ? undefined : values.sso_provider_id,
            all_environments: !!(selectedAdminOrPowerUserRole || (selectedETACEnabledRole && values.all_environments)),
        };

        const eTACFormData = {
            ...formData,
            environment_targeted_access_control: {
                environments:
                    values.all_environments === false ? values.environment_targeted_access_control?.environments : null,
            },
        };

        onSubmit(selectedETACEnabledRole === false ? formData : eTACFormData);
    };

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(handleOnSave, onError)}>
                <div className='flex gap-x-4 justify-center'>
                    <Card className='p-6 rounded shadow max-w-[600px] w-full'>
                        <DialogTitle>{'Edit User'}</DialogTitle>

                        <div className='flex flex-col mt-4 w-full' data-testid='update-user-dialog_dialog-content'>
                            {!hasSelectedSelf && (
                                <div className='mb-4'>
                                    <FormField
                                        name='roles.0'
                                        control={form.control}
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
                                                        <div
                                                            className='flex'
                                                            data-testid='update-user-dialog_select_role-tooltip'>
                                                            <Tooltip
                                                                tooltip='Only Read-Only and Users roles contain the environment target access control.'
                                                                contentProps={{
                                                                    className:
                                                                        'max-w-80 dark:bg-neutral-dark-5 dark:text-white border-0 !z-[2000]',
                                                                }}
                                                            />
                                                        </div>
                                                    </div>
                                                    <FormControl>
                                                        <Select
                                                            onValueChange={(field) => {
                                                                form.setValue('roles.0', Number(field));
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
                                    name='email_address'
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
                                    name='first_name'
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
                                    name='last_name'
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

                            {!hasSelectedSelf && (
                                <div className='mb-4'>
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
                                                    onValueChange={field.onChange}
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
                                <div>
                                    <FormField
                                        name='sso_provider_id'
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
                                                    defaultValue={field.value}
                                                    onValueChange={field.onChange}
                                                    value={field.value}>
                                                    <FormControl className='pointer-events-auto'>
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

                            {error && (
                                <div>
                                    <Alert severity='error'>An unexpected error occurred. Please try again.</Alert>
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
                    {showEnvironmentAccessControls && selectedETACEnabledRole && (
                        <EnvironmentSelectPanel form={form} initialData={initialData} />
                    )}
                </div>
            </form>
        </Form>
    );
};

export default UpdateUserForm;
