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

import React, { FC } from 'react';
import { Link, Box } from '@mui/material';

const References: FC = () => {
    const references = [
        {
            label: 'Abuse Elevation Control Mechanism',
            link: 'https://attack.mitre.org/techniques/T1548/',
        },
        {
            label: 'ADCS ESC13 Abuse Technique',
            link: 'https://posts.specterops.io/adcs-esc13-abuse-technique-fda4272fbd53',
        },
        {
            label: 'Certified Pre-Owned - Abusing Active Directory Certificate Services',
            link: 'https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf',
        },
        {
            label: 'Certipy',
            link: 'https://github.com/ly4k/Certipy',
        },
        {
            label: 'Certify',
            link: 'https://github.com/GhostPack/Certify',
        },
        {
            label: 'Rubeus',
            link: 'https://github.com/GhostPack/Rubeus',
        },
        {
            label: 'Authentication Mechanism Assurance for AD DS in Windows Server 2008 R2 Step-by-Step Guide',
            link: 'https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/dd378897(v=ws.10)?redirectedfrom=MSDN',
        },
        {
            label: 'Use Authentication Mechanism Assurance (AMA) to secure administrative account logins',
            link: 'https://www.gradenegger.eu/en/using-authentication-mechanism-assurance-ama-to-secure-the-login-of-administrative-accounts/',
        },
    ];
    return (
        <Box sx={{ overflowX: 'auto' }}>
            {references.map((reference) => {
                return (
                    <React.Fragment key={reference.link}>
                        <Link target='_blank' rel='noopener' href={reference.link}>
                            {reference.label}
                        </Link>
                        <br />
                    </React.Fragment>
                );
            })}
        </Box>
    );
};

export default References;
