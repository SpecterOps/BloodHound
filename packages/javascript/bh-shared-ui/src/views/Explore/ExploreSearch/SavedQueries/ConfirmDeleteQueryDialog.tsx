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
import { Button } from '@bloodhoundenterprise/doodleui';
import { Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import { FC } from 'react';

interface ConfirmDeleteQueryDialogProps {
    open: boolean;
    queryId?: number;
    handleClose: () => void;
    deleteHandler?: (id: number) => void;
}

const ConfirmDeleteQueryDialog: FC<ConfirmDeleteQueryDialogProps> = ({ open, queryId, handleClose, deleteHandler }) => {
    return (
        <Dialog open={open} onClose={handleClose} maxWidth={'xs'} fullWidth>
            <DialogTitle>Delete Query</DialogTitle>
            <DialogContent>
                <DialogContentText>Are you sure you want to delete this query?</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={handleClose}>
                    Cancel
                </Button>
                <Button
                    onClick={() => {
                        if (deleteHandler) deleteHandler(queryId!);
                        handleClose();
                    }}
                    color='primary'
                    autoFocus>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ConfirmDeleteQueryDialog;
