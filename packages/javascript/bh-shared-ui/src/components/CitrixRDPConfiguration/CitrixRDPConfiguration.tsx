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
import { Switch } from '@bloodhoundenterprise/doodleui';
import { Box, Paper, Typography } from '@mui/material';
import { useState } from 'react';

const CitrixRDPConfiguration = () => {
    const [enabled, setEnabled] = useState(false);

    return (
        <Paper sx={{ padding: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <Typography variant='h4'>Citrix RDP Support</Typography>
                <Switch label={enabled ? 'On' : 'Off'} onClick={() => setEnabled((prev) => !prev)}></Switch>
            </Box>
            <Typography variant='body1'>
                When enabled, post-processing for the CanRDP edge will look for the presence of the default &quot;Direct
                Access Users&quot; group and assume that only local Administrators and members of this group can RDP to
                the system without validation that Citrix VDA is present and correctly configured. Use with caution.
            </Typography>
        </Paper>
    );
};

export default CitrixRDPConfiguration;
