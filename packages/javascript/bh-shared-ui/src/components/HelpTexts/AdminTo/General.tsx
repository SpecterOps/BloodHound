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
import { groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} admin rights to the computer {targetName}.
            </Typography>

            <Typography variant='body2'>
                By default, administrators have several ways to perform remote code execution on Windows systems,
                including via RDP, WMI, WinRM, the Service Control Manager, and remote DCOM execution.
            </Typography>

            <Typography variant='body2'>
                Further, administrators have several options for impersonating other users logged onto the system,
                including plaintext password extraction, token impersonation, and injecting into processes running as
                another user.
            </Typography>

            <Typography variant='body2'>
                Finally, administrators can often disable host-based security controls that would otherwise prevent the
                aforementioned techniques.
            </Typography>
        </>
    );
};

export default General;
