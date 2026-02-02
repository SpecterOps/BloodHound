// Copyright 2023 Specter Ops, Inc.
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

import {Link, Typography} from '@mui/material';
import { FC } from 'react';

const AZGroupLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-group'>
        AZGroup
    </Link>
);

const AZServicePrincipalLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-service-principal'>
        AZServicePrincipal
    </Link>
);

const AZDeviceLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-device'>
        AZDevice
    </Link>
)

const General: FC = () => {
    return (
        <Typography variant='body2'>
            The principal is granted owner rights on the principal.
            <br/><br/>
            AZOwns targets resources in Entra ID (for example {AZGroupLink}, {AZServicePrincipalLink}, and {AZDeviceLink}) from various object-specific ownership.
        </Typography>
    );
};

export default General;
