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

import { Button } from '@bloodhoundenterprise/doodleui';
import { useState } from 'react';
import { Box, Typography } from '@mui/material';
import { useMutation, useQuery } from 'react-query';
import { PageWithTitle, apiClient } from 'bh-shared-ui';
import {
    CreateSAMLProviderDialog,
    CreateSAMLProviderFormInputs,
    ConfirmationDialog,
    DocumentationLinks,
    SAMLProviderTable,
} from 'bh-shared-ui';
import useToggle from 'src/hooks/useToggle';
import { useAppDispatch } from 'src/store';
import { addSnackbar } from 'src/ducks/global/actions';

const SAMLConfiguration: React.FC = () => {
    /* Hooks */
    const dispatch = useAppDispatch();
    const [selectedSAMLProvider, setSelectedSAMLProvider] = useState('');
    const [SAMLProviderDialogOpen, setSAMLProviderDialogOpen] = useState(false);
    const [deleteSAMLProviderDialogOpen, toggleDeleteSAMLProviderDialogOpen] = useToggle(false);
    const [createSAMLProviderError, setCreateSAMLProviderError] = useState('');
    const listSAMLProvidersQuery = useQuery(['listSAMLProviders'], ({ signal }) =>
        apiClient.listSAMLProviders({ signal }).then((res) => res.data.data.saml_providers)
    );
    const deleteSAMLProviderMutation = useMutation(
        (SAMLProviderId: string) => apiClient.deleteSAMLProvider(SAMLProviderId),
        {
            onSuccess: () => {
                dispatch(addSnackbar('SAML Provider successfully deleted!', 'deleteSAMLProviderSuccess'));
            },
        }
    );

    /* Event Handlers */

    const openSAMLProviderDialog = () => {
        setSAMLProviderDialogOpen(true);
    };

    const closeSAMLProviderDialog = () => {
        setSAMLProviderDialogOpen(false);
        setCreateSAMLProviderError('');
    };

    const createSAMLProvider = async (samlProvider: CreateSAMLProviderFormInputs) => {
        setCreateSAMLProviderError('');
        try {
            await apiClient.createSAMLProviderFromFile({ ...samlProvider, metadata: samlProvider.metadata[0] });
            listSAMLProvidersQuery.refetch();
            closeSAMLProviderDialog();
        } catch (error) {
            console.error(error);
            setCreateSAMLProviderError('Unable to create new SAML Provider configuration. Please try again.');
        }
    };

    /* Implementation */

    return (
        <>
            <PageWithTitle
                title='SAML Configuration'
                data-testid='saml-configuration'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        BloodHound supports SAML for single sign-on (SSO). Learn how to deploy SAML{' '}
                        {DocumentationLinks.samlConfigDocLink}.
                    </Typography>
                }>
                <Box>
                    <Box display='flex' justifyContent='space-between' mb={2}>
                        <div />
                        <Button onClick={openSAMLProviderDialog}>Create SAML Provider</Button>
                    </Box>
                    <SAMLProviderTable
                        SAMLProviders={listSAMLProvidersQuery.data || []}
                        loading={listSAMLProvidersQuery.isLoading}
                        onDeleteSAMLProvider={(SAMLProviderId) => {
                            setSelectedSAMLProvider(SAMLProviderId);
                            toggleDeleteSAMLProviderDialogOpen();
                        }}
                    />
                </Box>
            </PageWithTitle>
            <CreateSAMLProviderDialog
                open={SAMLProviderDialogOpen}
                error={createSAMLProviderError}
                onClose={closeSAMLProviderDialog}
                onSubmit={createSAMLProvider}
            />
            <ConfirmationDialog
                open={deleteSAMLProviderDialogOpen}
                title='Delete SAML Provider'
                text='Are you sure you wish to delete this SAML Provider? Any users which are currently configured to use this provider for authentication will no longer be able to access this application.'
                onClose={async (response) => {
                    if (response) {
                        try {
                            await deleteSAMLProviderMutation.mutateAsync(selectedSAMLProvider);
                            toggleDeleteSAMLProviderDialogOpen();
                            listSAMLProvidersQuery.refetch();
                        } catch (err) {
                            console.error(err);
                        }
                    } else {
                        toggleDeleteSAMLProviderDialogOpen();
                    }
                    deleteSAMLProviderMutation.reset();
                }}
                error={deleteSAMLProviderMutation.isError ? 'An unexpected error has occurred. Please try again.' : ''}
                isLoading={deleteSAMLProviderMutation.isLoading}
            />
        </>
    );
};

export default SAMLConfiguration;
