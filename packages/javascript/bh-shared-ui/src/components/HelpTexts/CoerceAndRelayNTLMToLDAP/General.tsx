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
                This edge indicates that the target computer has the WebClient service running. This enables an attacker
                with "Authenticated Users" access to trigger WebClient-based coercion from the target computer to their
                attacker-controlled host via NTLM. Since the connection originates from the WebClient instead of SMB,
                the attacker can relay the authentication attempt to the LDAP service of a domain controller that does
                not require LDAP signing. This relay can be used to abuse Active Directory permissions or obtain
                administrative access to the target computer using Resource-Based Constrained Delegation (RBCD) or
                Shadow Credentials.
            </Typography>

            <Typography variant='body2'>
                Click on Composition to view the domain controllers in the domain that do not require LDAP signing.
            </Typography>
        </>
    );
};

export default General;
