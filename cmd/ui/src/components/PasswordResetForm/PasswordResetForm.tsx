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

import { Alert, AlertTitle, Box, Button, Grid, TextField } from '@mui/material';
import React, { useState } from 'react';
import { PASSWD_REQS, testPassword } from 'bh-shared-ui';

interface PasswordResetFormProps {
    onSubmit: (password: string) => void;
    onCancel: () => void;
    loading?: boolean;
}

const PasswordResetForm: React.FC<PasswordResetFormProps> = ({ onSubmit, onCancel, loading = false }) => {
    const [password, setPassword] = useState('');
    const [passwordTouched, setPasswordTouched] = useState(false);
    const [confirmPassword, setConfirmPassword] = useState('');
    const [confirmPasswordTouched, setConfirmPasswordTouched] = useState(false);

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();
        if (password === confirmPassword) {
            onSubmit(password);
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
                        id='password'
                        name='password'
                        label='Password'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={password}
                        error={passwordTouched && !testPassword(password)}
                        onChange={(e) => setPassword(e.target.value)}
                        onBlur={() => setPasswordTouched(true)}
                        autoFocus
                    />
                </Grid>
                <Grid item xs={12}>
                    <TextField
                        id='confirmPassword'
                        name='confirmPassword'
                        label='Confirm Password'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        onBlur={() => setConfirmPasswordTouched(true)}
                        error={confirmPasswordTouched && password !== confirmPassword}
                        helperText={
                            confirmPasswordTouched && password !== confirmPassword ? 'Password does not match' : null
                        }
                    />
                </Grid>
                <Grid item xs={8}>
                    <Button
                        variant='contained'
                        color='primary'
                        size='large'
                        type='submit'
                        fullWidth
                        disableElevation
                        disabled={loading}>
                        {loading ? 'Resetting Password' : 'Reset Password'}
                    </Button>
                    <Box mt={2}>
                        <Button
                            variant='contained'
                            color='inherit'
                            size='large'
                            type='button'
                            onClick={handleCancel}
                            fullWidth
                            disableElevation
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
