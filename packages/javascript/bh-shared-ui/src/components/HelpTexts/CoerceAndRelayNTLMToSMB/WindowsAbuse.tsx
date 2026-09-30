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

import { Link } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body1'>1. Take Over the SMB Port on the Attacker Host</Typography>
            <Typography variant='body2'>
                To avoid a conflict with SMB running on the attacker-controlled Windows computer, it is necessary to
                takeover the SMB port. This can be achieved with{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/zyn3rgy/smbtakeover'>
                    smbtakeover
                </Link>
                .
            </Typography>
            <Typography variant='body1'>2. Start the Relay Server</Typography>
            <Typography variant='body2'>
                The NTLM relay can be executed with{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/Kevin-Robertson/Inveigh'>
                    Inveigh
                </Link>
                .
            </Typography>
            <Typography variant='body1'>3. Coerce the Target Computer</Typography>
            <Typography variant='body2'>
                Several coercion methods are documented here:{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/p0dalirius/windows-coerced-authentication-methods'>
                    Windows Coerced Authentication Methods
                </Link>
                . Examples of tools include:
                <ul style={{ paddingLeft: '1.5em' }}>
                    <li>
                        <Link
                            target='_blank'
                            rel='noopener noreferrer'
                            href='https://github.com/leechristensen/SpoolSample'>
                            SpoolSample
                        </Link>
                    </li>
                    <li>
                        <Link target='_blank' rel='noopener noreferrer' href='https://github.com/topotam/PetitPotam'>
                            PetitPotam
                        </Link>
                    </li>
                </ul>
            </Typography>
        </>
    );
};

export default WindowsAbuse;
