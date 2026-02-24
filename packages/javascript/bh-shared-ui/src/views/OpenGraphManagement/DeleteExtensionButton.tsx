// Copyright 2026 Specter Ops, Inc.
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
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
} from '@bloodhoundenterprise/doodleui';
import { type Extension } from 'js-client-library';
import { FC, useState } from 'react';
import { cn, ConditionalTooltip } from '../..';
import { AppIcon } from '../../components';
import { useDeleteExtension } from '../../hooks';
import { DEFAULT_NOTIFICATION, ERROR_NOTIFICATION, useNotifications } from '../../providers';

const ConfirmDeleteExtensionDialog: FC<{
    extensionName: string;
    onAccept: () => void;
    onCancel: () => void;
    isDeleting: boolean;
}> = ({ extensionName, onAccept, onCancel, isDeleting }) => {
    const [inputValue, setInputValue] = useState('');

    const isConfirmDisabled = isDeleting || inputValue !== extensionName;

    return (
        <Dialog open>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle>Delete selected extension</DialogTitle>
                    <DialogDescription asChild>
                        <div>
                            <div>This will permanently delete the selected extension.</div>
                            <div className='font-bold'>Warning: This change is irreversible.</div>
                            <div className='mt-3 text-xs'>Input "{extensionName}" in order to proceed.</div>
                            <Input
                                aria-label={`Type ${extensionName} to confirm deletion`}
                                disabled={isDeleting}
                                onChange={(e) => setInputValue(e.target.value)}
                                placeholder={extensionName}
                                value={inputValue}
                            />
                        </div>
                    </DialogDescription>
                    <DialogActions>
                        <Button variant='text' onClick={onCancel} disabled={isDeleting}>
                            Cancel
                        </Button>
                        <Button variant='text' fontColor='primary' onClick={onAccept} disabled={isConfirmDisabled}>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export const DeleteExtensionButton: FC<{ extension: Extension }> = ({ extension }) => {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const deleteExtensionMutation = useDeleteExtension();
    const { addNotification } = useNotifications();

    const { id: extensionId, name: extensionName, is_builtin: isUndeletable } = extension;

    const handleButtonClick = () => setIsDialogOpen(true);
    const handleDialogClose = () => setIsDialogOpen(false);

    const handleDelete = () => {
        deleteExtensionMutation.mutate(extensionId, {
            onSuccess: () => {
                addNotification(
                    `Extension "${extensionName}" was deleted successfully!`,
                    'deleteExtensionSuccess',
                    DEFAULT_NOTIFICATION
                );
            },
            onError: () => {
                addNotification(
                    `Failed to delete extension "${extensionName}". Please try again.`,
                    'deleteExtensionError',
                    ERROR_NOTIFICATION
                );
            },
            onSettled: handleDialogClose,
        });
    };

    return (
        <div className='flex content-center justify-center'>
            <ConditionalTooltip condition={isUndeletable} tooltip='Built-in extensions cannot be deleted.'>
                <button
                    aria-label={`Delete ${extensionName}`}
                    className={cn('cursor-pointer', { 'opacity-50': isUndeletable })}
                    onClick={handleButtonClick}
                    disabled={isUndeletable}
                    title={isUndeletable ? 'This is a default extension and cannot be deleted.' : undefined}>
                    <AppIcon.Trash size={18} />
                </button>
            </ConditionalTooltip>

            {isDialogOpen && (
                <ConfirmDeleteExtensionDialog
                    extensionName={extensionName}
                    isDeleting={deleteExtensionMutation.isLoading}
                    onAccept={handleDelete}
                    onCancel={handleDialogClose}
                />
            )}
        </div>
    );
};
