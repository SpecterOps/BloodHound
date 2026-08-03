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
        <>
            <Typography variant='body2'>
                Using the Entra user's control over the Entra group, add the Entra user or another controlled
                synchronized principal as a direct member of the Entra group. In Microsoft Graph PowerShell this can be
                done with:
            </Typography>
            <Typography component={'pre'} variant='body2'>
                {
                    'New-MgGroupMemberByRef -GroupId "<entra-group-object-id>" -OdataId "https://graph.microsoft.com/v1.0/directoryObjects/<member-object-id>"'
                }
            </Typography>
            <Typography variant='body2'>
                After Entra Domain Services synchronizes the direct membership change, the principal becomes a member of
                the corresponding Entra Domain Services group and inherits its access within the managed domain.
            </Typography>
        </>
    );
};

export default Abuse;
