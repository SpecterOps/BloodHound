// Copyright 2026 Specter Ops, Inc.
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

const LinuxAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/CravateRouge/bloodyAD'>
                    bloodyAD
                </Link>{' '}
                can be used to abuse this permission by reanimating a recently deleted high-privilege user or group and
                then using that access to further escalate privileges or maintain persistence in the environment.
            </Typography>

            <Typography component={'pre'}>
                {
                    "bloodyAD -h 'dc.domain.local' -d 'domain.local' -u 'controlledUser' -p 'ItsPassword' set restore targetObject"
                }
            </Typography>

            <Typography variant='body2'>
                If the controlled user does not have{' '}
                <pre style={{ display: 'inline', width: 'fit-content' }}>CreateChild</pre> permissions on the original
                parent container of the deleted object, a new parent container can be specified by adding it to the
                command:{' '}
                <pre style={{ display: 'inline', width: 'fit-content' }}>--newParent 'CN=...,DC=domain,DC=local'</pre>{' '}
            </Typography>
        </>
    );
};

export default LinuxAbuse;
