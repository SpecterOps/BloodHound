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
import { Link, Typography } from '@mui/material';
import { FC } from 'react';

const AZResourceGroupLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-resource-group'>
        AZResourceGroup
    </Link>
);

const AZSubscriptionLink = (
    <Link
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-subscription'>
        AZSubscription
    </Link>
);

const AZVMLink = (
    <Link target='_blank' rel='noopener noreferrer' href='https://bloodhound.specterops.io/resources/nodes/az-vm'>
        AZVM
    </Link>
);

const General: FC = () => {
    return (
        <Typography variant='body2'>
            The principal is granted the Owner role on the resource.
            <br />
            <br />
            AZOwner targets resources in AzureRM (for example {AZResourceGroupLink}, {AZSubscriptionLink} and {AZVMLink}
            ) through role assignment called “Owner”.
        </Typography>
    );
};

export default General;
