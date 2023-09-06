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
                You can abuse this privilege using BARK's Add-AZMemberToGroup function.
            </Typography>

            <Typography variant='body2'>
                This function requires you to supply an MS Graph-scoped JWT associated with the Service Principal that
                has the privilege to add principal to the target group. There are several ways to acquire a JWT. For
                example, you may use BARK’s Get-MSGraphTokenWithClientCredentials to acquire an MS Graph-scoped JWT by
                supplying a Service Principal Client ID and secret:
            </Typography>

            <Typography component={'pre'}>
                {'$MGToken = Get-MSGraphTokenWithClientCredentials `\n' +
                    '    -ClientID "34c7f844-b6d7-47f3-b1b8-720e0ecba49c" `\n' +
                    '    -ClientSecret "asdf..." `\n' +
                    '    -TenantName "contoso.onmicrosoft.com"'}
            </Typography>

            <Typography variant='body2'>
                Then use BARK’s Add-AZMemberToGroup function to add a new principial to the target group:
            </Typography>

            <Typography component={'pre'}>
                {'Add-AZMemberToGroup `\n' +
                    '    -PrincipalID = "028362ca-90ae-41f2-ae9f-1a678cc17391" `\n' +
                    '    -TargetGroupId "b9801b7a-fcec-44e2-a21b-86cb7ec718e4" `\n' +
                    '    -Token $MGToken.access_token'}
            </Typography>

            <Typography variant='body2'>
                Now you can re-authenticate as the principial you just added to the group and continue your attack path,
                now having whatever privileges the target group has.
            </Typography>
        </>
    );
};

export default Abuse;
