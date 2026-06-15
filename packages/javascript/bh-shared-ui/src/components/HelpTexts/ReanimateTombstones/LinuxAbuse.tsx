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
                This ExtendedRight does not guarantee the ability to reanimate deleted objects, but it is a necessary
                permission for doing so.{' '}
                <Link target='_blank' rel='noopener noreferrer' href='https://github.com/CravateRouge/bloodyAD'>
                    bloodyAD
                </Link>{' '}
                can be used to determine if the account has sufficient permissions to reanimate a deleted object and to
                perform the reanimation if it does. This can be abused by an attacker to regain access to a recently
                deleted object, such as a high-privilege user or group, and then use that access to further escalate
                privileges or maintain persistence in the environment. At a minimum, the account needs the ability to:
                CreateChild on a container for the reanimated object and either GenericAll or <strong>BOTH </strong>
                WriteRDN and WriteCommonName on the target to be reanimated.
            </Typography>
            <Typography component={'pre'}>
                {
                    "bloodyAD -h 'dc.domain.local' -d 'domain.local' -u 'controlledUser' -p 'ItsPassword' get writable --detail"
                }
            </Typography>
        </>
    );
};

export default LinuxAbuse;
