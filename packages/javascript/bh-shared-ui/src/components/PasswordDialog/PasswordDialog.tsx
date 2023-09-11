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
    AlertTitle,
    Button,
    Checkbox,
    Dialog,
    DialogTitle,
    FormControlLabel,
    Grid,
    TextField,
} from '@mui/material';
import { DialogActions, DialogContent } from '@mui/material';
import React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { PASSWD_REQS, passwordRegex } from '../../utils';

const passwdReqsList = PASSWD_REQS.map((req, i) => <li key={i}>{req}</li>);

type ChangePasswordFormInputs = {
    password: string;
    confirmPassword: string;
    needsPasswordReset: boolean;
};

const PasswordDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    userId: string;
    showNeedsPasswordReset?: boolean;
    onSave: ({
        userId,
        secret,
        needsPasswordReset,
    }: {
        userId: string;
        secret: string;
        needsPasswordReset: boolean;
    }) => void;
}> = ({ open, userId, onClose, showNeedsPasswordReset = false, onSave }) => {
    const {
        control,
        handleSubmit,
        getValues,
        formState: { errors },
        reset,
    } = useForm<ChangePasswordFormInputs>({
        defaultValues: {
            password: '',
            confirmPassword: '',
            needsPasswordReset: false,
        },
    });

    React.useEffect(() => {
        if (open) reset();
    }, [open, reset]);

    const handleOnSave = (data: { password: string; confirmPassword: string; needsPasswordReset: boolean }) => {
        return onSave({
            userId: userId,
            secret: data.password,
            needsPasswordReset: Boolean(data.needsPasswordReset),
        });
    };

    return (
        <Dialog
            open={open}
            fullWidth={true}
            maxWidth={'xs'}
            onClose={(event, reason) => {
                if (reason !== 'backdropClick' && reason !== 'escapeKeyDown') {
                    onClose();
                }
            }}
            PaperProps={{
                // @ts-ignore
                'data-testid': 'password-dialog',
            }}>
            <DialogTitle>{'Change Password'}</DialogTitle>
            <form autoComplete='off' onSubmit={handleSubmit(handleOnSave)}>
                <DialogContent>
                    <Grid container spacing={2}>
                        {!!errors.password && (
                            <Grid item xs={12}>
                                <Alert severity='error'>
                                    <AlertTitle>Password Requirements</AlertTitle>
                                    <ul>{passwdReqsList}</ul>
                                </Alert>
                            </Grid>
                        )}
                        <Grid item xs={12}>
                            <Controller
                                name='password'
                                control={control}
                                rules={{
                                    required: 'Password is required',
                                    pattern: passwordRegex,
                                }}
                                render={({ field }) => (
                                    <TextField
                                        {...field}
                                        variant='standard'
                                        id='password'
                                        label='Password'
                                        type='password'
                                        fullWidth
                                        error={!!errors.password}
                                        data-testid='password-dialog_input-password'
                                    />
                                )}
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <Controller
                                name='confirmPassword'
                                control={control}
                                rules={{
                                    required: 'Confirmation password is required',
                                    validate: (value) => getValues('password') === value || 'Password does not match',
                                }}
                                render={({ field }) => (
                                    <TextField
                                        {...field}
                                        variant='standard'
                                        id='confirmPassword'
                                        label='Confirmation Password'
                                        type='password'
                                        fullWidth
                                        error={!!errors.confirmPassword}
                                        helperText={errors.confirmPassword?.message}
                                        data-testid='password-dialog_input-password-confirmation'
                                    />
                                )}
                            />
                        </Grid>
                        {showNeedsPasswordReset && (
                            <Grid item xs={12}>
                                <Controller
                                    name='needsPasswordReset'
                                    control={control}
                                    render={({ field }) => (
                                        <FormControlLabel
                                            control={
                                                <Checkbox
                                                    {...field}
                                                    onChange={(e) => field.onChange(e.target.checked)}
                                                    color='primary'
                                                    data-testid='password-dialog_checkbox-needs-password-reset'
                                                />
                                            }
                                            label='Force Password Reset?'
                                        />
                                    )}
                                />
                            </Grid>
                        )}
                    </Grid>
                </DialogContent>

                <DialogActions>
                    <Button
                        autoFocus={true}
                        color='inherit'
                        onClick={onClose}
                        data-testid='password-dialog_button-close'>
                        Cancel
                    </Button>
                    <Button autoFocus={false} color='primary' type='submit' data-testid='password-dialog_button-save'>
                        Save
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default PasswordDialog;
