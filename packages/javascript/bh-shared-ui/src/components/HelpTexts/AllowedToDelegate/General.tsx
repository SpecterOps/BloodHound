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
                The {typeFormat(sourceType)} {sourceName} has the constrained delegation permission to the computer{' '}
                {targetName}.
            </Typography>

            <Typography variant='body2'>
                The constrained delegation primitive allows a principal to authenticate as any user to specific services
                (found in the msds-AllowedToDelegateTo LDAP property in the source node tab) on the target computer.
                That is, a node with this permission can impersonate any domain principal (including Domain Admins) to
                the specific service on the target host. One caveat- impersonated users can not be in the "Protected
                Users" security group or otherwise have delegation privileges revoked.
            </Typography>

            <Typography variant='body2'>
                An issue exists in the constrained delegation where the service name (sname) of the resulting ticket is
                not a part of the protected ticket information, meaning that an attacker can modify the target service
                name to any service of their choice. For example, if msds-AllowedToDelegateTo is "HTTP/host.domain.com",
                tickets can be modified for LDAP/HOST/etc. service names, resulting in complete server compromise,
                regardless of the specific service listed.
            </Typography>
        </>
    );
};

export default General;
