// Copyright 2025 Specter Ops, Inc.
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

const General: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                This edge indicates that an attacker with "Authenticated Users" access can trigger SMB-based coercion
                from the target computer to their attacker-controlled host via NTLM. The authentication attempt from the
                target computer can then be relayed to an ESC11-vulnerable RPC enrollment endpoint of an Active
                Directory Certificate Services (ADCS) enterprise CA server. This allows the attacker to obtain a
                certificate enabling domain authentication as the target computer.
            </Typography>

            <Typography variant='body2'>
                The ESC11 attack differs from the HTTP-based ADCS relay attack (ESC8) because it targets RPC endpoints
                instead of web enrollment endpoints. For ESC11 to succeed, the Enterprise CA must have RPC encryption
                disabled - specifically, the <code>IF_ENFORCEENCRYPTICERTREQUEST</code> CA flag must be turned off. In
                BloodHound, this configuration is revealed by the "RPC Encryption Enforced" (
                <code>rpcencryptionenforced</code>) property on the EnterpriseCA node.
            </Typography>

            <Typography variant='body2'>
                Click on Relay Targets to view vulnerable enterprise CA servers that enable certificate enrollment for
                the target computer via RPC endpoints.
            </Typography>
        </>
    );
};

export default General;
