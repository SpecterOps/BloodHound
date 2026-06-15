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
                While not usually modified directly, this attribute can be used in reanimating deleted objects out of
                the Active Directory Recycle Bin. This can be abused by an attacker to regain access to a recently
                deleted object, such as a high-privilege user or group, and then use that access to further escalate
                privileges or maintain persistence in the environment. This abuse is not specific to any particular
                source type, but may be more likely to be abused on user and group objects, as those are the most
                commonly deleted object types that would be reanimated out of the recycle bin.
            </Typography>
        </>
    );
};

export default WindowsAbuse;
