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
} from '@bloodhoundenterprise/doodleui';
import { Alert, Card, Checkbox, FormControl, FormControlLabel, Grid, TextField } from '@mui/material';
import { CreateUserRequest, SSOProvider } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { apiClient } from '../../utils';
import UserFormEnvironmentSelector from './UserFormEnvironmentSelector';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: CreateUserRequestForm) => void;
    isLoading: boolean;
    error: any;
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
    open?: boolean;
    /*
    className?: any;
    createUser?: boolean;
    disabled?: boolean;
    error: any;
    handleSubmit?: any;
    hasSelectedSelf?: boolean;
    initialData?: CreateUserRequestForm;
    isLoading: boolean;
    onCancel: () => void;
    onChange?: (value: string) => void;
    onSubmit: (user: CreateUserRequestForm) => void;
    open?: boolean;
    roles?: Role[];
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
    userId: string;
    value?: string;
    */
}> = ({
    //onCancel,
    onSubmit,
    isLoading,
    error,
    showEnvironmentAccessControls,
    open,
    /*
    disabled,
    error,
    isLoading,
    onCancel,
    onChange,
    onSubmit,
    open,
    roles,
    showEnvironmentAccessControls,
    userId,
    value,
    */
}) => {
    const {
        control,
        formState: { errors, isSubmitSuccessful },
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

    //console.log(selectedRole);

    const roleInputValue = watch('roles');
    //console.log(roleInputValue);

    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';

    console.log(errors);

    console.log(isSubmitSuccessful);

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                <div className='flex gap-x-4 justify-center'>
                    <Card className=' p-6 rounded shadow max-w-[600px]'>
                        <DialogTitle>{'Create User'}</DialogTitle>

                        <DialogDescription className='flex flex-col' data-testid='create-user-dialog'>
                            <Grid container spacing={2} className='min-h-[650px]'>
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
                                                data-testid='create-user-dialog_input-last-name'
                                            />
                                        )}
                                    />
                                </Grid>

                                <>
                                    <Grid item xs={12}>
                                        <FormControl>
                                            <Label className='text-base font-bold'>Authentication Method</Label>
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
                                                        {/* TODO: NEEDS TO BE IMPLEMENTED 
                                                                listSSOProvidersQuery.data && listSSOProvidersQuery.data?.length > 0 && (
                                                                    <SelectItem value='sso'>Single Sign-On (SSO)</SelectItem>
                                                                )*/}
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
                                            <FormControl>
                                                <Label id='SSOProviderId-label' className='text-base font-bold'>
                                                    SSO Provider
                                                </Label>
                                                <Select
                                                    data-testid='create-user-dialog_select-authentication-method'
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
                                                <Label className='text-base font-bold'>Role</Label>
                                                <Select
                                                    onValueChange={field.onChange}
                                                    value={isNaN(field.value) ? '' : field.value.toString()}
                                                    {...register('roles')}>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder={field.value} />
                                                    </SelectTrigger>
                                                    <SelectPortal>
                                                        <SelectContent>
                                                            {getRolesQuery.isLoading ? (
                                                                <SelectItem value={''}>Loading...</SelectItem>
                                                            ) : (
                                                                getRolesQuery.data?.map((role: any) => (
                                                                    <SelectItem
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
                                    data-testid='create-user-dialog_button-close'>
                                    Cancel
                                </Button>
                            </DialogClose>
                            <Button
                                type='submit'
                                disabled={isLoading}
                                data-testid='create-user-dialog_button-save'
                                //onClick={handleSubmit(onSubmit)}
                                onClick={() => {
                                    open;
                                }}>
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
