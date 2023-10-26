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

import { Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField, FormHelperText } from '@mui/material';
import { useState } from 'react';

const SaveQueryDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onSave: (data: { name: string }) => Promise<any>;
    isLoading?: boolean;
    error?: any;
}> = ({ open, onClose, onSave, isLoading = false, error = undefined }) => {
    const [name, setName] = useState('');

    const saveDisabled = name.trim() === '';

    const handleSave = () => {
        onSave({ name });
    };

    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            onClose={onClose}
            disableEscapeKeyDown
            TransitionProps={{
                onExited: () => {
                    setName('');
                },
            }}>
            <DialogTitle>Save Query</DialogTitle>
            <DialogContent>
                <TextField
                    variant='standard'
                    id='queryName'
                    value={name}
                    onChange={(e) => {
                        setName(e.target.value);
                    }}
                    label='Query Name'
                    fullWidth
                />
                {error ? (
                    <FormHelperText error>
                        An error ocurred while attempting to save this query. Please try again.
                    </FormHelperText>
                ) : null}
            </DialogContent>
            <DialogActions>
                <Button type='button' color='inherit' autoFocus onClick={onClose} disabled={isLoading}>
                    Cancel
                </Button>
                <Button type='button' color='primary' onClick={handleSave} disabled={saveDisabled || isLoading}>
                    Save
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default SaveQueryDialog;
