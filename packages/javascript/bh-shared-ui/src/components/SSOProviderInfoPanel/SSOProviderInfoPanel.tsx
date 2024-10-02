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
import { faAngleDown, faAngleUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React, { useState } from 'react';
import Icon from '../../components/Icon';
import { Field, FieldsContainer, usePaneStyles, useHeaderStyles } from '../../views/Explore';
import LabelWithCopy from '../LabelWithCopy';

const SAMLProviderInfoPanel: React.FC<{
    ssoProvider?: any;
}> = ({ ssoProvider }) => {
    return (
        <FieldsContainer>
            <Field
                label={<LabelWithCopy label='IdP SSO URL' valueToCopy={ssoProvider.idp_sso_uri} hoverOnly />}
                value={ssoProvider.idp_sso_uri}
            />
            <Field
                label={<LabelWithCopy label='BHE SSO URL' valueToCopy={ssoProvider.sp_sso_uri} hoverOnly />}
                value={ssoProvider.sp_sso_uri}
            />
            <Field
                label={<LabelWithCopy label='BHE ACS URL' valueToCopy={ssoProvider.sp_acs_uri} hoverOnly />}
                value={ssoProvider.sp_acs_uri}
            />
            <Field
                label={<LabelWithCopy label='BHE Metadata URL' valueToCopy={ssoProvider.sp_metadata_uri} hoverOnly />}
                value={ssoProvider.sp_metadata_uri}
            />
        </FieldsContainer>
    );
};

const SSOProviderInfoPanel: React.FC<{
    ssoProvider?: any;
}> = ({ ssoProvider }) => {
    const theme = useTheme();
    const paneStyles = usePaneStyles();
    const headerStyles = useHeaderStyles();
    const [expanded, setExpanded] = useState(true);

    if (!ssoProvider) {
        return null;
    }

    var infoPanel;
    switch (ssoProvider.type) {
        case 1:
            infoPanel = <SAMLProviderInfoPanel ssoProvider={ssoProvider} />;
            break;
        default:
            infoPanel = <SAMLProviderInfoPanel ssoProvider={ssoProvider} />;
    }

    return (
        <Box className={paneStyles.container} data-testid='sso_provider-info-panel'>
            <Paper elevation={0} classes={{ root: paneStyles.headerPaperRoot }}>
                <Box className={headerStyles.header}>
                    <Box
                        sx={{
                            backgroundColor: theme.palette.primary.main,
                            width: 10,
                            height: theme.spacing(5),
                        }}
                    />
                    <Icon
                        className={headerStyles.icon}
                        click={() => {
                            setExpanded(!expanded);
                        }}>
                        <FontAwesomeIcon icon={expanded ? faAngleUp : faAngleDown} />
                    </Icon>

                    <Typography
                        data-testid='sso_provider-info-panel_header-text'
                        variant='h5'
                        noWrap
                        className={headerStyles.headerText}>
                        {ssoProvider?.name}
                    </Typography>
                </Box>
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: paneStyles.contentPaperRoot }}
                sx={{ display: expanded ? 'initial' : 'none' }}>
                <Box flexShrink={0} flexGrow={1} fontWeight='bold' mr={1} fontSize={'small'}>
                    Provider Information:
                </Box>
                {infoPanel}
            </Paper>
        </Box>
    );
};

export default SSOProviderInfoPanel;
