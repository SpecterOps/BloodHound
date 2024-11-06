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
import React, { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import { SSOProvider, UpdateUserRequest } from 'js-client-library';

export type UpdateUserRequestForm = Omit<UpdateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const UpdateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdateUserRequestForm) => void;
    userId: string;
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, userId, isLoading, error }) => {
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
            isLoading={isLoading}
            error={error}
        />
    );
};

const UpdateUserFormInner: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdateUserRequestForm) => void;
    initialData: UpdateUserRequestForm;
    roles: any[];
    SSOProviders?: SSOProvider[];
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, initialData, roles, SSOProviders, isLoading, error }) => {
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
        watch,
    } = useForm<UpdateUserRequestForm & { authenticationMethod: 'sso' | 'password' }>({
        defaultValues: {
            ...initialData,
            authenticationMethod: initialData.SSOProviderId ? 'sso' : 'password',
        },
    });

    const authenticationMethod = watch('authenticationMethod');

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SSOProviderId', undefined);
        }
    }, [authenticationMethod, setValue]);

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            <DialogContent>
                <Grid container spacing={1}>
                    <Grid item xs={12}>
                        <Controller
                            name='emailAddress'
                            control={control}
                            rules={{ required: 'Email Address is required' }}
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
                            rules={{ required: 'Principal Name is required' }}
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
                            rules={{ required: 'First Name is required' }}
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
                            rules={{ required: 'Last Name is required' }}
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
                                        <InputLabel id='authenticationMethod-label' sx={{ ml: '-14px', mt: '8px' }}>
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
                                            data-testid='update-user-dialog_select-authentication-method'>
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
                                            <InputLabel id='SSOProviderId-label' sx={{ ml: '-14px', mt: '8px' }}>
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
                                                data-testid='update-user-dialog_select-sso-provider'>
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
                                    <InputLabel id='role-label' sx={{ ml: '-14px', mt: '8px' }}>
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
                                        data-testid='update-user-dialog_select-role'>
                                        {roles.map((role: any) => (
                                            <MenuItem key={role.id} value={role.id.toString()}>
                                                {role.name}
                                            </MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>
                            )}
                        />
                    </Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                {error && (
                    <FormHelperText error style={{ margin: 0 }}>
                        An unexpected error occurred. Please try again.
                    </FormHelperText>
                )}
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
