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

const SHORTCUTS = [
    {
        Global: [
            ['H', 'View keyboard shortcuts dialog'],
            ['M', 'Toggle Dark Mode'],
            ['<integer>', 'Navigate sidebar pages'],
        ],
    },
    {
        'Explore Page': [
            ['S', 'Jump to Node Search'],
            ['P', 'Jump to Pathfinding'],
            ['C', 'Focus Cypher Query Editor'],
            ['Shift + S', 'Save Current Query'],
            ['R', 'Run Current Cypher Query'],
            ['/', 'Search Current Nodes'],
            ['T', 'Toggle Table View'],
            ['I', 'TODO: Toggle Node Info Panel'],
            ['Shift + G', 'Reset Graph View'],
        ],
    },
];

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
                <DialogContent className='h-1/2 grid-rows[50px, 1fr, 50px] max-w-[95%]' onEscapeKeyDown={handleClose}>
                    <DialogTitle className='text-lg flex justify-center'>Keyboard Shortcuts</DialogTitle>
                    <DialogDescription className='overflow-auto grid grid-cols-2'>
                        {SHORTCUTS.map((shortCutMap) => {
                            const entries = Object.entries(shortCutMap);
                            const [heading, bindings] = entries[0];

                            return (
                                <div className='mb-5 h-full' key={heading}>
                                    <div className='font-bold text-center p-2'>{heading}</div>
                                    <div className='flex flex-col gap-2 text-sm'>
                                        {bindings.map((binding: string[]) => (
                                            <div key={`${binding[0]}-${heading}`} className='flex gap-2'>
                                                <div className='w-1/2 text-right p-2'> {binding[1]}</div>
                                                <div className='w-1/2 border-2 rounded-md p-2 text-center'>
                                                    Alt/Option + {binding[0]}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            );
                        })}
                    </DialogDescription>
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
