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

import { Button, Grid, TextField } from '@mui/material';
import React, { useState } from 'react';

interface LoginFormProps {
    onSubmit: (username: string, password: string) => void;
    onLoginViaSAML?: () => void;
    loading?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({ onSubmit, onLoginViaSAML, loading = false }) => {
    /* Hooks */
    const [username, setUsername] = useState('');

    const [password, setPassword] = useState('');

    /* Event Handlers */
    const handleLogin: React.FormEventHandler<HTMLFormElement> = async (e) => {
        e.preventDefault();
        onSubmit(username, password);
    };

    return (
        <form onSubmit={handleLogin}>
            <Grid container spacing={4} justifyContent='center'>
                <Grid item xs={12}>
                    <TextField
                        id='username'
                        name='username'
                        label='Email Address'
                        fullWidth
                        variant='outlined'
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        autoFocus
                    />
                </Grid>
                <Grid item xs={12}>
                    <TextField
                        id='password'
                        name='password'
                        label='Password'
                        type='password'
                        fullWidth
                        variant='outlined'
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
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
                        {loading ? 'LOGGING IN' : 'LOGIN'}
                    </Button>
                </Grid>
                {onLoginViaSAML !== undefined && (
                    <Grid item xs={8}>
                        <Button
                            variant='contained'
                            color='neutral'
                            size='large'
                            type='button'
                            fullWidth
                            disableElevation
                            onClick={onLoginViaSAML}
                            disabled={loading}>
                            LOGIN VIA SSO
                        </Button>
                    </Grid>
                )}
            </Grid>
        </form>
    );
};

export default LoginForm;
