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

const Abuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                To abuse this privilege with DirSync, first import DirSync into your agent session or into a PowerShell
                instance at the console. You must authenticate to the Domain Controller as{' '}
                {groupSpecialFormat(sourceType, sourceName)}. Then, execute the{' '}
                <Typography component={'pre'}>Sync-LAPS</Typography> function:
            </Typography>

            <Typography component={'pre'}>Sync-LAPS -LDAPFilter &quot;(samaccountname={targetName})&quot;</Typography>

            <Typography variant='body2'>
                You can target a specific domain controller using the <Typography component={'pre'}>-Server</Typography>{' '}
                parameter.
            </Typography>
        </>
    );
};

export default Abuse;
