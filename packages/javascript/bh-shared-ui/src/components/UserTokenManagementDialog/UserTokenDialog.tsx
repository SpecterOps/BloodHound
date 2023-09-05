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
import { NewAuthToken } from 'js-client-library';
import React from 'react';

const UserTokenDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    token?: NewAuthToken;
}> = ({ open, token, onClose }) => {
    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            PaperProps={{
                //@ts-ignore
                'data-testid': 'user-token-dialog',
            }}>
            <DialogTitle>Auth Token</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Below is the new authentication token. Make sure to save this key, it will not be displayed again.
                </DialogContentText>
                <DialogContentText variant={'body2'}>
                    Key: {token?.key}
                    <br />
                    ID: {token?.id}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button color='inherit' onClick={onClose} data-testid='user-token-dialog_button-close'>
                    Close
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default UserTokenDialog;
