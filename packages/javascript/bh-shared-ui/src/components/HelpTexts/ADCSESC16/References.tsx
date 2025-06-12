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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/ly4k/Certipy/wiki/06-%E2%80%90-Privilege-Escalation#esc16-security-extension-disabled-on-ca-globally'>
                ESC16: Security Extension Disabled on CA (Globally)
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/adcs-attack-paths-in-bloodhound-part-3-33efb00856ac'>
                ADCS Attack Paths in BloodHound — Part 3
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'>
                Certified Pre-Owned
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://research.ifcr.dk/certipy-4-0-esc9-esc10-bloodhound-gui-new-authentication-and-request-methods-and-more-7237d88061f7'>
                Certipy 4.0
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://book.hacktricks.xyz/windows-hardening/active-directory-methodology/ad-certificates/domain-escalation#editf_attributesubjectaltname2-esc6'>
                Domain Escalation Edit Attributes
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://techcommunity.microsoft.com/t5/ask-the-directory-services-team/preview-of-san-uri-for-certificate-strong-mapping-for-kb5014754/ba-p/3789785'>
                Preview of SAN URI for Certificate Strong Mapping for KB5014754
            </Link>
        </Box>
    );
};

export default References;
