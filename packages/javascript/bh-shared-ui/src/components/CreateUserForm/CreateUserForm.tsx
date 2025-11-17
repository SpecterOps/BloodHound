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
    Tooltip,
} from '@bloodhoundenterprise/doodleui';
import { Alert } from '@mui/material';
import { CreateUserRequest, Role, SSOProvider } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { useListDisplayRoles } from '../../hooks/useListDisplayRoles/useListDisplayRoles';
import { apiClient } from '../../utils';
import { getDefaultRoleId, isAdminRole, isETACRole } from '../../utils/roles';
import { mapFormFieldsToUserRequest } from '../../views/Users/utils';
import EnvironmentSelectPanel from '../EnvironmentSelectPanel/EnvironmentSelectPanel';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'sso_provider_id' | 'roles'> & {
    sso_provider_id: string | undefined;
    roles: number | undefined;
};

const CreateUserForm: React.FC<{
    error: any;
    isLoading: boolean;
    onSubmit: (user: CreateUserRequest) => void;
    showEnvironmentAccessControls?: boolean;
}> = (props) => {
    const getRolesQuery = useListDisplayRoles();

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data.data)
    );

    if (!getRolesQuery.isLoading && !listSSOProvidersQuery.isLoading) {
        return <CreateUserFormInner {...props} roles={getRolesQuery.data} SSOProviders={listSSOProvidersQuery.data} />;
    }
};

const CreateUserFormInner: React.FC<{
    error: any;
    isLoading: boolean;
    onSubmit: (user: CreateUserRequest) => void;
    showEnvironmentAccessControls?: boolean;
    roles?: Role[];
    SSOProviders?: SSOProvider[];
}> = ({ error, isLoading, onSubmit, showEnvironmentAccessControls, roles, SSOProviders }) => {
    const defaultValues = {
        email_address: '',
        principal: '',
        first_name: '',
        last_name: '',
        secret: '',
        needs_password_reset: false,
        roles: getDefaultRoleId(roles),
        sso_provider_id: '',
        all_environments: false,
        environment_targeted_access_control: {
            environments: null,
        },
    } satisfies CreateUserRequestForm;

    const form = useForm<CreateUserRequestForm>({ defaultValues });

    const [authenticationMethod, setAuthenticationMethod] = useState<string>('password');

    const selectedRoleId = form.watch('roles');

    const selectedETACEnabledRole = isETACRole(selectedRoleId, roles);
    const selectedAdminOrPowerUserRole = isAdminRole(selectedRoleId, roles);

    useEffect(() => {
        if (error) {
            const message = error.response?.data?.errors[0]?.message?.toLowerCase() ?? '';
            if (error?.response?.status === 409) {
                if (message.includes('principal name')) {
                    form.setError('principal', {
                        type: 'custom',
                        message: 'Principal name is already in use.',
                    });
                } else if (message.includes('email')) {
                    form.setError('email_address', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    form.setError('root.generic', { type: 'custom', message: `A conflict has occurred.` });
                }
            } else {
                form.setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    }, [error, form]);

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
            <form autoComplete='off' data-testid='create-user-dialog_form' onSubmit={form.handleSubmit(handleOnSave)}>
                <div className='flex gap-x-4 justify-center'>
                    <Card className='p-6 rounded shadow max-w-[600px] w-full'>
                        <DialogTitle>{'Create User'}</DialogTitle>

                        <div className='flex flex-col mt-4 w-full' data-testid='create-user-dialog_content'>
                            <div className='mb-4'>
                                <FormField
                                    name='roles'
                                    control={form.control}
                                    rules={{
                                        required: 'Role is required',
                                    }}
                                    render={({ field }) => (
                                        <FormItem>
                                            <div className='flex row'>
                                                <FormLabel className='mr-2 font-medium !text-sm' htmlFor='role'>
                                                    Role
                                                </FormLabel>

                                                <Tooltip
                                                    defaultOpen={false}
                                                    tooltip='Only Read-Only and Users roles contain the environment target access control.'
                                                    triggerProps={{ type: 'button' }}
                                                    contentProps={{
                                                        className:
                                                            'max-w-80 dark:bg-neutral-dark-5 dark:text-white border-0 !z-[2000]',
                                                    }}
                                                />
                                            </div>
                                            <Select
                                                onValueChange={(field) => {
                                                    form.setValue('roles', Number(field));
                                                }}
                                                value={String(selectedRoleId)}>
                                                <FormControl>
                                                    <SelectTrigger
                                                        className='bg-transparent'
                                                        data-testid='create-user-dialog_select_role'
                                                        id='role'
                                                        variant='underlined'>
                                                        <SelectValue placeholder={field.value} />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        {roles?.map((role: Role) => (
                                                            <SelectItem
                                                                data-testid={`create-user-dialog_select_role-${role.name}`}
                                                                className='hover:cursor-pointer'
                                                                key={role.id}
                                                                value={role.id.toString()}>
                                                                {role.name}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </SelectPortal>
                                            </Select>
                                        </FormItem>
                                    )}
                                />
                            </div>
                            <div className='mb-4'>
                                <FormField
                                    name='email_address'
                                    control={form.control}
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
                                                <Input
                                                    {...field}
                                                    type='email'
                                                    id='emailAddress'
                                                    placeholder='user@domain.com'
                                                />
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
                                                Principal Name{' '}
                                            </FormLabel>
                                            <FormControl>
                                                <Input {...field} id='principal' placeholder='Principal Name' />
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
                                        <>
                                            <FormItem>
                                                <FormLabel className='font-medium !text-sm' htmlFor='firstName'>
                                                    First Name
                                                </FormLabel>
                                                <FormControl>
                                                    <Input {...field} id='firstName' placeholder='First Name' />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        </>
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
                                        <>
                                            <FormItem>
                                                <FormLabel className='font-medium !text-sm' htmlFor='lastName'>
                                                    Last Name
                                                </FormLabel>
                                                <FormControl>
                                                    <Input {...field} id='lastName' placeholder='Last Name' />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        </>
                                    )}
                                />
                            </div>

                            <>
                                <div className='mb-4'>
                                    <FormItem>
                                        <FormLabel className='font-medium !text-sm' htmlFor='authenticationMethod'>
                                            Authentication Method
                                        </FormLabel>
                                        <Select onValueChange={setAuthenticationMethod} value={authenticationMethod}>
                                            <FormControl className='mt-2'>
                                                <SelectTrigger
                                                    className='bg-transparent'
                                                    data-testid='create-user-dialog_select_authentication-method'
                                                    id='authenticationMethod'
                                                    variant='underlined'>
                                                    <SelectValue placeholder={authenticationMethod} />
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

                                {authenticationMethod === 'password' ? (
                                    <>
                                        <div className='mb-4'>
                                            <FormField
                                                name='secret'
                                                control={form.control}
                                                defaultValue=''
                                                rules={{
                                                    required: 'Password is required',
                                                    minLength: {
                                                        value: 12,
                                                        message: 'Password must be at least 12 characters long',
                                                    },
                                                    pattern: {
                                                        value: /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%^&*])/,
                                                        message:
                                                            'Password must contain at least 1 lowercase character, 1 uppercase character, 1 number and 1 special character (!@#$%^&*)',
                                                    },
                                                    maxLength: {
                                                        value: 1000,
                                                        message: 'Password must be less than 1000 characters',
                                                    },
                                                }}
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <FormLabel className='font-medium !text-sm' htmlFor='secret'>
                                                            Initial Password
                                                        </FormLabel>
                                                        <FormControl>
                                                            <Input
                                                                {...field}
                                                                id='secret'
                                                                type='password'
                                                                placeholder='Initial Password'
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                )}
                                            />
                                        </div>
                                        <div className=''>
                                            <FormField
                                                name='needs_password_reset'
                                                control={form.control}
                                                defaultValue={false}
                                                render={({ field }) => (
                                                    <div className='flex flex-row items-center'>
                                                        <FormItem className='flex flex-row my-3'>
                                                            <FormControl>
                                                                <Checkbox
                                                                    id='needsPasswordReset'
                                                                    checked={field.value}
                                                                    onCheckedChange={field.onChange}
                                                                />
                                                            </FormControl>
                                                            <FormLabel
                                                                htmlFor='needsPasswordReset'
                                                                className='pl-2 font-medium !text-sm'>
                                                                Force Password Reset?
                                                            </FormLabel>
                                                        </FormItem>
                                                    </div>
                                                )}
                                            />
                                        </div>
                                    </>
                                ) : (
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
                                                        onValueChange={field.onChange}
                                                        defaultValue={field.value}
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
                            </>
                            {form.formState.errors?.root?.generic && (
                                <div>
                                    <Alert severity='error'>{form.formState.errors.root.generic.message}</Alert>
                                </div>
                            )}
                        </div>
                        <DialogActions className='mt-8 flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button
                                    type='button'
                                    disabled={isLoading}
                                    variant='tertiary'
                                    data-testid='create-user-dialog_button-cancel'>
                                    Cancel
                                </Button>
                            </DialogClose>
                            <Button
                                data-testid='create-user-dialog_button-save'
                                disabled={isLoading}
                                role='button'
                                type='submit'>
                                Save
                            </Button>
                        </DialogActions>
                    </Card>
                    {showEnvironmentAccessControls && selectedETACEnabledRole && <EnvironmentSelectPanel form={form} />}
                </div>
            </form>
        </Form>
    );
};

export default CreateUserForm;
