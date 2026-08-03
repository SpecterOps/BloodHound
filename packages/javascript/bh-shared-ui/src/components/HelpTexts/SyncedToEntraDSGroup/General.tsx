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

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                This relationship indicates that the Entra group and the Entra Domain Services group are the same group
                across the Entra ID and managed domain boundary.
            </Typography>
            <Typography variant='body2'>
                The Entra Domain Services group is created from the Entra group during synchronization and can be
                correlated through the BloodHound aadobjectid property, collected from the LDAP attribute
                msDS-aadObjectId. Membership changes made to the Entra group are synchronized into the corresponding
                Entra Domain Services group.
            </Typography>
            <Typography variant='body2'>
                Only direct membership is synchronized. Nested Entra groups do not become nested Entra Domain Services
                groups through this relationship.
            </Typography>
            <Typography variant='body2'>
                This relationship is informational. Control of the Entra group does not by itself provide a usable Entra
                Domain Services identity. The related AddEntraDSGroupMember edge captures the case where a synchronized
                Entra user can use control of a synchronized Entra group to gain effective membership in the Entra
                Domain Services group.
            </Typography>
        </>
    );
};

export default General;
