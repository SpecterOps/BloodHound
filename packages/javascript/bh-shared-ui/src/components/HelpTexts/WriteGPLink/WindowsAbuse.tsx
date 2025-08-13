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

const WindowsAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                From a domain-joined compromised Windows machine, the WriteGPLink permission may be abused through
                Powermad, PowerView and native Windows functionalities. For a detailed outline of exploit requirements
                and implementation, you can refer to{' '}
                <Link target='_blank' rel='noopener' href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                    this article
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

export default WindowsAbuse;
