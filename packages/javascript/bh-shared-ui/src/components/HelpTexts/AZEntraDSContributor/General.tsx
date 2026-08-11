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

import { Typography } from 'doodle-ui';
import { FC } from 'react';

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                AZEntraDSContributor records raw Azure Resource Manager authorization. The built-in Domain Services
                Contributor role, definition ID <code>eeaeda52-9324-47f6-8069-5d5bade478b2</code>, grants{' '}
                <code>Microsoft.AAD/domainServices/*</code> and related network permissions over the target AZEntraDS
                resource.
            </Typography>
            <Typography variant='body2'>
                The edge is not independently traversable. BloodHound creates AZManageEntraDS only when the same
                effective principal also has Application Administrator and Groups Administrator.
            </Typography>
        </>
    );
};

export default General;
