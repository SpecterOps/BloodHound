// Copyright 2026 Specter Ops, Inc.
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

import { Typography } from 'doodle-ui';
import { FC } from 'react';

const References: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                <a href='https://attack.mitre.org/techniques/T1558/003/' target='_blank' rel='noreferrer'>
                    MITRE ATT&CK T1558.003 - Steal or Forge Kerberos Tickets: Kerberoasting
                </a>
            </Typography>
            <Typography variant='body2'>
                <a href='https://github.com/fortra/impacket/blob/master/examples/GetUserSPNs.py' target='_blank' rel='noreferrer'>
                    Impacket GetUserSPNs.py
                </a>
            </Typography>
            <Typography variant='body2'>
                <a href='https://github.com/GhostPack/Rubeus#kerberoast' target='_blank' rel='noreferrer'>
                    Rubeus kerberoast
                </a>
            </Typography>
        </>
    );
};

export default References;
