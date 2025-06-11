// Copyright 2025 Specter Ops, Inc.
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
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1098/'>
                MITRE ATT&CK - Account Manipulation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/adcs-esc14-abuse-technique-333a004dc2b9'>
                ADCS ESC14 Abuse Technique
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'>
                Certified Pre-Owned - Abusing Active Directory Certificate Services
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/ff520074(v=ws.10)'>
                How to disable the Subject Alternative Name for UPN mapping
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/ly4k/Certipy'>
                Certipy
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Certify'>
                Certify
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                Rubeus
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/JonasBK/Powershell/blob/master/Add-AltSecIDMapping.ps1'>
                Add-AltSecIDMapping.ps1
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/JonasBK/Powershell/blob/master/Get-AltSecIDMapping.ps1'>
                Get-AltSecIDMapping.ps1
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/JonasBK/Powershell/blob/master/Remove-AltSecIDMapping.ps1'>
                Remove-AltSecIDMapping.ps1
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://linux.die.net/man/1/ldapsearch'>
                ldapsearch
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://linux.die.net/man/1/ldapmodify'>
                ldapmodify
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/certutil'>
                certutil
            </Link>
        </Box>
    );
};

export default References;
