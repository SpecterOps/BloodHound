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
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    LinearProgress,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { DateTime } from 'luxon';
import React, { useState } from 'react';
import { useQuery } from 'react-query';
import { AuthToken, NewAuthToken } from 'js-client-library';
import { LuxonFormat, apiClient } from '../../utils';
import { useNotifications } from '../../providers';
import CreateUserTokenDialog from './CreateUserTokenDialog';
import UserTokenDialog from './UserTokenDialog';
import TokenRevokeDialog from './TokenRevokeDialog';

const useStyles = makeStyles({
    revokeButton: {
        padding: '0',
        '& .MuiButton-label': {
            padding: 0,
            justifyContent: 'left',
        },
    },
});

const UserTokenManagementDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    userId: string;
}> = ({ open, onClose, userId }) => {
    const { isLoading, isIdle, error, data, refetch } = useQuery(
        ['getUserTokens', userId],
        ({ signal }) => apiClient.getUserTokens(userId, { signal }),
        {
            enabled: !!open,
        }
    );

    const { addNotification } = useNotifications();

    const styles = useStyles();

    const [newTokenDialogOpen, setNewTokenDialogOpen] = useState<boolean>(false);
    const [tokenDialogOpen, setTokenDialogOpen] = useState<boolean>(false);
    const [currentToken, setCurrentToken] = useState<AuthToken | NewAuthToken | undefined>(undefined);
    const [tokenRevokeDialogOpen, setTokenRevokeDialogOpen] = useState<boolean>(false);

    const openRevokeTokenDialog = (token: AuthToken) => {
        setCurrentToken(token);
        setTokenRevokeDialogOpen(true);
    };

    const handleTokenRevoke = async () => {
        if (currentToken == null) return;
        try {
            await apiClient.deleteUserToken(currentToken.id);
        } catch (error) {
            console.error(error);
            addNotification(`Error deleting token: ${error}:`, 'ErrorDeleteToken');
        }

        await refetch();
        clearToken();
    };

    const handleNewTokenSubmit = async (newToken: { token_name: string }) => {
        setNewTokenDialogOpen(false);
        try {
            const {
                data: { data: token },
            } = await apiClient.createUserToken(userId, newToken.token_name);
            setCurrentToken(token);
            setTokenDialogOpen(true);
        } catch (error) {
            console.error(error);
            addNotification(`Error creating token: ${error}:`, 'ErrorCreateToken');
        }

        await refetch();
    };

    const clearToken = () => {
        setCurrentToken(undefined);
        setTokenDialogOpen(false);
        setTokenRevokeDialogOpen(false);
    };

    const tableComponent = () => {
        if (isLoading || isIdle) {
            return (
                <TableRow>
                    <TableCell colSpan={4}>
                        <LinearProgress />
                    </TableCell>
                </TableRow>
            );
        } else if (error) {
            return (
                <TableRow>
                    <TableCell>
                        <div>Error loading user tokens</div>
                    </TableCell>
                </TableRow>
            );
        } else {
            const tokens = data?.data.data.tokens || [];
            if (tokens.length === 0) {
                return (
                    <TableRow>
                        <TableCell colSpan={4} align='center'>
                            No tokens available
                        </TableCell>
                    </TableRow>
                );
            } else {
                return tokens.map((row) => (
                    <TableRow key={`${row.id}`}>
                        <TableCell component={'th'} scope={'row'}>
                            {row.name}
                        </TableCell>
                        <TableCell>{DateTime.fromISO(row.created_at).toFormat(LuxonFormat.DATETIME)}</TableCell>
                        <TableCell>{DateTime.fromISO(row.last_access).toFormat(LuxonFormat.DATETIME)}</TableCell>
                        <TableCell>
                            <Button
                                variant={'text'}
                                color={'secondary'}
                                className={styles.revokeButton}
                                onClick={() => openRevokeTokenDialog(row)}>
                                Revoke
                            </Button>
                        </TableCell>
                    </TableRow>
                ));
            }
        }
    };

    return (
        <>
            <Dialog
                open={open}
                fullWidth={true}
                maxWidth={'md'}
                onClose={onClose}
                PaperProps={{
                    // @ts-ignore
                    'data-testid': 'user-token-management-dialog',
                }}>
                <DialogTitle>Generate/Revoke API Tokens</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Permanent Authentication Tokens are used for authenticating API calls. Tokens never expire and
                        will remain valid until revoked. Ensure tokens are stored securely.
                    </DialogContentText>

                    <Typography variant={'h6'}>Existing Tokens</Typography>

                    <Table data-testid='user-token-management-dialog_table'>
                        <TableHead>
                            <TableRow>
                                <TableCell>Description</TableCell>
                                <TableCell>Created</TableCell>
                                <TableCell>Last Use</TableCell>
                                <TableCell>Actions</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>{tableComponent()}</TableBody>
                    </Table>
                </DialogContent>
                <DialogActions>
                    <Button
                        autoFocus
                        color='inherit'
                        onClick={onClose}
                        data-testid='user-token-management-dialog_button-close'>
                        Close
                    </Button>
                    <Button
                        color='primary'
                        type='submit'
                        onClick={() => setNewTokenDialogOpen(true)}
                        data-testid='user-token-management-dialog_button-save'>
                        Create Token
                    </Button>
                </DialogActions>
            </Dialog>
            {newTokenDialogOpen && (
                <CreateUserTokenDialog
                    open={newTokenDialogOpen}
                    onCancel={() => setNewTokenDialogOpen(false)}
                    onSubmit={handleNewTokenSubmit}
                />
            )}
            {tokenDialogOpen && (
                <UserTokenDialog open={tokenDialogOpen} onClose={clearToken} token={currentToken as NewAuthToken} />
            )}
            {tokenRevokeDialogOpen && (
                <TokenRevokeDialog
                    open={tokenRevokeDialogOpen}
                    onCancel={() => setTokenRevokeDialogOpen(false)}
                    onConfirm={handleTokenRevoke}
                    token={currentToken}
                />
            )}
        </>
    );
};

export default UserTokenManagementDialog;
