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
import { Button } from '../../Button';
import { Dialog, DialogActions, DialogContent, DialogDescription, DialogPortal, DialogTitle } from '../../Dialog';

import React, { useCallback } from 'react';

const PinDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onConfirm: (activeId: string | number, overId: string | number) => void;
    pinDialogState: {
        action: 'pin' | 'unpin' | null;
        activeId: string | number;
        overId: string | number;
        label?: string;
    };
}> = ({ open, onClose, onConfirm, pinDialogState }) => {
    const handleClose = useCallback(() => {
        onClose();
    }, [onClose]);

    const { action, activeId, overId } = pinDialogState;

    const handleConfirm = useCallback(
        (activeId: string | number, overId: string | number) => {
            onConfirm(activeId, overId);
            onClose();
        },
        [onConfirm, onClose]
    );

    return (
        <Dialog open={open} data-testid='pin-dialog'>
            <DialogPortal>
                <DialogContent onEscapeKeyDown={handleClose}>
                    <DialogTitle>{action === 'unpin' ? 'Unpin' : 'Pin'} Column</DialogTitle>
                    {pinDialogState.label && (
                        <DialogDescription>
                            The {pinDialogState.label} column will {action === 'pin' ? 'now' : 'no longer'} be pinned.
                        </DialogDescription>
                    )}
                    <DialogActions className='flex justify-end gap-4'>
                        <Button variant='secondary' onClick={handleClose}>
                            Cancel
                        </Button>
                        <Button onClick={() => handleConfirm(activeId, overId)}>Confirm</Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default PinDialog;
