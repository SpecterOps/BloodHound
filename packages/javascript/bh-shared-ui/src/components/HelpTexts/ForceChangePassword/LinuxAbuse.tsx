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
                Use samba's net tool to change the user's password. The credentials can be supplied in cleartext or
                prompted interactively if omitted from the command line. The new password will be prompted if omitted
                from the command line.
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
                Now that you know the target user's plain text password, you can either start a new agent as that user,
                or use that user's credentials in conjunction with PowerView's ACL abuse functions, or perhaps even RDP
                to a system the target user has access to. For more ideas and information, see the references tab.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
