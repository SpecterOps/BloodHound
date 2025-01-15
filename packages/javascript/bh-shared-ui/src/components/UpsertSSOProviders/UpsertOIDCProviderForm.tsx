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
import { Alert, DialogContent, DialogActions, Grid, TextField } from '@mui/material';
import { useEffect, FC } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { OIDCProviderInfo, SSOProvider, UpsertOIDCProviderRequest, Role } from 'js-client-library';
import SSOProviderConfigForm, { backfillSSOProviderConfig } from './SSOProviderConfigForm';

const UpsertOIDCProviderForm: FC<{
    error: any;
    oldSSOProvider?: SSOProvider;
    roles?: Role[];
    onClose: () => void;
    onSubmit: (data: UpsertOIDCProviderRequest) => void;
}> = ({ error, oldSSOProvider, roles, onClose, onSubmit }) => {
    const readOnlyRoleId = roles?.find((role) => role.name === 'Read-Only')?.id;

    const defaultValues = {
        name: oldSSOProvider?.name ?? '',
        client_id: (oldSSOProvider?.details as OIDCProviderInfo)?.client_id ?? '',
        issuer: (oldSSOProvider?.details as OIDCProviderInfo)?.issuer ?? '',
        config: oldSSOProvider?.config ? oldSSOProvider.config : backfillSSOProviderConfig(readOnlyRoleId),
    };

    const {
        control,
        formState: { errors },
        handleSubmit,
        reset,
        resetField,
        setError,
        watch,
    } = useForm<UpsertOIDCProviderRequest>({ defaultValues });

    useEffect(() => {
        if (error) {
            if (error?.response?.status === 409) {
                if (error.response?.data?.errors[0]?.message.toLowerCase().includes('sso provider name')) {
                    setError('name', { type: 'custom', message: 'SSO Provider Name is already in use.' });
                } else {
                    setError('root.generic', {
                        type: 'custom',
                        message: 'A conflict has occured.',
                    });
                }
            } else {
                setError('root.generic', {
                    type: 'custom',
                    message: `Unable to ${oldSSOProvider ? 'update' : 'create new'} OIDC Provider configuration. Please try again.`,
                });
            }
        }
    }, [error, setError, oldSSOProvider]);

    const handleClose = () => {
        onClose();
        reset();
    };

    return (
        <form autoComplete='off' onSubmit={handleSubmit(onSubmit)}>
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
                            rules={{ required: 'Client ID is required' }}
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
                            rules={{ required: 'Issuer is required' }}
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
                    <SSOProviderConfigForm
                        control={control}
                        errors={errors}
                        readOnlyRoleId={readOnlyRoleId}
                        resetField={resetField}
                        roles={roles}
                        watch={watch}
                    />
                    {!!errors.root?.generic && (
                        <Grid item xs={12}>
                            <Alert severity='error'>{errors.root.generic.message}</Alert>
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
                    {oldSSOProvider ? 'Confirm Edits' : 'Submit'}
                </Button>
            </DialogActions>
        </form>
    );
};

export default UpsertOIDCProviderForm;
