// Copyright 2025 Specter Ops, Inc.
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
import ShortcutSection from './ShortcutSection';
import { EXPLORE_SHORTCUTS, GLOBAL_SHORTCUTS, POSTURE_PAGE_SHORTCUTS, type ShortCutsMap } from './shortcuts';

const parseShortcutEntries = (shortCutMap: ShortCutsMap) => {
    const entries = Object.entries(shortCutMap);
    const [heading, bindings] = entries[0];

    return { heading, bindings };
};

const KeyboardShortcutsDialog: React.FC<{
    open: boolean;
    onClose: () => void;
}> = ({ open, onClose }) => {
    const handleClose = useCallback(() => {
        onClose();
    }, [onClose]);

    return (
        <Dialog open={open} data-testid='keyboard-shortcuts-dialog'>
            <DialogPortal>
                <DialogContent className='h-3/4 grid-rows[50px, 1fr, 50px] max-w-[900px]' onEscapeKeyDown={handleClose}>
                    <DialogTitle className='text-lg flex justify-center'>Keyboard Shortcuts</DialogTitle>
                    <DialogDescription hidden>Keyboard Shortcuts List</DialogDescription>
                    <hr />
                    <div className='overflow-auto grid grid-cols-2 gap-3 justify-center'>
                        <div>
                            <ShortcutSection {...parseShortcutEntries(GLOBAL_SHORTCUTS)} />
                            <ShortcutSection {...parseShortcutEntries(POSTURE_PAGE_SHORTCUTS)} />
                        </div>
                        <ShortcutSection {...parseShortcutEntries(EXPLORE_SHORTCUTS)} />
                    </div>
                    <hr />
                    <DialogActions>
                        <Button onClick={handleClose} data-testid='keyboard-shortcuts-dialog_button-close'>
                            Close
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default KeyboardShortcutsDialog;
