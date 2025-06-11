// Copyright 2025 Specter Ops, Inc.
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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker with permission to modify the gPLink attribute can link GPOs to the object, affecting all
                contained users and computers. The GPO can be weaponized by injecting a malicious configuration, such as
                a scheduled task executing a malicious script.
            </Typography>
            <Typography variant='body2'>
                The GPO can be linked as enforced to bypass blocked GPO inheritance. WMI or security filtering can be
                used to limit the impact to specific accounts, which is important in environments with many users or
                computers under the affected scope.
            </Typography>
            <Typography variant='body2'>
                Refer to{' '}
                <Link target='_blank' rel='noopener' href='https://wald0.com/?p=179'>
                    A Red Teamer's Guide to GPOs and OUs
                </Link>
                for details about the abuse technique, and check out the following tools for practical exploitation:
                <ul>
                    <li>
                        Windows:{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/FSecureLABS/SharpGPOAbuse'>
                            GitHub: SharpGPOAbuse
                        </Link>
                    </li>
                    <li>
                        Linux:{' '}
                        <Link target='_blank' rel='noopener' href='https://github.com/Hackndo/pyGPOAbuse'>
                            GitHub: pyGPOAbuse
                        </Link>
                    </li>
                </ul>
            </Typography>
            <Typography variant='body2'>
                <b>Without control over a GPO</b>
                <br />
                An attacker can still execute the attack without control over a GPO by setting up a fake LDAP server to
                host a GPO. This approach requires the ability to add non-existent DNS records and create machine
                accounts, or access to a compromised domain-joined machine. However, this method is complex and requires
                significant setup.
                <p className='my-4'>
                    From a domain-joined compromised Windows machine, the write access to the gPLink attribute may be
                    abused through Powermad, PowerView and native Windows functionalities. For a detailed outline of
                    exploit requirements and implementation, you can refer to this article:{' '}
                    <Link
                        target='_blank'
                        rel='noopener'
                        href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                        OU having a laugh?
                    </Link>
                </p>
                From a Linux machine, the write access to the gPLink attribute may be abused using the{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/synacktiv/OUned'>
                    OUned.py
                </Link>{' '}
                exploitation tool. For a detailed outline of exploit requirements and implementation, you can refer to{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                    the article associated to the OUned.py tool
                </Link>
                .
            </Typography>
        </>
    );
};

export default Abuse;
