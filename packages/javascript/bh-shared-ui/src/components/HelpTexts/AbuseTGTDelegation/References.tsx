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
                href='https://posts.specterops.io/not-a-security-boundary-breaking-forest-trusts-cd125829518d'>
                Not A Security Boundary: Breaking Forest Trusts
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/hunting-in-active-directory-unconstrained-delegation-forests-trusts-71f2b33688e1'>
                Hunting in Active Directory: Unconstrained Delegation & Forests Trusts
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://exploit.ph/user-constrained-delegation.html'>
                Abusing Users Configured with Unconstrained Delegation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://blog.improsec.com/tech-blog/sid-filter-as-security-boundary-between-domains-part-7-trust-account-attack-from-trusting-to-trusted'>
                SID filter as security boundary between domains? (Part 7) - Trust account attack - from trusting to
                trusted
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://dirkjanm.io/krbrelayx-unconstrained-delegation-abuse-toolkit/'>
                “Relaying” Kerberos - Having fun with unconstrained delegation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                Windows Coerced Authentication Methods
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                Rubeus
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/leechristensen/SpoolSample'>
                SpoolSample
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/gentilkiwi/mimikatz'>
                mimikatz
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/dirkjanm/krbrelayx/'>
                krbrelayx.py
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/dirkjanm/krbrelayx/blob/master/printerbug.py'>
                printerbug.py
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/fortra/impacket/blob/master/examples/secretsdump.py'>
                secretsdump.py
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://support.microsoft.com/en-us/topic/updates-to-tgt-delegation-across-incoming-trusts-in-windows-server-1a6632ac-1599-0a7c-550a-a754796c291e'>
                Updates to TGT delegation across incoming trusts in Windows Server
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2003/cc755321(v=ws.10)'>
                Microsoft AD Trust Technical Documentation
            </Link>
        </Box>
    );
};

export default References;
