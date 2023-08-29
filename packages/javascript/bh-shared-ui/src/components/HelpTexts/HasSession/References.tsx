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

import { FC } from 'react';
import { Link, Typography, Box } from '@mui/material';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Typography variant='body1'>Gathering Credentials</Typography>
            <Link target='_blank' rel='noopener' href='http://blog.gentilkiwi.com/mimikatz'>
                http://blog.gentilkiwi.com/mimikatz
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/gentilkiwi/mimikatz'>
                https://github.com/gentilkiwi/mimikatz
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://adsecurity.org/?page_id=1821'>
                https://adsecurity.org/?page_id=1821
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/wiki/Credential_Access'>
                https://attack.mitre.org/wiki/Credential_Access
            </Link>

            <Typography variant='body1'>Token Impersonation</Typography>
            <Link
                target='_blank'
                rel='noopener'
                href='https://labs.mwrinfosecurity.com/assets/BlogFiles/mwri-security-implications-of-windows-access-tokens-2008-04-14.pdf'>
                https://labs.mwrinfosecurity.com/assets/BlogFiles/mwri-security-implications-of-windows-access-tokens-2008-04-14.pdf
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/PowerShellMafia/PowerSploit/blob/master/Exfiltration/Invoke-TokenManipulation.ps1'>
                https://github.com/PowerShellMafia/PowerSploit/blob/master/Exfiltration/Invoke-TokenManipulation.ps1
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/wiki/Technique/T1134'>
                https://attack.mitre.org/wiki/Technique/T1134
            </Link>
        </Box>
    );
};

export default References;
