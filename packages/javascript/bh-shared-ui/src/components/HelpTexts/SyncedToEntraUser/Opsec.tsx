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

const Opsec: FC = () => {
    return (
        <Typography variant='body2'>
            The attacker may create artifacts of abusing this relationship in both on-prem AD and in Entra. A password
            reset operation against the on-prem user may create a 4724 Windows event, along with a corresponding Entra
            activity log entry when the on-prem agent synchronizes the new password hash up to Entra.
        </Typography>
    );
};

export default Opsec;
