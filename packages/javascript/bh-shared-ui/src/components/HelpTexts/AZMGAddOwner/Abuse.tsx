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
                You can use BARK to add a new owner to the target object. The BARK function you use will depend on the
                target object type, but all of the functions follow a similar syntax.
            </Typography>

            <Typography variant='body2'>
                These functions require you to supply an MS Graph-scoped JWT associated with the Service Principal that
                has the privilege to add a new owner to the target object. There are several ways to acquire a JWT. For
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
                To add a new owner to a Service Principal, use BARK's New-ServicePrincipalOwner function:
            </Typography>

            <Typography component={'pre'}>
                {'New-ServicePrincipalOwner `\n' +
                    '    -ServicePrincipalObjectId "082cf9b3-24e2-427b-bcde-88ffdccb5fad" `\n' +
                    '    -NewOwnerObjectId "cea271c4-7b01-4f57-932d-99d752bbbc60" `\n' +
                    '    -Token $Token'}
            </Typography>

            <Typography variant='body2'>
                To add a new owner to an App Registration, use BARK's New-AppOwner function:
            </Typography>

            <Typography component={'pre'}>
                {'New-AppOwner `\n' +
                    '    -AppObjectId "52114a0d-fa5b-4ee5-9a29-2ba048d46eee" `\n' +
                    '    -NewOwnerObjectId "cea271c4-7b01-4f57-932d-99d752bbbc60" `\n' +
                    '    -Token $Token'}
            </Typography>

            <Typography variant='body2'>To add a new owner to a Group, use BARK's New-GroupOwner function:</Typography>

            <Typography component={'pre'}>
                {'New-AppOwner `\n' +
                    '    -GroupObjectId "352032bf-161d-4788-b77c-b6f935339770" `\n' +
                    '    -NewOwnerObjectId "cea271c4-7b01-4f57-932d-99d752bbbc60" `\n' +
                    '    -Token $Token'}
            </Typography>
        </>
    );
};

export default Abuse;
