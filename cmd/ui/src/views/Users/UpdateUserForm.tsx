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
import { apiClient } from 'bh-shared-ui';
import { UpdatedUser } from 'src/ducks/auth/types';

const UpdateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdatedUser) => void;
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

    const listSAMLProvidersQuery = useQuery(['listSAMLProviders'], ({ signal }) =>
        apiClient.listSAMLProviders({ signal }).then((res) => res.data.data.saml_providers)
    );

    if (getUserQuery.isLoading || getRolesQuery.isLoading || listSAMLProvidersQuery.isLoading) {
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
                            autoFocus
                            color='inherit'
                            onClick={onCancel}
                            data-testid='update-user-dialog_button-close'>
                            Close
                        </Button>
                    </DialogActions>
                </DialogActions>
            </>
        );
    }

    if (getUserQuery.isError || getRolesQuery.isError || listSAMLProvidersQuery.isError) {
        return (
            <>
                <DialogContent>
                    <DialogContentText>User not found.</DialogContentText>
                </DialogContent>
                <DialogActions>
                    <DialogActions>
                        <Button
                            autoFocus
                            color='inherit'
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
                SAMLProviderId: getUserQuery.data.saml_provider_id?.toString() || '',
                roles: getUserQuery.data.roles?.map((role: any) => role.id) || [],
            }}
            roles={getRolesQuery.data}
            SAMLProviders={listSAMLProvidersQuery.data}
            isLoading={isLoading}
            error={error}
        />
    );
};

const UpdateUserFormInner: React.FC<{
    onCancel: () => void;
    onSubmit: (user: UpdatedUser) => void;
    initialData: {
        emailAddress: string;
        principal: string;
        firstName: string;
        lastName: string;
        SAMLProviderId?: string;
        roles: number[];
    };
    roles: any[];
    SAMLProviders: any[];
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, initialData, roles, SAMLProviders, isLoading, error }) => {
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
        watch,
    } = useForm({
        defaultValues: {
            ...initialData,
            authenticationMethod: initialData.SAMLProviderId ? 'saml' : 'password',
        },
    });

    const authenticationMethod = watch('authenticationMethod');

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SAMLProviderId', '');
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
                                render={({ field: { onChange, onBlur, value, ref }, formState, fieldState }) => (
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
                                            {SAMLProviders.length > 0 && <MenuItem value='saml'>SAML</MenuItem>}
                                        </Select>
                                    </FormControl>
                                )}
                            />
                        </Grid>

                        {authenticationMethod === 'saml' && (
                            <Grid item xs={12}>
                                <Controller
                                    name='SAMLProviderId'
                                    control={control}
                                    rules={{
                                        required: 'SAML Provider is required',
                                    }}
                                    render={({ field: { onChange, onBlur, value, ref }, formState, fieldState }) => (
                                        <FormControl>
                                            <InputLabel id='SAMLProviderId-label' sx={{ ml: '-14px', mt: '8px' }}>
                                                SAML Provider
                                            </InputLabel>
                                            <Select
                                                onChange={onChange as (event: SelectChangeEvent<string>) => void}
                                                onBlur={onBlur}
                                                value={value}
                                                ref={ref}
                                                labelId='SAMLProviderId-label'
                                                id='SAMLProviderId'
                                                name='SAMLProviderId'
                                                variant='standard'
                                                fullWidth
                                                data-testid='update-user-dialog_select-saml-provider'>
                                                {SAMLProviders.map((SAMLProvider: any) => (
                                                    <MenuItem value={SAMLProvider.id.toString()} key={SAMLProvider.id}>
                                                        {SAMLProvider.name}
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
                    autoFocus
                    color='inherit'
                    onClick={onCancel}
                    disabled={isLoading}
                    data-testid='update-user-dialog_button-close'>
                    Cancel
                </Button>
                <Button color='primary' type='submit' disabled={isLoading} data-testid='update-user-dialog_button-save'>
                    Save
                </Button>
            </DialogActions>
        </form>
    );
};

export default UpdateUserForm;
