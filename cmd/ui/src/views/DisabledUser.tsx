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

import { Alert, AlertTitle, Box, Button, Grid } from '@mui/material';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { logout } from 'src/ducks/auth/authSlice';
import { ROUTE_LOGIN } from 'src/ducks/global/routes';
import { useAppDispatch } from 'src/store';
import LoginPage from 'src/components/LoginPage';

const DisabledUser: React.FC = () => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    return (
        <LoginPage>
            <Grid container spacing={4} justifyContent='center'>
                <Grid item xs={12}>
                    <Alert severity='warning'>
                        <AlertTitle>Your Account Has Been Disabled</AlertTitle>
                        Please contact your system administrator for assistance.
                    </Alert>
                </Grid>
                <Grid item xs={12}>
                    <Box mt={2}>
                        <Button
                            onClick={() => {
                                dispatch(logout());
                                navigate(ROUTE_LOGIN);
                            }}
                            data-testid='disabled-user-back-to-login'
                            color='inherit'
                            variant='contained'
                            size='large'
                            type='button'
                            fullWidth
                            disableElevation>
                            Back to Login
                        </Button>
                    </Box>
                </Grid>
            </Grid>
        </LoginPage>
    );
};

export default DisabledUser;
