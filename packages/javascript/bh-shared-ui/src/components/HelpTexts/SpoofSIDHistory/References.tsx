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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2003/cc755321(v=ws.10)'>
                Microsoft AD Trust Technical Documentation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2003/cc755321(v=ws.10)#how-sid-history-can-be-used-to-elevate-privileges'>
                How SID History can be used to elevate privileges
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://dirkjanm.io/active-directory-forest-trusts-part-one-how-does-sid-filtering-work/'>
                Active Directory forest trusts part 1 - How does SID filtering work?
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1134/005/'>
                T1134.005: Access Token Manipulation: SID-History Injection
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1558/'>
                T1558: Steal or Forge Kerberos Tickets
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1550/003/'>
                T1550.003: Use Alternate Authentication Material: Pass the Ticket
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://adsecurity.org/?p=1772'>
                Sneaky Active Directory Persistence #14: SID History
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/MichaelGrafnetter/DSInternals/blob/master/Documentation/PowerShell/Add-ADDBSidHistory.md'>
                Add-ADDBSidHistory
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                Rubeus
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/fortra/impacket/blob/master/examples/ticketer.py'>
                ticketer.py
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://www.thehacker.recipes/ad/persistence/sid-history'>
                The Hacker Recipes: SID History
            </Link>
            <br />
        </Box>
    );
};

export default References;
