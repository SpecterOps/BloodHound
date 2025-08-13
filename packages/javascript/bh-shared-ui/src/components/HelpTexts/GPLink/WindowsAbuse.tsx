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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                With full control of a GPO, you may make modifications to that GPO which will then apply to the users
                and computers affected by the GPO. Select the target object you wish to push an evil policy down to,
                then use the gpedit GUI to modify the GPO, using an evil policy that allows item-level targeting, such
                as a new immediate scheduled task. Then wait for the group policy client to pick up and execute the new
                evil policy. See the references tab for a more detailed write up on this abuse.
            </Typography>
        </>
    );
};

export default WindowsAbuse;
