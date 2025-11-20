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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';
import { groupSpecialFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the capability to remotely backup and restore files and
                registry keys on the computer {targetName}.
            </Typography>

            <Typography variant='body2'>
                The Backup Operators built-in group, and its members, by default are granted the SeBackupPrivilege and 
                SeRestorePrivilege rights on the host. These privileges allow users to access all files and registry 
                keys on the host, regardless of their permission, through backup and restore operations.
            </Typography>

            <Typography variant='body2'>
                In Active Directory, the Backup Operators AD group is granted the same user rights assignment on all 
                the domain controllers by default, allowing all the Backup Operators AD group members to compromise 
                domain controllers in various ways and gain domain dominance.
            </Typography>

            <Typography variant='body2'>
                Backup Operators are granted allow ACEs on the admin shares and remote registry named pipe by default,
                which is why Backup Operators can manipulate the registry and file system of a host remotely, whereas 
                other principals which are assigned the SeBackupPrivilege and SeRestorePrivilege cannot.
            </Typography>
        </>
    );
};

export default General;
