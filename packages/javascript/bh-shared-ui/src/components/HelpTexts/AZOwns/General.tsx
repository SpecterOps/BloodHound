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

const AZGroupLink = (
    <a target='_blank' rel='noopener noreferrer' href='https://bloodhound.specterops.io/resources/nodes/az-group'>
        AZGroup
    </a>
);

const AZServicePrincipalLink = (
    <a
        target='_blank'
        rel='noopener noreferrer'
        href='https://bloodhound.specterops.io/resources/nodes/az-service-principal'>
        AZServicePrincipal
    </a>
);

const AZDeviceLink = (
    <a target='_blank' rel='noopener noreferrer' href='https://bloodhound.specterops.io/resources/nodes/az-device'>
        AZDevice
    </a>
);

const General: FC = () => {
    return (
        <p className='edge-accordion-body2'>
            AZOwns means an Entra principal has been added as an owner over an Entra asset.
            <br />
            <br />
            AZOwns targets resources in Entra ID (for example {AZGroupLink}, {AZServicePrincipalLink}, and{' '}
            {AZDeviceLink}) from various object-specific ownership.
        </p>
    );
};

export default General;
