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

const Abuse: FC<EdgeInfoProps> = ({ targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The PowerShell script Invoke-DCOM implements lateral movement using a variety of different COM objects
                (ProgIds: MMC20.Application, ShellWindows, ShellBrowserWindow, ShellBrowserWindow, and ExcelDDE).
                LethalHTA implements lateral movement using the HTA COM object (ProgId: htafile).
            </Typography>

            <Typography variant='body2'>
                One can manually instantiate and manipulate COM objects on a remote machine using the following
                PowerShell code. If specifying a COM object by its CLSID:
            </Typography>
            <Typography component={'pre'}>
                {`$ComputerName = ${targetName}  # Remote computer\n` +
                    '$clsid = "{fbae34e8-bf95-4da8-bf98-6c6e580aa348}"      # GUID of the COM object\n' +
                    '$Type = [Type]::GetTypeFromCLSID($clsid, $ComputerName)\n' +
                    '$ComObject = [Activator]::CreateInstance($Type)'}
            </Typography>
            <Typography variant='body2'>If specifying a COM object by its ProgID:</Typography>
            <Typography component={'pre'}>
                {`$ComputerName = ${targetName}  # Remote computer\n` +
                    '$ProgId = "<NAME>"      # GUID of the COM object\n' +
                    '$Type = [Type]::GetTypeFromProgID($ProgId, $ComputerName)\n' +
                    '$ComObject = [Activator]::CreateInstance($Type)'}
            </Typography>
        </>
    );
};

export default Abuse;
