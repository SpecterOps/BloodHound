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

import { Typography } from '@mui/material';
import { FC } from 'react';

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Azure provides several systems and mechanisms for granting control of securable objects within Azure
                Active Directory, including tenant-scoped admin roles, object-scoped admin roles, explicit object
                ownership, and API permissions.
            </Typography>
            <Typography variant='body2'>
                When a principal has been granted "Cloud App Admin" or "App Admin" against the tenant, that principal
                gains the ability to add new secrets to all Service Principals* and App Registrations. Additionally, a
                principal that has been granted "Cloud App Admin" or "App Admin" against, or explicit ownership of a
                Service Principal* or App Registration gains the ability to add secrets to that particular object.

                * Secrets can only be added to the Service Principal if it is not protected by the "App instance property lock" configuration in the corresponding App Registration.
            </Typography>
        </>
    );
};

export default General;
