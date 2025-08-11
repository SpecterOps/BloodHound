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

import { DialogContent } from '@bloodhoundenterprise/doodleui';
import { CreateUserRequest } from 'js-client-library';
import React, { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';
import CreateUserFormLeftPanel from './CreateUserFormLeftPanel';
import CreateUserFormRightPanel from './CreateUserFormRightPanel';

export type CreateUserRequestForm = Omit<CreateUserRequest, 'SSOProviderId'> & { SSOProviderId: string | undefined };

const CreateUserForm: React.FC<{
    onCancel: () => void;
    onSubmit: (user: CreateUserRequestForm) => void;
    isLoading: boolean;
    error: any;
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
}> = ({
    //onCancel,
    onSubmit,
    //isLoading,
    error,
    showEnvironmentAccessControls = true,
}) => {
    const {
        //control,
        handleSubmit,
        setValue,
        //formState: { errors },
        setError,
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

    const [authenticationMethod] = React.useState<string>('password');
    //const [authenticationMethod, setAuthenticationMethod] = React.useState<string>('password');

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
                setError('root.generic', {
                    type: 'custom',
                    message: 'An unexpected error occurred. Please try again.',
                });
            }
        }
    }, [authenticationMethod, setValue, error, setError]);

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data?.data?.roles)
    );

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data?.data)
    );

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
            {!(getRolesQuery.isLoading || listSSOProvidersQuery.isLoading) && (
                <div className=''>
                    {showEnvironmentAccessControls ? (
                        <div className='flex gap-x-4 justify-center'>
                            <CreateUserFormLeftPanel />
                            <CreateUserFormRightPanel />
                        </div>
                    ) : (
                        <div className=''>
                            <DialogContent>
                                <CreateUserFormLeftPanel />
                            </DialogContent>
                        </div>
                    )}
                </div>
            )}
        </form>
    );
};

export default CreateUserForm;
