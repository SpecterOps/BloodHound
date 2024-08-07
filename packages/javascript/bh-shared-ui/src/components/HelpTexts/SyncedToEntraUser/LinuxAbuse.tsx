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
import { Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <Typography variant='body2'>
            An attacker may authenticate as the synchronized Entra user using the on-prem user’s credentials, for
            example by dumping the user’s plain-text credential from memory, key-logging the user’s password, or by
            changing the on-prem user’s password and authenticating with that new password.
        </Typography>
    );
};

export default Abuse;
