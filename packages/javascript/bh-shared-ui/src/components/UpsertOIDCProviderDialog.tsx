// Copyright 2024 Specter Ops, Inc.
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
import { Alert, Dialog, DialogTitle, DialogContent, DialogActions, Grid, TextField } from '@mui/material';
import { FC } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { UpsertOIDCProviderRequest } from 'js-client-library';

const UpsertOIDCProviderDialog: FC<{
    open: boolean;
    error?: string;
    isUpdate: boolean;
    onClose: () => void;
    onSubmit: (data: UpsertOIDCProviderRequest) => void;
}> = ({ open, error, isUpdate, onClose, onSubmit: _onSubmit }) => {
    const {
        control,
        handleSubmit,
        reset,
        formState: { errors },
    } = useForm<UpsertOIDCProviderRequest>({
        defaultValues: {
            name: '',
            client_id: '',
            issuer: '',
        },
    });

    const onSubmit = (data: UpsertOIDCProviderRequest) => {
        _onSubmit(data);
        reset();
    };

    const handleClose = () => {
        onClose();
        reset();
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth='sm'
            PaperProps={{
                // @ts-ignore
                'data-testid': 'create-oidc-provider-dialog',
            }}>
            <DialogTitle>{isUpdate ? 'Edit' : 'Create'} OIDC Provider</DialogTitle>
            <form onSubmit={handleSubmit(onSubmit)}>
                <DialogContent>
                    <Grid container spacing={2}>
                        <Grid item xs={12}>
                            <Controller
                                control={control}
                                name='name'
                                rules={{
                                    required: 'OIDC Provider Name is required',
                                    pattern: {
                                        value: /^[A-z0-9 ]+$/,
                                        message: 'OIDC Provider Name must be alphanumeric.',
                                    },
                                }}
                                render={({ field }) => (
                                    <TextField
                                        {...field}
                                        id={'name'}
                                        variant='standard'
                                        fullWidth
                                        name='name'
                                        label='OIDC Provider Name'
                                        error={!!errors.name}
                                        helperText={
                                            errors.name?.message || 'Choose a name for your OIDC Provider configuration'
                                        }
                                    />
                                )}
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <Controller
                                control={control}
                                name='client_id'
                                rules={{
                                    required: 'Client ID is required',
                                }}
                                render={({ field }) => (
                                    <TextField
                                        {...field}
                                        id={'clientId'}
                                        variant='standard'
                                        fullWidth
                                        name='clientId'
                                        label='Client ID'
                                        error={!!errors.client_id}
                                        helperText={errors.client_id?.message || 'OIDC Provider Client ID'}
                                    />
                                )}
                            />
                        </Grid>
                        <Grid item xs={12}>
                            <Controller
                                control={control}
                                name='issuer'
                                rules={{
                                    required: 'Issuer is required',
                                }}
                                render={({ field }) => (
                                    <TextField
                                        {...field}
                                        id={'issuer'}
                                        variant='standard'
                                        fullWidth
                                        name='issuer'
                                        label='Issuer'
                                        error={!!errors.issuer}
                                        helperText={errors.issuer?.message || 'OIDC Issuer'}
                                    />
                                )}
                            />
                        </Grid>
                        {error && (
                            <Grid item xs={12}>
                                <Alert severity='error'>{error}</Alert>
                            </Grid>
                        )}
                    </Grid>
                </DialogContent>
                <DialogActions>
                    <Button
                        type='button'
                        variant='tertiary'
                        onClick={handleClose}
                        data-testid='create-oidc-provider-dialog_button-close'>
                        Cancel
                    </Button>
                    <Button data-testid='create-oidc-provider-dialog_button-save' type='submit'>
                        Submit
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
    );
};

export default UpsertOIDCProviderDialog;
