// Copyright 2026 Specter Ops, Inc.
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
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles/identity#domain-services-contributor'>
                Domain Services Contributor built-in role
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/azure/templates/microsoft.aad/domainservices'>
                Microsoft.AAD/domainServices reference
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/entra/identity/domain-services/secure-your-domain'>
                Harden a Microsoft Entra Domain Services managed domain
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/entra/identity/domain-services/scoped-synchronization'>
                Configure scoped synchronization
            </Link>
            <br />
            <Link target='_blank' rel='noopener noreferrer' href='https://attack.mitre.org/techniques/T1484/'>
                MITRE ATT&amp;CK T1484: Domain or Tenant Policy Modification
            </Link>
        </Box>
    );
};

export default References;
