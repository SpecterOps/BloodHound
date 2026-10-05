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
import { EdgeInfoProps } from '../index';

const LinuxAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                If you control a GPO linked to a target object, you can modify that GPO to inject malicious
                configuration. For example, you can add an immediate scheduled task that runs on the computers or users
                that process the GPO, compromising those objects. Some settings, including scheduled tasks, support
                item-level targeting, which can limit execution to specific objects. GPOs apply every 90 minutes for
                standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
            </Typography>

            <Typography variant='body2'>
                The{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/synacktiv/GroupPolicyBackdoor'>
                    GroupPolicyBackdoor.py
                </Link>{' '}
                tool can perform the attack from Linux. First, define a module file that describes the configuration to
                inject. The example below defines a computer configuration with an immediate scheduled task that adds a
                domain user as a local administrator. The filter limits the configuration to a specific target.
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
                Save this configuration as Scheduled_task_add.ini, then inject it into the target GPO with the 'inject'
                command.
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
};

export default LinuxAbuse;
