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

//import { Dialog, DialogTitle } from '@mui/material';
import { DialogContent } from '@bloodhoundenterprise/doodleui';
import { CreateUserRequest } from 'js-client-library';
import React from 'react';
import CreateUserForm, { CreateUserRequestForm } from '../CreateUserForm';

const CreateUserDialog: React.FC<{
    createUser: boolean;
    updateUser: boolean;
    error: any;
    isLoading: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: CreateUserRequest) => Promise<any>;
    open: boolean;
    userId?: string;
    hasSelectedSelf?: boolean;
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
}> = ({
    open,
    onClose,
    onExited,
    onSave,
    isLoading,
    error,
    showEnvironmentAccessControls,
    createUser,
    updateUser,
    userId,
}) => {
    const handleOnSave = (user: CreateUserRequestForm) => {
        let parsedSSOProviderId: number | undefined = undefined;
        if (user.SSOProviderId) {
            parsedSSOProviderId = parseInt(user.SSOProviderId);
        }

        onSave({
            ...user,
            SSOProviderId: parsedSSOProviderId,
        })
            .then(() => {
                onClose();
            })
            .catch((err) => console.error(err));
    };

    return (
        <DialogContent maxWidth='lg' className='!bg-transparent'>
            <CreateUserForm
                createUser={createUser}
                error={error}
                isLoading={isLoading}
                onCancel={onClose}
                onSubmit={handleOnSave}
                open={open}
                showEnvironmentAccessControls={showEnvironmentAccessControls}
                updateUser={updateUser}
                userId={userId!}
            />
        </DialogContent>
    );
};

export default CreateUserDialog;

/*
        <>
            {showEnvironmentAccessControls ? (
                <DialogContent maxWidth='lg' className='!bg-transparent'>
                    <CreateUserForm
                        createUser={createUser}
                        error={error}
                        isLoading={isLoading}
                        onCancel={onClose}
                        onSubmit={handleOnSave}
                        open={open}
                        showEnvironmentAccessControls={true}
                        updateUser={updateUser}
                        userId={userId}
                    />
                </DialogContent>
            ) : (
                <DialogContent maxWidth='lg' className=''>
                    <CreateUserForm
                        //createUser={createUser} //TODO: NEEDED FOR BHCE?
                        //updateUser={updateUser} //TODO: NEEDED FOR BHCE?
                        error={error}
                        isLoading={isLoading}
                        onCancel={onClose}
                        onSubmit={handleOnSave}
                        showEnvironmentAccessControls={false}
                        userId={userId}
                    />
                </DialogContent>
            )}
        </>
*/

{
    /*
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            onClose={onClose}
            disableEscapeKeyDown
            PaperProps={{
                //@ts-ignore
                'data-testid': 'create-user-dialog',
            }}
            TransitionProps={{
                onExited,
            }}>
            <DialogTitle>{'Create User'}</DialogTitle>
            <CreateUserForm onCancel={onClose} onSubmit={handleOnSave} isLoading={isLoading} error={error} />
        </Dialog>
        */
}
