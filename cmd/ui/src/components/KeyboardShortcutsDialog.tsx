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
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
} from '@bloodhoundenterprise/doodleui';
import React, { useCallback } from 'react';

const KeyboardShortcutsDialog: React.FC<{
    open: boolean;
    title: string;
    text: string | JSX.Element;
    onCancel: () => void;
    onConfirm: () => void;
    challengeTxt?: string;
    isLoading?: boolean;
    error?: string;
}> = ({ open, onCancel, isLoading }) => {
    const handleClose = useCallback(() => {
        onCancel();
    }, [onCancel]);

    return (
        <Dialog open={open} data-testid='keyboard-shortcuts-dialog'>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle className='text-lg'>Keyboard Shortcuts</DialogTitle>
                    <DialogDescription className='text-lg'>See keyboard shortcuts</DialogDescription>
                    <DialogDescription className='text-lg'>[[Helpful table of shortcuts]]</DialogDescription>

                    <DialogActions>
                        <Button
                            variant='tertiary'
                            onClick={handleClose}
                            disabled={isLoading}
                            data-testid='keyboard-shortcuts-dialog_button-close'>
                            Okay
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default KeyboardShortcutsDialog;
