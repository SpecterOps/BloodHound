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
import { groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the ability to add{' '}
                {sourceType === 'Group' ? 'themselves' : 'itself'}, to the group {targetName}. Because of security group
                delegation, the members of a security group have the same privileges as that group.
            </Typography>

            <Typography variant='body2'>
                By adding itself to the group, {sourceName} will gain the same privileges that {targetName} already has.
            </Typography>
        </>
    );
};

export default General;
