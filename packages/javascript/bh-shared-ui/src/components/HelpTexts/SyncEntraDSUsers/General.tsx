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
                This relationship indicates that control of either a Microsoft Entra Domain Services managed domain or,
                in the filtered-sync case described below, the Domain Controller Services service principal can be used
                to create usable Entra Domain Services users with baseline Domain Users access.
            </Typography>
            <Typography variant='body2'>
                BloodHound emits this relationship from an AZDomainService because the managed domain controls the broad
                synchronization boundary through its filtered sync and sync scope settings. A principal that controls
                the resource can change those settings so an attacker-controlled identity is eligible for
                synchronization.
            </Typography>
            <Typography variant='body2'>
                BloodHound can also emit this relationship from the Domain Controller Services service principal with
                application ID 2565bd9d-da50-47d4-8b85-4c97f669dc36, but only when the related managed domain has
                filtered sync set to Enabled and sync scope set to All. In that state, a principal that controls the
                service principal can assign an attacker-controlled Entra security group to the filtered synchronization
                scope. The direct members of that group are then materialized as Entra Domain Services users.
            </Typography>
            <Typography variant='body2'>
                Every newly materialized Entra Domain Services user receives Domain Users as its primary group. In
                BloodHound, Domain Users is already nested into Authenticated Users, which is in turn nested into
                Everyone, so this relationship represents baseline authenticated access to the managed domain.
            </Typography>
            <Typography variant='body2'>
                Direct user assignments may appear in the portal, but they are not honored by the Entra Domain Services
                sync engine. Only explicitly scoped groups and their direct members are synchronized.
            </Typography>
        </>
    );
};

export default General;
