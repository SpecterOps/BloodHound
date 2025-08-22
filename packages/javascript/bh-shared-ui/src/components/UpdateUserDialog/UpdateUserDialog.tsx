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

import { DialogContent, DialogTitle, VisuallyHidden } from '@bloodhoundenterprise/doodleui';
import { UpdateUserRequest } from 'js-client-library';
import React from 'react';
import UpdateUserForm, { UpdateUserRequestForm } from '../UpdateUserForm';

const UpdateUserDialog: React.FC<{
    error: any;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: UpdateUserRequest) => Promise<any>;
    open?: boolean;
    showEnvironmentAccessControls?: boolean; //TODO: required or not?
    userId: string;
}> = ({ error, hasSelectedSelf, isLoading, onClose, onSave, open, showEnvironmentAccessControls, userId }) => {
    const handleOnSave = (user: UpdateUserRequestForm) => {
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
            <VisuallyHidden asChild>
                <DialogTitle>Edit User</DialogTitle>
            </VisuallyHidden>
            <UpdateUserForm
                error={error}
                isLoading={isLoading}
                onCancel={onClose}
                onSubmit={handleOnSave}
                hasSelectedSelf={hasSelectedSelf}
                open={open}
                showEnvironmentAccessControls={showEnvironmentAccessControls}
                userId={userId!}
            />
        </DialogContent>
    );
};

export default UpdateUserDialog;

{
    /*
<Dialog
        open={open}
        fullWidth={true}
        maxWidth={'sm'}
        onClose={onClose}
        disableEscapeKeyDown
        keepMounted={false}
        PaperProps={{
            // @ts-ignore
            'data-testid': 'update-user-dialog',
        }}
        TransitionProps={{
            onExited,
        }}>
        <DialogTitle>{'Update User'}</DialogTitle>
        <UpdateUserForm
            onCancel={onClose}
            onSubmit={handleOnSave}
            userId={userId}
            hasSelectedSelf={hasSelectedSelf}
            isLoading={isLoading}
            error={error}
        />
</Dialog>
*/
}
