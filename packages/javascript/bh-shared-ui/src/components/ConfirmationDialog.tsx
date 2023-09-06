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
    DialogContentText,
    DialogTitle,
    FormHelperText,
} from '@mui/material';
import React from 'react';

const ConfirmationDialog: React.FC<{
    open: boolean;
    title: string;
    text: string;
    onClose: (response: boolean) => void;
    isLoading?: boolean;
    error?: string;
}> = ({ open, title, text, onClose, isLoading, error }) => {
    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            onClose={() => onClose(false)}
            PaperProps={{
                // @ts-ignore
                'data-testid': 'confirmation-dialog',
            }}>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <DialogContentText>{text}</DialogContentText>
            </DialogContent>
            <DialogActions>
                {error && (
                    <FormHelperText error style={{ margin: 0 }}>
                        {error}
                    </FormHelperText>
                )}
                <Button
                    autoFocus
                    color='inherit'
                    onClick={() => onClose(false)}
                    disabled={isLoading}
                    data-testid='confirmation-dialog_button-no'>
                    {'No'}
                </Button>
                <Button
                    color='primary'
                    onClick={() => onClose(true)}
                    disabled={isLoading}
                    data-testid='confirmation-dialog_button-yes'>
                    {'Yes'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ConfirmationDialog;
