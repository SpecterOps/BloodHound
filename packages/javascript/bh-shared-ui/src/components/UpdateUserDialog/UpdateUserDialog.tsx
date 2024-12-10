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

import { Dialog, DialogTitle } from '@mui/material';
import React from 'react';
import UpdateUserForm, { UpdateUserRequestForm } from '../UpdateUserForm';
import { UpdateUserRequest } from 'js-client-library';

const UpdateUserDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: UpdateUserRequest) => Promise<any>;
    userId: string;
    hasSelectedSelf: boolean;
    isLoading: boolean;
    error: any;
}> = ({ open, onClose, onExited, userId, onSave, hasSelectedSelf, isLoading, error }) => {
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
    );
};

export default UpdateUserDialog;
