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

const General: FC<EdgeInfoProps> = () => {
    return (
        <>
            <Typography variant='body2'>
                This edge indicates that an attacker with "Authenticated Users" access can compromise the target
                computer by relaying the NTLM authentication of a victim computer with administrative rights on the
                target computer. The attack is possible because the attacker can trigger SMB-based coercion from the
                victim computer to their attacker-controlled host, and the target computer does not enforce SMB signing.
            </Typography>

            <Typography variant='body2'>
                Click on Composition to view victim computers with administrative rights on the target computer.
            </Typography>
        </>
    );
};

export default General;
