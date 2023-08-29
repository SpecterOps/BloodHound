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

import { Box, Button, Grid, TextField, Typography } from '@mui/material';
import React, { useState } from 'react';

interface OneTimePasscodeFormProps {
    onSubmit: (otp: string) => void;
    onCancel: () => void;
    loading?: boolean;
}

const OneTimePasscodeForm: React.FC<OneTimePasscodeFormProps> = ({ onSubmit, onCancel, loading = false }) => {
    /* Hooks */

    const [otp, setOtp] = useState('');

    /* Event Handlers */

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();
        onSubmit(otp);
    };

    /* Implementation */

    return (
        <form onSubmit={handleSubmit}>
            <Grid container spacing={4} justifyContent='center'>
                <Grid item xs={12}>
                    <Typography variant='body1'>
                        <strong>Multi-Factor Authentication Enabled</strong>
                    </Typography>
                    <Typography variant='body1'>Provide the 6 digit code from your authenticator app.</Typography>
                </Grid>
                <Grid item xs={12}>
                    <TextField
                        id='otp'
                        name='otp'
                        label='6-Digit Code'
                        type='text'
                        fullWidth
                        variant='outlined'
                        value={otp}
                        onChange={(e) => setOtp(e.target.value)}
                        autoFocus
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
                        {loading ? 'Checking Code' : 'Check Code'}
                    </Button>
                    <Box mt={2}>
                        <Button
                            onClick={onCancel}
                            color='inherit'
                            variant='contained'
                            size='large'
                            type='button'
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

export default OneTimePasscodeForm;
