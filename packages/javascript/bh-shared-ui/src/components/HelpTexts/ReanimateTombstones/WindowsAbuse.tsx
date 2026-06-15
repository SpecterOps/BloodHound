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

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                The <pre style={{ display: 'inline', width: 'fit-content' }}>Restore-ADObject</pre> PowerShell cmdlet
                can be used to reactivate a deleted object currently in the Active Directory Recycle Bin into a location
                where the user has direct or inherited{' '}
                <pre style={{ display: 'inline', width: 'fit-content' }}>CreateChild</pre> permissions. This includes
                the ability to specify a new relative distinguished name (RDN) for the object, which can be used to
                change the name of the object upon reactivation. This can be abused by an attacker to regain access to a
                recently deleted object, such as a high-privilege user or group, and then use that access to further
                escalate privileges or maintain persistence in the environment.
            </Typography>
            <Typography component={'pre'}>Restore-ADObject -Identity "613dc90a-2afd-49fb-8bd8-eac48c6ab59f"</Typography>
            <Typography variant='body2'>
                If the user does not have `CreateChild` permissions on the original parent container of the deleted
                object, a new parent container can be specified by adding the TargetPath to the command:
            </Typography>
            <Typography component={'pre'}>-TargetPath 'OU=...,DC=domain,DC=local'</Typography>
        </>
    );
};

export default WindowsAbuse;
