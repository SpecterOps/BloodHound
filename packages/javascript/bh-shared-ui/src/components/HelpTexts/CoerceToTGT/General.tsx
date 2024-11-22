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

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <>
            <Typography variant='body2'>
                The {typeFormat(sourceType)} {sourceName} is configured with Kerberos unconstrained delegation.
            </Typography>

            <Typography variant='body2'>
                Users and computers authenticating against {sourceName} will have their Kerberos TGT sent to{' '}
                {sourceName}, unless they are marked as sensitive or members of Protected Users.
            </Typography>

            <Typography variant='body2'>
                An attacker with control over {sourceName} can coerce a Tier Zero computer (e.g. DC) to authenticate
                against {sourceName} and obtain the target's TGT. With the TGT of a DC, the attacker can perform DCSync
                to compromise the domain. Alternatively, the TGT can be used to obtain admin access to the target host
                with a shadow credentials + silver ticket attack or a resource-based constrained delegation attack.
            </Typography>
        </>
    );
};

export default General;
