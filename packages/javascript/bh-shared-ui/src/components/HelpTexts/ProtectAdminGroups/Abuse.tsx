// Copyright 2025 Specter Ops, Inc.
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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Any modifications to the AdminSDHolder node's security descriptor via inbound Owns, WriteOwner, or
                WriteDACL edges will propagate to all nodes with an inbound ProtectAdminGroups edge from this
                AdminSDHolder node at the next run of the ProtectAdminGroups background task on the PDCe for the domain.
            </Typography>
            <Typography variant='body2'>
                The amount of time between ProtectAdminGroups run cycles defaults to 1 hour, but is controlled via a
                registry key setting on the PDCe and can be as little as 1 minute or as much as 120 minutes.
            </Typography>
        </>
    );
};

export default Abuse;