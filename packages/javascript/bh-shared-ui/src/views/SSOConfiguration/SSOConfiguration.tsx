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

import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Grid, Paper, TextField, Typography, useTheme } from '@mui/material';
import { SSOProvider, UpsertOIDCProviderRequest, UpsertSAMLProviderFormInputs } from 'js-client-library';
import { ChangeEvent, FC, useMemo, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import {
    ConfirmationDialog,
    CreateMenu,
    DocumentationLinks,
    PageWithTitle,
    SSOProviderInfoPanel,
    SSOProviderTable,
} from '../../components';
import { UpsertOIDCProviderDialog, UpsertSAMLProviderDialog } from '../../components/UpsertSSOProviders';
import { useFeatureFlag, useForbiddenNotifier } from '../../hooks';
import { useNotifications } from '../../providers';
import { Permission, SortOrder, apiClient } from '../../utils';

const SSOConfiguration: FC<{ permissions: Permission[] }> = ({ permissions }) => {
    /* Hooks */
    const [selectedSSOProviderId, setSelectedSSOProviderId] = useState<SSOProvider['id'] | undefined>();
    const [ssoProviderIdToDeleteOrUpdate, setSSOProviderIdToDeleteOrUpdate] = useState<SSOProvider['id'] | undefined>();
    const [dialogOpen, setDialogOpen] = useState<'SAML' | 'OIDC' | 'DELETE' | ''>('');
    const [nameFilter, setNameFilter] = useState<string>('');
    const [upsertProviderError, setUpsertProviderError] = useState<any>();
    const [typeSortOrder, setTypeSortOrder] = useState<SortOrder>();

    const { data: flag } = useFeatureFlag('oidc_support');
    const theme = useTheme();
    const { addNotification } = useNotifications();
    const forbidden = useForbiddenNotifier(
        Permission.AUTH_MANAGE_PROVIDERS,
        permissions,
        'Your role does not grant permission to manage SSO providers.',
        'manage-sso-permission'
    );

    const getRolesQuery = useQuery(['getRoles'], ({ signal }) =>
        apiClient.getRoles({ signal }).then((res) => res.data.data.roles)
    );

    const listSSOProvidersQuery = useQuery(
        ['listSSOProviders'],
        ({ signal }) => apiClient.listSSOProviders({ signal }).then((res) => res.data.data),
        { enabled: !forbidden }
    );

    const deleteSSOProviderMutation = useMutation(
        (ssoProviderId: SSOProvider['id']) => apiClient.deleteSSOProvider(ssoProviderId),
        {
            onSuccess: () => {
                addNotification('SSO Provider successfully deleted!', 'deleteSSOProviderSuccess', {
                    variant: 'success',
                });
            },
            onError: (err: any) => {
                addNotification(
                    err?.response?.status === 404
                        ? 'SSO Provider not found.'
                        : 'Unable to delete sso provider. Please try again.',
                    'deleteSSOProviderFailure',
                    { variant: 'error' }
                );
            },
        }
    );

    const ssoProviders = useMemo(() => {
        let ssoProviders = listSSOProvidersQuery.data ?? [];

        if (nameFilter) {
            ssoProviders = ssoProviders.filter((ssoProvider) => ssoProvider.name?.toLowerCase()?.includes(nameFilter));
        }

        if (typeSortOrder) {
            ssoProviders = ssoProviders.sort((a, b) => {
                switch (typeSortOrder) {
                    case 'asc':
                        return a.type.localeCompare(b.type);
                    case 'desc':
                        return b.type.localeCompare(a.type);
                }
            });
        }

        return ssoProviders;
    }, [nameFilter, typeSortOrder, listSSOProvidersQuery.data]);

    const selectedSSOProvider = useMemo(() => {
        return listSSOProvidersQuery.data?.find(({ id }) => id === selectedSSOProviderId);
    }, [selectedSSOProviderId, listSSOProvidersQuery.data]);

    const selectedSSOProviderToUpdate = useMemo(() => {
        return listSSOProvidersQuery.data?.find(({ id }) => id === ssoProviderIdToDeleteOrUpdate);
    }, [ssoProviderIdToDeleteOrUpdate, listSSOProvidersQuery.data]);

    /* Event Handlers */

    const openSAMLProviderDialog = () => {
        setDialogOpen('SAML');
    };

    const openOIDCProviderDialog = () => {
        setDialogOpen('OIDC');
    };

    const openDeleteProviderDialog = () => {
        setDialogOpen('DELETE');
    };

    const closeDialog = () => {
        setUpsertProviderError(null);
        setDialogOpen('');
        setTimeout(() => setSSOProviderIdToDeleteOrUpdate(undefined), 500);
    };

    const onClickSSOProvider = (ssoProviderId: SSOProvider['id']) => {
        setSelectedSSOProviderId(ssoProviderId);
    };

    const onSelectDeleteOrUpdateSSOProvider = (action: 'DELETE' | 'UPDATE') => (ssoProvider: SSOProvider) => {
        setSSOProviderIdToDeleteOrUpdate(ssoProvider.id);
        switch (action) {
            case 'DELETE':
                openDeleteProviderDialog();
                break;
            case 'UPDATE':
                switch (ssoProvider.type) {
                    case 'SAML':
                        openSAMLProviderDialog();
                        break;
                    case 'OIDC':
                        openOIDCProviderDialog();
                        break;
                }
                break;
        }
    };

    const onDeleteSSOProvider = async (response: boolean) => {
        let errored = false;
        if (response && ssoProviderIdToDeleteOrUpdate) {
            try {
                await deleteSSOProviderMutation.mutateAsync(ssoProviderIdToDeleteOrUpdate);
            } catch (err: any) {
                if (err?.response?.status !== 404) {
                    errored = true;
                    console.error(err);
                }
            }
        }
        if (!errored) {
            closeDialog();
            deleteSSOProviderMutation.reset();
            listSSOProvidersQuery.refetch();
        }
    };

    const toggleTypeSortOrder = () => {
        if (!typeSortOrder || typeSortOrder === 'desc') {
            setTypeSortOrder('asc');
        } else {
            setTypeSortOrder('desc');
        }
    };

    const upsertSAMLProvider = async (samlProvider: UpsertSAMLProviderFormInputs) => {
        setUpsertProviderError(null);
        try {
            const payload = {
                name: samlProvider.name,
                metadata: samlProvider.metadata && samlProvider.metadata[0],
                config: samlProvider.config,
            };
            if (ssoProviderIdToDeleteOrUpdate) {
                await apiClient.updateSAMLProviderFromFile(ssoProviderIdToDeleteOrUpdate, payload);
            } else {
                if (payload.name && payload.metadata && payload.config) {
                    await apiClient.createSAMLProviderFromFile({
                        name: payload.name,
                        metadata: payload.metadata,
                        config: payload.config,
                    });
                }
            }
            listSSOProvidersQuery.refetch();
            closeDialog();
        } catch (error: any) {
            console.error(error);
            setUpsertProviderError(error);
        }
    };

    const upsertOIDCProvider = async (oidcProvider: UpsertOIDCProviderRequest) => {
        setUpsertProviderError(null);
        try {
            if (ssoProviderIdToDeleteOrUpdate) {
                await apiClient.updateOIDCProvider(ssoProviderIdToDeleteOrUpdate, oidcProvider);
            } else {
                if (oidcProvider.name && oidcProvider.client_id && oidcProvider.issuer && oidcProvider.config) {
                    await apiClient.createOIDCProvider({
                        name: oidcProvider.name,
                        client_id: oidcProvider.client_id,
                        issuer: oidcProvider.issuer,
                        config: oidcProvider.config,
                    });
                }
            }
            listSSOProvidersQuery.refetch();
            closeDialog();
        } catch (error) {
            console.error(error);
            setUpsertProviderError(error);
        }
    };

    const onChangeNameFilter = (e: ChangeEvent<HTMLInputElement>) => {
        setNameFilter(e.target.value.toLowerCase());
    };

    /* Implementation */

    return (
        <>
            <PageWithTitle
                title='SSO Configuration'
                data-testid='sso-configuration'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        BloodHound supports SAML {flag?.enabled ? 'and OIDC ' : ''}for single sign-on (SSO). Learn how
                        to deploy {flag?.enabled ? 'SSO' : 'SAML'} with BloodHound{' '}
                        {DocumentationLinks.samlConfigDocLink}.
                    </Typography>
                }>
                <Grid container spacing={theme.spacing(2)}>
                    <Grid item display='flex' alignItems='center' justifyContent='end' minHeight='24px' mb={2} xs={12}>
                        <CreateMenu
                            disabled={forbidden}
                            createMenuTitle={`Create ${flag?.enabled ? '' : 'SAML '}Provider`}
                            featureFlag='oidc_support'
                            featureFlagEnabledMenuItems={[
                                { title: 'SAML Provider', onClick: openSAMLProviderDialog },
                                { title: 'OIDC Provider', onClick: openOIDCProviderDialog },
                            ]}
                            menuItems={[{ title: 'Create SAML Provider', onClick: openSAMLProviderDialog }]}
                        />
                    </Grid>
                    <Grid item xs={6}>
                        <Paper>
                            <Box display='flex' justifyContent='space-between'>
                                <Box display='flex' alignItems='center' ml={theme.spacing(3)} pt={theme.spacing(2)}>
                                    <Typography fontWeight='bold' variant='h5'>
                                        Providers
                                    </Typography>
                                </Box>
                                <Box display='flex' alignItems='center' mr={theme.spacing(3)}>
                                    <TextField
                                        onChange={onChangeNameFilter}
                                        variant='standard'
                                        label={
                                            <Box>
                                                Search
                                                <FontAwesomeIcon
                                                    icon={faSearch}
                                                    size='sm'
                                                    style={{ marginLeft: theme.spacing(1) }}
                                                />
                                            </Box>
                                        }
                                    />
                                </Box>
                            </Box>
                            <SSOProviderTable
                                ssoProviders={ssoProviders}
                                loading={listSSOProvidersQuery.isLoading}
                                onClickSSOProvider={onClickSSOProvider}
                                onDeleteSSOProvider={onSelectDeleteOrUpdateSSOProvider('DELETE')}
                                onUpdateSSOProvider={onSelectDeleteOrUpdateSSOProvider('UPDATE')}
                                typeSortOrder={typeSortOrder}
                                onToggleTypeSortOrder={toggleTypeSortOrder}
                            />
                        </Paper>
                    </Grid>
                    {selectedSSOProvider && (
                        <Grid item xs={6}>
                            <SSOProviderInfoPanel ssoProvider={selectedSSOProvider} roles={getRolesQuery.data} />
                        </Grid>
                    )}
                </Grid>
            </PageWithTitle>

            {/* Dialogs */}
            <UpsertSAMLProviderDialog
                open={dialogOpen === 'SAML'}
                oldSSOProvider={selectedSSOProviderToUpdate}
                error={upsertProviderError}
                onClose={closeDialog}
                onSubmit={upsertSAMLProvider}
                roles={getRolesQuery.data}
            />
            <UpsertOIDCProviderDialog
                open={dialogOpen === 'OIDC'}
                oldSSOProvider={selectedSSOProviderToUpdate}
                error={upsertProviderError}
                onClose={closeDialog}
                onSubmit={upsertOIDCProvider}
                roles={getRolesQuery.data}
            />
            <ConfirmationDialog
                open={dialogOpen === 'DELETE'}
                title='Delete SSO Provider'
                text='Are you sure you wish to delete this SSO Provider? Any users which are currently configured to use this provider for authentication will no longer be able to access this application.'
                onClose={onDeleteSSOProvider}
                error={deleteSSOProviderMutation.isError ? 'An unexpected error has occurred. Please try again.' : ''}
                isLoading={deleteSSOProviderMutation.isLoading}
            />
        </>
    );
};

export default SSOConfiguration;
