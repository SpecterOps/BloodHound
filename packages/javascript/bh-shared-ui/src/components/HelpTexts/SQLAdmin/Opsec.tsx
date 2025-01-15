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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Prior to executing operating system commands through SQL Server, review the audit configuration and
                choose a command execution method that is not being monitored.
            </Typography>
            <Typography variant='body2'>View audits:</Typography>
            <Typography component={'pre'}>{'SELECT * FROM sys.dm_server_audit_status'}</Typography>
            <Typography variant='body2'>View server specifications:</Typography>
            <Typography component={'pre'}>
                {'SELECT audit_id, \n' +
                    'a.name as audit_name, \n' +
                    's.name as server_specification_name, \n' +
                    'd.audit_action_name, \n' +
                    's.is_state_enabled, \n' +
                    'd.is_group, \n' +
                    'd.audit_action_id, \n' +
                    's.create_date, \n' +
                    's.modify_date \n' +
                    'FROM sys.server_audits AS a \n' +
                    'JOIN sys.server_audit_specifications AS s \n' +
                    'ON a.audit_guid = s.audit_guid \n' +
                    'JOIN sys.server_audit_specification_details AS d \n' +
                    'ON s.server_specification_id = d.server_specification_id'}
            </Typography>
            <Typography variant='body2'>View database specifications:</Typography>
            <Typography component={'pre'}>
                {'SELECT a.audit_id, \n' +
                    'a.name as audit_name, \n' +
                    's.name as database_specification_name, \n' +
                    'd.audit_action_name, \n' +
                    'd.major_id,\n' +
                    'OBJECT_NAME(d.major_id) as object,\n' +
                    's.is_state_enabled, \n' +
                    'd.is_group, s.create_date, \n' +
                    's.modify_date, \n' +
                    'd.audited_result \n' +
                    'FROM sys.server_audits AS a \n' +
                    'JOIN sys.database_audit_specifications AS s \n' +
                    'ON a.audit_guid = s.audit_guid \n' +
                    'JOIN sys.database_audit_specification_details AS d \n' +
                    'ON s.database_specification_id = d.database_specification_id'}
            </Typography>
            <Typography variant='body2'>
                If server audit specifications are configured on the SQL Server, event ID 15457 logs may be created in
                the Windows Application log when SQL Server level configurations are changed to facilitate OS command
                execution.
            </Typography>
            <Typography variant='body2'>
                If database audit specifications are configured on the SQL Server, event ID 33205 logs may be created in
                the Windows Application log when Agent and database level configuration changes are made.
            </Typography>
            <Typography variant='body2'>
                A summary of the what will show up in the logs, along with the TSQL queries for viewing and configuring
                audit configurations can be found at
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/NetSPI/PowerUpSQL/blob/master/templates/tsql/Audit%20Command%20Execution%20Template.sql'>
                    https://github.com/NetSPI/PowerUpSQL/blob/master/templates/tsql/Audit%20Command%20Execution%20Template.sql
                </Link>
                .
            </Typography>
            <Typography variant='body2'>Author: Scott Sutherland</Typography>
        </>
    );
};

export default Opsec;
