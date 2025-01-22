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

import { Button, Label } from '@bloodhoundenterprise/doodleui';
import { Box, Paper, Typography, useTheme } from '@mui/material';
import { OIDCProviderInfo, Role, SAMLProviderInfo, SSOProvider } from 'js-client-library';
import fileDownload from 'js-file-download';
import { FC, useMemo } from 'react';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { Field, FieldsContainer, useHeaderStyles, usePaneStyles } from '../../views/Explore';
import LabelWithCopy from '../LabelWithCopy';

const SAMLProviderInfoPanel: FC<{
    samlProviderDetails: SAMLProviderInfo;
}> = ({ samlProviderDetails }) => (
    <>
        <Field
            label={<LabelWithCopy label='IdP SSO URL' valueToCopy={samlProviderDetails.idp_sso_uri} hoverOnly />}
            value={samlProviderDetails.idp_sso_uri}
        />
        <Field
            label={<LabelWithCopy label='BHE SSO URL' valueToCopy={samlProviderDetails.sp_sso_uri} hoverOnly />}
            value={samlProviderDetails.sp_sso_uri}
        />
        <Field
            label={<LabelWithCopy label='BHE ACS URL' valueToCopy={samlProviderDetails.sp_acs_uri} hoverOnly />}
            value={samlProviderDetails.sp_acs_uri}
        />
        <Field
            label={
                <LabelWithCopy label='BHE Metadata URL' valueToCopy={samlProviderDetails.sp_metadata_uri} hoverOnly />
            }
            value={samlProviderDetails.sp_metadata_uri}
        />
    </>
);

const OIDCProviderInfoPanel: FC<{
    ssoProvider: SSOProvider;
}> = ({ ssoProvider }) => {
    const oidcProviderDetails = ssoProvider.details as OIDCProviderInfo;
    return (
        <>
            <Field
                label={<LabelWithCopy label='Client ID' valueToCopy={oidcProviderDetails.client_id} hoverOnly />}
                value={oidcProviderDetails.client_id}
            />
            <Field
                label={<LabelWithCopy label='Issuer' valueToCopy={oidcProviderDetails.issuer} hoverOnly />}
                value={oidcProviderDetails.issuer}
            />
            <Field
                label={<LabelWithCopy label='Callback URL' valueToCopy={ssoProvider.callback_uri} hoverOnly />}
                value={ssoProvider.callback_uri}
            />
        </>
    );
};

const SSOProviderInfoPanel: FC<{
    ssoProvider: SSOProvider;
    roles?: Role[];
}> = ({ ssoProvider, roles }) => {
    const theme = useTheme();
    const paneStyles = usePaneStyles();
    const headerStyles = useHeaderStyles();
    const { addNotification } = useNotifications();

    const defaultRoleName = useMemo(
        () => roles?.find((role) => role.id === ssoProvider.config?.auto_provision?.default_role_id)?.name,
        [roles, ssoProvider.config?.auto_provision?.default_role_id]
    );

    if (!ssoProvider.type) {
        return null;
    }

    let innerInfoPanel;
    switch (ssoProvider.type.toLowerCase()) {
        case 'saml':
            innerInfoPanel = <SAMLProviderInfoPanel samlProviderDetails={ssoProvider.details as SAMLProviderInfo} />;
            break;
        case 'oidc':
            innerInfoPanel = <OIDCProviderInfoPanel ssoProvider={ssoProvider} />;
            break;
        default:
            innerInfoPanel = null;
    }

    const downloadSAMLSigningCertificate = () => {
        if (ssoProvider.type.toLowerCase() == 'oidc') {
            addNotification('Only SAML providers support signing certificates.', 'errorDownloadSAMLSigningCertificate');
        } else {
            apiClient
                .getSAMLProviderSigningCertificate(ssoProvider.id)
                .then((res) => {
                    const filename =
                        res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] ||
                        `${ssoProvider.name}-signing-certificate`;

                    fileDownload(res.data, filename);
                })
                .catch((err) => {
                    console.error(err);
                    addNotification(
                        'This file could not be downloaded. Please try again.',
                        'downloadSAMLSigningCertificate'
                    );
                });
        }
    };

    return (
        <>
            <Box className={paneStyles.container} data-testid='sso_provider-info-panel'>
                <Paper>
                    <Box className={headerStyles.header} sx={{ backgroundColor: theme.palette.neutral.quinary }}>
                        <Box
                            sx={{
                                backgroundColor: theme.palette.primary.main,
                                width: 10,
                                height: theme.spacing(7),
                                mr: theme.spacing(1),
                            }}
                        />
                        <Typography
                            data-testid='sso_provider-info-panel_header-text'
                            variant={'h5'}
                            noWrap
                            sx={{
                                color: theme.palette.text.primary,
                                flexGrow: 1,
                            }}>
                            {ssoProvider?.name}
                        </Typography>
                    </Box>
                    <Paper
                        elevation={0}
                        sx={{
                            backgroundColor: theme.palette.neutral.secondary,
                            overflowX: 'hidden',
                            overflowY: 'auto',
                            padding: theme.spacing(1, 2),
                            pointerEvents: 'auto',
                            '& > div.node:nth-of-type(odd)': {
                                background: theme.palette.neutral.tertiary,
                            },
                        }}>
                        <Box flexShrink={0} flexGrow={1} fontWeight='bold' ml={theme.spacing(1)} fontSize={'small'}>
                            Provider Information:
                        </Box>
                        <FieldsContainer>
                            {innerInfoPanel}
                            <Field
                                label={<Label>Automatically create new users on login</Label>}
                                value={ssoProvider.config?.auto_provision?.enabled ? 'Yes' : 'No'}
                            />
                            {ssoProvider.config?.auto_provision?.enabled && (
                                <>
                                    <Field
                                        label={<Label>Allow SSO provider to manage roles for new users</Label>}
                                        value={ssoProvider.config?.auto_provision?.role_provision ? 'Yes' : 'No'}
                                    />
                                    <Field
                                        label={<Label>Default role when creating new users</Label>}
                                        value={defaultRoleName ?? 'Read-Only'}
                                    />
                                </>
                            )}
                        </FieldsContainer>
                    </Paper>
                </Paper>
            </Box>
            {ssoProvider.type.toLowerCase() === 'saml' && (
                <Box mt={theme.spacing(1)} justifyContent='center' display='flex'>
                    <Button
                        aria-label={`Download ${ssoProvider.name} SP Certificate`}
                        variant='secondary'
                        onClick={downloadSAMLSigningCertificate}>
                        Download SAML SP Certificate
                    </Button>
                </Box>
            )}
        </>
    );
};

export default SSOProviderInfoPanel;
