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
                From an AZDomainService source, change the managed domain synchronization settings so an
                attacker-controlled identity is eligible for synchronization, then wait for synchronization.
            </Typography>
            <Typography variant='body2'>
                From a Domain Controller Services service principal source, assign an attacker-controlled Entra security
                group to the filtered synchronization scope, add an attacker-controlled Entra user as a direct member of
                that group, and wait for synchronization. The user is then materialized in Entra Domain Services with
                Domain Users access. A cloud-only user may need an Entra password change before password material is
                available for authentication.
            </Typography>
        </>
    );
};

export default Abuse;
