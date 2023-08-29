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
import { Link, Box } from '@mui/material';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link target='_blank' rel='noopener' href='https://github.com/NetSPI/PowerUpSQL/wiki'>
                https://github.com/NetSPI/PowerUpSQL/wiki
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.slideshare.net/nullbind/powerupsql-2018-blackhat-usa-arsenal-presentation'>
                https://www.slideshare.net/nullbind/powerupsql-2018-blackhat-usa-arsenal-presentation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://sqlwiki.netspi.com/attackQueries/executingOSCommands/#sqlserver'>
                https://sqlwiki.netspi.com/attackQueries/executingOSCommands/#sqlserver
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-windows-service-accounts-and-permissions?view=sql-server-2017'>
                https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-windows-service-accounts-and-permissions?view=sql-server-2017
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://blog.netspi.com/finding-sensitive-data-domain-sql-servers-using-powerupsql/'>
                https://blog.netspi.com/finding-sensitive-data-domain-sql-servers-using-powerupsql/
            </Link>
        </Box>
    );
};

export default References;
