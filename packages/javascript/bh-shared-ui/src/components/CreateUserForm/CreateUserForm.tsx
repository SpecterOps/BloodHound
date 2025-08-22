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
import { Card, Checkbox, FormControl, FormControlLabel, Grid } from '@mui/material';
import { CreateUserRequest, Role, SSOProvider } from 'js-client-library';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
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
    /*
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

    */

    const defaultValues = {
        emailAddress: '',
        principal: '',
        firstName: '',
        lastName: '',
        password: '',
        needsPasswordReset: false,
        roles: [3],
        SSOProviderId: '',
    };

    const form = useForm<CreateUserRequestForm>({ defaultValues });

    const [authenticationMethod, setAuthenticationMethod] = useState<string>('password');
    const [selectedRoleValue, setSelectedRoleValue] = useState([3]);

    const roleInputValue = form.watch('roles');
    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data?.data?.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    /*
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
    */

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(onSubmit)}>
                {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                    <div className='flex gap-x-4 justify-center'>
                        <Card className=' p-6 rounded shadow max-w-[600px]'>
                            <DialogTitle>{'Create User'}</DialogTitle>

                            <DialogDescription className='flex flex-col' data-testid='create-user-dialog_content'>
                                <Grid container spacing={2} className='min-h-[650px] mt-4'>
                                    <Grid item xs={12}>
                                        <FormField
                                            name='emailAddress'
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
                                                    <FormLabel
                                                        aria-labelledby='emailAddress'
                                                        data-testid='create-user-dialog_label-email-address'>
                                                        Email Address
                                                    </FormLabel>
                                                    <FormControl>
                                                        <Input
                                                            {...field}
                                                            type='email'
                                                            id='emailAddress'
                                                            data-testid='create-user-dialog_input-email-address'
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
                                                    <FormLabel>Principal Name</FormLabel>
                                                    <FormControl>
                                                        <Input
                                                            {...field}
                                                            data-testid='create-user-dialog_input-principal-name'
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
                                                <>
                                                    <FormItem>
                                                        <FormLabel>First Name</FormLabel>
                                                        <FormControl>
                                                            <Input
                                                                {...field}
                                                                data-testid='create-user-dialog_input-first-name'
                                                                id='firstName'
                                                            />
                                                        </FormControl>
                                                        <FormMessage />
                                                    </FormItem>
                                                </>
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
                                                <>
                                                    <FormItem>
                                                        <FormLabel>Last Name</FormLabel>
                                                        <FormControl>
                                                            <Input
                                                                {...field}
                                                                data-testid='create-user-dialog_input-last-name'
                                                                id='lastName'
                                                            />
                                                        </FormControl>
                                                    </FormItem>
                                                </>
                                            )}
                                        />
                                    </Grid>

                                    <>
                                        <Grid item xs={12}>
                                            <FormItem>
                                                <FormLabel>Authentication Method</FormLabel>
                                                <Select
                                                    data-testid='create-user-dialog_select-authentication-method'
                                                    onValueChange={(value) => setAuthenticationMethod(value as string)}
                                                    value={authenticationMethod}>
                                                    <FormControl>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder={authenticationMethod} />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectPortal>
                                                        <SelectContent>
                                                            <SelectItem value='password'>
                                                                Username / Password
                                                            </SelectItem>
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
                                            </FormItem>
                                        </Grid>

                                        {authenticationMethod === 'password' ? (
                                            <>
                                                <Grid item xs={12}>
                                                    <FormField
                                                        name='password'
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
                                                                <FormLabel>Password</FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        {...field}
                                                                        data-testid='create-user-dialog_input-password'
                                                                        id='password'
                                                                        type='password'
                                                                    />
                                                                </FormControl>
                                                            </FormItem>
                                                        )}
                                                    />
                                                </Grid>
                                                <Grid item xs={12}>
                                                    <FormField
                                                        name='needsPasswordReset'
                                                        control={form.control}
                                                        defaultValue={false}
                                                        render={({ field }) => (
                                                            <FormControlLabel
                                                                control={
                                                                    <Checkbox
                                                                        {...field}
                                                                        color='primary'
                                                                        data-testid='create-user-dialog_checkbox-needs-password-reset'
                                                                        onChange={(e, checked) =>
                                                                            field.onChange(checked)
                                                                        }
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
                                                <FormItem>
                                                    <FormLabel>SSO Provider</FormLabel>
                                                    <Select
                                                        data-testid='create-user-dialog_sso-provider'
                                                        onValueChange={(value) =>
                                                            setAuthenticationMethod(value as string)
                                                        }
                                                        value={authenticationMethod}>
                                                        <FormControl>
                                                            <SelectTrigger>
                                                                <SelectValue placeholder='SSO Provider' />
                                                            </SelectTrigger>
                                                        </FormControl>
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
                                                </FormItem>
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
                                                <FormItem>
                                                    <div className='flex row'>
                                                        <FormLabel className='mr-2'>Role</FormLabel>
                                                        <Tooltip
                                                            tooltip='Only User, Read-Only, Upload-Only roles contain the limited access functionality.'
                                                            contentProps={{
                                                                className: 'max-w-80 dark:bg-neutral-dark-5 border-0',
                                                            }}
                                                        />
                                                    </div>
                                                    <Select
                                                        {...form.register('roles')}
                                                        data-testid='create-user-dialog_role'
                                                        onValueChange={(field) => {
                                                            form.setValue('roles', [Number(field)]);
                                                            setSelectedRoleValue([Number([field])]);
                                                        }}
                                                        value={String(selectedRoleValue)}>
                                                        <FormControl>
                                                            <SelectTrigger>
                                                                <SelectValue placeholder={field.value} />
                                                            </SelectTrigger>
                                                        </FormControl>
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
                                                </FormItem>
                                            )}
                                        />
                                    </Grid>
                                    {/*
                                    {!!errors.root?.generic && (
                                        <Grid item xs={12}>
                                            <Alert severity='error'>{errors.root.generic.message}</Alert>
                                        </Grid>
                                    )}
                                    */}
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
        </Form>
    );
};

export default CreateUserForm;
