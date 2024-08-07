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
import { Alert, AlertTitle, Box, Grid, TextField } from '@mui/material';
import React, { useState } from 'react';
import { PASSWD_REQS, testPassword } from '../../utils';
import { PutUserAuthSecretRequest } from 'js-client-library';

interface PasswordResetFormProps {
    onSubmit: (payload: PutUserAuthSecretRequest) => void;
    onCancel: () => void;
    loading?: boolean;
}

const PasswordResetForm: React.FC<PasswordResetFormProps> = ({ onSubmit, onCancel, loading = false }) => {
    const [currentPassword, setCurrentPassword] = useState('');
    const [currentPasswordTouched, setCurrentPasswordTouched] = useState(false);
    const [password, setPassword] = useState('');
    const [passwordTouched, setPasswordTouched] = useState(false);
    const [confirmPassword, setConfirmPassword] = useState('');
    const [confirmPasswordTouched, setConfirmPasswordTouched] = useState(false);

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();
        if (password === confirmPassword && password !== currentPassword && currentPassword !== '') {
            onSubmit({ secret: password, needsPasswordReset: false, currentSecret: currentPassword });
            return;
        }
    };

    const handleCancel: React.MouseEventHandler<HTMLButtonElement> = (e) => {
        e.preventDefault();
        onCancel();
    };

    return (
        <form onSubmit={handleSubmit}>
            <Grid container spacing={4} justifyContent='center'>
                <Grid item xs={12}>
                    <Alert severity='info'>
                        <AlertTitle>Your Account Password Has Expired</AlertTitle>
                        Please provide a new password for this account to continue.
                    </Alert>
                </Grid>
                {passwordTouched && !testPassword(password) && (
                    <Grid item xs={12}>
                        <Alert severity='error'>
                            <AlertTitle>Password Requirements</AlertTitle>
                            <ul>
                                {PASSWD_REQS.map((req, i) => (
                                    <li key={i}>{req}</li>
                                ))}
                            </ul>
                        </Alert>
                    </Grid>
                )}
                <Grid item xs={12}>
                    <TextField
                        id='currentPassword'
                        name='currentPassword'
                        label='Expired password'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={currentPassword}
                        error={currentPasswordTouched && currentPassword === ''}
                        onChange={(e) => setCurrentPassword(e.target.value)}
                        helperText={
                            currentPasswordTouched && currentPassword === '' ? 'Expired password is required' : null
                        }
                        onBlur={() => setCurrentPasswordTouched(true)}
                        autoFocus
                    />
                </Grid>
                <Grid item xs={12}>
                    <TextField
                        id='password'
                        name='password'
                        label='New Password'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={password}
                        error={passwordTouched && (!testPassword(password) || password === currentPassword)}
                        onChange={(e) => setPassword(e.target.value)}
                        helperText={
                            passwordTouched && password === currentPassword
                                ? 'New password must not match expired password'
                                : null
                        }
                        onBlur={() => setPasswordTouched(true)}
                    />
                </Grid>
                <Grid item xs={12}>
                    <TextField
                        id='confirmPassword'
                        name='confirmPassword'
                        label='New Password Confirmation'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        onBlur={() => setConfirmPasswordTouched(true)}
                        error={confirmPasswordTouched && password !== confirmPassword}
                        helperText={
                            confirmPasswordTouched && password !== confirmPassword
                                ? 'New password does not match'
                                : null
                        }
                    />
                </Grid>
                <Grid item xs={8}>
                    <Button size='large' type='submit' style={{ width: '100%' }} disabled={loading}>
                        {loading ? 'Resetting Password' : 'Reset Password'}
                    </Button>
                    <Box mt={2}>
                        <Button
                            variant='tertiary'
                            size='large'
                            type='button'
                            onClick={handleCancel}
                            style={{ width: '100%' }}
                            disabled={loading}>
                            Return to Login
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </form>
    );
};

export default PasswordResetForm;
