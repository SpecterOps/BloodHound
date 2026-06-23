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

const WindowsAbuse: FC<EdgeInfoProps> = ({ targetType }) => {
    switch (targetType) {
        case 'Domain':
            return (
                <>
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
        case 'OU':
            return (
                <>
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
        case 'Site':
            return (
                <>
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
