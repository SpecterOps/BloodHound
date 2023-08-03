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

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, TextField } from '@mui/material';
import React, { useState } from 'react';

const Disable2FADialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onCancel: () => void;
    onSave: (secret: string) => void;
    error?: string;
}> = ({ open, onClose, onCancel, onSave, error }) => {
    const [secret, setSecret] = useState('');

    const handleOnClose = () => {
        onClose();
    };

    const handleOnSave: React.FormEventHandler = (e) => {
        e.preventDefault();
        onSave(secret);
    };

    return (
        <Dialog open={open} onClose={handleOnClose} maxWidth='sm' fullWidth>
            <DialogTitle>Disable Two-Factor Authentication?</DialogTitle>
            <form onSubmit={handleOnSave}>
                <DialogContent>
                    <DialogContentText>
                        To stop using two-factor authentication, please enter your password for security purposes.
                    </DialogContentText>

                    <TextField
                        id='secret'
                        name='secret'
                        value={secret}
                        onChange={(e) => {
                            setSecret(e.target.value);
                        }}
                        type='password'
                        label='Password'
                        variant='outlined'
                        margin='dense'
                        fullWidth
                        autoFocus
                        error={!!error}
                        helperText={error}
                    />
                </DialogContent>
                <DialogActions>
                    <Button color='inherit' onClick={onCancel}>
                        Cancel
                    </Button>
                    <Button color='primary' type='submit'>
                        Disable Two-Factor Authentication
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default Disable2FADialog;
