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

const Opsec: FC = () => {
    return (
        <Typography variant='body2'>
            Authentication via a trust account is unusual and can be detected by Windows security events with the
            account name of a trust account. Specifically, monitor for:
            <ul>
                <li>Event ID 4768 - A Kerberos authentication ticket (TGT) was requested</li>
                <li>Event ID 4624 - A successful account login</li>
            </ul>
        </Typography>
    );
};

export default Opsec;
