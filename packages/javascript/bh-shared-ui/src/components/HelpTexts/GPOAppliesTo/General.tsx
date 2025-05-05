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

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                The {sourceType} {sourceName} applies to the {targetType} {targetName}.
            </Typography>
            <Typography variant='body2'>
                This relationship is created when the GPO is linked to the domain or an organizational unit (OU)
                containing the account. Click on the Composition accordion item to see where the GPO is linked.
            </Typography>
            <Typography variant='body2'>
                BloodHound will not generate this edge if GPO inheritance is blocked and prevents the GPO from applying.
                However, it does not consider WMI or security filtering, since an attacker with control over the GPO can
                modify those filters to apply the policy.
            </Typography>
        </>
    );
};

export default General;
