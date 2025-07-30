// Copyright 2025 Specter Ops, Inc.
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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body1'>Guest Account</Typography>
            <Typography variant='body2'>
                The Guest user account allows users without a personal account to log in. The account has no password by
                default.
                <br />
                If enabled, anyone with AD access can log in with the Guest account.
            </Typography>

            <Typography variant='body1'>Network Identity</Typography>
            <Typography variant='body2'>
                Any user or computer accessing a Windows system via a network has the Network identity in their access
                token.
            </Typography>

            <Typography variant='body1'>Authentication Authority Asserted Identity</Typography>
            <Typography variant='body2'>
                Included in access tokens when an account is authenticated directly against a domain controller and not
                through Kerberos constrained delegation (service asserted identity).
            </Typography>

            <Typography variant='body1'>Key Trust</Typography>
            <Typography variant='body2'>
                Included in access tokens when authentication is based on public key credentials via key trust objects.
                <br />
                Anyone with key trust credentials (e.g., from a Shadow Credentials attack) can obtain Key Trust identity
                access through PKINIT authentication.
            </Typography>

            <Typography variant='body1'>MFA Key Property</Typography>
            <Typography variant='body2'>
                Similar to Key Trust but requires the MFA property on the key trust credentials.
                <br />A Shadow Credentials attack enables anyone to obtain the MFA Key Property identity access through
                PKINIT authentication.
            </Typography>

            <Typography variant='body1'>NTLM Authentication</Typography>
            <Typography variant='body2'>
                Included in an access token when authentication occurs via NTLM protocol.
                <br />
                Any AD account can obtain NTLM authentication identity access, assuming NTLM is available.
            </Typography>

            <Typography variant='body1'>Schannel Authentication</Typography>
            <Typography variant='body2'>
                Included in an access token when authentication occurs via Schannel protocol.
                <br />
                Any AD account can obtain the Schannel Authentication identity, for example by performing certificate
                authentication over Schannel.
            </Typography>

            <Typography variant='body1'>This Organization Identity</Typography>
            <Typography variant='body2'>
                Assigned to all accounts within the same Active Directory forest and trusted forests without selective
                authentication.
            </Typography>

            <Typography variant='body1'>This Organization Certificate Identity</Typography>
            <Typography variant='body2'>
                Assigned to all accounts within the same Active Directory forest and trusted forests without selective
                authentication, when the Kerberos PAC contains an NTLM_SUPPLEMENTAL_CREDENTIAL structure.
                <br />
                Authentication using an ADCS certificate ensures the required PAC structure.
            </Typography>
        </>
    );
};

export default Abuse;
