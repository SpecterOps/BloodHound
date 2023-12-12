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
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
} from '@mui/material';
import React from 'react';

const Disable2FADialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onCancel: () => void;
    onSave: (secret: string) => void;
    error?: string;
    secret: string;
    onSecretChange: (e: any) => void;
    contentText: string;
}> = ({ open, onClose, onCancel, onSave, error, secret, onSecretChange, contentText }) => {
    const handleOnSave: React.FormEventHandler = (e) => {
        e.preventDefault();
        onSave(secret);
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth='sm' fullWidth>
            <DialogTitle>Disable Multi-Factor Authentication?</DialogTitle>
            <form onSubmit={handleOnSave}>
                <DialogContent>
                    <Alert severity='warning' style={{ marginBottom: '10px', alignItems: 'center' }}>
                        Disabling MFA increases the risk of unauthorized access. For optimal account security, we highly
                        recommend keeping MFA enabled.
                    </Alert>
                    <DialogContentText>{contentText}</DialogContentText>

                    <TextField
                        id='secret'
                        name='secret'
                        value={secret}
                        onChange={onSecretChange}
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
                        Disable Multi-Factor Authentication
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default Disable2FADialog;
