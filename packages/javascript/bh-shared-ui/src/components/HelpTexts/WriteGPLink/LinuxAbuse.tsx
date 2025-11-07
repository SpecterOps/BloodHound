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
        case 'Domain':
            return (
                <>
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
        case 'OU':
            return (
                <>
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
        case 'Site':
            return (
                <>
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
