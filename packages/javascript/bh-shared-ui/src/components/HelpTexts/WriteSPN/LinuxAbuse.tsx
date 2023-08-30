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
                A targeted kerberoast attack can be performed using{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/ShutdownRepo/targetedKerberoast'>
                    targetedKerberoast.py
                </Link>
                .
            </Typography>

            <Typography component={'pre'}>
                {"targetedKerberoast.py -v -d 'domain.local' -u 'controlledUser' -p 'ItsPassword'"}
            </Typography>

            <Typography variant='body2'>
                The tool will automatically attempt a targetedKerberoast attack, either on all users or against a
                specific one if specified in the command line, and then obtain a crackable hash. The cleanup is done
                automatically as well.
            </Typography>

            <Typography variant='body2'>
                The recovered hash can be cracked offline using the tool of your choice.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
