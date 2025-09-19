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

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Information for <code>RoleManagementAlert.ReadWrite.Directory</code>.
            </Typography>
            <ul><li><strong>Application:</strong> Allows the app to read and manage all role-based access control (RBAC) alerts for your company's directory, without a signed-in user. This includes managing alert settings, initiating alert scans, dismissing alerts, remediating alert incidents, and reading alert statuses, alert definitions, alert configurations and incidents that lead to an alert.</li></ul>
        </>
    );
};

export default General;