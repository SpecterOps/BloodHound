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

import { FC } from 'react';
import { Typography } from '@mui/material';

const General: FC = () => {
    return (
        <Typography variant='body2'>
            Azure resources like Virtual Machines, Logic Apps, and Automation Accounts can be assigned to either System-
            or User-Assigned Managed Identities. This assignment allows the Azure resource to authenticate to Azure
            services as the Managed Identity without needing to know the credential for that Managed Identity. Managed
            Identities, whether System- or User-Assigned, are AzureAD Service Principals.
        </Typography>
    );
};

export default General;
