// Copyright 2025 Specter Ops, Inc.
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

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The cross-forest trust from {targetName} to {sourceName} has TGT delegation enabled.
            </Typography>
            <Typography variant='body2'>
                When a resource in {sourceName} is configured with unconstrained delegation, principals from{' '}
                {targetName} will automatically forward their Ticket Granting Ticket (TGT) to that resource upon access.
            </Typography>
        </>
    );
};

export default General;
