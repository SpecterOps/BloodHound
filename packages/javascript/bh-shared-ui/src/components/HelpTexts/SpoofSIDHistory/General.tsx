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
                The cross-forest trust from {targetName} to {sourceName} has a weak SID filtering configuration (Spoof
                SID History Blocked = False).
            </Typography>
            <Typography variant='body2'>
                The {targetName} domain allows principals of {sourceName} access by SIDs of {targetName} in their SID
                history. An attacker with control over the {sourceName} domain can craft access requests with
                manipulated SID history containing SIDs of privileged principals of {targetName} to gain control over
                the {targetName} domain.
            </Typography>
        </>
    );
};

export default General;
