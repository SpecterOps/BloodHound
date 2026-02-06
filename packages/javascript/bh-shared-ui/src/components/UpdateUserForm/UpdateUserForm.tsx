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
    Tooltip,
} from '@bloodhoundenterprise/doodleui';
import { Alert, CircularProgress } from '@mui/material';
import { Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { useAvailableEnvironments } from '../../hooks';
import { useGetUser } from '../../hooks/useBloodHoundUsers';
import { useListDisplayRoles } from '../../hooks/useListDisplayRoles/useListDisplayRoles';
import { useSSOProviders } from '../../hooks/useSSOProviders';
import { isAdminRole, isETACRole } from '../../utils/roles';
import { mapFormFieldsToUserRequest, mapUserResponseToForm } from '../../views/Users/utils';
import EnvironmentSelectPanel from '../EnvironmentSelectPanel';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'sso_provider_id' | 'roles'> & {
    sso_provider_id: string | undefined;
    roles: number | undefined;
};

const UpdateUserForm: React.FC<{
    onSubmit: (user: UpdateUserRequest) => void;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
    showEnvironmentAccessControls?: boolean;
}> = ({ onSubmit, userId, hasSelectedSelf, isLoading, error, showEnvironmentAccessControls }) => {
    const getUserQuery = useGetUser(userId);
    const getRolesQuery = useListDisplayRoles();
    const getSSOProvidersQuery = useSSOProviders();
    const getEnvironmentsQuery = useAvailableEnvironments();

    if (
        getUserQuery.isLoading ||
        getRolesQuery.isLoading ||
        getSSOProvidersQuery.isLoading ||
        getEnvironmentsQuery.isLoading
    ) {
        return (
            <div className='w-full h-full text-center'>
                <CircularProgress />
            </div>
        );
    }

    if (getUserQuery.isError || getRolesQuery.isError || getSSOProvidersQuery.isError || getEnvironmentsQuery.isError) {
        return (
            <Card className='p-6 rounded shadow w-[600px] m-auto h-[800px] flex flex-col justify-center'>
                <div>Unable to load data required to edit this user.</div>

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
            initialData={mapUserResponseToForm(getUserQuery.data)}
            error={error}
            hasSelectedSelf={hasSelectedSelf}
            isLoading={isLoading}
            onSubmit={onSubmit}
            roles={getRolesQuery.data}
            SSOProviders={getSSOProvidersQuery.data}
            showEnvironmentAccessControls={showEnvironmentAccessControls}
        />
    );
};

const UpdateUserFormInner: React.FC<{
    error: any;
    hasSelectedSelf: boolean;
    initialData: UpdateUserRequestForm;
    isLoading: boolean;
    onSubmit: (user: UpdateUserRequest) => void;
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
    const form = useForm<UpdateUserRequestForm>({
        defaultValues: initialData,
    });

    const [authenticationMethod, setAuthenticationMethod] = useState<string>(
        initialData.sso_provider_id ? 'sso' : 'password'
    );

    const selectedRoleId = form.watch('roles');

    const selectedETACEnabledRole = isETACRole(selectedRoleId, roles);
    const selectedAdminOrPowerUserRole = isAdminRole(selectedRoleId, roles);

    const selectedSSOProviderHasRoleProvisionEnabled = !!SSOProviders?.find(
        ({ id }) => id === Number(form.watch('sso_provider_id'))
    )?.config?.auto_provision?.role_provision;

    useEffect(() => {
        if (error) {
            const message = error.response?.data?.errors[0]?.message?.toLowerCase() ?? '';
            if (error.response?.status === 400 && message.includes('role provision enabled')) {
                form.setError('root.generic', {
                    type: 'custom',
                    message: 'Cannot modify user roles for role provision enabled SSO providers.',
                });
            } else if (error.response?.status === 409) {
                if (message.includes('principal name')) {
                    form.setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (message.includes('email')) {
                    form.setError('email_address', { type: 'custom', message: 'Email is already in use.' });
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
        // ignoring so we don't have to include the form element
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [error]);

    const handleOnSave = () => {
        const user = mapFormFieldsToUserRequest(
            form.getValues(),
            authenticationMethod,
            selectedAdminOrPowerUserRole,
            selectedETACEnabledRole
        );

        onSubmit(user);
    };

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(handleOnSave)}>
                <div className='flex gap-x-4 justify-center'>
                    <Card className='p-6 flex flex-col rounded shadow max-w-[600px] w-full'>
                        <DialogTitle>{'Edit User'}</DialogTitle>

                        <div className='flex flex-col mt-4 mb-8 w-full' data-testid='update-user-dialog_dialog-content'>
                            {!hasSelectedSelf && (
                                <div className='mb-4'>
                                    <FormField
                                        name='roles'
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
                                                                triggerProps={{ type: 'button' }}
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
                                                                form.setValue('roles', Number(field));
                                                            }}
                                                            value={String(selectedRoleId)}>
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
                                    <FormItem>
                                        <FormLabel className='font-medium !text-sm' htmlFor='authenticationMethod'>
                                            Authentication Method
                                        </FormLabel>

                                        <Select onValueChange={setAuthenticationMethod} value={authenticationMethod}>
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
                                                    <SelectItem value='password'>Username / Password</SelectItem>
                                                    {SSOProviders && SSOProviders.length > 0 && (
                                                        <SelectItem value='sso'>Single Sign-On (SSO)</SelectItem>
                                                    )}
                                                </SelectContent>
                                            </SelectPortal>
                                        </Select>
                                    </FormItem>
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
                                                    value={field.value || ''}>
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

                            {form.formState.errors?.root?.generic && (
                                <div>
                                    <Alert severity='error'>{form.formState.errors.root.generic.message}</Alert>
                                </div>
                            )}
                        </div>
                        <DialogActions className='mt-auto flex justify-end gap-4'>
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
