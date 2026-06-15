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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Executing this from within Windows requires command line execution. If your target organization has
                command line logging enabled, this is a detection opportunity for their analysts.
            </Typography>
            <Typography variant='body2'>
                If the organization has PowerShell logging enabled, this is an even bigger detection opportunity, as the
                command line arguments will be logged in the PowerShell logs, which are more likely to be monitored by
                defenders. Event IDs can include 5136 (a directory service object was modified) and4738 (a user account
                was changed). If the organization has enabled module logging, the specific cmdlet used to perform the
                write operation will also be logged, which can further increase the chances of detection.
            </Typography>
        </>
    );
};

export default Opsec;
