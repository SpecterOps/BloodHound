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
                net.exe binary in Windows (e.g.: net user dfm.a Password123! /domain). See the opsec considerations tab
                for why this may be a bad idea. The second, and highly recommended method, is by using the
                Set-DomainUserPassword function in PowerView. This function is superior to using the net.exe binary in
                several ways. For instance, you can supply alternate credentials, instead of needing to run a process as
                or logon as the user with the ForceChangePassword permission. Additionally, you have much safer
                execution options than you do with spawning net.exe (see the opsec tab).
            </Typography>
            <Typography variant='body2'>
                To abuse this permission with PowerView's Set-DomainUserPassword, first import PowerView into your agent
                session or into a PowerShell instance at the console. You may need to authenticate to the Domain
                Controller as
                {sourceType === 'User'
                    ? `${sourceName} if you are not running a process as that user`
                    : `a member of ${sourceName} if you are not running a process as a member`}
                . To do this in conjunction with Set-DomainUserPassword, first create a PSCredential object (these
                examples comes from the PowerView help documentation):
            </Typography>
            <Typography component={'pre'}>
                {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                    "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
            </Typography>
            <Typography variant='body2'>
                Then create a secure string object for the password you want to set on the target user:
            </Typography>
            <Typography component={'pre'}>
                {"$UserPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force"}
            </Typography>
            <Typography variant='body2'>
                Finally, use Set-DomainUserPassword, optionally specifying $Cred if you are not already running a
                process as {sourceName}:
            </Typography>
            <Typography component={'pre'}>
                {'Set-DomainUserPassword -Identity andy -AccountPassword $UserPassword -Credential $Cred'}
            </Typography>
            <Typography variant='body2'>
                Now that you know the target user's plain text password, you can either start a new agent as that user,
                or use that user's credentials in conjunction with PowerView's ACL abuse functions, or perhaps even RDP
                to a system the target user has access to. For more ideas and information, see the references tab.
            </Typography>
        </>
    );
};

export default WindowsAbuse;
