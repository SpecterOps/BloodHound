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
import { Link, Typography } from '@mui/material';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps & { haslaps: boolean }> = ({
    sourceName,
    sourceType,
    targetName,
    targetType,
    haslaps,
}) => {
    switch (targetType) {
        case 'User':
            return (
                <>
                    <Typography variant='body2'>
                        The AllExtendedRights permission grants {sourceName} the ability to change the password of the
                        user {targetName} without knowing their current password. This is equivalent to the
                        "ForceChangePassword" edge in BloodHound.
                    </Typography>

                    <Typography variant='body2'>
                        There are at least two ways to execute this attack. The first and most obvious is by using the
                        built-in net.exe binary in Windows (e.g.: net user dfm.a Password123! /domain). See the opsec
                        considerations tab for why this may be a bad idea. The second, and highly recommended method, is
                        by using the Set-DomainUserPassword function in PowerView. This function is superior to using
                        the net.exe binary in several ways. For instance, you can supply alternate credentials, instead
                        of needing to run a process as or logon as the user with the ForceChangePassword permission.
                        Additionally, you have much safer execution options than you do with spawning net.exe (see the
                        opsec tab).
                    </Typography>

                    <Typography variant='body2'>
                        To abuse this permission with PowerView's Set-DomainUserPassword, first import PowerView into
                        your agent session or into a PowerShell instance at the console. You may need to authenticate to
                        the Domain Controller as{' '}
                        {sourceType === 'User'
                            ? `${sourceName} if you are not running a process as that user`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                        . To do this in conjunction with Set-DomainUserPassword, first create a PSCredential object
                        (these examples comes from the PowerView help documentation):
                    </Typography>

                    <Typography component={'pre'}>
                        {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                            "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
                    </Typography>

                    <Typography variant='body2'>
                        Then create a secure string object for the password you want to set on the target user:
                    </Typography>

                    <Typography component={'pre'}>
                        {"$UserPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force"}
                    </Typography>

                    <Typography variant='body2'>
                        Finally, use Set-DomainUserPassword, optionally specifying $Cred if you are not already running
                        a process as {sourceName}:
                    </Typography>

                    <Typography component={'pre'}>
                        {'Set-DomainUserPassword -Identity andy -AccountPassword $UserPassword -Credential $Cred'}
                    </Typography>

                    <Typography variant='body2'>
                        Now that you know the target user's plain text password, you can either start a new agent as
                        that user, or use that user's credentials in conjunction with PowerView's ACL abuse functions,
                        or perhaps even RDP to a system the target user has access to. For more ideas and information,
                        see the references tab.
                    </Typography>
                </>
            );
        case 'Computer':
            return (
                <>
                    <Typography variant='body2'>
                        The AllExtendedRights permission allows {sourceName} to retrieve the LAPS (RID 500
                        administrator) password for {targetName}.
                    </Typography>
                    <Typography variant='body2'>
                        For systems using legacy LAPS, the following AD computer object properties are relevant:
                        <br />
                        <b>- ms-Mcs-AdmPwd</b>: The plaintext LAPS password
                        <br />
                        <b>- ms-Mcs-AdmPwdExpirationTime</b>: The LAPS password expiration time
                        <br />
                    </Typography>
                    <Typography variant='body2'>
                        For systems using Windows LAPS (2023 edition), the following AD computer object properties are
                        relevant:
                        <br />
                        <b>- msLAPS-Password</b>: The plaintext LAPS password
                        <br />
                        <b>- msLAPS-PasswordExpirationTime</b>: The LAPS password expiration time
                        <br />
                        <b>- msLAPS-EncryptedPassword</b>: The encrypted LAPS password
                        <br />
                        <b>- msLAPS-EncryptedPasswordHistory</b>: The encrypted LAPS password history
                        <br />
                        <b>- msLAPS-EncryptedDSRMPassword</b>: The encrypted Directory Services Restore Mode (DSRM)
                        password
                        <br />
                        <b>- msLAPS-EncryptedDSRMPasswordHistory</b>: The encrypted DSRM password history
                        <br />
                    </Typography>
                    <Typography variant='body2'>
                        Plaintext attributes can be read using a simple LDAP client. For example, with PowerView:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'Get-DomainComputer "MachineName" -Properties "cn","ms-mcs-admpwd","ms-mcs-admpwdexpirationtime"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Encrypted attributes can be decrypted using Microsoft's LAPS PowerShell module. For example:
                    </Typography>
                    <Typography component={'pre'}>{'Get-LapsADPassword "WIN10" -AsPlainText'}</Typography>
                    <Typography variant='body2'>
                        The encrypted attributes can also be retrieved and decrypted using{' '}
                        <Link
                            target='_blank'
                            rel='noopener'
                            href='https://github.com/xpn/RandomTSScripts/tree/master/lapsv2decrypt'>
                            lapsv2decrypt
                        </Link>{' '}
                        (dotnet or BOF).
                    </Typography>
                </>
            );
        case 'Domain':
            return (
                <Typography variant='body2'>
                    The AllExtendedRights permission grants {sourceName} both the DS-Replication-Get-Changes and
                    DS-Replication-Get-Changes-All privileges, which combined allow a principal to replicate objects
                    from the domain {targetName}. This can be abused using the lsadump::dcsync command in mimikatz.
                </Typography>
            );
        case 'CertTemplate':
            return (
                <>
                    <Typography variant='body2'>
                        The AllExtendedRights permission grants {sourceName} enrollment rights on the certificate
                        template {targetName}.
                    </Typography>
                    <Typography variant='body2'>Certify can be used to enroll a certificate:</Typography>
                    <Typography component={'pre'}>
                        {'Certify.exe request /ca:SERVER\\CA-NAME /template:TEMPLATE'}
                    </Typography>
                    <Typography variant='body2'>
                        The following additional requirements must be met for a principal to be able to enroll a
                        certificate:
                        <br />
                        1) The certificate template is published on an enterprise CA
                        <br />
                        2) The principal has Enroll permission on the enterprise CA
                        <br />
                        3) The principal meets the issuance requirements and the requirements for subject name and
                        subject alternative name defined by the template
                    </Typography>
                </>
            );
        default:
            return (
                <>
                    <Typography variant='body2'>No abuse information available for this node type.</Typography>
                </>
            );
    }
};

export default WindowsAbuse;
