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
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <Typography variant='body2'>
            The {sourceType} {sourceName} has the privileges to perform the ADCS ESC13 abuse against the target AD
            group. The principal has enrollment rights on a certificate template configured with an issuance policy
            extension. The issuance policy has an OID group link to an AD group. The principal also has enrollment
            permission for an enterprise CA with the necessary template published. This enterprise CA is trusted for NT
            authentication and chains up to a root CA for the forest. This setup allows the principal to enroll a
            certificate that the principal can use to obtain access to the environment as a member of the group
            specified in the OID group link.
        </Typography>
    );
};

export default General;
