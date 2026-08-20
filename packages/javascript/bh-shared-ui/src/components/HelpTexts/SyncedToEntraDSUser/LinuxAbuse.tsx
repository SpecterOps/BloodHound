// Copyright 2024 Specter Ops, Inc.
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

const Abuse: FC = () => {
    return (
        <Typography variant='body2' component='div'>
            An attacker may authenticate as the Microsoft Entra Domain Services (Entra DS) user using the Entra user's
            credentials:
            <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                <li>
                    Obtain the Entra user's current password, or use the control represented by the path to change or
                    reset it to a known value.
                </li>
                <li>
                    If the account is cloud-only and has not completed a qualifying password change while the managed
                    domain is active, perform that change. When a reset operation permits it, set{' '}
                    <code>forceChangePasswordNextSignIn</code> to <code>false</code>; otherwise complete the required
                    interactive password change before proceeding.
                </li>
                <li>
                    Wait for the legacy Kerberos and NTLM password material to synchronize. Do not treat the AD user's
                    existence alone as proof that its credential is usable; poll with a harmless authentication attempt
                    when <code>pwdLastSet</code> or equivalent synchronization evidence is unavailable.
                </li>
                <li>
                    Authenticate with the Entra UPN and the changed password through Kerberos, NTLM, LDAP, or a domain
                    logon.
                </li>
            </ol>
        </Typography>
    );
};

export default Abuse;
