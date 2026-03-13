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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box className='overflow-x-auto'>
            <Link target='_blank' rel='noopener noreferrer' href='https://attack.mitre.org/techniques/T1098/'>
                ATT&amp;CK T1098: Account Manipulation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://posts.specterops.io/azure-privilege-escalation-via-service-principal-abuse-210ae2be2a5'>
                Andy Robbins - Azure Privilege Escalation via Service Principal Abuse
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://docs.microsoft.com/en-us/azure/active-directory/roles/assign-roles-different-scopes'>
                Assign Azure AD roles at different scopes
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/entra/identity-platform/howto-configure-app-instance-property-locks'>
                Protect Service Principals using App Instance Property Lock
            </Link>
        </Box>
    );
};

export default References;
