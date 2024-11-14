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
import { CreateOIDCProviderRequest, SSOProvider } from 'js-client-library';
import { ChangeEvent, FC, useMemo, useState } from 'react';
import { useMutation, useQuery } from 'react-query';
import {
    ConfirmationDialog,
    CreateMenu,
    CreateSAMLProviderDialog,
    CreateSAMLProviderFormInputs,
    DocumentationLinks,
    PageWithTitle,
    SSOProviderInfoPanel,
    SSOProviderTable,
} from '../../components';
import CreateOIDCProviderDialog from '../../components/CreateOIDCProviderDialog';
import { useFeatureFlag } from '../../hooks';
import { useNotifications } from '../../providers';
import { SortOrder, apiClient } from '../../utils';

const SSOConfiguration: FC = () => {
    /* Hooks */
    const theme = useTheme();
    const { addNotification } = useNotifications();
    const { data: flag } = useFeatureFlag('oidc_support');

    const [selectedSSOProviderId, setSelectedSSOProviderId] = useState<SSOProvider['id'] | undefined>();
    const [ssoProviderIdToDelete, setSSOProviderIdToDelete] = useState<SSOProvider['id'] | undefined>();
    const [dialogOpen, setDialogOpen] = useState<'SAML' | 'OIDC' | 'DELETE' | ''>('');
    const [nameFilter, setNameFilter] = useState<string>('');
    const [createProviderError, setCreateProviderError] = useState<string>('');
    const [typeSortOrder, setTypeSortOrder] = useState<SortOrder>();

    const listSSOProvidersQuery = useQuery(['listSSOProviders'], ({ signal }) =>
        apiClient.listSSOProviders({ signal }).then((res) => res.data.data)
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
        setDialogOpen('');
        setCreateProviderError('');
    };

    const onClickSSOProvider = (ssoProviderId: SSOProvider['id']) => {
        setSelectedSSOProviderId(ssoProviderId);
    };

    const onSelectDeleteSSOProvider = (ssoProviderId: SSOProvider['id']) => {
        setSSOProviderIdToDelete(ssoProviderId);
        openDeleteProviderDialog();
    };

    const onDeleteSSOProvider = async (response: boolean) => {
        let errored = false;
        if (response && ssoProviderIdToDelete) {
            try {
                await deleteSSOProviderMutation.mutateAsync(ssoProviderIdToDelete);
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

    const createSAMLProvider = async (samlProvider: CreateSAMLProviderFormInputs) => {
        setCreateProviderError('');
        try {
            await apiClient.createSAMLProviderFromFile({ ...samlProvider, metadata: samlProvider.metadata[0] });
            listSSOProvidersQuery.refetch();
            closeDialog();
        } catch (error) {
            console.error(error);
            setCreateProviderError('Unable to create new SAML Provider configuration. Please try again.');
        }
    };

    const createOIDCProvider = async (oidcProvider: CreateOIDCProviderRequest) => {
        setCreateProviderError('');
        try {
            await apiClient.createOIDCProvider(oidcProvider);
            listSSOProvidersQuery.refetch();
            closeDialog();
        } catch (error) {
            console.error(error);
            setCreateProviderError('Unable to create new OIDC Provider configuration. Please try again.');
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
                                onDeleteSSOProvider={onSelectDeleteSSOProvider}
                                typeSortOrder={typeSortOrder}
                                onToggleTypeSortOrder={toggleTypeSortOrder}
                            />
                        </Paper>
                    </Grid>
                    {selectedSSOProvider && (
                        <Grid item xs={6}>
                            <SSOProviderInfoPanel ssoProvider={selectedSSOProvider} />
                        </Grid>
                    )}
                </Grid>
            </PageWithTitle>
            <CreateSAMLProviderDialog
                open={dialogOpen === 'SAML'}
                error={createProviderError}
                onClose={closeDialog}
                onSubmit={createSAMLProvider}
            />
            <CreateOIDCProviderDialog
                open={dialogOpen === 'OIDC'}
                error={createProviderError}
                onClose={closeDialog}
                onSubmit={createOIDCProvider}
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
