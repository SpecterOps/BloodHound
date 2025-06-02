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

const General: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                The ClaimSpecialIdentity edge represents the ability to obtain an access token containing a special
                identity (group) SID. Unlike regular groups, membership in special identities is determined at
                authentication rather than via an explicit member list.
            </Typography>
            <Typography variant='body2'>See the Abuse section for specific cases.</Typography>
        </>
    );
};

export default General;
