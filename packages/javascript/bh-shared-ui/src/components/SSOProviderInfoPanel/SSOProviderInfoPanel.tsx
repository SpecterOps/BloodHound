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

import { Paper, Box, Typography, useTheme } from '@mui/material';
import { FC } from 'react';
import { OIDCProviderInfo, SAMLProviderInfo, SSOProvider } from 'js-client-library';
import { Field, FieldsContainer, usePaneStyles, useHeaderStyles } from '../../views/Explore';
import LabelWithCopy from '../LabelWithCopy';

const SAMLProviderInfoPanel: FC<{
    samlProviderDetails: SAMLProviderInfo;
}> = ({ samlProviderDetails }) => (
    <FieldsContainer>
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
    </FieldsContainer>
);

const OIDCProviderInfoPanel: FC<{
    oidcProviderDetails: OIDCProviderInfo;
}> = ({ oidcProviderDetails }) => (
    <FieldsContainer>
        <Field
            label={<LabelWithCopy label='Client ID' valueToCopy={oidcProviderDetails.client_id} hoverOnly />}
            value={oidcProviderDetails.client_id}
        />
        <Field
            label={<LabelWithCopy label='Issuer' valueToCopy={oidcProviderDetails.issuer} hoverOnly />}
            value={oidcProviderDetails.issuer}
        />
    </FieldsContainer>
);

const SSOProviderInfoPanel: FC<{
    ssoProvider: SSOProvider;
}> = ({ ssoProvider }) => {
    const theme = useTheme();
    const paneStyles = usePaneStyles();
    const headerStyles = useHeaderStyles();

    if (!ssoProvider.type) {
        return null;
    }

    let infoPanel;
    switch (ssoProvider.type.toLowerCase()) {
        case 'saml':
            infoPanel = <SAMLProviderInfoPanel samlProviderDetails={ssoProvider.details as SAMLProviderInfo} />;
            break;
        case 'oidc':
            infoPanel = <OIDCProviderInfoPanel oidcProviderDetails={ssoProvider.details as OIDCProviderInfo} />;
            break;
        default:
            infoPanel = null;
    }

    return (
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
                    {infoPanel}
                </Paper>
            </Paper>
        </Box>
    );
};

export default SSOProviderInfoPanel;
