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
import { UpdatedUser } from 'src/ducks/auth/types';
import UpdateUserForm from './UpdateUserForm';

const UpdateUserDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: UpdatedUser) => Promise<any>;
    userId: string;
    isLoading: boolean;
    error: any;
}> = ({ open, onClose, onExited, userId, onSave, isLoading, error }) => {
    const handleOnSave = (user: UpdatedUser) => {
        onSave(user)
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
                isLoading={isLoading}
                error={error}
            />
        </Dialog>
    );
};

export default UpdateUserDialog;
