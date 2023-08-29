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
        <>
            <Typography variant='body2'>
                This edge is created to link Azure Kubernetes Service Managed Clusters to the Virtual Machine Scale Sets
                they use to execute commands on.
            </Typography>

            <Typography variant='body2'>
                The system-assigned identity for the AKS Cluster will have the Contributor role against the target
                Resource Group and its child Virtual Machine Scale Sets.
            </Typography>
        </>
    );
};

export default General;
