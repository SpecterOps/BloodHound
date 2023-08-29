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
                The artifacts generated when using DCOM vary depending on the specific COM object used.
            </Typography>
            <Typography variant='body2'>
                DCOM is built on top of the TCP/IP RPC protocol (TCP ports 135 + high ephemeral ports) and may leverage
                several different RPC interface UUIDs(outlined here). In order to use DCOM, one must be authenticated.
                Consequently, logon events and authentication-specific logs(Kerberos, NTLM, etc.) will be generated when
                using DCOM.
            </Typography>
            <Typography variant='body2'>
                Processes may be spawned as the user authenticating to the remote system, as a user already logged into
                the system, or may take advantage of an already spawned process.
            </Typography>
            <Typography variant='body2'>
                Many DCOM servers spawn under the process "svchost.exe -k DcomLaunch" and typically have a command line
                containing the string " -Embedding" or are executing inside of the DLL hosting process "DllHost.exe
                /Processid:{'{<AppId>}'}" (where AppId is the AppId the COM object is registered to use). Certain COM
                services are implemented as service executables; consequently, service-related event logs may be
                generated.
            </Typography>
        </>
    );
};

export default Opsec;
