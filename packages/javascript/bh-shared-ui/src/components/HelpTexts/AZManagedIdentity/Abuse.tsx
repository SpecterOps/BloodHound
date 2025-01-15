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
                You can modify the Azure RM resource to execute actions against Azure with the privileges of the Managed
                Identity Service Principal.
            </Typography>
            <Typography variant='body2'>
                It is also possible to extract a JSON Web Token (JWT) for the Service Principal, then use that JWT to
                authenticate as the Service Principal outside the scope of the Azure RM resource. Here is how you
                extract the JWT using PowerShell:
            </Typography>
            <Typography component={'pre'}>
                {'$tokenAuthURI = $env:MSI_ENDPOINT + "?resource=https://graph.microsoft.com/&api-version=2017-09-01"\n' +
                    '$tokenResponse = Invoke-RestMethod -Method Get -Headers @{"Secret"="$env:MSI_SECRET"} -Uri $tokenAuthURI\n' +
                    '$tokenResponse.access_token'}
            </Typography>
            <Typography variant='body2'>
                We can then use this JWT to authenticate as the Service Principal to the Microsoft Graph APIs using BARK
                for example.
            </Typography>
        </>
    );
};

export default Abuse;
