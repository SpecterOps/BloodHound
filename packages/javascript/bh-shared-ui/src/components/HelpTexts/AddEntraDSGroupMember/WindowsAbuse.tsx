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

const Abuse: FC = () => {
    return (
        <Typography variant='body2' component='div'>
            <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                <li>
                    Add a user as a <strong>direct</strong> member of the Entra AZGroup that correlates to the
                    destination Microsoft Entra Domain Services (Entra DS) group. Nested group membership does not reach
                    Entra DS. Submit the following Microsoft Graph request:
                    <Typography component={'pre'} variant='body2'>
                        {
                            'POST https://graph.microsoft.com/v1.0/groups/{entra-group-object-id}/members/$ref\nContent-Type: application/json\n\n{\n  "@odata.id": "https://graph.microsoft.com/v1.0/directoryObjects/{controlled-user-object-id}"\n}'
                        }
                    </Typography>
                    A successful request returns <code>204 No Content</code> with no response body. The same request can
                    be sent with Microsoft Graph PowerShell:
                    <Typography component={'pre'} variant='body2'>
                        {
                            'New-MgGroupMemberByRef -GroupId "<entra-group-object-id>" -OdataId "https://graph.microsoft.com/v1.0/directoryObjects/<member-object-id>"'
                        }
                    </Typography>
                </li>
                <li>
                    Wait for Entra DS to synchronize the membership and verify the destination group's direct{' '}
                    <code>member</code> value when LDAP read access is available.
                </li>
                <li>
                    Reauthenticate the controlled user to Entra DS so its logon session or Kerberos ticket contains the
                    newly synchronized group SID.
                </li>
            </ol>
        </Typography>
    );
};

export default Abuse;
