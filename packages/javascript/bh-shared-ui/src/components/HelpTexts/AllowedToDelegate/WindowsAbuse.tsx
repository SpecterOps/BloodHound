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

import { FC } from 'react';
import { Typography } from '@mui/material';

const WindowsAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Abusing this privilege can utilize Benjamin Delpy's Kekeo project, proxying in traffic generated from
                the Impacket library, or using the Rubeus project's s4u abuse.
            </Typography>

            <Typography variant='body2'>
                In the following example, *victim* is the attacker-controlled account (i.e. the hash is known) that is
                configured for constrained delegation. That is, *victim* has the "HTTP/PRIMARY.testlab.local" service
                principal name (SPN) set in its msds-AllowedToDelegateTo property. The command first requests a TGT for
                the *victim* user and executes the S4U2self/S4U2proxy process to impersonate the "admin" user to the
                "HTTP/PRIMARY.testlab.local" SPN. The alternative sname "cifs" is substituted in to the final service
                ticket and the ticket is submitted to the current logon session. This grants the attacker the ability to
                access the file system of PRIMARY.testlab.local as the "admin" user.
            </Typography>

            <Typography component={'pre'}>
                {
                    'Rubeus.exe s4u /user:victim /rc4:2b576acbe6bcfda7294d6bd18041b8fe /impersonateuser:admin /msdsspn:"HTTP/PRIMARY.testlab.local" /altservice:cifs /ptt'
                }
            </Typography>
        </>
    );
};

export default WindowsAbuse;
