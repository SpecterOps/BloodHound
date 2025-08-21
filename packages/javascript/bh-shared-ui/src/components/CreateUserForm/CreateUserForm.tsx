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
import { Alert, Card, Checkbox, FormControl, FormControlLabel, Grid, TextField } from '@mui/material';
import { CreateUserRequest, Role, SSOProvider } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { apiClient } from '../../utils';
import UserFormEnvironmentSelector from './UserFormEnvironmentSelector';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserForm: React.FC<{
    error: any;
    isLoading: boolean;
    onCancel: () => void;
    onSubmit: (user: CreateUserRequestForm) => void;
    open?: boolean;
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
}> = ({ error, isLoading, onSubmit, open, showEnvironmentAccessControls }) => {
    const {
        control,
        formState: { errors },
        handleSubmit,
        setError,
        setValue,
        register,
        watch,
    } = useForm<CreateUserRequestForm>({
        defaultValues: {
            emailAddress: '',
            principal: '',
            firstName: '',
            lastName: '',
            password: '',
            needsPasswordReset: false,
            roles: [3],
            SSOProviderId: '',
        },
    });

    const [authenticationMethod, setAuthenticationMethod] = useState<string>('password');
    const [selectedRoleValue, setSelectedRoleValue] = useState([3]);

    const roleInputValue = watch('roles');
    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data?.data?.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    useEffect(() => {
        if (authenticationMethod === 'password') {
            setValue('SSOProviderId', undefined);
        }

        if (error) {
            if (error?.response?.status === 409) {
                if (error.response?.data?.errors[0]?.message.toLowerCase().includes('principal name')) {
                    setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (error.response?.data?.errors[0]?.message.toLowerCase().includes('email')) {
                    setError('emailAddress', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            } else {
                setError('root.oeneric', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    }, [authenticationMethod, setValue, error, setError]);

    if (error) {
        console.log(error);
    }

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                <div className='flex gap-x-4 justify-center'>
                    <Card className=' p-6 rounded shadow max-w-[600px]'>
                        <DialogTitle>{'Create User'}</DialogTitle>

                        <DialogDescription className='flex flex-col' data-testid='create-user-dialog_content'>
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
                                            <>
                                                <Label size='small'>Email Address</Label>
                                                <TextField
                                                    {...field}
                                                    data-testid='create-user-dialog_input-email-address'
                                                    error={!!errors.emailAddress}
                                                    fullWidth
                                                    helperText={errors.emailAddress?.message}
                                                    id='emailAddress'
                                                    label='Email Address'
                                                    type='email'
                                                    variant='standard'
                                                />
                                            </>
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
                                            <>
                                                <Label size='small'>Principal Name</Label>
                                                <TextField
                                                    {...field}
                                                    data-testid='create-user-dialog_input-principal-name'
                                                    error={!!errors.principal}
                                                    fullWidth
                                                    helperText={errors.principal?.message}
                                                    id='principal'
                                                    label='Principal Name'
                                                    variant='standard'
                                                />
                                            </>
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
                                            <>
                                                <Label size='small'>First Name</Label>
                                                <TextField
                                                    {...field}
                                                    data-testid='create-user-dialog_input-first-name'
                                                    error={!!errors.firstName}
                                                    fullWidth
                                                    helperText={errors.firstName?.message}
                                                    id='firstName'
                                                    label='First Name'
                                                    variant='standard'
                                                />
                                            </>
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
                                            <>
                                                <Label size='small'>Last Name</Label>
                                                <TextField
                                                    {...field}
                                                    data-testid='create-user-dialog_input-last-name'
                                                    error={!!errors.lastName}
                                                    fullWidth
                                                    helperText={errors.lastName?.message}
                                                    id='lastName'
                                                    label='Last Name'
                                                    variant='standard'
                                                />
                                            </>
                                        )}
                                    />
                                </Grid>

                                <>
                                    <Grid item xs={12}>
                                        <FormControl>
                                            <Label size='small'>Authentication Method</Label>
                                            <Select
                                                data-testid='create-user-dialog_select-authentication-method'
                                                onValueChange={(value) => setAuthenticationMethod(value as string)}
                                                value={authenticationMethod}>
                                                <SelectTrigger>
                                                    <SelectValue placeholder={authenticationMethod} />
                                                </SelectTrigger>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        <SelectItem value='password'>Username / Password</SelectItem>
                                                        <SelectItem value='sso'>Single Sign-On (SSO)</SelectItem>
                                                        {listSSOProvidersQuery.data &&
                                                            listSSOProvidersQuery.data?.length > 0 && (
                                                                <SelectItem value='sso'>
                                                                    Single Sign-On (SSO)
                                                                </SelectItem>
                                                            )}
                                                    </SelectContent>
                                                </SelectPortal>
                                            </Select>
                                            {/* TODO: REMOVE
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
                                            */}
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
                                                        maxLength: {
                                                            value: 1000,
                                                            message: 'Password must be less than 1000 characters',
                                                        },
                                                    }}
                                                    render={({ field }) => (
                                                        <>
                                                            <Label size='small'>Password</Label>
                                                            <TextField
                                                                {...field}
                                                                data-testid='create-user-dialog_input-password'
                                                                error={!!errors.password}
                                                                fullWidth
                                                                helperText={errors.password?.message}
                                                                id='password'
                                                                label='Initial Password'
                                                                type='password'
                                                                variant='standard'
                                                            />
                                                        </>
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
                                                                    color='primary'
                                                                    data-testid='create-user-dialog_checkbox-needs-password-reset'
                                                                    onChange={(e, checked) => field.onChange(checked)}
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
                                            <FormControl>
                                                <Label size='small'>SSO Provider</Label>
                                                <Select
                                                    data-testid='create-user-dialog_sso-provider'
                                                    onValueChange={(value) => setAuthenticationMethod(value as string)}
                                                    value={authenticationMethod}>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder='SSO Provider' />
                                                    </SelectTrigger>
                                                    <SelectPortal>
                                                        <SelectContent>
                                                            {listSSOProvidersQuery.data?.map(
                                                                (SSOProvider: SSOProvider) => (
                                                                    <SelectItem
                                                                        value={SSOProvider.id.toString()}
                                                                        key={SSOProvider.id}>
                                                                        {SSOProvider.name}
                                                                    </SelectItem>
                                                                )
                                                            )}
                                                        </SelectContent>
                                                    </SelectPortal>
                                                </Select>
                                            </FormControl>
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
                                                            {getRolesQuery.isLoading ? (
                                                                <SelectItem value={''}>Loading...</SelectItem>
                                                            ) : (
                                                                getRolesQuery.data?.map((role: Role) => (
                                                                    <SelectItem
                                                                        className='hover:cursor-pointer'
                                                                        key={role.id}
                                                                        value={role.id.toString()}>
                                                                        {role.name}
                                                                    </SelectItem>
                                                                ))
                                                            )}
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
            )}
        </form>
    );
};

export default CreateUserForm;
