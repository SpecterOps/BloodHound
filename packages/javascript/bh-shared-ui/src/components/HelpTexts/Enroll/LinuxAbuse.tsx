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

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>Certipy can be used to enroll a certificate:</Typography>
            <Typography component={'pre'}>
                {'certipy req -u USER@CORP.LOCAL -p PWD -ca CA-NAME -target SERVER -template TEMPLATE'}
            </Typography>
            <Typography variant='body2'>
                The following requirements must be met for a principal to be able to enroll a certificate:
                <br />
                1) The principal has enrollment rights on a certificate template
                <br />
                2) The certificate template is published on an enterprise CA
                <br />
                3) The principal has Enroll permission on the enterprise CA
                <br />
                4) The principal meets the issuance requirements and the requirements for subject name and subject
                alternative name defined by the template
            </Typography>
        </>
    );
};

export default LinuxAbuse;
