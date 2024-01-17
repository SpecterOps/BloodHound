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
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Grid,
    TextField,
} from '@mui/material';
import withStyles from '@mui/styles/withStyles';
import React, { useState } from 'react';
import SetupKeyDialog from '../SetupKeyDialog';

const BarcodeButton = withStyles((theme) => ({
    root: {
        backgroundColor: '#fff',
        padding: theme.spacing(1),
        borderRadius: theme.shape.borderRadius,
        '&:hover': {
            backgroundColor: '#fff',
        },
    },
    label: {
        padding: 0,
    },
}))(Button);

const Enable2FADialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onCancel: () => void;
    onSavePassword: (password: string) => Promise<void>;
    onSaveOTP: (OTP: string) => Promise<void>;
    onSave: () => void;
    QRCode: string;
    TOTPSecret: string;
    error?: string;
}> = ({ open, onClose, onCancel, onSavePassword, onSaveOTP, onSave, QRCode, TOTPSecret, error }) => {
    const [secret, setSecret] = useState('');
    const [OTP, setOTP] = useState('');
    const [secretAccepted, setSecretAccepted] = useState(false);
    const [OTPAccepted, setOTPAccepted] = useState(false);
    const [setupKeyDialogOpen, setSetupKeyDialogOpen] = useState(false);

    const handleOnClose = () => {
        onClose();
    };

    const handleOnExited = () => {
        setSecret('');
        setOTP('');
        setSecretAccepted(false);
        setOTPAccepted(false);
    };

    const handleOnSave: React.FormEventHandler = async (e) => {
        e.preventDefault();

        if (!secretAccepted) {
            await onSavePassword(secret)
                .then(() => {
                    setSecretAccepted(true);
                })
                .catch(() => {});
        } else if (!OTPAccepted) {
            await onSaveOTP(OTP)
                .then(() => {
                    setOTPAccepted(true);
                })
                .catch(() => {});
        } else {
            onSave();
        }
    };

    return (
        <>
            <Dialog
                open={open}
                onClose={handleOnClose}
                maxWidth='sm'
                fullWidth
                TransitionProps={{ onExited: handleOnExited }}
                PaperProps={{
                    // @ts-ignore
                    'data-testid': 'enable-2fa-dialog',
                }}>
                <DialogTitle>
                    {!(secretAccepted && OTPAccepted)
                        ? 'Configure Multi-Factor Authentication'
                        : 'Multi-Factor Authentication Configured Successfully'}
                </DialogTitle>
                <form onSubmit={handleOnSave}>
                    <DialogContent>
                        {!secretAccepted && !OTPAccepted && (
                            <>
                                <DialogContentText>
                                    To set up multi-factor authentication, you'll need to download an authenticator app.
                                </DialogContentText>
                                <DialogContentText>To get started, first enter your password.</DialogContentText>
                                <TextField
                                    name='secret'
                                    type='password'
                                    value={secret}
                                    onChange={(e) => {
                                        setSecret(e.target.value);
                                    }}
                                    aria-label='Password'
                                    label='Password'
                                    variant='outlined'
                                    margin='dense'
                                    fullWidth
                                    autoFocus
                                    error={!!error}
                                    helperText={error}
                                    data-testid='enable-2fa-dialog_input-password'
                                />
                            </>
                        )}

                        {secretAccepted && !OTPAccepted && (
                            <>
                                <Grid container spacing={2}>
                                    <Grid item xs={7}>
                                        <DialogContentText>
                                            <strong>Step 1:</strong> Visit your phone's App Store to download and
                                            install an authenticator app like Google Authenticator or Authy, then follow
                                            the app's instructions to set up an account with them.
                                        </DialogContentText>
                                        <DialogContentText>
                                            <strong>Step 2:</strong> Use your authenticator app to scan the barcode and
                                            enter the 6-digit verification code. Alternatively, click the barcode to
                                            reveal a setup key for manual entry.
                                        </DialogContentText>
                                        <TextField
                                            name='otp'
                                            type='text'
                                            value={OTP}
                                            onChange={(e) => {
                                                setOTP(e.target.value);
                                            }}
                                            aria-label='One-Time Password'
                                            label='One-Time Password'
                                            variant='outlined'
                                            margin='dense'
                                            fullWidth
                                            autoFocus
                                            error={!!error}
                                            helperText={error}
                                            data-testid='enable-2fa-dialog_input-one-time-password'
                                        />
                                    </Grid>
                                    <Grid item xs={5}>
                                        <Box height={200} width={200}>
                                            <BarcodeButton
                                                type='button'
                                                variant='outlined'
                                                color='primary'
                                                disableElevation
                                                disableRipple
                                                disableFocusRipple
                                                disableTouchRipple
                                                title='Click to reveal setup key'
                                                onClick={() => {
                                                    setSetupKeyDialogOpen(true);
                                                }}
                                                data-testid='enable-2fa-dialog_button-barcode'>
                                                <img
                                                    src={QRCode}
                                                    height='100%'
                                                    alt='QR Code for Configuring Multi-Factor Authentication'></img>
                                            </BarcodeButton>
                                        </Box>
                                    </Grid>
                                </Grid>
                            </>
                        )}

                        {secretAccepted && OTPAccepted && (
                            <>
                                <DialogContentText>
                                    Next time you log in, you'll need to use your password and authentication code.
                                </DialogContentText>
                                <DialogContentText>
                                    If you lose your authentication code, you'll need to contact your account's
                                    Administrator to reset your password and then go through the multi-factor
                                    authentication setup again.
                                </DialogContentText>
                            </>
                        )}
                    </DialogContent>
                    <DialogActions>
                        {!(secretAccepted && OTPAccepted) ? (
                            <>
                                <Button color='inherit' onClick={onCancel} data-testid='enable-2fa-dialog_button-close'>
                                    Cancel
                                </Button>
                                <Button color='primary' type='submit' data-testid='enable-2fa-dialog_button-next'>
                                    Next
                                </Button>
                            </>
                        ) : (
                            <>
                                <Button
                                    color='primary'
                                    onClick={handleOnClose}
                                    data-testid='enable-2fa-dialog_button-close'>
                                    Close
                                </Button>
                            </>
                        )}
                    </DialogActions>
                </form>
            </Dialog>

            <SetupKeyDialog
                open={setupKeyDialogOpen}
                onClose={() => setSetupKeyDialogOpen(false)}
                setupKey={TOTPSecret}
            />
        </>
    );
};

export default Enable2FADialog;
