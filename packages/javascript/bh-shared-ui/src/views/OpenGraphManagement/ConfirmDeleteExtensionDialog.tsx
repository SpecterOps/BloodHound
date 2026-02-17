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
import { FC, useState, useEffect } from 'react';

const ConfirmDeleteExtensionDialog: FC<{
    extensionName: string;
    onAccept: () => void;
    onCancel: () => void;
    open: boolean;
    isDeleting: boolean;
}> = ({ extensionName, onAccept, onCancel, open, isDeleting }) => {
    const [inputValue, setInputValue] = useState('');

    useEffect(() => {
        if (!open) {
            setInputValue('');
        }
    }, [open]);

    const isConfirmDisabled = isDeleting || inputValue !== extensionName;

    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent>
                    <DialogTitle>Delete selected extension</DialogTitle>
                    <DialogDescription asChild>
                        <div className='flex flex-col gap-4'>
                            <p>
                                This will permanently delete the selected extension.{' '}
                                <span className='font-bold'>Warning: This change is irreversible.</span>
                            </p>
                            <div>
                                <p className='mb-2'>Input "{extensionName}" in order to proceed.</p>
                                <Input
                                    value={inputValue}
                                    onChange={(e) => setInputValue(e.target.value)}
                                    placeholder={extensionName}
                                    disabled={isDeleting}
                                    className='w-full'
                                />
                            </div>
                        </div>
                    </DialogDescription>
                    <DialogActions>
                        <Button variant='tertiary' onClick={onCancel} disabled={isDeleting}>
                            Cancel
                        </Button>
                        <Button variant='primary' onClick={onAccept} disabled={isConfirmDisabled}>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default ConfirmDeleteExtensionDialog;