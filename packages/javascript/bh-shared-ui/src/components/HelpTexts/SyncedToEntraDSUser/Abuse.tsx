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

const SyncedToEntraDSUserAbuse: FC = () => {
    return (
        <Typography variant='body2' component='div'>
            An attacker may authenticate as the Microsoft Entra Domain Services (Entra DS) user using the Entra user's
            credentials:
            <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                <li>
                    Obtain the Entra user's current password.
                </li>
                <li>
                    If the account is cloud-only and has not completed a qualifying password change while the managed
                    domain is active, perform that change.
                </li>
                <li>
                    Wait for the legacy Kerberos and NTLM password material to synchronize.
                </li>
                <li>
                    Authenticate with the Entra UPN and the changed password.
                </li>
            </ol>
        </Typography>
    );
};

export default SyncedToEntraDSUserAbuse;
