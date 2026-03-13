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
    Dialog,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogPortal,
    DialogTitle,
    DialogTrigger,
    VisuallyHidden,
} from 'doodle-ui';
import { CreateUserRequest } from 'js-client-library';
import React, { useState } from 'react';
import { usePermissions } from '../../hooks';
import { Permission } from '../../utils';
import CreateUserForm from '../CreateUserForm';

const CreateUserDialog: React.FC<{
    error: any;
    isLoading: boolean;
    onSave: (user: CreateUserRequest) => Promise<any>;
    showEnvironmentAccessControls: boolean;
}> = ({ error, isLoading, onSave, showEnvironmentAccessControls }) => {
    const handleOnSave = (user: CreateUserRequest) => {
        onSave(user)
            .then(() => setIsOpen(false))
            .catch((err) => console.error(err));
    };

    const [isOpen, setIsOpen] = useState(false);
    const { checkPermission } = usePermissions();

    const hasPermission = checkPermission(Permission.AUTH_MANAGE_USERS);

    return (
        <Dialog open={isOpen} onOpenChange={setIsOpen} data-testid='manage-users_create-user-dialog'>
            <DialogTrigger asChild>
                <Button disabled={!hasPermission} data-testid='manage-users_button-create-user'>
                    Create User
                </Button>
            </DialogTrigger>
            <DialogPortal>
                <DialogOverlay>
                    <DialogContent maxWidth='lg' className='!bg-transparent shadow-none max-h-screen overflow-y-auto'>
                        <VisuallyHidden asChild>
                            <DialogTitle>Create User</DialogTitle>
                        </VisuallyHidden>
                        <VisuallyHidden asChild>
                            <DialogDescription>Create User</DialogDescription>
                        </VisuallyHidden>
                        <CreateUserForm
                            error={error}
                            isLoading={isLoading}
                            onSubmit={handleOnSave}
                            showEnvironmentAccessControls={showEnvironmentAccessControls}
                        />
                    </DialogContent>
                </DialogOverlay>
            </DialogPortal>
        </Dialog>
    );
};

export default CreateUserDialog;
