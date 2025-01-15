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

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                There are several ways to abuse the ability to read the GMSA password. The most straight forward abuse
                is possible when the GMSA is currently logged on to a computer, which is the intended behavior for a
                GMSA. If the GMSA is logged on to the computer account which is granted the ability to retrieve the
                GMSA's password, simply steal the token from the process running as the GMSA, or inject into that
                process.
            </Typography>
            <Typography variant='body2'>
                If the GMSA is not logged onto the computer, you may create a scheduled task or service set to run as
                the GMSA. The computer account will start the sheduled task or service as the GMSA, and then you may
                abuse the GMSA logon in the same fashion you would a standard user running processes on the machine (see
                the "HasSession" help modal for more details).
            </Typography>
            <Typography variant='body2'>
                Finally, it is possible to remotely retrieve the password for the GMSA and convert that password to its
                equivalent NT hash.
                <Link target='_blank' rel='noopener' href='https://github.com/micahvandeusen/gMSADumper'>
                    gMSADumper.py
                </Link>{' '}
                can be used for that purpose.
            </Typography>

            <Typography component={'pre'}>{"gMSADumper.py -u 'user' -p 'password' -d 'domain.local'"}</Typography>

            <Typography variant='body2'>
                At this point you are ready to use the NT hash the same way you would with a regular user account. You
                can perform pass-the-hash, overpass-the-hash, or any other technique that takes an NT hash as an input.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
