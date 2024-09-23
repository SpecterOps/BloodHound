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
import { Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                The TGT delegation enable an attacker to caputure TGTs of privileged users or computers of the target
                domain as they authenticate against a computer configured with unconstrained delegation.
            </Typography>
            <Typography variant='body2'>
                A common way for attackers to execute the attack is to log in on DC of the source domain and coerce a DC
                of the target domain. DCs have unconstrained delegation enabled by default. That gives the attacker a
                TGT of a DC of the target domain, which enable the attacker to DCSync the target domain.
            </Typography>
            <Typography variant='body1'>Step 1: Start monitoring for TGTs</Typography>
            <Typography variant='body2'>
                Log in on a DC (or other host configured with unconstrained delegation) of the source domain and open
                CMD as Administrator.
            </Typography>
            <Typography variant='body2'>Start monitoring for incomming TGTs using Rubeus:</Typography>
            <Typography component={'pre'}>{'Rubeus.exe request monitor /user:targetdc.domain.local'}</Typography>
            <Typography variant='body1'>Step 2: Coerce target DC</Typography>
            <Typography variant='body2'>Coerce the target DC using SpoolSample:</Typography>
            <Typography component={'pre'}>
                {'SpoolSample.exe targetdc.domain.local compromiseddc.otherdomain.local'}
            </Typography>
        </>
    );
};

export default Abuse;
