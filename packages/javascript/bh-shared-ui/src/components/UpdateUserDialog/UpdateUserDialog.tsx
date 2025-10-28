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
    Dialog,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogPortal,
    DialogTitle,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
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
    showEnvironmentAccessControls?: boolean;
    userId: string;
}> = ({ error, hasSelectedSelf, isLoading, onClose, onSave, open, showEnvironmentAccessControls, userId }) => {
    const handleOnSave = (user: UpdateUserRequestForm) => {
        let parsedSSOProviderId: number | undefined = undefined;
        if (user.sso_provider_id) {
            parsedSSOProviderId = parseInt(user.sso_provider_id);
        }

        onSave({
            ...user,
            sso_provider_id: parsedSSOProviderId,
        })
            .then(() => {
                onClose();
            })
            .catch((err) => console.error(err));
    };

    return (
        <Dialog open={open} onOpenChange={onClose} data-testid='manage-users_update-user-dialog'>
            <DialogPortal>
                <DialogOverlay>
                    <DialogContent
                        maxWidth='lg'
                        className='!bg-transparent !pointer-events-auto overflow-y-auto max-h-screen'
                        data-testid='update-user-dialog'>
                        <VisuallyHidden asChild>
                            <>
                                <DialogTitle />
                                <DialogDescription />
                            </>
                        </VisuallyHidden>
                        <UpdateUserForm
                            error={error}
                            isLoading={isLoading}
                            onSubmit={handleOnSave}
                            hasSelectedSelf={hasSelectedSelf}
                            showEnvironmentAccessControls={showEnvironmentAccessControls}
                            userId={userId!}
                        />
                    </DialogContent>
                </DialogOverlay>
            </DialogPortal>
        </Dialog>
    );
};

export default UpdateUserDialog;
