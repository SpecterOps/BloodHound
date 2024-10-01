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
    Paper,
    Box,
    Typography
} from '@mui/material';
import { faAngleDown, faAngleUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React, { useState } from 'react';
import Icon from '../../components/Icon';
import { Field, FieldsContainer, usePaneStyles, useHeaderStyles } from '../../views/Explore';

const SAMLProviderInfoPanel: React.FC<{
    ssoProvider?: any;
}> = ({ ssoProvider }) => {
    return (
    <FieldsContainer>
        <Field label={'IdP SSO URL'} value={ssoProvider.idp_sso_uri}></Field>
        <Field label={'BHE SSO URL'} value={ssoProvider.sp_sso_uri}></Field>
        <Field label={'BHE ACS URL'} value={ssoProvider.sp_acs_uri}></Field>
        <Field label={'BHE Metadata URL'} value={ssoProvider.sp_metadata_uri}></Field>
    </FieldsContainer>
    )
}

const SSOProviderInfoPanel: React.FC<{
    ssoProvider?: any;
}> = ({ ssoProvider }) => {
    const paneStyles = usePaneStyles();
    const headerStyles = useHeaderStyles();
    const [expanded, setExpanded] = useState(true);

    if (!ssoProvider) {
        return null;
    }
    console.log(ssoProvider)
    return (
        <Box className={paneStyles.container} data-testid='sso_provider-info-panel'>
            <Paper elevation={0} classes={{ root: paneStyles.headerPaperRoot }}>
            <Box className={headerStyles.header}>
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
                style={{
                    display: expanded ? 'initial' : 'none',
                }}>
                    <Box flexShrink={0} flexGrow={1} fontWeight='bold' mr={1} fontSize={'small'}>
                        Provider Information:
                    </Box>
                    <SAMLProviderInfoPanel ssoProvider={ssoProvider} />
            </Paper>
        </Box>
    );
};

export default SSOProviderInfoPanel;
