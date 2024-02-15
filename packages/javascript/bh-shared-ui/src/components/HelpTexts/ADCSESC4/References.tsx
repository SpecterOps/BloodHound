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
            label: 'Steal or Forge Authentication Certificates',
            link: 'https://attack.mitre.org/techniques/T1649/',
        },
        {
            label: 'Vulnerable Certificate Template Access Control - ESC4',
            link: 'https://book.hacktricks.xyz/windows-hardening/active-directory-methodology/ad-certificates/domain-escalation#vulnerable-certificate-template-access-control-esc4',
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
            label: 'Impacket',
            link: 'https://github.com/fortra/impacket',
        },
        {
            label: 'Rubeus',
            link: 'https://github.com/GhostPack/Rubeus',
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
