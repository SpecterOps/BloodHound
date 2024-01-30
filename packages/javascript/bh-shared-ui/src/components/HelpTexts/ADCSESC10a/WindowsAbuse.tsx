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
import { Typography, Link, List, ListItem, Box } from '@mui/material';
import { useHelpTextStyles } from '../utils';

const WindowsAbuse: FC = () => {
    const classes = useHelpTextStyles();
    const step1 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 1: </b>Create .exe version of Certipy.
                <br />
                <br />
                Install PyInstaller on a host with python installed, clone down Certipy from GitHub, and run this cmdlet
                from the root of the GitHub repo to bundle the python project into an .exe binary which can be used on
                Windows computer where Python is not installed:
            </Typography>
            <Typography component={'pre'}>{'pyinstaller ./Certipy.spec'}</Typography>
            <Typography variant='body2' className={classes.containsCodeEl}>
                The Certipy.exe will be in the <code>dist</code> folder.
            </Typography>
        </>
    );

    const step2 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                <b>Step 2: </b> Set UPN of victim to targeted principal's <code>sAMAccountName</code> followed by @ and
                the domain name.
                <br />
                <br />
                Set the UPN of the victim principal using Certipy:
            </Typography>
            <Typography component={'pre'}>
                {'Certipy.exe account update -u ATTACKER@CORP.LOCAL -p PWD -user VICTIM -upn Target@CORP.LOCAL'}
            </Typography>
        </>
    );

    const step3 = (
        <>
            <Typography variant='body2' className={classes.containsCodeEl} sx={{ marginBottom: '-8px' }}>
                <b>Step 3: </b>Check if <code>mail</code> attribute of victim must be set and set it if required.
                <br />
                <br />
                If the certificate template is of schema version 2 or above and its attribute{' '}
                <code>msPKI-CertificateNameFlag</code> contains the flag <code>SUBJECT_REQUIRE_EMAIL</code> and/or{' '}
                <code>SUBJECT_ALT_REQUIRE_EMAIL</code> then the victim principal must have their mail attribute set for
                the certificate enrollment. The CertTemplate BloodHound node will have <em>"Subject Require Email"</em>{' '}
                or <em>"Subject Alternative Name Require Email"</em> set to true if any of the flags are present.
                <br />
                <br />
                If the certificate template is of schema version 1 or does not have any of the email flags, then
                continue to Step 4.
                <br />
                <br />
                If any of the two flags are present, you will need the victim’s mail attribute to be set. The value of
                the attribute will be included in the issues certificate but it is not used to identify the target
                principal why it can be set to any arbitrary string.
                <br />
                <br />
                Check if the victim has the mail attribute set using PowerView:
            </Typography>
            <Typography component='pre'>{'Get-DomainObject -Identity VICTIM -Properties mail'}</Typography>
            <Typography variant='body2'>
                If the victim has the mail attribute set, continue to Step 4.
                <br />
                <br />
                If the victim does not has the mail attribute set, set it to a dummy mail using PowerView:
            </Typography>
            <Typography component='pre'>
                {"Set-DomainObject -Identity VICTIM -Set @{'mail'='dummy@mail.com'}"}
            </Typography>
        </>
    );

    const step4 = (
        <Box
            sx={{
                borderRadius: '4px',
                backgroundColor: '#eee',
            }}>
            <Typography variant='body2'>
                <b>Step 4: </b>Obtain a session as victim.
                <br />
                <br />
                There are several options for this step.
                <br />
                <br />
                If the victim is a computer, you can obtain the credentials of the computer account using the Shadow
                Credentials attack (see{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17358104809499-AddKeyCredentialLink'>
                    AddKeyCredentialLink edge documentation
                </Link>
                ). Alternatively, you can obtain a session as SYSTEM on the host, which allows you to interact with AD
                as the computer account, by abusing control over the computer AD object (see{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17312347318043-GenericAll'>
                    GenericAll edge documentation
                </Link>
                ).
            </Typography>
            <Typography variant='body2' className={classes.containsCodeEl}>
                If the victim is a user, you have the following options for obtaining the credentials:
            </Typography>
            <List sx={{ fontSize: '12px' }}>
                <ListItem>
                    Shadow Credentials attack (see{' '}
                    <Link
                        target='blank'
                        rel='noopener'
                        href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17358104809499-AddKeyCredentialLink'>
                        AddKeyCredentialLink edge documentation
                    </Link>
                    )
                </ListItem>
                <ListItem>
                    Password reset (see{' '}
                    <Link
                        target='blank'
                        rel='noopener'
                        href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17223286750747-ForceChangePassword'>
                        ForceChangePassword edge documentation
                    </Link>
                    )
                </ListItem>
                <ListItem>
                    Targeted Kerberoasting (see{' '}
                    <Link
                        target='blank'
                        rel='noopener'
                        href='https://support.bloodhoundenterprise.io/hc/en-us/articles/17222775975195-WriteSPN'>
                        WriteSPN edge documentation
                    </Link>
                    )
                </ListItem>
            </List>
        </Box>
    );

    const step5 = (
        <>
            <Typography variant='body2'>
                <b>Step 5: </b>Enroll certificate as victim.
                <br />
                <br />
                Use Certipy as the victim principal to request enrollment in the affected template, specifying the
                affected EnterpriseCA:
            </Typography>
            <Typography component={'pre'}>
                {'Certipy.exe req -u VICTIM@CORP.LOCAL -p PWD -ca CA-NAME -target CA-SERVER -template TEMPLATE'}
            </Typography>
            <Typography variant='body2'>
                The issued certificate will be saved to disk with the name of the targeted user.
            </Typography>
        </>
    );
    const step6 = (
        <>
            <Typography variant='body2'>
                <b>Step 6: </b>Set UPN of victim to arbitrary value.
                <br />
                <br />
                Set the UPN of the victim principal using Certipy:
            </Typography>
            <Typography component={'pre'}>
                {'Certipy.exe account update -u ATTACKER@CORP.LOCAL -p PWD -user VICTIM -upn victim@corp.local'}
            </Typography>
        </>
    );
    const step7 = (
        <>
            <Typography variant='body2'>
                <b>Step 7: </b>Perform Schannel authentication as targeted principal against affected DC using
                certificate.
                <br />
                <br />
                Open an LDAP shell as the victim using Certipy by specifying the certificate created in Step 5 and the
                IP of an affected DC:
            </Typography>
            <Typography component={'pre'}>{'Certipy.exe auth -pfx TARGET.pfx -dc-ip IP -ldap-shell'}</Typography>
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

export default WindowsAbuse;
