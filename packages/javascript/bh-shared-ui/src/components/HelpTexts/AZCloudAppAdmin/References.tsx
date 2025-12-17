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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Typography variant='body2'>
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#cloud-application-administrator'>
                Microsoft Entra built-in roles: Cloud Application Administrator
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener noreferrer'
                href='https://dirkjanm.io/azure-ad-privilege-escalation-application-admin/'>
                https://dirkjanm.io/azure-ad-privilege-escalation-application-admin/
            </Link>
        </Typography>
    );
};

export default References;
