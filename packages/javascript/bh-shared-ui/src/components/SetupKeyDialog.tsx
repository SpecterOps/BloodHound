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

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import React from 'react';

const SetupKeyDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    setupKey: string;
}> = ({ open, onClose, setupKey }) => {
    return (
        <Dialog
            open={open}
            onClose={onClose}
            PaperProps={{
                // @ts-ignore
                'data-testid': 'setup-key-dialog',
            }}>
            <DialogTitle>Multi-Factor Authentication Setup Key</DialogTitle>
            <DialogContent>
                <DialogContentText data-testid='setup-key'>{setupKey}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button
                    type='button'
                    color='inherit'
                    autoFocus
                    onClick={onClose}
                    data-testid={'setup-key-dialog_button-close'}>
                    Close
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default SetupKeyDialog;
