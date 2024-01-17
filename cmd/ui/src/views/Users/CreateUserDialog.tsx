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
import { NewUser } from 'src/ducks/auth/types';
import CreateUserForm from './CreateUserForm';

const CreateUserDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: NewUser) => Promise<any>;
    isLoading: boolean;
    error: any;
}> = ({ open, onClose, onExited, onSave, isLoading, error }) => {
    const handleOnSave = (user: NewUser) => {
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
    );
};

export default CreateUserDialog;
