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

import { Button } from '@bloodhoundenterprise/doodleui';
import {
    Alert,
    DialogActions,
    DialogContent,
    DialogContentText,
    FormControl,
    FormHelperText,
    Grid,
    InputLabel,
    MenuItem,
    Select,
    SelectChangeEvent,
    Skeleton,
    TextField,
} from '@mui/material';
import { Role, SSOProvider, UpdateUserRequest } from 'js-client-library';
import React, { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { apiClient } from '../../utils';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const UpdateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdateUserRequestForm) => void;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
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
            onCancel={onCancel}
            onSubmit={onSubmit}
            initialData={{
                emailAddress: getUserQuery.data.email_address || '',
                principal: getUserQuery.data.principal_name || '',
                firstName: getUserQuery.data.first_name || '',
                lastName: getUserQuery.data.last_name || '',
                SSOProviderId: getUserQuery.data.sso_provider_id?.toString() || '',
                roles: getUserQuery.data.roles?.map((role: any) => role.id) || [],
            }}
            roles={getRolesQuery.data}
            SSOProviders={listSSOProvidersQuery.data}
            hasSelectedSelf={hasSelectedSelf}
            isLoading={isLoading}
            error={error}
        />
    );
};

const UpdateUserFormInner: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdateUserRequestForm) => void;
    initialData: UpdateUserRequestForm;
    roles?: Role[];
    SSOProviders?: SSOProvider[];
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, initialData, roles, SSOProviders, hasSelectedSelf, isLoading, error }) => {
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
        setError,
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

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            <DialogContent>
                <Grid container spacing={1}>
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
                                        <InputLabel
                                            id='authenticationMethod-label'
                                            sx={{ ml: '-14px', mt: '8px' }}
                                            hidden={hasSelectedSelf}>
                                            Authentication Method
                                        </InputLabel>
                                        <Select
                                            onChange={onChange as (event: SelectChangeEvent<string>) => void}
                                            onBlur={onBlur}
                                            value={value}
                                            ref={ref}
                                            labelId='authenticationMethod-label'
                                            id='authenticationMethod'
                                            name='authenticationMethod'
                                            variant='standard'
                                            fullWidth
                                            data-testid='update-user-dialog_select-authentication-method'
                                            hidden={hasSelectedSelf}>
                                            <MenuItem value='password'>Username / Password</MenuItem>
                                            {SSOProviders && SSOProviders.length > 0 && (
                                                <MenuItem value='sso'>Single Sign-On (SSO)</MenuItem>
                                            )}
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
                                            <InputLabel
                                                id='SSOProviderId-label'
                                                sx={{ ml: '-14px', mt: '8px' }}
                                                hidden={hasSelectedSelf}>
                                                SSO Provider
                                            </InputLabel>
                                            <Select
                                                onChange={onChange as (event: SelectChangeEvent<string>) => void}
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
                                                hidden={hasSelectedSelf}>
                                                {SSOProviders?.map((SSOProvider: SSOProvider) => (
                                                    <MenuItem value={SSOProvider.id.toString()} key={SSOProvider.id}>
                                                        {SSOProvider.name}
                                                    </MenuItem>
                                                ))}
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
                                <FormControl>
                                    <InputLabel
                                        id='role-label'
                                        sx={{ ml: '-14px', mt: '8px' }}
                                        hidden={hasSelectedSelf}>
                                        Role
                                    </InputLabel>
                                    <Select
                                        labelId='role-label'
                                        id='role'
                                        name='role'
                                        onChange={(e) => {
                                            const output = parseInt(e.target.value as string, 10);
                                            field.onChange(isNaN(output) ? 1 : output);
                                        }}
                                        value={isNaN(field.value) ? '' : field.value.toString()}
                                        variant='standard'
                                        fullWidth
                                        disabled={selectedSSOProviderHasRoleProvisionEnabled}
                                        data-testid='update-user-dialog_select-role'
                                        hidden={hasSelectedSelf}>
                                        {roles?.map((role: Role) => (
                                            <MenuItem key={role.id} value={role.id.toString()}>
                                                {role.name}
                                            </MenuItem>
                                        ))}
                                    </Select>
                                    {selectedSSOProviderHasRoleProvisionEnabled && (
                                        <FormHelperText id='role-helper-text'>
                                            SSO Provider has enabled role provision.
                                        </FormHelperText>
                                    )}
                                </FormControl>
                            )}
                        />
                    </Grid>
                    {!!errors.root?.generic && (
                        <Grid item xs={12}>
                            <Alert severity='error'>{errors.root.generic.message}</Alert>
                        </Grid>
                    )}
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button
                    type='button'
                    variant={'tertiary'}
                    onClick={onCancel}
                    disabled={isLoading}
                    data-testid='update-user-dialog_button-close'>
                    Cancel
                </Button>
                <Button type='submit' disabled={isLoading} data-testid='update-user-dialog_button-save'>
                    Save
                </Button>
            </DialogActions>
        </form>
    );
};

export default UpdateUserForm;
