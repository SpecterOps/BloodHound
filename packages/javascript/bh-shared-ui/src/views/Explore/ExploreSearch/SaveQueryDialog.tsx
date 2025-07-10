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

import { Button } from '@bloodhoundenterprise/doodleui';
// import { Dialog, DialogActions, DialogContent, DialogTitle, FormHelperText, TextField } from '@mui/material';
import { useState } from 'react';

import {
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
} from '@bloodhoundenterprise/doodleui';

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
        <>
            <Dialog open={open} onOpenChange={onClose}>
                {/* <DialogTrigger asChild>
                    <Button variant='primary'>Default Dialog</Button>
                </DialogTrigger> */}
                <DialogPortal>
                    <DialogContent
                        DialogOverlayProps={{
                            blurBackground: false,
                        }}
                        maxWidth='sm'>
                        <DialogTitle>Save Query</DialogTitle>
                        {/* <VisuallyHidden>
                            something that we want to hide visually but still want in the DOM for accessibility
                        </VisuallyHidden> */}

                        <Input
                            type='text'
                            id='queryName'
                            value={name}
                            onChange={(e) => {
                                setName(e.target.value);
                            }}
                        />
                        {error ? (
                            <div>An error ocurred while attempting to save this query. Please try again.</div>
                        ) : null}

                        <DialogDescription>
                            To save your query to the Pre-built Query, add a name, optional description, and set sharing
                            permissions.
                        </DialogDescription>
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Submit</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>

            {/* OLD */}

            {/* <Dialog
                open={open}
                onClose={onClose}
                disableEscapeKeyDown
                TransitionProps={{
                    onExited: () => {
                        setName('');
                    },
                }}>
                <DialogTitle>Save Query</DialogTitle>
                <DialogContent maxWidth='sm'>
                    <Input
                        type='text'
                        id='queryName'
                        value={name}
                        onChange={(e) => {
                            setName(e.target.value);
                        }}
                    />
                    {error ? (
                        <FormHelperText error>
                            An error ocurred while attempting to save this query. Please try again.
                        </FormHelperText>
                    ) : null}
                </DialogContent>
                <DialogActions>
                    <Button type='button' variant={'tertiary'} onClick={onClose} disabled={isLoading}>
                        Cancel
                    </Button>
                    <Button type='button' onClick={handleSave} disabled={saveDisabled || isLoading}>
                        Save
                    </Button>
                </DialogActions>
            </Dialog> */}
        </>
    );
};

export default SaveQueryDialog;
