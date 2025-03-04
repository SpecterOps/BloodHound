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
            <Link target='_blank' rel='noopener' href='https://en.hackndo.com/ntlm-relay/'>
                Hackndo: NTLM relay
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/windows-server/security/kerberos/ntlm-overview'>
                Microsoft: NTLM Overview
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/relay-your-heart-away-an-opsec-conscious-approach-to-445-takeover-1c9b4666c8ac'>
                Relay Your Heart Away: An OPSEC-Conscious Approach to 445 Takeover
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/Kevin-Robertson/Inveigh'>
                Inveigh
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                Windows Coerced Authentication Methods
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/topotam/PetitPotam'>
                PetitPotam
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/leechristensen/SpoolSample'>
                SpoolSample
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.guidepointsecurity.com/blog/beyond-the-basics-exploring-uncommon-ntlm-relay-attack-techniques/'>
                Beyond the Basics: Exploring Uncommon NTLM Relay Attack Techniques
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/dirkjanm/krbrelayx/blob/master/printerbug.py'>
                printerbug.py
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://trustedsec.com/blog/a-comprehensive-guide-on-relaying-anno-2022'>
                I’m bringing relaying back: A comprehensive guide on relaying anno 2022
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'>
                ntlmrelayx.py
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://support.microsoft.com/en-us/topic/2020-2023-and-2024-ldap-channel-binding-and-ldap-signing-requirements-for-windows-kb4520412-ef185fb8-00f7-167d-744c-f299a66fc00a'>
                2020, 2023, and 2024 LDAP channel binding and LDAP signing requirements for Windows (KB4520412)
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.bluraven.io/detecting-ntlm-relay-attacks-d92e99e68fb9'>
                Detecting NTLM Relay Attacks
            </Link>
        </Box>
    );
};

export default References;
