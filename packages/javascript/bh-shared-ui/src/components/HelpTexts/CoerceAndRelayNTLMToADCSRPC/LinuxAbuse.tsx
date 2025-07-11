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

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant={'body2'}>
                1. Start the Relay Server The NTLM relay can be executed with{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'>
                    ntlmrelayx.py
                </Link>
                . To relay to the enterprise CA via RPC endpoints and enroll a certificate, specify the RPC endpoint as
                the target and use the arguments:
            </Typography>
            <Typography component={'pre'}>
                {'-t rpc://<CA_IP> -rpc-mode ICPR -icpr-ca-name <CA_NAME> -smb2support'}
            </Typography>

            <Typography variant={'body2'}>
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
            <Typography variant='body2'>
                3: Authenticate using the certificate obtained as the target principal, for example by using{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/ly4k/Certipy'>
                    Certipy
                </Link>
                .
            </Typography>
        </>
    );
};

export default LinuxAbuse;
