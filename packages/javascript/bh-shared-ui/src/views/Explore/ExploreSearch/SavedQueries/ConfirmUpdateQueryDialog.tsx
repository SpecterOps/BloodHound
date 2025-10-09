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
import { Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import { Button } from 'doodle-ui';
import { FC } from 'react';

const ConfirmUpdateQueryDialog: FC<{
    open: boolean;
    handleCancel: () => void;
    handleApply: () => void;
    dialogContent: string;
}> = ({ open, handleApply, handleCancel, dialogContent }) => {
    return (
        <Dialog open={open}>
            <DialogContent>
                <DialogTitle className='px-0'>Update Query</DialogTitle>
                <DialogContentText>{dialogContent}</DialogContentText>
                <DialogActions className='px-0'>
                    <Button variant='text' onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button variant='text' onClick={handleApply}>
                        Ok
                    </Button>
                </DialogActions>
            </DialogContent>
        </Dialog>
    );
};

export default ConfirmUpdateQueryDialog;
