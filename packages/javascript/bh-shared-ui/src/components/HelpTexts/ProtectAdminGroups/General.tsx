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
import { EdgeInfoProps } from '../index';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                AdminSDHolder is an object and an associated background task which runs on the PDCe FSMO role holder
                domain controller. Any modifications made to the security descriptor of {sourceName} will be tattooed on
                the {targetType} {targetName} by the ProtectAdminGroups background task (not SDProp) every hour by
                default.
            </Typography>
            <Typography variant='body2'>
                Any Owner or DACL abuse on {targetName} will be overwritten by the ProtectAdminGroups task on its next
                cycle. Any modifications to the Owner or DACL of {sourceName} will be tattooed on all protected objects
                in that domain during the next ProtectAdminGroups task cycle.
            </Typography>
        </>
    );
};

export default General;
