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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>Read the LAPS password attributes listed in the General section.</Typography>
            <Typography variant='body2'>
                Plaintext attributes can be read using a simple LDAP client. For example, with PowerView:
            </Typography>
            <Typography component={'pre'}>
                {'Get-DomainComputer "MachineName" -Properties "cn","ms-mcs-admpwd","ms-mcs-admpwdexpirationtime"'}
            </Typography>

            <Typography variant='body2'>
                Encrypted attributes can be decrypted using Microsoft's LAPS PowerShell module. For example:
            </Typography>
            <Typography component={'pre'}>{'Get-LapsADPassword "WIN10" -AsPlainText'}</Typography>

            <Typography variant='body2'>
                The encrypted attributes can also be retrieved and decrypted using{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/xpn/RandomTSScripts/tree/master/lapsv2decrypt'>
                    lapsv2decrypt
                </Link>{' '}
                (dotnet or BOF).
            </Typography>
        </>
    );
};

export default WindowsAbuse;
