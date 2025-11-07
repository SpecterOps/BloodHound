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

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                If you control a GPO linked to a target object, you may make modifications to that GPO in order to inject malicious configurations into it. 
                You could for instance add a Scheduled Task that will then be executed by all of the computers and/or users to which the GPO applies, 
                thus compromising them. Note that some configurations (such as Scheduled Tasks) implement item-level targeting, allowing 
                to precisely target a specific object.
                GPOs are applied every 90 minutes for standard objects (with a random offset of 0 to 30 minutes), and every 5 minutes for domain controllers.
                See the references tab for a more detailed write up on this abuse.
            </Typography>

            <Typography variant='body2'>
                On a domain-joined Windows machine, the native Group Policy Management Console (GPMC) may be used to edit GPOs. 
                On a non-domain joined Windows Machine, the{' '}
                <Link
                    target='_blank'
                    rel='noopener noreferrer'
                    href='https://github.com/CCob/DRSAT'>
                    DRSAT (Disconnected RSAT)
                </Link> tool can be used.
            </Typography>
        </>
    );
};

export default WindowsAbuse;
