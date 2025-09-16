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
    Card,
    Checkbox,
    DialogActions,
    DialogClose,
    DialogTitle,
    Form,
    FormControl,
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
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert } from '@mui/material';
import { CreateUserRequest, Environment, Role, SSOProvider } from 'js-client-library';
import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { MAX_EMAIL_LENGTH, MAX_NAME_LENGTH, MIN_NAME_LENGTH } from '../../constants';
import { useAvailableEnvironments } from '../../hooks/useAvailableEnvironments/useAvailableEnvironments';
import { apiClient } from '../../utils';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserForm: React.FC<{
    error: any;
    isLoading: boolean;
    onSubmit: (user: CreateUserRequestForm) => void;
    open?: boolean;
    showEnvironmentAccessControls?: boolean;
}> = ({
    error,
    isLoading,
    onSubmit,
    //open,
    showEnvironmentAccessControls,
}) => {
    const defaultValues = {
        emailAddress: '',
        principal: '',
        firstName: '',
        lastName: '',
        password: '',
        needsPasswordReset: false,
        roles: [3],
        SSOProviderId: '',
        environment_control_list: {
            environments: [],
            all_environments: false,
        },
    };

    const form = useForm<CreateUserRequestForm>({ defaultValues });

    const [authenticationMethod, setAuthenticationMethod] = React.useState<string>('password');
    const [selectedRoleValue, setSelectedRoleValue] = useState([3]);

    const roleInputValue = form.watch('roles');
    const selectedRole = roleInputValue.toString() === '2' || roleInputValue.toString() === '3';

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data?.data?.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    const { data: availableEnvironments } = useAvailableEnvironments();

    const [searchInput, setSearchInput] = useState<string>('');
    const [selectedEnvironments, setSelectedEnvironments] = useState<string[]>([]);
    const [needsPasswordReset, setNeedsPasswordReset] = useState(false);

    const handleCheckedChange = (checked: boolean | 'indeterminate') => {
        setNeedsPasswordReset(checked === true);
        if (checked === true) {
            form.setValue('needsPasswordReset', true);
        } else {
            form.setValue('needsPasswordReset', false);
        }
    };

    const filteredEnvironments = availableEnvironments?.filter((environment: Environment) =>
        environment.name.toLowerCase().includes(searchInput.toLowerCase())
    );

    const handleSelectAllChange = (checked: any) => {
        if (checked) {
            const returnMappedEnvironments: string[] | undefined = availableEnvironments?.map((item) => item.id);
            setSelectedEnvironments(returnMappedEnvironments || []);
        } else {
            setSelectedEnvironments([]);
        }
    };

    const handleItemChange = (itemId: any, checked: any) => {
        if (checked) {
            setSelectedEnvironments((prevSelected) => [...prevSelected, itemId]);
        } else {
            setSelectedEnvironments((prevSelected) => prevSelected.filter((id) => id !== itemId));
        }
    };

    const isAllSelected =
        selectedEnvironments.length === availableEnvironments?.length && availableEnvironments.length > 0;

    useEffect(() => {
        if (authenticationMethod === 'password') {
            form.setValue('SSOProviderId', undefined);
        }

        if (error) {
            if (error?.response?.status === 409) {
                if (error.response?.data?.errors[0]?.message.toLowerCase().includes('principal name')) {
                    form.setError('principal', { type: 'custom', message: 'Principal name is already in use.' });
                } else if (error.response?.data?.errors[0]?.message.toLowerCase().includes('email')) {
                    form.setError('emailAddress', { type: 'custom', message: 'Email is already in use.' });
                } else {
                    form.setError('root.generic', { type: 'custom', message: `A conflict has occured.` });
                }
            } else {
                form.setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    }, [authenticationMethod, form, form.setValue, error, form.setError]);

    return (
        <Form {...form}>
            <form autoComplete='off' onSubmit={form.handleSubmit(onSubmit)}>
                {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                    <div className='flex gap-x-4 justify-center'>
                        <Card className='p-6 rounded shadow max-w-[600px] w-full'>
                            <DialogTitle>{'Create User'}</DialogTitle>

                            <div className='flex flex-col mt-4 w-full' data-testid='create-user-dialog_content'>
                                <div className='mb-4'>
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
                                                <FormLabel className='font-medium !text-sm' htmlFor='emailAddress'>
                                                    Email Address
                                                </FormLabel>
                                                <FormControl>
                                                    <Input
                                                        {...field}
                                                        type='email'
                                                        id='emailAddress'
                                                        placeholder='user@domain.com'
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>

                                <div className='mb-4'>
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
                                                <FormLabel className='font-medium !text-sm' htmlFor='principal'>
                                                    Principal Name{' '}
                                                </FormLabel>
                                                <FormControl>
                                                    <Input {...field} id='principal' placeholder='Principal Name' />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                                <div className='mb-4'>
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
                                                    <FormLabel className='font-medium !text-sm' htmlFor='firstName'>
                                                        First Name
                                                    </FormLabel>
                                                    <FormControl>
                                                        <Input {...field} id='firstName' placeholder='First Name' />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            </>
                                        )}
                                    />
                                </div>
                                <div className='mb-4'>
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
                                                    <FormLabel className='font-medium !text-sm' htmlFor='lastName'>
                                                        Last Name
                                                    </FormLabel>
                                                    <FormControl>
                                                        <Input {...field} id='lastName' placeholder='Last Name' />
                                                    </FormControl>
                                                    <FormMessage />
                                                </FormItem>
                                            </>
                                        )}
                                    />
                                </div>

                                <>
                                    <div className='mb-4'>
                                        <FormItem>
                                            <FormLabel className='font-medium !text-sm' htmlFor='authenticationMethod'>
                                                Authentication Method
                                            </FormLabel>
                                            <Select
                                                onValueChange={(value) => setAuthenticationMethod(value as string)}
                                                value={authenticationMethod}>
                                                <FormControl className='mt-2'>
                                                    <SelectTrigger
                                                        variant='underlined'
                                                        className='bg-transparent'
                                                        id='authenticationMethod'>
                                                        <SelectValue placeholder={authenticationMethod} />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        <SelectItem value='password'>Username / Password</SelectItem>
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
                                    </div>

                                    {authenticationMethod === 'password' ? (
                                        <>
                                            <div className='mb-4'>
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
                                                            <FormLabel
                                                                className='font-medium !text-sm'
                                                                htmlFor='password'>
                                                                Initial Password
                                                            </FormLabel>
                                                            <FormControl>
                                                                <Input
                                                                    {...field}
                                                                    id='password'
                                                                    type='password'
                                                                    placeholder='Initial Password'
                                                                />
                                                            </FormControl>
                                                            <FormMessage />
                                                        </FormItem>
                                                    )}
                                                />
                                            </div>
                                            <div className='mb-4'>
                                                <FormField
                                                    name='needsPasswordReset'
                                                    control={form.control}
                                                    defaultValue={false}
                                                    render={() => (
                                                        <div className='flex flex-row items-center'>
                                                            <FormItem className='flex flex-row my-3'>
                                                                <FormControl>
                                                                    <Checkbox
                                                                        id='needsPasswordReset'
                                                                        onCheckedChange={handleCheckedChange}
                                                                        checked={needsPasswordReset}
                                                                    />
                                                                </FormControl>
                                                                <FormLabel
                                                                    htmlFor='needsPasswordReset'
                                                                    className='pl-2 font-medium !text-sm'>
                                                                    Force Password Reset?
                                                                </FormLabel>
                                                            </FormItem>
                                                        </div>
                                                    )}
                                                />
                                            </div>
                                        </>
                                    ) : (
                                        <div className='mb-4'>
                                            <FormItem>
                                                <FormLabel htmlFor='sso'>SSO Provider</FormLabel>
                                                <Select
                                                    onValueChange={(value) => setAuthenticationMethod(value as string)}
                                                    value={authenticationMethod}>
                                                    <FormControl>
                                                        <SelectTrigger
                                                            variant='underlined'
                                                            className='bg-transparent'
                                                            id='sso'>
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
                                        </div>
                                    )}
                                </>

                                <div className=''>
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
                                                    <FormLabel className='mr-2 font-medium !text-sm' htmlFor='role'>
                                                        Role
                                                    </FormLabel>
                                                    <Tooltip
                                                        tooltip='Only User, Read-Only, Upload-Only roles contain the limited access functionality.'
                                                        contentProps={{
                                                            className: 'dark:bg-neutral-dark-5 border-0',
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
                                                        <SelectTrigger
                                                            variant='underlined'
                                                            className='bg-transparent'
                                                            id='role'>
                                                            <SelectValue placeholder={field.value} />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectPortal>
                                                        <SelectContent>
                                                            {getRolesQuery.isLoading ? (
                                                                <SelectItem value={'loading'}>Loading...</SelectItem>
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
                                </div>
                                {form.formState.errors.root?.generic && (
                                    <div>
                                        <Alert severity='error'>{form.formState.errors.root.generic.message}</Alert>
                                    </div>
                                )}
                            </div>
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
                        {showEnvironmentAccessControls && selectedRole && (
                            <Card className='flex-1 p-4 rounded shadow max-w-[400px]'>
                                <DialogTitle>Environmental Access Control</DialogTitle>
                                <div
                                    className='flex flex-col h-full pb-6'
                                    data-testid='create-user-dialog_environments-checkboxes-dialog'>
                                    <div className='border border-color-[#CACFD3] mt-3 box-border h-full overflow-y-auto'>
                                        <div className='border border-solid border-color-[#CACFD3] flex'>
                                            <FontAwesomeIcon className='ml-4 mt-3' icon={faSearch} />
                                            <Input
                                                variant='underlined'
                                                className='w-full ml-3'
                                                id='search'
                                                type='text'
                                                placeholder='Search'
                                                onChange={(e) => {
                                                    setSearchInput(e.target.value);
                                                }}
                                            />
                                        </div>
                                        <div
                                            className='flex flex-row ml-4 mt-6 mb-2 items-center'
                                            data-testid='create-user-dialog_environments-checkboxes-select-all'>
                                            <FormField
                                                name='allEnvironments'
                                                //control={form.control}  // TODO: uncomment when form controls available via api
                                                defaultValue={false}
                                                render={() => (
                                                    <FormItem className='flex flex-row items-center'>
                                                        <Checkbox
                                                            checked={isAllSelected}
                                                            id='allEnvironments'
                                                            onCheckedChange={handleSelectAllChange}
                                                        />
                                                        <FormLabel
                                                            htmlFor='allEnvironments'
                                                            className='ml-3 w-full cursor-pointer font-normal'>
                                                            Select All Environments
                                                        </FormLabel>
                                                    </FormItem>
                                                )}
                                            />
                                        </div>
                                        <div
                                            className='flex flex-col'
                                            data-testid='create-user-dialog_environments-checkboxes'>
                                            {filteredEnvironments &&
                                                filteredEnvironments?.map((item) => {
                                                    return (
                                                        <div
                                                            className='flex justify-start items-center ml-5'
                                                            data-testid='create-user-dialog_environments-checkbox'>
                                                            <FormField
                                                                name='environments'
                                                                //control={form.control} // TODO: uncomment when form controls available via api
                                                                defaultValue={false}
                                                                render={({ field }) => (
                                                                    <FormItem className='flex flex-row items-center'>
                                                                        <Checkbox
                                                                            {...field}
                                                                            checked={selectedEnvironments.includes(
                                                                                item.id
                                                                            )}
                                                                            className='m-3'
                                                                            id='environments'
                                                                            onCheckedChange={(checked) =>
                                                                                handleItemChange(item.id, checked)
                                                                            }
                                                                            value={item.name} // environment_control_list.environments
                                                                        />
                                                                        <FormLabel
                                                                            htmlFor='environments'
                                                                            className='mr-3 w-full cursor-pointer font-normal'>
                                                                            {item.name}
                                                                        </FormLabel>
                                                                    </FormItem>
                                                                )}
                                                            />
                                                        </div>
                                                    );
                                                })}
                                        </div>
                                    </div>
                                </div>
                            </Card>
                        )}
                    </div>
                )}
            </form>
        </Form>
    );
};

export default CreateUserForm;
