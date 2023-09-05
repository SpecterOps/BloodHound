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
import { typeFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {targetName} is a Group Managed Service Account. The {typeFormat(sourceType)} {sourceName} can retrieve
                the password for the GMSA {targetName}.
            </Typography>
            <Typography variant='body2'>
                Group Managed Service Accounts are a special type of Active Directory object, where the password for
                that object is mananaged by and automatically changed by Domain Controllers on a set interval (check the
                MSDS-ManagedPasswordInterval attribute).
            </Typography>
            <Typography variant='body2'>
                The intended use of a GMSA is to allow certain computer accounts to retrieve the password for the GMSA,
                then run local services as the GMSA. An attacker with control of an authorized principal may abuse that
                privilege to impersonate the GMSA.
            </Typography>
        </>
    );
};

export default General;
