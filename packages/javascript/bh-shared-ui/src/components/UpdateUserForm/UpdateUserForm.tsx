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
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogTitle,
    Form,
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
import { Card, DialogContentText, FormControl, Grid, Skeleton } from '@mui/material';
import { Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { apiClient } from '../../utils';
import UserFormEnvironmentSelector from '../CreateUserForm/UserFormEnvironmentSelector';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const UpdateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdateUserRequestForm) => void;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
    open?: boolean;
    showEnvironmentAccessControls?: boolean;
}> = ({ onCancel, onSubmit, userId, hasSelectedSelf, isLoading, error, showEnvironmentAccessControls }) => {
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
            <>
                <DialogContent>
                    <DialogContentText>
                        <Skeleton />
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <DialogActions>
                        <Button
                            type='button'
                            variant='tertiary'
                            onClick={onCancel}
                            data-testid='update-user-dialog_button-close'>
                            Close
                        </Button>
                    </DialogActions>
                </DialogActions>
            </>
        );
    }

    if (getUserQuery.isError || getRolesQuery.isError || listSSOProvidersQuery.isError) {
        return (
            <>
                <DialogContent>
                    <DialogContentText>User not found.</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <DialogActions>
                        <Button
                            type='button'
                            variant='tertiary'
                            onClick={onCancel}
                            data-testid='update-user-dialog_button-close'>
                            Close
                        </Button>
                    </DialogActions>
                </DialogActions>
            </>
        );
    }

    return (
        <UpdateUserFormInner
            initialData={{
                emailAddress: getUserQuery.data.email_address || '',
                principal: getUserQuery.data.principal_name || '',
                firstName: getUserQuery.data.first_name || '',
                lastName: getUserQuery.data.last_name || '',
                SSOProviderId: getUserQuery.data.sso_provider_id?.toString() || '',
                roles: getUserQuery.data.roles?.map((role: any) => role.id) || [],
            }}
            error={error}
            hasSelectedSelf={hasSelectedSelf}
            isLoading={isLoading}
            onCancel={onCancel}
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
    onCancel: () => void;
    open?: boolean;
    onSubmit: (user: UpdateUserRequestForm) => void;
    roles?: Role[];
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
    SSOProviders?: SSOProvider[];
}> = ({
    error,
    hasSelectedSelf,
    initialData,
    isLoading,
    //onCancel,
    onSubmit,
    //open,
    roles,
    showEnvironmentAccessControls,
    SSOProviders,
}) => {
    /*
    const {
        control,
        formState: { errors },
        handleSubmit,
        register,
        setError,
        setValue,
        watch,
    } = useForm<UpdateUserRequestForm & { authenticationMethod: 'sso' | 'password' }>({
        defaultValues: {
            ...initialData,
            authenticationMethod: initialData.SSOProviderId ? 'sso' : 'password',
        },
    });
    */

    const defaultValues = {
        ...initialData,
        //authenticationMethod: initialData.SSOProviderId ? 'sso' : 'password',
    };

    const form = useForm<UpdateUserRequestForm & { authenticationMethod: 'sso' | 'password' }>({ defaultValues });

    const [authenticationMethod, setAuthenticationMethod] = useState(initialData.SSOProviderId ? 'sso' : 'password');

    console.log(authenticationMethod);

    const [selectedRoleValue, setSelectedRoleValue] = useState<number[]>(initialData.roles);
    //console.log(selectedRoleValue);

    const rolesWithEnvironmentPermissions =
        selectedRoleValue.toString() === '2' || selectedRoleValue.toString() === '3';

    /*
    const selectedSSOProviderHasRoleProvisionEnabled = !!SSOProviders?.find(
    ({ id }) => id === Number(('SSOProviderId'))
    */
    /*
    const roleInputValue = watch('roles');

    const authenticationMethod = watch('authenticationMethod');


    )?.config?.auto_provision?.role_provision;

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SSOProviderId', undefined);
        }

        if (error) {
            const errMsg = error.response?.data?.errors[0]?.message.toLowerCase();
            if (error.response?.status === 400) {
                if (errMsg.includes('role provision enabled')) {
                    setError('root.generic', {
                        type: 'custom',
                        message: 'Cannot modify user roles for role provision enabled SSO providers.',
                    });
                }
            } else if (error.response?.status === 409) {
                if (errMsg.includes('principal name')) {
                    setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (errMsg.includes('email')) {
                    setError('emailAddress', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            } else {
                setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    }, [authenticationMethod, setValue, error, setError]);
    */

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(onSubmit)}>
                <div className='flex gap-x-4 justify-center'>
                    <Card className=' p-6 rounded shadow max-w-[600px]'>
                        <DialogTitle>{'Edit User'}</DialogTitle>

                        <DialogDescription className='flex flex-col' data-testid='update-user-dialog_dialog-content'>
                            <Grid container spacing={2} className='min-h-[650px] mt-4'>
                                <Grid item xs={12}>
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
                                                <FormLabel
                                                    aria-labelledby='emailAddress'
                                                    data-testid='update-user-dialog_label-email-address'>
                                                    Email Address
                                                </FormLabel>
                                                <FormControl hiddenLabel>
                                                    <Input
                                                        {...field}
                                                        data-testid='update-user-dialog_input-email-address'
                                                        id='emailAddress'
                                                        type='email'
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </Grid>

                                <Grid item xs={12}>
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
                                                <FormLabel
                                                    aria-labelledby='principal'
                                                    data-testid='update-user-dialog_label-email-address'>
                                                    Principal Name
                                                </FormLabel>
                                                <FormControl hiddenLabel>
                                                    <Input
                                                        {...field}
                                                        data-testid='update-user-dialog_input-principal-name'
                                                        id='principal'
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </Grid>

                                <Grid item xs={12}>
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
                                                <FormLabel data-testid='update-user-dialog_label-first-name'>
                                                    First Name
                                                </FormLabel>
                                                <FormControl>
                                                    <Input
                                                        {...field}
                                                        id='firstName'
                                                        data-testid='update-user-dialog_input-first-name'
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </Grid>

                                <Grid item xs={12}>
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
                                                <FormLabel data-testid='update-user-dialog_label-last-name'>
                                                    Last Name
                                                </FormLabel>
                                                <FormControl>
                                                    <Input
                                                        {...field}
                                                        data-testid='update-user-dialog_input-last-name'
                                                        id='lastName'
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </Grid>

                                <>
                                    <Grid item xs={12}>
                                        <FormField
                                            name='authenticationMethod'
                                            control={form.control}
                                            rules={{
                                                required: 'Authentication Method is required',
                                            }}
                                            render={({ field }) => (
                                                <>
                                                    <FormLabel
                                                        data-testid='update-user-dialog_label-authentication-method'
                                                        //hidden={hasSelectedSelf}
                                                        className='mb-4'>
                                                        Authentication Method
                                                    </FormLabel>

                                                    <Select
                                                        data-testid='update-user-dialog_select-authentication-method'
                                                        //defaultValue={field.value}
                                                        onValueChange={(field) => {
                                                            form.setValue('authenticationMethod', field);
                                                            setAuthenticationMethod(field);
                                                        }}
                                                        value={field.value}
                                                        //hidden={hasSelectedSelf}
                                                    >
                                                        <FormControl>
                                                            <SelectTrigger className='mt-3'>
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
                                                                <SelectItem value='sso'>
                                                                    Single Sign-On (SSO)
                                                                </SelectItem>
                                                                {SSOProviders && SSOProviders.length > 0 && (
                                                                    <SelectItem value='sso'>
                                                                        Single Sign-On (SSO)
                                                                    </SelectItem>
                                                                )}
                                                            </SelectContent>
                                                        </SelectPortal>
                                                    </Select>
                                                </>
                                            )}
                                        />
                                    </Grid>

                                    {authenticationMethod === 'sso' && (
                                        <Grid item xs={12}>
                                            <FormField
                                                name='SSOProviderId'
                                                control={form.control}
                                                rules={{
                                                    required: 'SSO Provider is required',
                                                }}
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <FormLabel
                                                            data-testid='update-user-dialog_label-sso-provider'
                                                            id='SSOProviderId-label'
                                                            //hidden={hasSelectedSelf}
                                                        >
                                                            SSO Provider
                                                        </FormLabel>

                                                        <Select
                                                            data-testid='update-user-dialog_sso-provider'
                                                            onValueChange={(field) => {
                                                                form.setValue('authenticationMethod', field.value);
                                                                setAuthenticationMethod(field.value);
                                                            }}
                                                            value={field.value}
                                                            //hidden={hasSelectedSelf}
                                                        >
                                                            <FormControl>
                                                                <SelectTrigger className='mt-3'>
                                                                    <SelectValue placeholder='SSO Provider' />
                                                                </SelectTrigger>
                                                            </FormControl>
                                                            <SelectPortal>
                                                                <SelectContent>
                                                                    {SSOProviders?.map((SSOProvider: SSOProvider) => (
                                                                        <SelectItem
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
                                        </Grid>
                                    )}
                                </>

                                <Grid item xs={12}>
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
                                                        <FormLabel
                                                            data-testid='update-user-dialog_label-role'
                                                            className='mr-2'>
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
                                                            data-testid='update-user-dialog_input-role'
                                                            onValueChange={(field) => {
                                                                form.setValue('roles', [Number(field)]);
                                                                setSelectedRoleValue([Number(field)]);
                                                            }}
                                                            value={String(selectedRoleValue)}>
                                                            <FormControl>
                                                                <SelectTrigger className='mt-3'>
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
                                                </FormItem>
                                            </>
                                        )}
                                    />
                                </Grid>
                            </Grid>
                        </DialogDescription>
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
                                //onSubmit={form.handleSubmit(onSubmit)}
                                role='button'
                                type='submit'>
                                Save
                            </Button>
                        </DialogActions>
                    </Card>
                    {showEnvironmentAccessControls && rolesWithEnvironmentPermissions && (
                        <UserFormEnvironmentSelector />
                    )}
                </div>
            </form>
        </Form>
    );
};

export default UpdateUserForm;
