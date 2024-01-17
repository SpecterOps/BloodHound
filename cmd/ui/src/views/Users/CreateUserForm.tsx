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
import { apiClient } from 'bh-shared-ui';
import { NewUser } from 'src/ducks/auth/types';

const CreateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: NewUser) => void;
    isLoading: boolean;
    error: any;
}> = ({ onCancel, onSubmit, isLoading, error }) => {
    const {
        control,
        handleSubmit,
        setValue,
        formState: { errors },
    } = useForm({
        defaultValues: {
            emailAddress: '',
            principal: '',
            firstName: '',
            lastName: '',
            password: '',
            needsPasswordReset: false,
            SAMLProviderId: '',
            roles: [1],
        },
    });

    const [authenticationMethod, setAuthenticationMethod] = React.useState<string>('password');

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SAMLProviderId', '');
        }
    }, [authenticationMethod, setValue]);

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data.data.roles)
    );

    const listSAMLProvidersQuery = useQuery(['listSAMLProviders'], ({ signal }) =>
        apiClient.listSAMLProviders({ signal }).then((res) => res.data.data.saml_providers)
    );

    const handleCancel: React.MouseEventHandler<HTMLButtonElement> = (e) => {
        e.preventDefault();
        onCancel();
    };

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            {!(getRolesQuery.isLoading || listSAMLProvidersQuery.isLoading) && (
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
                                            {listSAMLProvidersQuery.data.length > 0 && (
                                                <MenuItem value='saml'>SAML</MenuItem>
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
                                            name='SAMLProviderId'
                                            control={control}
                                            defaultValue=''
                                            rules={{
                                                required: 'SAML Provider is required',
                                            }}
                                            render={({
                                                field: { onChange, onBlur, value, ref },
                                                formState,
                                                fieldState,
                                            }) => (
                                                <FormControl error={!!errors.SAMLProviderId}>
                                                    <InputLabel
                                                        id='SAMLProviderId-label'
                                                        sx={{ ml: '-14px', mt: '8px' }}>
                                                        SAML Provider
                                                    </InputLabel>
                                                    <Select
                                                        onChange={
                                                            onChange as (event: SelectChangeEvent<string>) => void
                                                        }
                                                        onBlur={onBlur}
                                                        value={value}
                                                        ref={ref}
                                                        labelId='SAMLProviderId-label'
                                                        id='SAMLProviderId'
                                                        name='SAMLProviderId'
                                                        variant='standard'
                                                        fullWidth
                                                        data-testid='create-user-dialog_select-saml-provider'>
                                                        {listSAMLProvidersQuery.data.map((SAMLProvider: any) => (
                                                            <MenuItem
                                                                value={SAMLProvider.id.toString()}
                                                                key={SAMLProvider.id}>
                                                                {SAMLProvider.name}
                                                            </MenuItem>
                                                        ))}
                                                    </Select>
                                                    <FormHelperText>{errors.SAMLProviderId?.message}</FormHelperText>
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
                                                    getRolesQuery.data.map((role: any) => (
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
                            autoFocus
                            color='inherit'
                            onClick={handleCancel}
                            disabled={isLoading}
                            data-testid='create-user-dialog_button-close'>
                            Cancel
                        </Button>
                        <Button
                            color='primary'
                            type='submit'
                            disabled={isLoading}
                            data-testid='create-user-dialog_button-save'>
                            Save
                        </Button>
                    </DialogActions>
                </>
            )}
        </form>
    );
};

export default CreateUserForm;
