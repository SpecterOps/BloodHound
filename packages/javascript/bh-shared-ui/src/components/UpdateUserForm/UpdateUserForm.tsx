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
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Tooltip,
} from '@bloodhoundenterprise/doodleui';
import { Alert, Card, DialogContentText, FormControl, Grid, Skeleton, TextField } from '@mui/material';
import { Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
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
    //open?: boolean;
    //showEnvironmentAccessControls?: boolean;
}> = ({ onCancel, onSubmit, userId, hasSelectedSelf, isLoading, error }) => {
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
    onCancel,
    onSubmit,
    open,
    roles,
    showEnvironmentAccessControls = true,
    SSOProviders,
}) => {
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

    const authenticationMethod = watch('authenticationMethod');

    const selectedSSOProviderHasRoleProvisionEnabled = !!SSOProviders?.find(
        ({ id }) => id === Number(watch('SSOProviderId'))
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

    const [selectedRoleValue, setSelectedRoleValue] = useState<number[]>(initialData.roles);

    const roleInputValue = watch('roles');
    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            <div className='flex gap-x-4 justify-center'>
                <Card className=' p-6 rounded shadow max-w-[600px]'>
                    <DialogTitle>{'Edit User'}</DialogTitle>

                    <DialogDescription className='flex flex-col' data-testid='update-user-dialog_dialog-content'>
                        <Grid container spacing={2} className='min-h-[650px] mt-4'>
                            <Grid item xs={12}>
                                <Controller
                                    name='emailAddress'
                                    control={control}
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
                                        <TextField
                                            {...field}
                                            variant='standard'
                                            id='emailAddress'
                                            label='Email Address'
                                            type='email'
                                            fullWidth
                                            error={!!errors.emailAddress}
                                            helperText={errors.emailAddress?.message}
                                            data-testid='update-user-dialog_input-email-address'
                                        />
                                    )}
                                />
                            </Grid>

                            <Grid item xs={12}>
                                <Controller
                                    name='principal'
                                    control={control}
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
                                        <TextField
                                            {...field}
                                            variant='standard'
                                            id='principal'
                                            label='Principal Name'
                                            fullWidth
                                            error={!!errors.principal}
                                            helperText={errors.principal?.message}
                                            data-testid='update-user-dialog_input-principal-name'
                                        />
                                    )}
                                />
                            </Grid>
                            <Grid item xs={12}>
                                <Controller
                                    name='firstName'
                                    control={control}
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
                                        <TextField
                                            {...field}
                                            variant='standard'
                                            id='firstName'
                                            label='First Name'
                                            fullWidth
                                            error={!!errors.firstName}
                                            helperText={errors.firstName?.message}
                                            data-testid='update-user-dialog_input-first-name'
                                        />
                                    )}
                                />
                            </Grid>
                            <Grid item xs={12}>
                                <Controller
                                    name='lastName'
                                    control={control}
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
                                        <TextField
                                            {...field}
                                            variant='standard'
                                            id='lastName'
                                            label='Last Name'
                                            fullWidth
                                            error={!!errors.lastName}
                                            helperText={errors.lastName?.message}
                                            data-testid='update-user-dialog_input-last-name'
                                        />
                                    )}
                                />
                            </Grid>

                            <>
                                <Grid item xs={12}>
                                    <Controller
                                        name='authenticationMethod'
                                        control={control}
                                        rules={{
                                            required: 'Authentication Method is required',
                                        }}
                                        render={({ field: { onChange, onBlur, value, ref } }) => (
                                            <FormControl>
                                                <Label
                                                    id='authenticationMethod-label'
                                                    //sx={{ ml: '-14px', mt: '8px' }}
                                                    hidden={hasSelectedSelf}>
                                                    Authentication Method
                                                </Label>
                                                <Select
                                                    data-testid='update-user-dialog_select-authentication-method'
                                                    onValueChange={
                                                        onChange as (event: SelectChangeEvent<string>) => void
                                                    }
                                                    value={value}

                                                    /*
                                                    onChange={onChange as (event: SelectChangeEvent<string>) => void}
                                                    onBlur={onBlur}
                                                    data-testid='update-user-dialog_select-authentication-method'
                                                    fullWidth
                                                    hidden={hasSelectedSelf}
                                                    id='authenticationMethod'
                                                    labelId='authenticationMethod-label'
                                                    name='authenticationMethod'
                                                    ref={ref}
                                                    value={value}
                                                    variant='standard'
                                                */
                                                >
                                                    <SelectTrigger>
                                                        <SelectValue placeholder={value} />
                                                    </SelectTrigger>
                                                    <SelectPortal>
                                                        <SelectContent>
                                                            <SelectItem value='password'>
                                                                Username / Password
                                                            </SelectItem>
                                                            <SelectItem value='sso'>Single Sign-On (SSO)</SelectItem>
                                                            {SSOProviders && SSOProviders.length > 0 && (
                                                                <SelectItem value='sso'>
                                                                    Single Sign-On (SSO)
                                                                </SelectItem>
                                                            )}
                                                        </SelectContent>
                                                    </SelectPortal>
                                                </Select>
                                            </FormControl>
                                        )}
                                    />
                                </Grid>

                                {authenticationMethod === 'sso' && (
                                    <Grid item xs={12}>
                                        <Controller
                                            name='SSOProviderId'
                                            control={control}
                                            rules={{
                                                required: 'SSO Provider is required',
                                            }}
                                            render={({ field: { onChange, onBlur, value, ref } }) => (
                                                <FormControl>
                                                    <Label
                                                        id='SSOProviderId-label'
                                                        //sx={{ ml: '-14px', mt: '8px' }}
                                                        hidden={hasSelectedSelf}>
                                                        SSO Provider
                                                    </Label>
                                                    <Select
                                                    /*
                                                        onChange={
                                                            onChange as (event: SelectChangeEvent<string>) => void
                                                        }
                                                        onBlur={onBlur}
                                                        value={value?.toString()}
                                                        ref={ref}
                                                        defaultValue={''}
                                                        labelId='SSOProviderId-label'
                                                        id='SSOProviderId'
                                                        name='SSOProviderId'
                                                        variant='standard'
                                                        fullWidth
                                                        data-testid='update-user-dialog_select-sso-provider'
                                                    hidden={hasSelectedSelf}
                                                        */
                                                    >
                                                        <SelectTrigger>
                                                            <SelectValue placeholder='SSO Provider' />
                                                        </SelectTrigger>
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
                                                </FormControl>
                                            )}
                                        />
                                    </Grid>
                                )}
                            </>

                            <Grid item xs={12}>
                                <Controller
                                    name='roles.0'
                                    control={control}
                                    defaultValue={1}
                                    rules={{
                                        required: 'Role is required',
                                    }}
                                    render={({ field }) => (
                                        <>
                                            <div className='flex row'>
                                                <Label className='mr-2' size='small'>
                                                    Role
                                                </Label>
                                                <Tooltip
                                                    tooltip='Only User, Read-Only, Upload-Only roles contain the limited access functionality.'
                                                    contentProps={{
                                                        className: 'max-w-80 dark:bg-neutral-dark-5 border-0',
                                                    }}
                                                />
                                            </div>
                                            <Select
                                                {...register('roles')}
                                                disabled={selectedSSOProviderHasRoleProvisionEnabled}
                                                data-testid='create-user-dialog_role'
                                                onValueChange={(field) => {
                                                    setValue('roles', [Number(field)]);
                                                    setSelectedRoleValue([Number([field])]);
                                                }}
                                                value={String(selectedRoleValue)}>
                                                <SelectTrigger>
                                                    <SelectValue placeholder={field.value} />
                                                </SelectTrigger>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        {roles?.map((role: Role) => (
                                                            <SelectItem
                                                                className='hover:cursor-pointer'
                                                                key={role.id}
                                                                value={role.id.toString()}>
                                                                {role.name}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </SelectPortal>
                                            </Select>
                                        </>
                                    )}
                                />
                            </Grid>
                            {!!errors.root?.generic && (
                                <Grid item xs={12}>
                                    <Alert severity='error'>{errors.root.generic.message}</Alert>
                                </Grid>
                            )}
                        </Grid>
                    </DialogDescription>
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
                        <Button type='submit' disabled={isLoading} data-testid='create-user-dialog_button-save'>
                            Save
                        </Button>
                    </DialogActions>
                </Card>
                {showEnvironmentAccessControls && selectedRole && <UserFormEnvironmentSelector />}
            </div>
        </form>
    );
};

export default UpdateUserForm;
