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
                From a Linux machine, the WriteGPLink permission may be abused using the{' '}
                <Link target='_blank' rel='noopener' href='https://github.com/synacktiv/OUned'>
                    OUned.py
                </Link>{' '}
                exploitation tool. For a detailed outline of exploit requirements and implementation, you can refer to{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                    the article associated to the OUned.py tool
                </Link>
                .
            </Typography>

            <Typography variant='body2'>
                Be mindful of the number of users and computers that are in the given domain as they all will attempt to
                fetch and apply the malicious GPO.
            </Typography>
        </>
    );
};

export default LinuxAbuse;
