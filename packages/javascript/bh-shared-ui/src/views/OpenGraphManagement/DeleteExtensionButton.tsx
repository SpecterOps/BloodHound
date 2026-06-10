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
} from 'doodle-ui';
import { type Extension } from 'js-client-library';
import { FC, useCallback, useState } from 'react';
import { AppIcon, ConditionalTooltip } from '../../components';
import { usePermissions } from '../../hooks';
import { Permission } from '../../utils';
import { cn } from '../../utils/theme';

export const ConfirmDeleteExtensionDialog: FC<{
    open: boolean;
    extensionName: string;
    onAccept: () => void;
    onCancel: () => void;
    isDeleting: boolean;
}> = ({ open, extensionName, onAccept, onCancel, isDeleting }) => {
    const [inputValue, setInputValue] = useState('');

    const handleCancel = useCallback(() => {
        onCancel();
        setTimeout(() => {
            setInputValue('');
        }, 1000);
    }, [onCancel]);

    const handleAccept = useCallback(() => {
        onAccept();
        setTimeout(() => {
            setInputValue('');
        }, 1000);
    }, [onAccept]);

    const isConfirmDisabled = isDeleting || inputValue !== extensionName;

    return (
        <Dialog open={open}>
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
                        <Button variant='text' onClick={handleCancel} disabled={isDeleting}>
                            Cancel
                        </Button>
                        <Button variant='text' fontColor='primary' onClick={handleAccept} disabled={isConfirmDisabled}>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export const DeleteExtensionButton: FC<{ extension: Extension; onDeleteClick: (extension: Extension) => void }> = ({
    extension,
    onDeleteClick,
}) => {
    const { checkPermission } = usePermissions();
    const hasDeletePermission = checkPermission(Permission.OPENGRAPH_WRITE);

    const { name: extensionName, is_builtin: isUndeletable } = extension;

    return (
        <div className='flex content-center justify-center'>
            <ConditionalTooltip
                condition={isUndeletable || !hasDeletePermission}
                tooltip={
                    isUndeletable
                        ? 'Built-in extensions cannot be deleted.'
                        : 'You do not have permission to delete this extension.'
                }>
                <button
                    aria-label={`Delete ${extensionName}`}
                    className={cn({
                        'cursor-pointer': !isUndeletable && hasDeletePermission,
                        'opacity-50 cursor-not-allowed': isUndeletable || !hasDeletePermission,
                    })}
                    onClick={() => onDeleteClick(extension)}
                    disabled={isUndeletable || !hasDeletePermission}>
                    <AppIcon.Trash size={18} />
                </button>
            </ConditionalTooltip>
        </div>
    );
};
