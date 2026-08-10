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
                This relationship indicates that a synchronized Entra user can effectively add or remove members from an
                Entra Domain Services group by controlling the corresponding synchronized Entra group.
            </Typography>
            <Typography variant='body2'>
                The relationship is composed from three conditions: the Entra user is synchronized to Entra Domain
                Services; the Entra user owns or can add and remove members from an Entra group; and the Entra group is
                synchronized to an Entra Domain Services group.
            </Typography>
            <Typography variant='body2'>
                Because the Entra user already has a usable Entra Domain Services identity, they can add themselves or
                another controlled synchronized principal to the Entra group, remove existing members, and wait for the
                membership change to synchronize into the Entra Domain Services group. Adding membership effectively
                grants the Entra user any privileges held by the Entra Domain Services group; removing membership can
                revoke those privileges from another principal.
            </Typography>
            <Typography variant='body2'>
                Only direct membership in the source Entra group is synchronized. Nested Entra groups do not satisfy
                this relationship.
            </Typography>
        </>
    );
};

export default General;
