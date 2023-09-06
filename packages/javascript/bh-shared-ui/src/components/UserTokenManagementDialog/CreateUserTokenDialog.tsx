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

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, TextField } from '@mui/material';
import React from 'react';
import { Controller, useForm } from 'react-hook-form';

const CreateUserTokenDialog: React.FC<{
    open: boolean;
    onCancel: () => void;
    onSubmit: (newToken: { token_name: string }) => void;
}> = ({ open, onCancel, onSubmit }) => {
    const {
        control,
        handleSubmit,
        formState: { errors },
    } = useForm({
        defaultValues: {
            token_name: '',
        },
    });

    const handleCancel: React.MouseEventHandler<HTMLButtonElement> = (e) => {
        e.preventDefault();
        onCancel();
    };

    return (
        <Dialog
            aria-labelledby='createUserTokenDialogTitle'
            open={open}
            fullWidth={true}
            maxWidth={'sm'}
            onClose={onCancel}
            PaperProps={{
                // @ts-ignore
                'data-testid': 'create-user-token-dialog',
            }}>
            <DialogTitle id='createUserTokenDialogTitle'>Create User Token</DialogTitle>
            <form autoComplete={'off'} onSubmit={handleSubmit(onSubmit)}>
                <DialogContent>
                    <Grid container spacing={1}>
                        <Controller
                            name={'token_name'}
                            control={control}
                            defaultValue={''}
                            rules={{ required: 'Token name is required' }}
                            render={({ field }) => (
                                <TextField
                                    {...field}
                                    variant='standard'
                                    label={'Token Name'}
                                    fullWidth
                                    error={!!errors.token_name}
                                    helperText={errors.token_name?.message}
                                    data-testid='create-user-token-dialog_input-token-name'
                                />
                            )}
                        />
                    </Grid>
                </DialogContent>
                <DialogActions>
                    <Button
                        autoFocus
                        color='inherit'
                        onClick={handleCancel}
                        data-testid='create-user-token-dialog_button-close'>
                        Cancel
                    </Button>
                    <Button color='primary' type='submit' data-testid='create-user-token-dialog_button-save'>
                        Save
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default CreateUserTokenDialog;
