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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                When you create a new secret for an App or Service Principal, Azure creates an event called “Update
                application – Certificates and secrets management”. This event describes who added the secret to which
                application or service principal.
            </Typography>
            <Typography variant='body2'>
                When resetting a user’s password and letting Azure set a new random password, Azure will log two events:
            </Typography>

            <Typography variant='body2'>
                “Reset user password” and “Reset password (by admin)”. These logs describe who performed the password
                reset, against which user, and at what time.
            </Typography>

            <Typography variant='body2'>
                When setting a specified new password for the user, Azure will log two events:
            </Typography>

            <Typography variant='body2'>
                “Reset user password” and “Update user”. The first log will describe who changed the target’s password
                and when.
            </Typography>
        </>
    );
};

export default Opsec;
