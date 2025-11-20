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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                The registry paths targetted by commodity Backup Operator tooling do not have SACLs configured on the keys by default.
                Advanced Audit Policy settings required to generate a 4672 Event ID related to special logon privileges are not enabled
                by default. Creating a Windows service by manually manipulation of the Windows registry does not appear to generate a 
                4697 Event ID for service creation upon system reboot.
            </Typography>
            <Typography variant='body2'>
                Connecting to the Remote Registry named pipe or Admin Shares over SMB will generate a logon event on the target computer.
            </Typography>
        </>
    );
};

export default Opsec;
