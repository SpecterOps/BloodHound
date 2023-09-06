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
            The Azure Kubernetes Service Contributor role grants full control of the target Azure Kubernetes Service
            Managed Cluster. This includes the ability to remotely fetch administrator credentials for the cluster as
            well as the ability to execute arbitrary commands on compute nodes associated with the AKS Managed Cluster.
        </Typography>
    );
};

export default General;
