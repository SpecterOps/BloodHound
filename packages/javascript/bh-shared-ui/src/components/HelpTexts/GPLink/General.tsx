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
import { typeFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                The GPO {sourceName} is linked to the {typeFormat(targetType)} {targetName}.
            </Typography>
            <Typography variant='body2'>
                The GPO's settings apply to AD accounts (users and computers) within {targetName}.
            </Typography>
            <Typography variant='body2'>
                The Enforced property on the edge indicates whether the GPO link is enforced, meaning it still applies
                if an OU has blocked GPO inheritance.
            </Typography>
        </>
    );
};

export default General;
