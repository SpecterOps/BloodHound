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

import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@bloodhoundenterprise/doodleui';
import { TextField } from '@mui/material';
import { FC, useState } from 'react';

type CertifyMembersConfirmDialogProps = {
    onConfirm: (withNote: boolean, note?: string) => void;
    onClose: () => void;
    open: boolean;
};

const CertifyMembersConfirmDialog: FC<CertifyMembersConfirmDialogProps> = ({ onConfirm, open, onClose }) => {
    const [note, setNote] = useState('');
    const handleChange = (event: any) => {
        setNote(event.target.value);
    };

    return (
        <Dialog open={open} modal={true}>
            <DialogContent
                className=''
                aria-describedby='textNote'
                maxWidth='xs'
                DialogOverlayProps={{ blurBackground: true }}>
                <DialogTitle className='flex'>Add note</DialogTitle>
                <TextField
                    multiline
                    rows={8}
                    maxRows={16}
                    fullWidth
                    id='textNote'
                    value={note}
                    onChange={handleChange}></TextField>
                <DialogActions>
                    <div className='flex w-full justify-between items-center'>
                        <Button onClick={onClose} variant='text'>
                            Cancel
                        </Button>
                        <div className='flex'>
                            <Button onClick={() => onConfirm(false)} variant='text'>
                                Skip Note
                            </Button>
                            <Button onClick={() => onConfirm(true, note)} variant='text'>
                                Save Note
                            </Button>
                        </div>
                    </div>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default CertifyMembersConfirmDialog;
