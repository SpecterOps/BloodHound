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

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                This relationship indicates that a synchronized Entra user can effectively add or remove members from a
                Microsoft Entra Domain Services (Entra DS) group by controlling the corresponding synchronized Entra
                group.
            </Typography>
            <Typography variant='body2'>
                The relationship is composed from three conditions: BloodHound correlates the Entra user with an Entra
                DS user; the Entra user owns or can add and remove members from an Entra group; and the Entra group is
                synchronized to an Entra DS group.
            </Typography>
            <Typography variant='body2'>
                The user can add themselves or another controlled synchronized principal to the Entra group, remove
                existing members, and wait for the membership change to synchronize into the Entra DS group. Adding
                membership can grant privileges held by the Entra DS group; removing membership can revoke those
                privileges from another principal.
            </Typography>
            <Typography variant='body2'>
                User correlation relies on the BloodHound aadobjectid property. Current collection does not include the
                Entra user&apos;s identities, creationType, or externalUserState properties, so B2B external identities
                can be misclassified. BloodHound also does not verify synchronized password material or runtime
                credential usability, which can make this composed relationship a false positive for direct
                exploitation.
            </Typography>
            <Typography variant='body2'>
                Only direct membership in the source Entra group is synchronized. Nested Entra groups do not satisfy
                this relationship.
            </Typography>
        </>
    );
};

export default General;
