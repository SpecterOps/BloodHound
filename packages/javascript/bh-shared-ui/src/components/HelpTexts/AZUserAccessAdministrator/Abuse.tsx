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
import { Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                This role can be used to grant yourself or another principal any privilege you want against Automation
                Accounts, VMs, Key Vaults, and Resource Groups. For example, you can make yourself an administrator of 
                an Azure Subscription by assigning the Owner role at the Subscription scope.
            </Typography>

            <Typography variant='body2'>
                The simplest way to execute this attack is to use the Azure portal to add a new, abusable role 
                assignment against the target object for yourself. 
            </Typography>

            <Typography variant='body2'>
                If this role is assigned to a Service Principal, you won't be able to authenticate directly to the 
                Azure portal. In this case:
            </Typography>

            <Typography variant='body2'>
                You'll need to acquire a bearer token for the service principal with AzureRM as the audience. 
                This can be done using BARK's Get-AzureRMTokenWithClientCredentials cmdlet.
                
            </Typography>

            <Typography variant='body2'>
                Using that token, you can make a call to the AzureRM API to create a new role assignment on the
                target object, such as assigning yourself the Owner role. This can be done using BARK's 
                New-AzureRMRoleAssignment cmdlet.
            </Typography>
        </>
    );
};

export default Abuse;
