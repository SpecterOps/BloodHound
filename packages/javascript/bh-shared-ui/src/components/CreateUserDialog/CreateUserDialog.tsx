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

import { DialogContent, DialogOverlay, DialogTitle, VisuallyHidden } from '@bloodhoundenterprise/doodleui';
import { CreateUserRequest } from 'js-client-library';
import React from 'react';
import CreateUserForm, { CreateUserRequestForm } from '../CreateUserForm';

const CreateUserDialog: React.FC<{
    error: any;
    isLoading: boolean;
    onClose: () => void;
    onExited?: () => void;
    onSave: (user: CreateUserRequest) => Promise<any>;
    open: boolean;
    showEnvironmentAccessControls: boolean;
}> = ({ error, isLoading, onClose, onSave, open, showEnvironmentAccessControls }) => {
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
        <DialogOverlay>
            <DialogContent maxWidth='lg' className='!bg-transparent overflow-y-auto max-h-screen'>
                <VisuallyHidden asChild>
                    <DialogTitle>Create User</DialogTitle>
                </VisuallyHidden>
                <CreateUserForm
                    error={error}
                    isLoading={isLoading}
                    onSubmit={handleOnSave}
                    open={open}
                    showEnvironmentAccessControls={showEnvironmentAccessControls}
                />
            </DialogContent>
        </DialogOverlay>
    );
};

export default CreateUserDialog;
