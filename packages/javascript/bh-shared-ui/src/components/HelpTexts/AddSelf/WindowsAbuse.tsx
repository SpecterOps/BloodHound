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
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <>
            <Typography variant='body2'>
                There are at least two ways to execute this attack. The first and most obvious is by using the built-in
                net.exe binary in Windows (e.g.: net group "Domain Admins" dfm.a /add /domain). See the opsec
                considerations tab for why this may be a bad idea. The second, and highly recommended method, is by
                using the Add-DomainGroupMember function in PowerView. This function is superior to using the net.exe
                binary in several ways. For instance, you can supply alternate credentials, instead of needing to run a
                process as or logon as the user with the AddMember permission. Additionally, you have much safer
                execution options than you do with spawning net.exe (see the opsec tab).
            </Typography>

            <Typography variant='body2'>
                To abuse this permission with PowerView's Add-DomainGroupMember, first import PowerView into your agent
                session or into a PowerShell instance at the console.{' '}
            </Typography>

            <Typography variant='body2'>
                You may need to authenticate to the Domain Controller as{' '}
                {sourceType === 'User'
                    ? `${sourceName} if you are not running a process as that user`
                    : `a member of ${sourceName} if you are not running a process as a member`}
                . To do this in conjunction with Add-DomainGroupMember, first create a PSCredential object (these
                examples comes from the PowerView help documentation):
            </Typography>

            <Typography component={'pre'}>
                {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                    "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
            </Typography>

            <Typography variant='body2'>
                Then, use Add-DomainGroupMember, optionally specifying $Cred if you are not already running a process as{' '}
                {sourceName}:
            </Typography>

            <Typography component={'pre'}>
                {"Add-DomainGroupMember -Identity 'Domain Admins' -Members 'harmj0y' -Credential $Cred"}
            </Typography>

            <Typography variant='body2'>
                Finally, verify that the user was successfully added to the group with PowerView's
                Get-DomainGroupMember:
            </Typography>

            <Typography component={'pre'}>{"Get-DomainGroupMember -Identity 'Domain Admins'"}</Typography>
        </>
    );
};

export default WindowsAbuse;
