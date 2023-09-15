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

import { useState } from 'react';
import { Box, Button, CircularProgress, Grid, Switch, Typography } from '@mui/material';
import { useMutation, useQuery } from 'react-query';
import { Alert, AlertTitle } from '@mui/material';
import { useNotifications } from '../../providers';
import { apiClient, getUsername } from '../../utils';
import {
    Disable2FADialog,
    Enable2FADialog,
    PageWithTitle,
    PasswordDialog,
    TextWithFallback,
    UserTokenManagementDialog,
} from '../../components';

const UserProfile = () => {
    const { addNotification } = useNotifications();
    const [changePasswordDialogOpen, setChangePasswordDialogOpen] = useState(false);
    const [userTokenManagementDialogOpen, setUserTokenManagementDialogOpen] = useState(false);
    const [enable2FADialogOpen, setEnable2FADialogOpen] = useState(false);
    const [disable2FADialogOpen, setDisable2FADialogOpen] = useState(false);
    const [TOTPSecret, setTOTPSecret] = useState('');
    const [QRCode, setQRCode] = useState('');
    const [enable2FAError, setEnable2FAError] = useState('');
    const [disable2FAError, setDisable2FAError] = useState('');
    const [disable2FASecret, setDisable2FASecret] = useState('');

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data.data)
    );

    const updateUserPasswordMutation = useMutation(
        ({ userId, secret, needsPasswordReset }: { userId: string; secret: string; needsPasswordReset: boolean }) =>
            apiClient.putUserAuthSecret(userId, {
                needs_password_reset: needsPasswordReset,
                secret: secret,
            }),
        {
            onSuccess: () => {
                addNotification('Password updated successfully!', 'updateUserPasswordSuccess');
                setChangePasswordDialogOpen(false);
            },
        }
    );

    if (getSelfQuery.isLoading) {
        return (
            <PageWithTitle title='My Profile' data-testid='my-profile'>
                <Typography variant='h2'>User Information</Typography>
                <Box p={4} textAlign='center'>
                    <CircularProgress />
                </Box>
            </PageWithTitle>
        );
    }

    if (getSelfQuery.isError) {
        return (
            <PageWithTitle title='My Profile' data-testid='my-profile'>
                <Typography variant='h2'>User Information</Typography>

                <Alert severity='error'>
                    <AlertTitle>Error</AlertTitle>
                    Sorry, there was a problem fetching your user information.
                    <br />
                    Please try refreshing the page or logging in again.
                </Alert>
            </PageWithTitle>
        );
    }

    const user = getSelfQuery.data;

    return (
        <>
            <PageWithTitle title='My Profile' data-testid='my-profile'>
                <Typography variant='h2'>User Information</Typography>

                <Grid container spacing={2} alignItems='center'>
                    <Grid item xs={3}>
                        <Typography variant='body1'>Email</Typography>
                    </Grid>
                    <Grid item xs={9}>
                        <Typography variant='body1'>{user?.email_address}</Typography>
                    </Grid>

                    <Grid item xs={3}>
                        <Typography variant='body1'>Name</Typography>
                    </Grid>
                    <Grid item xs={9}>
                        <Typography variant='body1'>
                            <TextWithFallback text={getUsername(user)} fallback='Unknown' />
                        </Typography>
                    </Grid>

                    <Grid item xs={3}>
                        <Typography variant='body1'>Role</Typography>
                    </Grid>
                    <Grid item xs={9}>
                        <Typography variant='body1'>
                            <TextWithFallback text={user?.roles?.[0]?.name} fallback='Unknown' />
                        </Typography>
                    </Grid>
                </Grid>

                <Box mt={2}>
                    <Typography variant='h2'>Authentication</Typography>
                </Box>
                <Grid container spacing={2} alignItems='center'>
                    <Grid container item>
                        <Grid item xs={3}>
                            <Typography variant='body1'>API Key Management</Typography>
                        </Grid>
                        <Grid item xs={2}>
                            <Button
                                variant='contained'
                                color='primary'
                                size='small'
                                disableElevation
                                fullWidth
                                onClick={() => setUserTokenManagementDialogOpen(true)}
                                data-testid='my-profile_button-api-key-management'>
                                API Key Management
                            </Button>
                        </Grid>
                    </Grid>
                    {user.saml_provider_id === null && (
                        <>
                            <Grid container item>
                                <Grid item xs={3}>
                                    <Typography variant='body1'>Password</Typography>
                                </Grid>
                                <Grid item xs={2}>
                                    <Button
                                        variant='contained'
                                        color='primary'
                                        size='small'
                                        disableElevation
                                        fullWidth
                                        onClick={() => setChangePasswordDialogOpen(true)}
                                        data-testid='my-profile_button-reset-password'>
                                        Reset Password
                                    </Button>
                                </Grid>
                            </Grid>

                            <Grid container item>
                                <Grid item xs={3}>
                                    <Typography variant='body1'>Multi-Factor Authentication</Typography>
                                </Grid>
                                <Grid item xs={9}>
                                    <Box display='flex' alignItems='center'>
                                        <Switch
                                            inputProps={{
                                                'aria-label': 'Multi-Factor Authentication Enabled',
                                            }}
                                            checked={user.AuthSecret?.totp_activated}
                                            onChange={() => {
                                                if (!user.AuthSecret?.totp_activated) setEnable2FADialogOpen(true);
                                                else setDisable2FADialogOpen(true);
                                            }}
                                            color='primary'
                                            data-testid='my-profile_switch-multi-factor-authentication'
                                        />
                                        {user.AuthSecret?.totp_activated && (
                                            <Typography variant='body1'>Enabled</Typography>
                                        )}
                                    </Box>
                                </Grid>
                            </Grid>
                        </>
                    )}
                </Grid>
            </PageWithTitle>

            <PasswordDialog
                open={changePasswordDialogOpen}
                onClose={() => setChangePasswordDialogOpen(false)}
                userId={user.id}
                showNeedsPasswordReset={false}
                onSave={updateUserPasswordMutation.mutate}
            />

            <UserTokenManagementDialog
                open={userTokenManagementDialogOpen}
                onClose={() => setUserTokenManagementDialogOpen(false)}
                userId={user.id}
            />

            <Enable2FADialog
                open={enable2FADialogOpen}
                onClose={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onCancel={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onSavePassword={(password) => {
                    setEnable2FAError('');
                    return apiClient
                        .enrollMFA(user.id, {
                            secret: password,
                        })
                        .then((response) => {
                            setQRCode(response.data.data.qr_code);
                            setTOTPSecret(response.data.data.totp_secret);
                            setEnable2FAError('');
                        })
                        .catch((err) => {
                            setEnable2FAError('Unable to verify password. Please try again.');
                            throw err;
                        });
                }}
                onSaveOTP={(OTP) => {
                    setEnable2FAError('');
                    return apiClient
                        .activateMFA(user.id, {
                            otp: OTP,
                        })
                        .then(() => {
                            setEnable2FAError('');
                        })
                        .catch((err) => {
                            setEnable2FAError('Unable to verify one-time password. Please try again.');
                            throw err;
                        });
                }}
                onSave={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    getSelfQuery.refetch();
                }}
                TOTPSecret={TOTPSecret}
                QRCode={QRCode}
                error={enable2FAError}
            />

            <Disable2FADialog
                open={disable2FADialogOpen}
                onClose={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    getSelfQuery.refetch();
                }}
                onCancel={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    getSelfQuery.refetch();
                }}
                onSave={(secret: string) => {
                    setDisable2FAError('');
                    apiClient
                        .disenrollMFA(user.id, { secret })
                        .then(() => {
                            setDisable2FADialogOpen(false);
                            setDisable2FAError('');
                            setDisable2FASecret('');
                            getSelfQuery.refetch();
                        })
                        .catch(() => {
                            setDisable2FAError('Unable to verify password. Please try again.');
                        });
                }}
                error={disable2FAError}
                secret={disable2FASecret}
                onSecretChange={(e: any) => setDisable2FASecret(e.target.value)}
                contentText='To stop using multi-factor authentication, please enter your password for security purposes.'
            />
        </>
    );
};

export default UserProfile;
