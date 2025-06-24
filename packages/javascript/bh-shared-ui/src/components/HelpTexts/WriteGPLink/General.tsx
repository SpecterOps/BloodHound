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

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} has the permissions to modify the gPLink attribute of{' '}
                {targetType} {targetName}.
            </Typography>

            <Typography variant='body2'>
                The ability to alter the gPLink attribute may allow an attacker to apply a malicious Group Policy Object
                (GPO) to all child user and computer objects (including the ones located in nested OUs). This can be
                exploited to make said child objects execute arbitrary commands through an immediate scheduled task,
                thus compromising them.
            </Typography>
        </>
    );
};

export default General;
