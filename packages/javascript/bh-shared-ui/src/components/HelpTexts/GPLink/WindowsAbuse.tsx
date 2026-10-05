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

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                If you control a GPO linked to a target object, you can modify that GPO to inject malicious
                configuration. For example, you can add an immediate scheduled task that runs on the computers or users
                that process the GPO, compromising those objects. Some settings, including scheduled tasks, support
                item-level targeting, which can limit execution to specific objects. GPOs apply every 90 minutes for
                standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
                See the References tab for more detail.
            </Typography>

            <Typography variant='body2'>
                On a domain-joined Windows machine, you can edit GPOs with the native Group Policy Management Console
                (GPMC). On a non-domain-joined Windows machine, use the{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/CCob/DRSAT'>
                    DRSAT (Disconnected RSAT)
                </Link>{' '}
                tool.
            </Typography>
        </>
    );
};

export default WindowsAbuse;
