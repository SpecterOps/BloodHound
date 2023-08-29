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

const WindowsAbuse: FC<{ sourceName: string }> = ({ sourceName }) => {
    return (
        <>
            <Typography variant='body2'>
                Abusing this primitive is currently only possible through the Rubeus project.
            </Typography>

            <Typography variant='body2'>
                To use this attack, the controlled account MUST have a service principal name set, along with access to
                either the plaintext or the RC4_HMAC hash of the account.
            </Typography>

            <Typography variant='body2'>
                If the plaintext password is available, you can hash it to the RC4_HMAC version using Rubeus:
            </Typography>

            <Typography component={'pre'}>{'Rubeus.exe hash /password:Summer2018!'}</Typography>

            <Typography variant='body2'>
                Use Rubeus' *s4u* module to get a service ticket for the service name (sname) we want to "pretend" to be
                "admin" for. This ticket is injected (thanks to /ptt), and in this case grants us access to the file
                system of the TARGETCOMPUTER:
            </Typography>

            <Typography component={'pre'}>
                {`Rubeus.exe s4u /user:${sourceName}$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt`}
            </Typography>
        </>
    );
};

export default WindowsAbuse;
