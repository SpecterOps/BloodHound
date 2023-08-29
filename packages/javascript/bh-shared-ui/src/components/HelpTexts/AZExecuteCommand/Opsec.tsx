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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                When the Intune agent pulls down and executes PowerShell scripts, a number of artifacts are created on
                the endpoint — some permanent and some ephemeral.
            </Typography>

            <Typography variant='body2'>
                Two files are created on the endpoint when a PowerShell script is executed in the following locations: -
                C:\Program files (x86)\Microsoft Intune Management Extension\Policies\Scripts - C:\Program files
                (x86)\Microsoft Intune Management Extension\Policies\Results
            </Typography>

            <Typography variant='body2'>
                The file under the “Scripts” folder will be a local copy of the PS1 stored in Azure, and the file under
                the “Results” folder will be the output of the PS1; however, both of these files are automatically
                deleted as soon as the script finishes running.
            </Typography>

            <Typography variant='body2'>
                There are also permanent artifacts left over (assuming the attacker doesn’t tamper with them). A full
                copy of the contents of the PS1 can be found in this log file: -
                C:\ProgramData\Microsoft\IntuneManagementExtension\Logs\_IntuneManagementExtension.txt
            </Typography>
        </>
    );
};

export default Opsec;
