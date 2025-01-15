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
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The user {targetName} has a session on the computer {sourceName}.
            </Typography>
            <Typography variant='body2'>
                When a user authenticates to a computer, they often leave credentials exposed on the system, which can
                be retrieved through LSASS injection, token manipulation/theft, or injecting into a user's process.
            </Typography>
            <Typography variant='body2'>
                Any user that is an administrator to the system has the capability to retrieve the credential material
                from memory if it still exists.
            </Typography>
            <Typography variant='body2'>
                Note: A session does not guarantee credential material is present, only possible.
            </Typography>
        </>
    );
};

export default General;
