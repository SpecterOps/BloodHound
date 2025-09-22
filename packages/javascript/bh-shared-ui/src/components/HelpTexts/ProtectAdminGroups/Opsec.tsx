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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Modifications to the AdminSDHolder security descriptor can be detected via SACLs if configured and
                ingested, as can modifications to the security descriptor of any objects that AdminSDHolder protects.
            </Typography>
            <Typography variant='body2'>
                If auditing is properly configured in the environment, Event ID 4780 will be generated if the
                ProtectAdminGroups background task tattoos the AdminSDHolder security descriptor on a protected object.
            </Typography>
        </>
    );
};

export default Opsec;
