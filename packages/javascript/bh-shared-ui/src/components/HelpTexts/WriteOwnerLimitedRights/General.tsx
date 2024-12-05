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

const General: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                When specific privileges on an object's DACL are explicitly granted to the "OWNER RIGHTS" SID (S-1-3-4),
                and inheritance is configured for those permissions, they are inherited by the new object owner after a
                change in ownership. In this case, implicit owner rights are blocked, and the new owner is granted only
                the specific inherited privileges granted to OWNER RIGHTS.
            </Typography>
        </>
    );
};

export default General;
