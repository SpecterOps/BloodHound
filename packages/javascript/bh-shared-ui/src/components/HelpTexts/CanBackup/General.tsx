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
import { groupSpecialFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the capability to remotely backup and restore files and registry keys on the computer {targetName}.
            </Typography>

            <Typography variant='body2'>
                Remote Desktop access allows you to enter an interactive session with the target computer. If
                authenticating as a low privilege user, a privilege escalation may allow you to gain high privileges on
                the system.
            </Typography>
            <Typography variant='body2'>Note: This edge does not guarantee privileged execution.</Typography>
        </>
    );
};

export default General;
