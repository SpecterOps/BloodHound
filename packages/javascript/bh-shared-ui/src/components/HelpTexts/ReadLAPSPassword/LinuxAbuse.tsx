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
                Sufficient control on a computer object is abusable when the computer's local admin account credential
                is controlled with LAPS. The clear-text password for the local administrator account is stored in an
                extended attribute on the computer object called ms-Mcs-AdmPwd.
            </Typography>

            <Typography variant='body2'>
                <Link target='_blank' rel='noopener' href='https://github.com/p0dalirius/pyLAPS'>
                    pyLAPS
                </Link>{' '}
                can be used to retrieve LAPS passwords:
            </Typography>

            <Typography component={'pre'}>
                {'pyLAPS.py --action get -d "DOMAIN" -u "ControlledUser" -p "ItsPassword"'}
            </Typography>
        </>
    );
};

export default LinuxAbuse;
