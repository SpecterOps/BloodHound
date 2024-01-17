// Copyright 2024 Specter Ops, Inc.
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
import { groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the privileges to perform the ADCS ESC6 Scenario B attack
                against the target domain.
            </Typography>
            <Typography variant='body2'>
                The principal has permission to enroll on one or more certificate templates allowing for authentication.
                They also have enrollment permission for an enterprise CA with the necessary templates published. This
                enterprise CA is trusted for NT authentication in the forest, and chains up to a root CA for the forest.
                The enterprise CA is configured with the EDITF_ATTRIBUTESUBJECTALTNAME2 flag allowing enrollees to
                specify a Subject Alternate Name (SAN) identifying another principal during certificate enrollment of
                any published certificate template. This setup allow an attacker principal to obtain a malicious
                certificate as another principal. There is an affected Domain Controller configured to allow weak
                certificate mapping enforcement, which enables the attacker principal to authenticate with the malicious
                certificate and thereby impersonating any AD forest user or computer without their credentials.
            </Typography>
        </>
    );
};

export default General;
