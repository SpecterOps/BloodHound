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

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The user {sourceName} is a SQL admin on the computer {targetName}.
            </Typography>
            <Typography variant='body2'>
                There is at least one MSSQL instance running on {targetName} where the user {sourceName} is the account
                configured to run the SQL Server instance. The typical configuration for MSSQL is to have the local
                Windows account or Active Directory domain account that is configured to run the SQL Server service (the
                primary database engine for SQL Server) have sysadmin privileges in the SQL Server application. As a
                result, the SQL Server service account can be used to log into the SQL Server instance remotely, read
                all of the databases (including those protected with transparent encryption), and run operating systems
                command through SQL Server (as the service account) using a variety of techniques.
            </Typography>
            <Typography variant='body2'>
                For Windows systems that have been joined to an Active Directory domain, the SQL Server instances and
                the associated service account can be identified by executing a LDAP query for a list of "MSSQLSvc"
                Service Principal Names (SPN) as a domain user. In short, when the Database Engine service starts, it
                attempts to register the SPN, and the SPN is then used to help facilitate Kerberos authentication.
            </Typography>
            <Typography variant='body2'>Author: Scott Sutherland</Typography>
        </>
    );
};

export default General;
