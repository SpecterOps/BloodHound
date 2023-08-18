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
import { AuthToken } from 'js-client-library';
import React from 'react';

const TokenRevokeDialog: React.FC<{
    open: boolean;
    onCancel: () => void;
    onConfirm: () => void;
    token?: AuthToken;
}> = ({ open, onCancel, onConfirm, token }) => {
    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'xs'}
            PaperProps={{
                //@ts-ignore
                'data-testid': 'token-revoke-dialog',
            }}>
            <DialogTitle>Revoke "{token?.name}" Auth Token</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Are you sure you want to revoke the permanent token? Applications using this token will be unable to
                    contact the API.
                </DialogContentText>
                <DialogActions>
                    <Button color='inherit' onClick={onCancel} data-testid='token-revoke-dialog_button-close'>
                        Cancel
                    </Button>
                    <Button onClick={onConfirm} data-testid='token-revoke-dialog_button-save'>
                        Confirm
                    </Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default TokenRevokeDialog;
