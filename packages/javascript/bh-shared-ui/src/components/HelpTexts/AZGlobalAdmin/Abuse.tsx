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
import { Link, Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                As a Global Admin, you can change passwords, run commands on VMs, read key vault secrets, activate roles
                for other users, etc.
            </Typography>
            <Typography variant='body2'>Via PowerZure</Typography>
            <Link
                target='_blank'
                rel='noopener'
                href='https://powerzure.readthedocs.io/en/latest/Functions/operational.html'>
                https://powerzure.readthedocs.io/en/latest/Functions/operational.html
            </Link>
            <Typography variant='body2'>
                For Global Admin to be able to abuse Azure resources, you must first grant yourself the 'User Access
                Administrator' role in Azure RBAC. This is done through a toggle button in the portal, or via the
                PowerZure function Set-AzureElevatedPrivileges.
            </Typography>
            <Typography variant='body2'>
                Once that role is applied to account, you can then add yourself as an Owner to all subscriptions in the
                tenant
            </Typography>
        </>
    );
};
export default Abuse;
