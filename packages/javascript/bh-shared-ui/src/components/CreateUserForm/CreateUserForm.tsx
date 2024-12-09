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
    Checkbox,
    DialogActions,
    DialogContent,
    FormControl,
    FormControlLabel,
    FormHelperText,
    Grid,
    InputLabel,
    MenuItem,
    Select,
    SelectChangeEvent,
    TextField,
} from '@mui/material';
import React, { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import { CreateUserRequest, SSOProvider } from 'js-client-library';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: CreateUserRequestForm) => void;
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, isLoading, error }) => {
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
    } = useForm<CreateUserRequestForm>({
        defaultValues: {
            emailAddress: '',
            principal: '',
            firstName: '',
            lastName: '',
            password: '',
            needsPasswordReset: false,
            roles: [1],
            SSOProviderId: '',
        },
    });

    const [authenticationMethod, setAuthenticationMethod] = React.useState<string>('password');

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SSOProviderId', undefined);
        }
    }, [authenticationMethod, setValue]);

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data?.data?.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    const handleCancel: React.MouseEventHandler<HTMLButtonElement> = (e) => {
        e.preventDefault();
        onCancel();
    };

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                <>
                    <DialogContent>
                        <Grid container spacing={2}>
                            <Grid item xs={12}>
                                <Controller
                                    name='emailAddress'
                                    control={control}
                                    rules={{
                                        required: 'Email Address is required',
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
                                            data-testid='create-user-dialog_input-email-address'
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
                                            data-testid='create-user-dialog_input-principal-name'
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
                                            data-testid='create-user-dialog_input-first-name'
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
                                            data-testid='create-user-dialog_input-last-name'
                                        />
                                    )}
                                />
                            </Grid>

                            <>
                                <Grid item xs={12}>
                                    <FormControl>
                                        <InputLabel id='authenticationMethod-label' sx={{ ml: '-14px', mt: '8px' }}>
                                            Authentication Method
                                        </InputLabel>
                                        <Select
                                            labelId='authenticationMethod-label'
                                            id='authenticationMethod'
                                            name='authenticationMethod'
                                            onChange={(e) => setAuthenticationMethod(e.target.value as string)}
                                            value={authenticationMethod}
                                            variant='standard'
                                            fullWidth
                                            data-testid='create-user-dialog_select-authentication-method'>
                                            <MenuItem value='password'>Username / Password</MenuItem>
                                            {listSSOProvidersQuery.data && listSSOProvidersQuery.data?.length > 0 && (
                                                <MenuItem value='sso'>Single Sign-On (SSO)</MenuItem>
                                            )}
                                        </Select>
                                    </FormControl>
                                </Grid>

                                {authenticationMethod === 'password' ? (
                                    <>
                                        <Grid item xs={12}>
                                            <Controller
                                                name='password'
                                                control={control}
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
                                                }}
                                                render={({ field }) => (
                                                    <TextField
                                                        {...field}
                                                        variant='standard'
                                                        id='password'
                                                        label='Initial Password'
                                                        type='password'
                                                        fullWidth
                                                        error={!!errors.password}
                                                        helperText={errors.password?.message}
                                                        data-testid='create-user-dialog_input-password'
                                                    />
                                                )}
                                            />
                                        </Grid>
                                        <Grid item xs={12}>
                                            <Controller
                                                name='needsPasswordReset'
                                                control={control}
                                                defaultValue={false}
                                                render={({ field }) => (
                                                    <FormControlLabel
                                                        control={
                                                            <Checkbox
                                                                {...field}
                                                                onChange={(e, checked) => field.onChange(checked)}
                                                                color='primary'
                                                                data-testid='create-user-dialog_checkbox-needs-password-reset'
                                                            />
                                                        }
                                                        label='Force Password Reset?'
                                                    />
                                                )}
                                            />
                                        </Grid>
                                    </>
                                ) : (
                                    <Grid item xs={12}>
                                        <Controller
                                            name='SSOProviderId'
                                            control={control}
                                            rules={{
                                                required: 'SSO Provider is required',
                                            }}
                                            render={({ field: { onChange, onBlur, value, ref } }) => (
                                                <FormControl error={!!errors.SSOProviderId}>
                                                    <InputLabel
                                                        id='SSOProviderId-label'
                                                        sx={{ ml: '-14px', mt: '8px' }}>
                                                        SSO Provider
                                                    </InputLabel>
                                                    <Select
                                                        onChange={
                                                            onChange as (event: SelectChangeEvent<string>) => void
                                                        }
                                                        defaultValue={''}
                                                        onBlur={onBlur}
                                                        value={value}
                                                        ref={ref}
                                                        labelId='SSOProviderId-label'
                                                        id='SSOProviderId'
                                                        name='SSOProviderId'
                                                        variant='standard'
                                                        fullWidth
                                                        data-testid='create-user-dialog_select-sso-provider'>
                                                        {listSSOProvidersQuery.data?.map((SSOProvider: SSOProvider) => (
                                                            <MenuItem value={SSOProvider.id} key={SSOProvider.id}>
                                                                {SSOProvider.name}
                                                            </MenuItem>
                                                        ))}
                                                    </Select>
                                                    <FormHelperText>{errors.SSOProviderId?.message}</FormHelperText>
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
                                                data-testid='create-user-dialog_select-role'>
                                                {getRolesQuery.isLoading ? (
                                                    <MenuItem value={1}>Loading...</MenuItem>
                                                ) : (
                                                    getRolesQuery.data?.map((role: any) => (
                                                        <MenuItem key={role.id} value={role.id.toString()}>
                                                            {role.name}
                                                        </MenuItem>
                                                    ))
                                                )}
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
                            onClick={handleCancel}
                            disabled={isLoading}
                            data-testid='create-user-dialog_button-close'>
                            Cancel
                        </Button>
                        <Button type='submit' disabled={isLoading} data-testid='create-user-dialog_button-save'>
                            Save
                        </Button>
                    </DialogActions>
                </>
            )}
        </form>
    );
};

export default CreateUserForm;
