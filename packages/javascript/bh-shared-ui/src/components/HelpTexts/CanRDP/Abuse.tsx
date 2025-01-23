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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const Abuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                Abuse of this privilege will depend heavily on the type of access you have.
            </Typography>
            <Typography variant='body1'>PlainText Credentials with Interactive Access</Typography>
            <Typography variant='body2'>
                With plaintext credentials, the easiest way to exploit this privilege is using the built in Windows
                Remote Desktop Client (mstsc.exe). Open mstsc.exe and input the computer {targetName}. When prompted for
                credentials, input the credentials for{' '}
                {sourceType === 'Group' ? `a member of ${sourceName}` : `${sourceName}`} to initiate the remote desktop
                connection.
            </Typography>
            <Typography variant='body1'>Password Hash with Interactive Access</Typography>
            <Typography variant='body2'>
                With a password hash, exploitation of this privilege will require local administrator privileges on a
                system, and the remote server must allow Restricted Admin Mode.
            </Typography>
            <Typography variant='body2'>
                First, inject the NTLM credential for the user you're abusing into memory using mimikatz:
            </Typography>
            <Typography component={'pre'}>
                {
                    'lsadump::pth /user:dfm /domain:testlab.local /ntlm:&lt;ntlm hash&gt; /run:"mstsc.exe /restrictedadmin"'
                }
            </Typography>
            <Typography variant='body2'>
                This will open a new RDP window. Input the computer {targetName} to initiate the remote desktop
                connection. If the target server does not support Restricted Admin Mode, the session will fail.
            </Typography>
            <Typography variant='body1'>Plaintext Credentials without Interactive Access</Typography>
            <Typography variant='body2'>
                This method will require some method of proxying traffic into the network, such as the socks command in
                cobaltstrike, or direct internet connection to the target network, as well as the xfreerdp (suggested
                because of support of Network Level Authentication (NLA)) tool, which can be installed from the
                freerdp-x11 package. If using socks, ensure that proxychains is configured properly. Initiate the remote
                desktop connection with the following command:
            </Typography>
            <Typography component={'pre'}>
                {'(proxychains) xfreerdp /u:dfm /d:testlab.local /v:<computer ip>'}
            </Typography>
            <Typography variant='body2'>
                xfreerdp will prompt you for a password, and then initiate the remote desktop connection.
            </Typography>
            <Typography variant='body1'>Password Hash without Interactive Access</Typography>
            <Typography variant='body2'>
                This method will require some method of proxying traffic into the network, such as the socks command in
                cobaltstrike, or direct internet connection to the target network, as well as the xfreerdp (suggested
                because of support of Network Level Authentication (NLA)) tool, which can be installed from the
                freerdp-x11 package. Additionally, the target computer must allow Restricted Admin Mode. If using socks,
                ensure that proxychains is configured properly. Initiate the remote desktop connection with the
                following command:
            </Typography>
            <Typography component={'pre'}>
                {'(proxychains) xfreerdp /pth:<ntlm hash> /u:dfm /d:testlab.local /v:<computer ip>'}
            </Typography>
            <Typography variant='body2'>
                This will initiate the remote desktop connection, and will fail if Restricted Admin Mode is not enabled.
            </Typography>
        </>
    );
};

export default Abuse;
