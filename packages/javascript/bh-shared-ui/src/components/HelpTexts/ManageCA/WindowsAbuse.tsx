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
            <Typography variant='body2'>
                CA administrators can perform the following actions that may enable and ADCS escalation:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>Grant CA officer (ManageCertificates) on enterprise CA</li>
                    <li>Publish certificate templates</li>
                    <li>Grant enroll on enterprise CA</li>
                    <li>Enable configurations like the ESC6 flag (EDITF_ATTRIBUTESUBJECTALTNAME2)</li>
                </ol>
            </Typography>
            <Typography variant='body2'>
                An attacker can identify an ADCS escalation where one or more requirements are not met and enable the
                abuse by performing one or more of the actions above.
            </Typography>
            <Typography variant='body2'>
                The combination of the ManageCA and ManageCertificates permissions allow the attacker to approve
                certificate requests that were denied becuase lack of enrollment rights on the certificate template or
                on the enterprise CA. A common action is therefore to grant this permission, enroll in an ESC1 template
                that the attacker does not have enrollment rights on, issue the certificate of the denied request. This
                attack fail if role separation is enabled on the CA, as it enforces that users cannot have both ManageCA
                and ManageCertificates. However, this setting is very rare.
            </Typography>
            <Typography variant='body1'>Grant CA officer (ManageCertificates)</Typography>
            <Typography variant='body2'>See Linux abuse.</Typography>
            <Typography variant='body1'>Publish certificate template</Typography>
            <Typography variant='body2'>Certificate templates can be published to the CA using certutil:</Typography>
            <Typography component={'pre'}>
                {'certutil -config "caserver.fabricam.com\\Fabricam Issuing CA" -SetCAtemplates +TemplateCN'}
            </Typography>
            <Typography variant='body1'>Approve certificate request (pending or denied)</Typography>
            <Typography variant='body2'>
                Certificate requsts can be approved using certutil (requires ManageCertificates):
            </Typography>
            <Typography component={'pre'}>
                {'certutil -config "caserver.fabricam.com\\Fabricam Issuing CA" -resubmit 12345'}
            </Typography>
            <Typography variant='body2'>Approved certificate can be downloaded using Certify:</Typography>
            <Typography component={'pre'}>
                {'Certify.exe download /ca:"caserver.fabricam.com\\Fabricam Issuing CA" /id:ReqID'}
            </Typography>
            <Typography variant='body1'>Enable ESC6 flag (requires restart of CA service)</Typography>
            <Typography variant='body2'>
                The EDITF_ATTRIBUTESUBJECTALTNAME2 flag can be enabled with certutil:
            </Typography>
            <Typography component={'pre'}>
                {
                    'certutil -config "caserver.fabricam.com\\Fabricam Issuing CA" -setreg policy\\EditFlags +EDITF_ATTRIBUTESUBJECTALTNAME2'
                }
            </Typography>
            <Typography variant='body2'>
                The setting will not apply until the CA service has been restarted. This cannot be performed by the CA
                administrator remotely, only locally.
            </Typography>
        </>
    );
};

export default Abuse;
